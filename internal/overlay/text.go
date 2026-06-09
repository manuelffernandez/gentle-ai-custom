package overlay

import (
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


