package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	whirl "github.com/balacode/zr-whirl"
)

type (
	Type       int
	Length     int
	Characters string
)

const (
	GeneratePassword Type = iota
	GenerateSalt
	GenerateUsername
)

const (
	UsernameCharacters Characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	SaltCharacters     Characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/."
	PasswordCharacters Characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@*()&"
)

const (
	PasswordLength Length = 12
	SaltLength     Length = 22
	UsernameLength Length = 20
)

// HashPass hashes a password using a Whirlpool hash.
// Passwords are presumed to be hashed.
func HashPass(password string) string {
	var iter = 1000

	var next string

	for i := 0; i < iter; i++ {
		next += password
		tmp := whirl.HashOfBytes([]byte(next), []byte(""))
		next = hex.EncodeToString(tmp)
	}

	return next
}

func GenerateRandomLength(length int, randomType Type) (string, error) {
	if length < 6 || length > 40 {
		return "", errors.New("length must be between 6 and 40")
	}
	switch randomType {
	case GeneratePassword:
		b, err := rangeLoop(PasswordCharacters, Length(length))
		if err != nil {
			return "", fmt.Errorf("error generating random password: %w", err)
		}

		return b, nil
	case GenerateSalt:
		b, err := rangeLoop(SaltCharacters, Length(length))
		if err != nil {
			return "", fmt.Errorf("error generating random salt: %w", err)
		}

		return "$2a$06$" + b, nil
	case GenerateUsername:
		b, err := rangeLoop(UsernameCharacters, Length(length))
		if err != nil {
			return "", fmt.Errorf("error generating random username: %w", err)
		}

		return b, nil
	default:
		return "", fmt.Errorf("invalid type: %d", randomType)
	}
}

// GenerateRandom generates a random string for either password or salt
func GenerateRandom(randomType Type) (string, error) {
	switch randomType {
	case GeneratePassword:
		b, err := rangeLoop(PasswordCharacters, PasswordLength)
		if err != nil {
			return "", fmt.Errorf("error generating random password: %w", err)
		}

		return b, nil
	case GenerateSalt:
		b, err := rangeLoop(SaltCharacters, SaltLength)
		if err != nil {
			return "", fmt.Errorf("error generating random salt: %w", err)
		}

		return "$2a$06$" + b, nil
	case GenerateUsername:
		b, err := rangeLoop(UsernameCharacters, UsernameLength)
		if err != nil {
			return "", fmt.Errorf("error generating random username: %w", err)
		}

		return b, nil
	default:
		return "", fmt.Errorf("invalid type: %d", randomType)
	}
}

// rangeLoop creates the random string that will be used
func rangeLoop(characters Characters, randomLength Length) (string, error) {
	bytes := make([]byte, randomLength)

	length := big.NewInt(int64(len(characters)))

	for i := range bytes {
		randInt, err := rand.Int(rand.Reader, length)
		if err != nil {
			return "", err
		}

		bytes[i] = characters[randInt.Int64()]
	}

	return string(bytes), nil
}
