package storage

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type core struct {
	cfg *Config
}

type Config struct {
	Name                     string
	Location                 string
	ForceDestroy             bool
	LifecycleRules           []LifecycleRule
	VersioningEnabled        bool
	Labels                   map[string]pulumi.StringInput
	UniformBucketLevelAccess bool
	PublicAccessPrevention   PublicAccessPrevention
	StorageClass             StorageClass
	SoftDeleteEnabled        bool
}

func DefaultConfig() *Config {
	return &Config{
		Name:                     fmt.Sprintf("%s-storage-%s", storageNamePrefix, "au"),
		Location:                 "australia-southeast2",
		ForceDestroy:             false,
		LifecycleRules:           []LifecycleRule{{DeleteAfterDays: 7}},
		VersioningEnabled:        false,
		Labels:                   make(map[string]pulumi.StringInput),
		UniformBucketLevelAccess: true,
		PublicAccessPrevention:   PublicAccessPreventionEnforced,
		StorageClass:             StorageClassStandard,
		SoftDeleteEnabled:        false,
	}
}

type LifecycleRule struct {
	DeleteAfterDays int
}

type PublicAccessPrevention string

const (
	PublicAccessPreventionEnforced  PublicAccessPrevention = "enforced"
	PublicAccessPreventionInherited PublicAccessPrevention = "inherited"
	PublicAccessPreventionOff       PublicAccessPrevention = "off"
)

type StorageClass string

const (
	StorageClassStandard   StorageClass = "standard"
	StorageClassInfrequent StorageClass = "infrequent"
	StorageClassCold       StorageClass = "cold"
	StorageClassArchive    StorageClass = "archive"
)
