// Package crypter Шифрует и расшифровывает сообщения при помощи публичного и приватного RSA ключа
// Большие сообщения разбиваются на части по 446 байт и шифруются отдельно
//
// Добавлена поддержка многопоточности в шифрование / расшифровку сообщений в методах
// KeyManager.EncryptBigMessage и KeyManager.DecryptBigMessage
//
// Варианты оптимизации:
//
// 1. Использовать для шифрования метрик симметричный алгоритм
// Ключ генерировать случайным образом и шифровать его ассиметрично с помощью rsa.EncryptOAEP
// Передавать зашифрованный ключ вместе с данными [512 байт ключ][симметрично зашифрованное тело сообщения]
package crypter

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math"
	"os"
	"sync"
)

var _ Crypter = (*KeyManager)(nil)

const (
	// MessageLenLimit = 512 (длинна ключа) - 2 * 32 -2 = 446
	// По ограничениям см. rsa.EncryptOAEP
	// k := pub.Size()
	//	if len(msg) > k-2*hash.Size()-2 {
	//		return nil, ErrMessageTooLong
	//	}
	MessageLenLimit    = 446
	EncryptedBlockSize = 512
)

type KeyManager struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// chunk используется для шифрования данных в многопоточном режиме.
// в зависимости от размера сообщения данные разбиваются на блоки длинной MessageLenLimit (446 байт) при шифровке
// и EncryptedBlockSize (512 байт) при расшифровке
type chunk struct {
	Index uint   // индекс блока
	Data  []byte // данные блока
}

func New() *KeyManager {
	km := KeyManager{}
	return &km
}

// ReadPrivateKey читает приватный ключ с диска, возвращает указатель на структуру rsa.PrivateKey
// использует ParsePrivateKey для парсинга ключа
func (k *KeyManager) ReadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	// получаем данные из файла
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	cert, err := k.ParsePrivateKey(data)
	// возвращаем готовый к работе ключ
	return cert, err
}

// ParsePrivateKey парсит приватный ключ из слайса байт, возвращает возвращает указатель на rsa.PrivateKey
func (k *KeyManager) ParsePrivateKey(key []byte) (*rsa.PrivateKey, error) {
	// парсим закрытый ключ
	block, _ := pem.Decode(key)
	cert, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	return cert, err
}

// RememberPrivateKey кеширование приватного ключа
func (k *KeyManager) RememberPrivateKey(key *rsa.PrivateKey) {
	k.PrivateKey = key
}

// ReadyForDecrypt возвращает true если приватный ключ загружен в менеджер
func (k *KeyManager) ReadyForDecrypt() bool {
	return k.PrivateKey != nil
}

// ReadPublicKey читает публичный ключ с диска, возвращает указатель на структуру rsa.PublicKey
// использует ParsePublicKey для парсинга ключа
func (k *KeyManager) ReadPublicKey(keyPath string) (*rsa.PublicKey, error) {
	// получаем данные из файла
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	cert, err := k.ParsePublicKey(data)
	// возвращаем готовый к работе ключ
	return cert, err
}

// ParsePublicKey парсит публичный ключ из слайса байт, возвращает возвращает указатель на rsa.PublicKey
func (k *KeyManager) ParsePublicKey(key []byte) (*rsa.PublicKey, error) {
	// парсим публичный ключ
	block, _ := pem.Decode(key)
	cert, err := x509.ParsePKCS1PublicKey(block.Bytes)

	return cert, err
}

// RememberPublicKey кеширование публичного ключа
func (k *KeyManager) RememberPublicKey(key *rsa.PublicKey) {
	k.PublicKey = key
}

// ReadyForEncrypt возвращает true если публичный ключ загружен в менеджер
func (k *KeyManager) ReadyForEncrypt() bool {
	return k.PublicKey != nil
}

// GenerateKeys генерирует пару приватного и публичного ключа длиной 4096 бит
// возвращает тело []byte ключей в формате в формате PEM
func (k *KeyManager) GenerateKeys() (publicKey []byte, privateKey []byte, err error) {
	// создаём новый приватный RSA-ключ длиной 4096 бит
	// для генерации ключей используется rand.Reader в качестве источника случайных данных
	private, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// кодируем ключи в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&private.PublicKey),
	})
	if err != nil {
		return nil, nil, err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	})
	if err != nil {
		return nil, nil, err
	}

	return publicKeyPEM.Bytes(), privateKeyPEM.Bytes(), nil
}

// Encrypt шифрует сообщение с кешированным публичным ключом
func (k *KeyManager) Encrypt(message []byte) ([]byte, error) {
	return k.EncryptBigMessage(message, k.PublicKey)
}

// EncryptBigMessage разбивает сообщение на чанки по размеру MessageLenLimit
// шифрует и собирает зашифрованное сообщение целиком
func (k *KeyManager) EncryptBigMessage(message []byte, key *rsa.PublicKey) ([]byte, error) {
	chunks := int(math.Ceil(float64(len(message)) / MessageLenLimit))

	result := make([]byte, 0, chunks*EncryptedBlockSize)

	// wg ждем завершения всех горутин
	wg := sync.WaitGroup{}
	// errCh канал для отлова ошибки. Если произошла ошибка, останавливаем остальные горутины
	errCh := make(chan error)
	// stopCh канал для сигналов остановки
	stopCh := make(chan struct{})
	// chunkCh канал для записи результатов
	chunkCh := make(chan chunk, chunks)

	for i, chunkIndex := 0, 0; i < chunks*MessageLenLimit; i, chunkIndex = i+MessageLenLimit, chunkIndex+1 {
		chunkEnd := len(message)
		if i+MessageLenLimit < chunkEnd {
			chunkEnd = i + MessageLenLimit
		}

		// добавляем горутину для шифрования чанка
		wg.Add(1)
		go k.runEncryptProcess(&wg, errCh, stopCh, chunkCh, chunkIndex, message[i:chunkEnd], key)
	}

	// в отдельной горутине ждём завершения всех шифровальщиков
	// после этого закрываем канал errCh — больше записей не будет
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Если есть ошибки шифрования закрываем канал stopCh (завершаем остальные горутины)
	if err := <-errCh; err != nil {
		log.Println(err)
		close(stopCh)
		return nil, err
	}

	// считываем данные из канала с чанками в слайс восстанавливая порядок частей сообщения
	resultBuffer := make([][]byte, chunks)
	for i := 0; i < chunks; i++ {
		part := <-chunkCh
		resultBuffer[part.Index] = part.Data
	}

	// закарываем канал
	close(chunkCh)

	// собираем из частей шифрованное сообщение
	result = bytes.Join(resultBuffer, nil)

	return result, nil
}

// runEncryptProcess для шифрования частей сообщения в многопоточном режиме
// errCh - пишет в канал ошибку в случае возникновения
// stopCh - остановка других шифровальщиков в случае ошибки
// chunkCh - канал для сброса обработанных данных
// index - индекс части данных в шифруемом сообщении
// messagePart - часть открытого сообщения длинной не более MessageLenLimit
// key - ключ для шифрования rsa.PublicKey
func (k *KeyManager) runEncryptProcess(
	wg *sync.WaitGroup,
	errCh chan<- error,
	stopCh <-chan struct{},
	chunkCh chan<- chunk,
	index int,
	messagePart []byte,
	key *rsa.PublicKey) {

	// Обработка ошибок после завершения шифрования чанка
	var defErr error
	defer func() {
		if defErr != nil {
			select {
			// первая горутина, поймавшая ошибку, сможет записать в канал
			case errCh <- defErr:
			// остальные завершат работу, провалившись в этот case
			case <-stopCh:
				log.Println("aborting chunk encryption")
			}
		}
		wg.Done()
	}()

	// шифруем кусок сообщения
	encPart, defErr := k.EncryptWithKey(messagePart, key)
	chunkCh <- chunk{
		Index: uint(index),
		Data:  encPart,
	}
}

// EncryptWithKey шифрует сообщение с предоставленным публичным ключом
func (k *KeyManager) EncryptWithKey(message []byte, key *rsa.PublicKey) ([]byte, error) {
	// OAEP is parameterised by a hash function that is used as a random oracle.
	// Encryption and decryption of a given message must use the same hash function
	// and sha256.New() is a reasonable choice.
	hash := sha256.New()

	encryptedMessage, err := rsa.EncryptOAEP(hash, rand.Reader, key, message, nil)
	if err != nil {
		return nil, err
	}

	return encryptedMessage, nil
}

// Decrypt расшифровывает сообщение с кешированным приватным ключом
func (k *KeyManager) Decrypt(message []byte) ([]byte, error) {
	return k.DecryptBigMessage(message, k.PrivateKey)
}

// DecryptBigMessage разбивает зашифрованное сообщение на чанки по размеру EncryptedBlockSize
// расшифровывает и собирает расшифрованное сообщение целиком
func (k *KeyManager) DecryptBigMessage(message []byte, key *rsa.PrivateKey) ([]byte, error) {
	chunks := int(math.Ceil(float64(len(message)) / EncryptedBlockSize))

	result := make([]byte, 0, len(message))

	// wg ждем завершения всех горутин
	wg := sync.WaitGroup{}
	// errCh канал для отлова ошибки. Если произошла ошибка, останавливаем остальные горутины
	errCh := make(chan error)
	// stopCh канал для сигналов остановки
	stopCh := make(chan struct{})
	// chunkCh канал для записи результатов
	chunkCh := make(chan chunk, chunks)

	for i, chunkIndex := 0, 0; i < chunks*EncryptedBlockSize; i, chunkIndex = i+EncryptedBlockSize, chunkIndex+1 {
		chunkEnd := len(message)
		if i+EncryptedBlockSize < chunkEnd {
			chunkEnd = i + EncryptedBlockSize
		}

		// добавляем горутину для шифрования чанка
		wg.Add(1)
		go k.runDecryptProcess(&wg, errCh, stopCh, chunkCh, chunkIndex, message[i:chunkEnd], key)
	}

	// в отдельной горутине ждём завершения всех шифровальщиков
	// после этого закрываем канал errCh — больше записей не будет
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Если есть ошибки шифрования закрываем канал stopCh (завершаем остальные горутины)
	if err := <-errCh; err != nil {
		log.Println(err)
		close(stopCh)
		return nil, err
	}

	// считываем данные из канала с чанками в слайс восстанавливая порядок частей сообщения
	resultBuffer := make([][]byte, chunks)
	for i := 0; i < chunks; i++ {
		part := <-chunkCh
		resultBuffer[part.Index] = part.Data
	}

	// закарываем канал
	close(chunkCh)

	// собираем из частей шифрованное сообщение
	result = bytes.Join(resultBuffer, nil)

	return result, nil
}

// runDecryptProcess для расшифровки частей сообщения в многопоточном режиме
// errCh - пишет в канал ошибку в случае возникновения
// stopCh - остановка других шифровальщиков в случае ошибки
// chunkCh - канал для сброса обработанных данных
// index - индекс части данных в шифруемом сообщении
// messagePart - часть шифрованного сообщения длинной не более EncryptedBlockSize
// key - ключ для расшифровки rsa.PrivateKey
func (k *KeyManager) runDecryptProcess(
	wg *sync.WaitGroup,
	errCh chan<- error,
	stopCh <-chan struct{},
	chunkCh chan<- chunk,
	index int,
	messagePart []byte,
	key *rsa.PrivateKey) {

	// Обработка ошибок после завершения шифрования чанка
	var defErr error
	defer func() {
		if defErr != nil {
			select {
			// первая горутина, поймавшая ошибку, сможет записать в канал
			case errCh <- defErr:
			// остальные завершат работу, провалившись в этот case
			case <-stopCh:
				log.Println("aborting chunk decryption")
			}
		}
		wg.Done()
	}()

	// шифруем кусок сообщения
	encPart, defErr := k.DecryptWithKey(messagePart, key)
	chunkCh <- chunk{
		Index: uint(index),
		Data:  encPart,
	}
}

// DecryptWithKey расшифровывает сообщение с предоставленным приватным ключом
func (k *KeyManager) DecryptWithKey(message []byte, key *rsa.PrivateKey) ([]byte, error) {
	// OAEP is parameterised by a hash function that is used as a random oracle.
	// Encryption and decryption of a given message must use the same hash function
	// and sha256.New() is a reasonable choice.
	hash := sha256.New()

	decryptedMessage, err := rsa.DecryptOAEP(hash, rand.Reader, key, message, nil)
	if err != nil {
		return nil, err
	}

	return decryptedMessage, nil
}
