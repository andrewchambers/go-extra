package rand

import (
	"crypto/rand"
	"time"
)

// Fill buf with random bytes returning buf.
func RandomBytes(buf []byte) []byte {
	for {
		_, err := rand.Read(buf)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	return buf
}
