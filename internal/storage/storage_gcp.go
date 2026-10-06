package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ adapter.TensorRigStorageInterface = &storageGCP{}

type storageGCP struct {
	core *core
}

func NewGCPStorage(
	opts ...coreOption,
) adapter.TensorRigStorageInterface {
	core := &core{
		cfg: DefaultConfig(),
	}

	for _, opt := range opts {
		opt(core)
	}

	return &storageGCP{
		core: core,
	}
}

func (b *storageGCP) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	return storage.NewBucket(ctx, b.core.cfg.Name, &storage.BucketArgs{
		Name:                     pulumi.String(b.core.cfg.Name),
		Location:                 pulumi.String(b.core.cfg.Location),
		ForceDestroy:             pulumi.Bool(b.core.cfg.ForceDestroy),
		UniformBucketLevelAccess: pulumi.Bool(b.core.cfg.UniformBucketLevelAccess),
		LifecycleRules:           gcpLifecycleRules(b.core.cfg.LifecycleRules),
		Versioning: &storage.BucketVersioningArgs{
			Enabled: pulumi.Bool(b.core.cfg.VersioningEnabled),
		},
		PublicAccessPrevention: pulumi.String(string(b.core.cfg.PublicAccessPrevention)),
		StorageClass:           pulumi.String(gcpStorageClass(b.core.cfg.StorageClass)),
		Labels:                 pulumi.StringMap(b.core.cfg.Labels),
		SoftDeletePolicy:       getSoftDeletePolicy(b.core.cfg.SoftDeleteEnabled),
	}, opts...)
}

func gcpLifecycleRules(rules []LifecycleRule) storage.BucketLifecycleRuleArray {
	out := make(storage.BucketLifecycleRuleArray, 0, len(rules))
	for _, rule := range rules {
		if rule.DeleteAfterDays <= 0 {
			continue
		}
		out = append(out, &storage.BucketLifecycleRuleArgs{
			Action: &storage.BucketLifecycleRuleActionArgs{
				Type: pulumi.String("Delete"),
			},
			Condition: &storage.BucketLifecycleRuleConditionArgs{
				Age: pulumi.Int(rule.DeleteAfterDays),
			},
		})
	}
	return out
}

func gcpStorageClass(class StorageClass) string {
	switch class {
	case StorageClassInfrequent:
		return "NEARLINE"
	case StorageClassCold:
		return "COLDLINE"
	case StorageClassArchive:
		return "ARCHIVE"
	default:
		return "STANDARD"
	}
}

func getSoftDeletePolicy(enabled bool) *storage.BucketSoftDeletePolicyArgs {
	if enabled {
		return &storage.BucketSoftDeletePolicyArgs{
			RetentionDurationSeconds: pulumi.Int(7 * 24 * 60 * 60),
		}
	}

	return &storage.BucketSoftDeletePolicyArgs{
		RetentionDurationSeconds: pulumi.Int(0),
	}
}
