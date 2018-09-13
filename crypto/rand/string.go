package rand

import (
	"encoding/hex"
)

// Return a randomly generated string of characters
// of length n from the set "0123456789abcdef"
func RandomHexString(n int) string {
	buf := make([]byte, (n/2)+1)
	return hex.EncodeToString(RandomBytes(buf))[:n]
}
