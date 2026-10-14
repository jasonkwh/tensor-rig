package tailscale

import (
	"os"
	"strings"
)

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func loadTailscaleAuthKey(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", ErrLoadAuthKey(err)
	}

	for line := range strings.Lines(string(content)) {
		field, value, ok := strings.Cut(strings.TrimRight(line, "\r\n"), "=")
		if !ok || field != authKeyField || len(value) == 0 {
			continue
		}

		return value, nil
	}

	return "", ErrNoAuthKey
}
