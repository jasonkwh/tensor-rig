package tailscale

import (
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
)

var _ = adapter.TailscaleCloudConfig(&tsCloudConfig{})

type tsCloudConfig struct {
	filePath string
	authKey  string
}

func New(
	opts ...tsOptions,
) (adapter.TailscaleCloudConfig, error) {
	tsc := &tsCloudConfig{
		filePath: DefaultFilePath,
	}

	for _, opt := range opts {
		opt(tsc)
	}

	var err error

	tsc.authKey, err = loadTailscaleAuthKey(tsc.filePath)
	if err != nil {
		return tsc, err
	}

	return tsc, nil
}

func (t *tsCloudConfig) Execute(hostname, loginUser string) (string, error) {
	var body strings.Builder

	if err := cloudConfigTmpl.Execute(&body, struct {
		AuthKey         string
		Hostname        string
		LoginUser       string
		CreateLoginUser bool
	}{
		AuthKey:         shellSingleQuote(t.authKey),
		Hostname:        shellSingleQuote(hostname),
		LoginUser:       loginUser,
		CreateLoginUser: loginUser != ImageLoginUser,
	}); err != nil {
		return "", ErrRenderCloudConfig(err)
	}

	return body.String(), nil
}

func (t *tsCloudConfig) FilePath() string {
	return t.filePath
}
