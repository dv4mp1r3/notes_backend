package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func HashPassword(password, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(salt + password))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func VerifyPassword(hash, password, salt string) bool {
	return hash == HashPassword(password, salt)
}
