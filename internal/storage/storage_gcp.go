package storage

import (
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ TensorRigStorage = &storageGCP{}

type storageGCP struct {
	core *core
}

func NewGCPStorage(
	opts ...coreOption,
) TensorRigStorage {
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
) (*storage.Bucket, error) {
	return storage.NewBucket(ctx, b.core.cfg.Name, &storage.BucketArgs{
		Name:                     pulumi.String(b.core.cfg.Name),
		Location:                 pulumi.String(b.core.cfg.Location),
		ForceDestroy:             pulumi.Bool(b.core.cfg.ForceDestroy),
		UniformBucketLevelAccess: pulumi.Bool(b.core.cfg.UniformBucketLevelAccess),
		LifecycleRules:           b.core.cfg.LifecycleRules,
		Versioning: &storage.BucketVersioningArgs{
			Enabled: pulumi.Bool(b.core.cfg.VersioningEnabled),
		},
		PublicAccessPrevention: pulumi.String(string(b.core.cfg.PublicAccessPrevention)),
		StorageClass:           pulumi.String(string(b.core.cfg.StorageClass)),
		Labels:                 pulumi.StringMap(b.core.cfg.Labels),
		SoftDeletePolicy:       getSoftDeletePolicy(b.core.cfg.SoftDeleteEnabled),
	}, opts...)
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
