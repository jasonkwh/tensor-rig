package cloudinit

import (
	"errors"
	"fmt"
)

var (
	ErrNoAuthKey         = errors.New("no auth key")
	ErrLoadAuthKey       = errors.New("load auth key")
	ErrRenderCloudConfig = errors.New("render cloud-config")
)

func errLoadAuthKey(err error) error {
	return fmt.Errorf("%w: %w", ErrLoadAuthKey, err)
}

func errRenderCloudConfig(err error) error {
	return fmt.Errorf("%w: %w", ErrRenderCloudConfig, err)
}
