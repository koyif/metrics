package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPublicKey loads an RSA public key from a PEM file.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA public key")
	}

	return rsaPub, nil
}

// LoadPrivateKey loads an RSA private key from a PEM file.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		rsaPriv, ok := privKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA private key")
		}
		return rsaPriv, nil
	}

	return priv, nil
}

// EncryptData encrypts data using RSA-OAEP with SHA-256.
// For large data, it encrypts in chunks since RSA can only encrypt data smaller than the key size.
func EncryptData(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	chunkSize := publicKey.Size() - 2*hash.Size() - 2

	var encrypted []byte

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[i:end]
		encryptedChunk, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt chunk: %w", err)
		}

		encrypted = append(encrypted, encryptedChunk...)
	}

	return encrypted, nil
}

// DecryptData decrypts data that was encrypted using RSA-OAEP with SHA-256.
// It handles data that was encrypted in chunks.
func DecryptData(privateKey *rsa.PrivateKey, encrypted []byte) ([]byte, error) {
	hash := sha256.New()
	chunkSize := privateKey.Size()

	var decrypted []byte

	for i := 0; i < len(encrypted); i += chunkSize {
		end := i + chunkSize
		if end > len(encrypted) {
			end = len(encrypted)
		}

		chunk := encrypted[i:end]
		decryptedChunk, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt chunk: %w", err)
		}

		decrypted = append(decrypted, decryptedChunk...)
	}

	return decrypted, nil
}
