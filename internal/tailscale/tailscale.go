package tailscale

import (
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
)

var _ adapter.TailscaleCloudConfig = &tsCloudConfig{}

type tsCloudConfig struct {
	authKey  string
	hostname string
}

func New(
	filePath, hostname string,
) (adapter.TailscaleCloudConfig, error) {
	if len(filePath) == 0 {
		return nil, ErrNoFilePath
	}

	if len(hostname) == 0 {
		return nil, ErrNoHostname
	}

	authKey, err := loadTailscaleAuthKey(filePath)
	if err != nil {
		return nil, err
	}

	return &tsCloudConfig{
		authKey:  authKey,
		hostname: hostname,
	}, nil
}

func (t *tsCloudConfig) Execute() (string, error) {
	var body strings.Builder

	if err := cloudConfigTmpl.Execute(&body, struct {
		AuthKey  string
		Hostname string
	}{
		AuthKey:  shellSingleQuote(t.authKey),
		Hostname: shellSingleQuote(t.hostname),
	}); err != nil {
		return "", ErrRenderCloudConfig(err)
	}

	return body.String(), nil
}
