package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// encryptToken encrypts a plaintext string using RSA (for AES key) and AES (for data)
func EncryptToken(plainText string, publicKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", errors.New("failed to decode public key PEM block")
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not a valid RSA public key")
	}

	// Generate AES key
	aesKey := make([]byte, 32) // AES-256
	if _, err := rand.Read(aesKey); err != nil {
		return "", fmt.Errorf("failed to generate AES key: %v", err)
	}

	// Encrypt AES key with RSA
	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt AES key: %v", err)
	}

	// Encrypt data with AES
	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	encryptedData := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Combine AES key and encrypted data
	combined := base64.StdEncoding.EncodeToString(encryptedAESKey) + ":" + base64.StdEncoding.EncodeToString(encryptedData)
	return combined, nil
}

// decryptToken decrypts the token directly
func DecryptToken(encryptedToken string, privateKeyPem string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPem))
	if block == nil {
		return "", errors.New("failed to decode private key PEM block")
	}

	// Handle PKCS#1 and PKCS#8 private keys
	var privKey *rsa.PrivateKey
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}

		var ok bool
		privKey, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("not a valid RSA private key")
		}
	} else {
		privKey = key
	}

	// Split the token into AES key and encrypted data
	parts := strings.Split(encryptedToken, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid encrypted token format")
	}

	encryptedAESKey, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode AES key: %v", err)
	}

	encryptedData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	// Decrypt AES key with RSA
	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encryptedAESKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt AES key: %v", err)
	}

	// Decrypt data with AES
	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", errors.New("invalid encrypted data")
	}

	nonce, cipherText := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(plainText), nil
}