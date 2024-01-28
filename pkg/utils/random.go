package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	CharsLowerAlpha = "abcdefghijklmnopqrstuvwxyz"
	CharsUpperAlpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsNumeric    = "0123456789"
)

// GenerateRandomBytes generates given number of random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("GenerateRandomBytes failed to rand.Read: %w", err)
	}

	return b, nil
}

// GenerateRandomString generates random string of given length.
// allowLowerAlpha, allowUpperAlpha, and allowNumeric can be used to include/exclude
// lowercase, uppercase, and numeric characters respectively.
func GenerateRandomString(
	n int,
	allowLowerAlpha bool,
	allowUpperAlpha bool,
	allowNumeric bool,
) (string, error) {
	lowerAlpha := ""
	if allowLowerAlpha {
		lowerAlpha = CharsLowerAlpha
	}

	upperAlpha := ""
	if allowUpperAlpha {
		upperAlpha = CharsUpperAlpha
	}

	numeric := ""
	if allowNumeric {
		numeric = CharsNumeric
	}

	allowed := []rune(lowerAlpha + upperAlpha + numeric)
	randRunes := make([]rune, n)

	for i := range randRunes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowed))))
		if err != nil {
			return "", fmt.Errorf("GenerateRandomString failed to rand.Int: %w", err)
		}

		randRunes[i] = allowed[n.Int64()]
	}

	randString := string(randRunes)

	return randString, nil
}
