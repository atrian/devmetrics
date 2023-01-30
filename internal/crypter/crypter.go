// Package crypter Шифрование сообщений между агентом и сервером
package crypter

import "crypto/rsa"

// Crypter осуществляет шифрование и расшифровку сообщений между агентом и сервером
// агент подписывает сообщения открытым ключом, сервер расшифровывает сообщения закрытым ключом
type Crypter interface {
	CryptoParser
	CryptoKeyKeeper
	// Encrypt шифрует сообщение с предварительно кешированным публичным ключом
	Encrypt(message []byte) ([]byte, error)
	// Decrypt расшифровывает сообщение с предварительно кешированным приватным ключом
	Decrypt(message []byte) ([]byte, error)
	// EncryptWithKey шифрует сообщение с предоставленным публичным ключом
	EncryptWithKey(message []byte, key *rsa.PublicKey) ([]byte, error)
	// DecryptWithKey расшифровывает сообщение с предоставленным приватным ключом
	DecryptWithKey(message []byte, key *rsa.PrivateKey) ([]byte, error)
}

// CryptoParser осуществляет чтение ключей по переданному пути
type CryptoParser interface {
	// ReadPrivateKey получает путь к закрытому ключу считывает его и возвращает rsa.PrivateKey
	ReadPrivateKey(keyPath string) (*rsa.PrivateKey, error)
	// ReadPublicKey получает путь к публичному ключу считывает его и возвращает rsa.PublicKey
	ReadPublicKey(keyPath string) (*rsa.PublicKey, error)
	// ParsePrivateKey получает тело приватного ключа []byte и возвращает rsa.PrivateKey
	ParsePrivateKey(key []byte) (*rsa.PrivateKey, error)
	// ParsePublicKey получает тело публичного ключа []byte и возвращает rsa.PublicKey
	ParsePublicKey(key []byte) (*rsa.PublicKey, error)
}

type CryptoKeyKeeper interface {
	// GenerateKeys генерирует пару ключей - публичный и приватный
	GenerateKeys() (publicKey []byte, privateKey []byte, err error)
	// RememberPrivateKey кеширование приватного ключа
	RememberPrivateKey(key *rsa.PrivateKey)
	// RememberPublicKey кеширование публичного ключа
	RememberPublicKey(key *rsa.PublicKey)
}
