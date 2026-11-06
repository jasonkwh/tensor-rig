package cloudinit

import (
	"errors"
	"fmt"
)

var (
	ErrNoAuthKey          = errors.New("cloudinit: auth key field missing")
	ErrNoLadderUUID       = errors.New("cloudinit: ladder uuid missing")
	ErrNoLadderPrivateKey = errors.New("cloudinit: ladder private key missing")

	ErrLoadAuthKey = func(err error) error {
		return fmt.Errorf("cloudinit: load auth key: %w", err)
	}
	ErrLoadLadderUUID = func(err error) error {
		return fmt.Errorf("cloudinit: load ladder uuid: %w", err)
	}
	ErrLoadLadderPrivateKey = func(err error) error {
		return fmt.Errorf("cloudinit: load ladder private key: %w", err)
	}
	ErrRenderCloudConfig = func(err error) error {
		return fmt.Errorf("cloudinit: render cloud-config: %w", err)
	}
)
