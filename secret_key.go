package storage

import (
	"fmt"
	"os"
)

const minSecretKeySize = 32

func GetSecretKey() ([]byte, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("Must provide a secret key under env variable SECRET_KEY, length must be %d or more", minSecretKeySize)
	}

	return []byte(secretKey), nil
}
