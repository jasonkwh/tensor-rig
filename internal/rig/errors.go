package rig

import "fmt"

var (
	ErrUnknownCloud = func(name string) error {
		return fmt.Errorf("rig: unknown cloud: %s", name)
	}
)
