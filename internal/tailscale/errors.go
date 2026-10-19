package tailscale

import (
	"errors"
	"fmt"
)

var (
	ErrNoAuthKey = errors.New("tailscale: auth key field missing")

	ErrLoadAuthKey = func(err error) error {
		return fmt.Errorf("tailscale: load auth key: %w", err)
	}
	ErrRenderCloudConfig = func(err error) error {
		return fmt.Errorf("tailscale: render cloud-config: %w", err)
	}
)
