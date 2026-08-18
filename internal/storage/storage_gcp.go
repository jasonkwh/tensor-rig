package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ adapter.TensorRigStorageBucket = &storageGCP{}

type storageGCP struct {
	core *core
}

func NewGCPStorage(
	opts ...coreOption,
) adapter.TensorRigStorageBucket {
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
) (*adapter.TensorRigStorageBucket, error) {
	_, err := storage.NewBucket(ctx, b.core.cfg.Name, &storage.BucketArgs{
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
	}, opts...)
}
