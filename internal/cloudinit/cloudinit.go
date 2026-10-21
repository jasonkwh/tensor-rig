package cloudinit

import (
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
)

var _ = adapter.CloudInit(&tsCloudConfig{})

type tsCloudConfig struct {
	filePath  string
	authKey   string
	hostname  string
	loginUser string
}

func New(
	opts ...tsOptions,
) (adapter.CloudInit, error) {
	tsc := &tsCloudConfig{
		filePath:  DefaultFilePath,
		hostname:  defaultHostname,
		loginUser: DefaultLoginUser,
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

func (t *tsCloudConfig) Execute() (string, error) {
	var body strings.Builder

	if err := cloudConfigTmpl.Execute(&body, struct {
		AuthKey         string
		Hostname        string
		LoginUser       string
		CreateLoginUser bool
	}{
		AuthKey:         shellSingleQuote(t.authKey),
		Hostname:        shellSingleQuote(t.hostname),
		LoginUser:       t.loginUser,
		CreateLoginUser: t.loginUser != DefaultLoginUser,
	}); err != nil {
		return "", ErrRenderCloudConfig(err)
	}

	return body.String(), nil
}

func (t *tsCloudConfig) FilePath() string {
	return t.filePath
}
