package tailscale

import (
	"errors"
	"fmt"
)

var (
	ErrNoFilePath = errors.New("tailscale: empty file path")
	ErrNoAuthKey  = errors.New("tailscale: auth key field missing")
	ErrNoHostname = errors.New("tailscale: empty hostname")

	ErrLoadAuthKey = func(err error) error {
		return fmt.Errorf("tailscale: load auth key: %w", err)
	}
	ErrRenderCloudConfig = func(err error) error {
		return fmt.Errorf("tailscale: render cloud-config: %w", err)
	}
)
