package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)

	// Menggunakan regex yang aman dan terhindar dari error compile
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Gunakan MatchString langsung agar lebih ringkas dan aman
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return false
	}

	return matched
}

func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	hasNumber, _ := regexp.MatchString(`[0-9]`, password)
	hasSymbol, _ := regexp.MatchString(`[!@#\$%\^&\*(),\.?":\{\}|<>_\-\+=~`+"`"+`]`, password)

	return hasLetter && hasNumber && hasSymbol
}
