// Package crypter Шифрование сообщений между агентом и сервером
package crypter

import "crypto/rsa"

// Crypter осуществляет шифрование и расшифровку сообщений между агентом и сервером
// агент подписывает сообщения открытым ключом, сервер расшифровывает сообщения закрытым ключом
type Crypter interface {
	CryptoParser
	CryptoKeyKeeper
	Encrypt(message []byte, key *rsa.PublicKey) ([]byte, error)
	Decrypt(message []byte, key *rsa.PrivateKey) ([]byte, error)
}

// CryptoParser осуществляет чтение ключей по переданному пути
type CryptoParser interface {
	// ReadPrivateKey получает путь к закрытому ключу считывает его и возвращает rsa.PrivateKey
	ReadPrivateKey(keyPath string) (*rsa.PrivateKey, error)
	// ReadPublicKey получает путь к публичному ключу считывает его и возвращает rsa.PublicKey
	ReadPublicKey(keyPath string) (*rsa.PublicKey, error)
}

type CryptoKeyKeeper interface {
	GenerateKeys() (publicKey []byte, privateKey []byte, err error)
}
