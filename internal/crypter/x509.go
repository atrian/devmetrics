package crypter

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"math"
	"os"
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
	for i := 0; i < chunks*MessageLenLimit; i += MessageLenLimit {
		chunkEnd := len(message)
		if i+MessageLenLimit < chunkEnd {
			chunkEnd = i + MessageLenLimit
		}

		encPart, err := k.EncryptWithKey(message[i:chunkEnd], key)
		if err != nil {
			return nil, err
		}

		result = append(result, encPart...)
	}

	return result, nil
}

// EncryptWithKey шифрует сообщение с предоставленным публичным ключом
func (k *KeyManager) EncryptWithKey(message []byte, key *rsa.PublicKey) ([]byte, error) {
	// OAEP is parameterised by a hash function that is used as a random oracle.
	// Encryption and decryption of a given message must use the same hash function
	// and sha256.New() is a reasonable choice.
	hash := sha256.New()

	size := hash.Size()
	_ = size

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

	for i := 0; i < chunks*EncryptedBlockSize; i += EncryptedBlockSize {
		chunkEnd := len(message)
		if i+EncryptedBlockSize < chunkEnd {
			chunkEnd = i + EncryptedBlockSize
		}

		decPart, err := k.DecryptWithKey(message[i:chunkEnd], key)
		if err != nil {
			return nil, err
		}

		result = append(result, decPart...)
	}

	return result, nil
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
