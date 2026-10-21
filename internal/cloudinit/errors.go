package cloudinit

import (
	"errors"
	"fmt"
)

var (
	ErrNoAuthKey = errors.New("cloudinit: auth key field missing")

	ErrLoadAuthKey = func(err error) error {
		return fmt.Errorf("cloudinit: load auth key: %w", err)
	}
	ErrRenderCloudConfig = func(err error) error {
		return fmt.Errorf("cloudinit: render cloud-config: %w", err)
	}
)
