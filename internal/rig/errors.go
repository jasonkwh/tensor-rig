package rig

import "fmt"

var (
	ErrUnknownCloud = func(name string) error {
		return fmt.Errorf("rig: unknown cloud: %s", name) //nolint:err113
	}
	ErrUnknownCloudRigKind = func(name string) error {
		return fmt.Errorf("rig: unknown cloud rig kind: %s", name) //nolint:err113
	}
)
