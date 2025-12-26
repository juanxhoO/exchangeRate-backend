package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"os"
)

type IAPIService interface {
	GenerateApiKey(length int) (string, error)
	EncryptApiKey(apiKey string) (string, error)
	DecryptApiKey(cipherKey string) (string, error)
}

type APIService struct {
}

func NewAPIService() IAPIService {
	return &APIService{}
}

func (s *APIService) GenerateApiKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	apiKey := base64.URLEncoding.EncodeToString(key)

	return apiKey, nil
}

func (s *APIService) EncryptApiKey(apiKey string) (string, error) {

	block, err := aes.NewCipher([]byte(os.Getenv("SECRET_API_KEY_GENERATOR")))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (s *APIService) DecryptApiKey(cipherKey string) (string, error) {

	cipherText, err := base64.URLEncoding.DecodeString(cipherKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(os.Getenv("SECRET_API_KEY_GENERATOR")))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", err
	}

	nonce, ciphertext := cipherText[:nonceSize], cipherText[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	print(string(plaintext))

	return string(plaintext), nil
}
