package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ adapter.TensorRigStorageBucket = &storageGCP{}

type storageGCP struct {
	core *storageCore
}

func NewGCPStorage(
	opts ...storageCoreOption,
) *storageGCP {
	core := &storageCore{
		cfg: DefaultStorageConfig(),
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
	name string,
	opts ...pulumi.ResourceOption,
) (*adapter.TensorRigStorageBucket, error) {
	_, err := storage.NewBucket(ctx, b.core.cfg.Name, &storage.BucketArgs{
		Name:                     pulumi.String(b.core.cfg.Name),
		Location:                 pulumi.String(b.core.cfg.Location),
		ForceDestroy:             pulumi.Bool(b.core.cfg.ForceDestroy),
		UniformBucketLevelAccess: pulumi.Bool(true),
		RetentionPolicy: &storage.BucketRetentionPolicyArgs{
			RetentionPeriod: pulumi.String(b.core.cfg.RetentionPeriod.String()),
			IsLocked:        pulumi.Bool(false),
		},
		Versioning: &storage.BucketVersioningArgs{
			Enabled: pulumi.Bool(b.core.cfg.VersioningEnabled),
		},
		PublicAccessPrevention: pulumi.String("enforced"),
		StorageClass:           pulumi.String("STANDARD"),
	}, opts...)
}
