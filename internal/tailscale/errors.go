package tailscale

import (
	"errors"
	"fmt"
)

var (
	ErrNoAuthKey  = errors.New("tailscale: empty auth key")
	ErrNoHostname = errors.New("tailscale: empty hostname")

	ErrRenderCloudConfig = func(err error) error {
		return fmt.Errorf("tailscale: render cloud-config: %w", err)
	}
)
