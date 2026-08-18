package storage

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type core struct {
	cfg *Config
}

type Config struct {
	Name                     string
	Location                 string
	ForceDestroy             bool
	LifecycleRules           storage.BucketLifecycleRuleArray
	VersioningEnabled        bool
	Labels                   map[string]pulumi.StringInput
	UniformBucketLevelAccess bool
	PublicAccessPrevention   PublicAccessPrevention
	StorageClass             StorageClass
}

func DefaultConfig() *Config {
	return &Config{
		Name:         fmt.Sprintf("%s-storage-%s", storageNamePrefix, "au"),
		Location:     "australia-southeast2",
		ForceDestroy: false,
		LifecycleRules: storage.BucketLifecycleRuleArray{
			&storage.BucketLifecycleRuleArgs{
				Action: &storage.BucketLifecycleRuleActionArgs{
					Type: pulumi.String("Delete"),
				},
				Condition: &storage.BucketLifecycleRuleConditionArgs{
					DaysSinceNoncurrentTime: pulumi.Int(7),
				},
			},
		},
		VersioningEnabled:        true,
		Labels:                   make(map[string]pulumi.StringInput),
		UniformBucketLevelAccess: true,
		PublicAccessPrevention:   PublicAccessPreventionEnforced,
		StorageClass:             StorageClassRegional,
	}
}

type PublicAccessPrevention string

const (
	PublicAccessPreventionEnforced  PublicAccessPrevention = "enforced"
	PublicAccessPreventionInherited PublicAccessPrevention = "inherited"
	PublicAccessPreventionOff       PublicAccessPrevention = "off"
)

type StorageClass string

const (
	StorageClassStandard      StorageClass = "STANDARD"
	StorageClassMultiRegional StorageClass = "MULTI_REGIONAL"
	StorageClassRegional      StorageClass = "REGIONAL"
	StorageClassNearline      StorageClass = "NEARLINE"
	StorageClassColdline      StorageClass = "COLDLINE"
	StorageClassArchive       StorageClass = "ARCHIVE"
)
