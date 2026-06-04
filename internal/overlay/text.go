package overlay

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func normalizeLF(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

func normalizeLFTerminated(text string) string {
	normalized := normalizeLF(text)
	if strings.HasSuffix(normalized, "\n") {
		return normalized
	}
	return normalized + "\n"
}

func sha256Text(text string) string {
	sum := sha256.Sum256([]byte(normalizeLFTerminated(text)))
	return hex.EncodeToString(sum[:])
}
