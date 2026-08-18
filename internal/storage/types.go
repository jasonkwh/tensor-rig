package storage

import (
	"fmt"
	"time"
)

type core struct {
	cfg *Config
}

type Config struct {
	Name                   string
	Location               string
	ForceDestroy           bool
	RetentionPeriod        time.Duration
	VersioningEnabled      bool
	Labels                 map[string]string
	PublicAccessPrevention PublicAccessPrevention
}

func DefaultConfig() *Config {
	return &Config{
		Name:                   fmt.Sprintf("%s-storage-%s", storageNamePrefix, "au"),
		Location:               "australia-southeast2",
		ForceDestroy:           false,
		RetentionPeriod:        7 * 24 * time.Hour,
		VersioningEnabled:      true,
		Labels:                 make(map[string]string),
		PublicAccessPrevention: PublicAccessPreventionEnforced,
	}
}

type PublicAccessPrevention string

const (
	PublicAccessPreventionEnforced  PublicAccessPrevention = "enforced"
	PublicAccessPreventionInherited PublicAccessPrevention = "inherited"
	PublicAccessPreventionOff       PublicAccessPrevention = "off"
)
