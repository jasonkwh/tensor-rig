package rig

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownCloud = errors.New("unknown cloud name")
)

func errUnknownCloud(name string) error {
	return fmt.Errorf("%w: %s", ErrUnknownCloud, name)
}
