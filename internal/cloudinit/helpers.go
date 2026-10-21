package cloudinit

import (
	"os"
	"path/filepath"
	"strings"
)

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func loadTailscaleAuthKey(filePath string) (string, error) {
	path, err := expandHomePath(filePath)
	if err != nil {
		return "", ErrLoadAuthKey(err)
	}

	content, err := os.ReadFile(path)
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

func expandHomePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
}
