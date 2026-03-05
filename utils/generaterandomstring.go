package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/toolchain/src/math/rand"
)

func GenerateRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateTRNCode(prefix, deviceID string, userID uint) string {
	return fmt.Sprintf(
		"%s-%d-%s-%d-%s",
		prefix,
		time.Now().UnixNano(), //nanosecond
		deviceID,
		userID,
		uuid.NewString()[:8], // uuid untuk unieq
	)
}
