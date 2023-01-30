package crypter

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
)

var _ Crypter = (*CertManager)(nil)

type CertManager struct {
}

func (c CertManager) ReadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	// получаем данные из файла
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// парсим закрытый ключ
	block, _ := pem.Decode(data)
	cert, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// возвращаем готовый к работе ключ
	return cert, err
}

func (c CertManager) ReadPublicKey(keyPath string) (*rsa.PublicKey, error) {
	// получаем данные из файла
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	// парсим публичный ключ
	block, _ := pem.Decode(data)
	cert, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// возвращаем готовый к работе ключ
	return cert, err
}

func (c CertManager) GenerateKeys() (publicKey []byte, privateKey []byte, err error) {
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

	return publicKeyPEM.Bytes(), publicKeyPEM.Bytes(), nil
}

func (c CertManager) Encrypt(message []byte, key *rsa.PublicKey) ([]byte, error) {
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

func (c CertManager) Decrypt(message []byte, key *rsa.PrivateKey) ([]byte, error) {
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
