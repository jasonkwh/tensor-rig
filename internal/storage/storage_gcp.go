package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ = adapter.Storage(&storageGCP{})

type storageGCP struct {
	cfg *Config
}

func newGCPStorage(cfg *Config) adapter.Storage {
	return &storageGCP{
		cfg: cfg,
	}
}

func (b *storageGCP) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	bucket, err := storage.NewBucket(ctx, b.cfg.Name, &storage.BucketArgs{
		Name:                     pulumi.String(b.cfg.Name),
		Location:                 pulumi.String(b.cfg.Location),
		ForceDestroy:             pulumi.Bool(b.cfg.ForceDestroy),
		UniformBucketLevelAccess: pulumi.Bool(b.cfg.UniformBucketLevelAccess),
		LifecycleRules:           gcpLifecycleRules(b.cfg.LifecycleRules),
		Versioning: &storage.BucketVersioningArgs{
			Enabled: pulumi.Bool(b.cfg.VersioningEnabled),
		},
		PublicAccessPrevention: pulumi.String(string(b.cfg.PublicAccessPrevention)),
		StorageClass:           pulumi.String(gcpStorageClass(b.cfg.StorageClass)),
		Labels:                 pulumi.StringMap(b.cfg.Labels),
		SoftDeletePolicy:       getSoftDeletePolicy(b.cfg.SoftDeleteEnabled),
	}, opts...)
	if err != nil {
		return nil, err
	}

	account, err := compute.GetDefaultServiceAccount(ctx, &compute.GetDefaultServiceAccountArgs{})
	if err != nil {
		return nil, err
	}

	_, err = storage.NewBucketIAMMember(ctx, b.cfg.Name+"-instance", &storage.BucketIAMMemberArgs{
		Bucket: bucket.Name,
		Role:   pulumi.String("roles/storage.objectAdmin"),
		Member: pulumi.String(account.Member),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return bucket, nil
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
