// Package password generates cryptographically-random passwords.
package password

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// Random returns a cryptographically-random alphanumeric password of length n.
func Random(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	var b strings.Builder
	for _, v := range buf {
		b.WriteByte(chars[int(v)%len(chars)])
	}
	return b.String()
}
