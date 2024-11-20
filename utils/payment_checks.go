package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GeneratePassword() (string, error) {
	password := make([]byte, 6)
	_, err := rand.Read(password)
	if err != nil {
		return "", err
	}

	// Convert to a 6-digit string
	return fmt.Sprintf("%06d", int(password[0])%1000000), nil
}
func HashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func CheckPassword_P(inputPassword string, hashedPassword string) bool {
	hashedInputPassword := HashPassword(inputPassword)
	return hashedInputPassword == hashedPassword
}

func validatePaymentPassword(inputPassword string, storedHashedPassword string) bool {
	if CheckPassword_P(inputPassword, storedHashedPassword) {
		return true // Password is correct
	}
	return false // Password is incorrect
}
