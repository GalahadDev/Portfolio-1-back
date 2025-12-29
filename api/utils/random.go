package utils

import (
	"math/rand"
	"time"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateFleetCode crea un string aleatorio de 6 caracteres
func GenerateFleetCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	// Formato: ABC-123
	return string(b[:3]) + "-" + string(b[3:])
}
