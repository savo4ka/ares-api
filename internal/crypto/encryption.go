package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionService предоставляет методы для шифрования и расшифровки данных
type EncryptionService struct {
	key []byte
}

// EncryptedData содержит зашифрованные данные и IV
type EncryptedData struct {
	Ciphertext string // base64-закодированный шифртекст
	IV         string // base64-закодированный initialization vector
}

// NewEncryptionService создаёт новый сервис шифрования с заданным ключом
func NewEncryptionService(key string) (*EncryptionService, error) {
	keyBytes := []byte(key)

	// Проверяем длину ключа для AES-128 (16 байт)
	if len(keyBytes) != 16 {
		return nil, fmt.Errorf("encryption key must be exactly 16 bytes for AES-128, got %d bytes", len(keyBytes))
	}

	return &EncryptionService{
		key: keyBytes,
	}, nil
}

// Encrypt шифрует plaintext с использованием AES-128-CBC
func (e *EncryptionService) Encrypt(plaintext string) (*EncryptedData, error) {
	plaintextBytes := []byte(plaintext)

	// Создаём AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Добавляем PKCS7 padding
	plaintextBytes = pkcs7Pad(plaintextBytes, aes.BlockSize)

	// Генерируем случайный IV (initialization vector)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Создаём CBC mode encrypter
	mode := cipher.NewCBCEncrypter(block, iv)

	// Шифруем данные
	ciphertext := make([]byte, len(plaintextBytes))
	mode.CryptBlocks(ciphertext, plaintextBytes)

	// Кодируем в base64 для хранения
	return &EncryptedData{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		IV:         base64.StdEncoding.EncodeToString(iv),
	}, nil
}

// Decrypt расшифровывает данные с использованием AES-128-CBC
func (e *EncryptionService) Decrypt(encrypted *EncryptedData) (string, error) {
	// Декодируем из base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encrypted.IV)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	// Создаём AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Проверяем длину IV
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid IV length: %d", len(iv))
	}

	// Проверяем, что ciphertext кратен размеру блока
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	// Создаём CBC mode decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// Расшифровываем данные
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Убираем PKCS7 padding
	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %w", err)
	}

	return string(plaintext), nil
}

// pkcs7Pad добавляет PKCS7 padding к данным
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// pkcs7Unpad убирает PKCS7 padding из данных
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	if length%blockSize != 0 {
		return nil, fmt.Errorf("data length is not a multiple of block size")
	}

	padding := int(data[length-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	// Проверяем, что все байты padding имеют правильное значение
	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:length-padding], nil
}
