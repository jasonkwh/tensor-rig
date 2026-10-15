package tailscale

import (
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
)

var _ = adapter.TailscaleCloudConfig(&tsCloudConfig{})

type tsCloudConfig struct {
	authKey string
}

func New(
	filePath string,
) (adapter.TailscaleCloudConfig, error) {
	if len(filePath) == 0 {
		return nil, ErrNoFilePath
	}

	authKey, err := loadTailscaleAuthKey(filePath)
	if err != nil {
		return nil, err
	}

	return &tsCloudConfig{
		authKey: authKey,
	}, nil
}

func (t *tsCloudConfig) Execute(hostname string) (string, error) {
	var body strings.Builder

	if err := cloudConfigTmpl.Execute(&body, struct {
		AuthKey  string
		Hostname string
	}{
		AuthKey:  shellSingleQuote(t.authKey),
		Hostname: shellSingleQuote(hostname),
	}); err != nil {
		return "", ErrRenderCloudConfig(err)
	}

	return body.String(), nil
}
