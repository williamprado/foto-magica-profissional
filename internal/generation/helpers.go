package generation

import (
	"encoding/base64"
	"strings"
)

func decodeBase64(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "data:image/jpeg;base64,")
	value = strings.TrimPrefix(value, "data:image/png;base64,")
	value = strings.TrimPrefix(value, "data:image/webp;base64,")
	return base64.StdEncoding.DecodeString(value)
}
