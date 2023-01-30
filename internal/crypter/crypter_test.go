package crypter_test

import (
	"fmt"
	"log"

	"github.com/atrian/devmetrics/internal/crypter"
)

func ExampleKeyManager_Encrypt() {
	keyManager := crypter.New()

	// Генерируем тестовые ключи
	pubKeyBody, privateKeyBody, err := keyManager.GenerateKeys()
	if err != nil {
		log.Fatal(err.Error())
	}

	// шифруем сообщение Test message публичным ключом
	message := "Test message"

	pubKey, err := keyManager.ParsePublicKey(pubKeyBody)
	if err != nil {
		log.Fatal("ParsePublicKey err:", err.Error())
	}
	keyManager.RememberPublicKey(pubKey)

	encryptedMessage, _ := keyManager.Encrypt([]byte(message))

	// расшифровываем Test message сообщение приватным ключом
	secret, err := keyManager.ParsePrivateKey(privateKeyBody)
	if err != nil {
		log.Fatal("ParsePrivateKey err:", err.Error())
	}
	keyManager.RememberPrivateKey(secret)

	decryptedMessage, _ := keyManager.Decrypt(encryptedMessage)

	fmt.Println(string(decryptedMessage))
	// Output:
	// Test message
}

func ExampleKeyManager_EncryptWithKey() {
	cm := crypter.New()

	// Генерируем тестовые ключи
	pubKeyBody, privateKeyBody, err := cm.GenerateKeys()
	if err != nil {
		log.Fatal(err.Error())
	}

	pubKey, err := cm.ParsePublicKey(pubKeyBody)
	if err != nil {
		log.Fatal("ParsePublicKey err:", err.Error())
	}

	// шифруем сообщение Test message публичным ключом
	message := "Test message"
	encryptedMessage, _ := cm.EncryptWithKey([]byte(message), pubKey)

	// расшифровываем Test message сообщение приватным ключом
	secret, err := cm.ParsePrivateKey(privateKeyBody)
	if err != nil {
		log.Fatal("ParsePrivateKey err:", err.Error())
	}
	decryptedMessage, _ := cm.DecryptWithKey(encryptedMessage, secret)

	fmt.Println(string(decryptedMessage))
	// Output:
	// Test message
}
