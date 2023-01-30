package crypter_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

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

func ExampleKeyManager_GenerateKeys() {
	// Генерируем пару приватного и публичного ключей
	cm := crypter.New()

	// Генерируем тестовые ключи
	pubKeyBody, privateKeyBody, err := cm.GenerateKeys()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Подготовка к сохранению файлов. Создаем папку, определяем пути
	cwd, err := os.Getwd()
	keysDir := "/../../.keys"

	keyPath := filepath.Join(cwd, keysDir)

	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		err := os.RemoveAll(keyPath)
		if err != nil {
			log.Fatal("Can't remove keys dir:", err.Error())
		}
	}

	err = os.Mkdir(keyPath, os.ModePerm)
	if err != nil {
		log.Fatal("Can't create keys dir:", err.Error())
	}

	// Сохраняем публичный ключ в файл
	pubPath := filepath.Join(cwd, keysDir, "pub.pem")
	pub, err := os.Create(pubPath)
	if err != nil {
		log.Fatal("Can't create file:", err.Error())
	}
	defer func(pub *os.File) {
		cErr := pub.Close()
		if cErr != nil {
			log.Fatal("Can't close pub:", cErr.Error())
		}
	}(pub)

	_, err = pub.Write(pubKeyBody)
	if err != nil {
		log.Fatal("Can't write to PUB file:", err.Error())
	}

	// Сохраняем приватный ключ в файл
	secretPath := filepath.Join(cwd, keysDir, "secret.pem")
	secret, err := os.Create(secretPath)
	if err != nil {
		log.Fatal("Can't create file:", err.Error())
	}
	defer func(secret *os.File) {
		cErr := secret.Close()
		if cErr != nil {
			log.Fatal("Can't close pub:", cErr.Error())
		}
	}(secret)

	_, err = secret.Write(privateKeyBody)
	if err != nil {
		log.Fatal("Can't write to SECRET file:", err.Error())
	}

	fmt.Println(err == nil)
	// Output:
	// true
}

func TestKeyManager_EncryptBigMessage(t *testing.T) {
	keyManager := crypter.New()

	// Генерируем тестовые ключи
	pubKeyBody, privateKeyBody, err := keyManager.GenerateKeys()
	if err != nil {
		log.Fatal(err.Error())
	}

	// шифруем длинное сообщение публичным ключом
	message := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

	pubKey, err := keyManager.ParsePublicKey(pubKeyBody)
	if err != nil {
		log.Fatal("ParsePublicKey err:", err.Error())
	}
	keyManager.RememberPublicKey(pubKey)

	encryptedMessage, err := keyManager.Encrypt([]byte(message))
	if err != nil {
		log.Fatal("Encrypt message err:", err.Error())
	}

	// расшифровываем длинное сообщение приватным ключом
	secret, err := keyManager.ParsePrivateKey(privateKeyBody)
	if err != nil {
		log.Fatal("ParsePrivateKey err:", err.Error())
	}
	keyManager.RememberPrivateKey(secret)

	decryptedMessage, err := keyManager.Decrypt(encryptedMessage)
	if err != nil {
		log.Fatal("Decrypt message error:", err.Error())
	}

	result := string(decryptedMessage)

	if message != result {
		t.Errorf("Encrypt / Decrypt big message fails = %v, want %v", result, message)
	}
}
