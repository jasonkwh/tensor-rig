package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ adapter.TensorRigStorage = &storageAWS{}

type storageAWS struct {
	core *core
}

func NewAWSStorage(
	opts ...coreOption,
) adapter.TensorRigStorage {
	core := &core{
		cfg: DefaultConfig(),
	}

	for _, opt := range opts {
		opt(core)
	}

	return &storageAWS{
		core: core,
	}
}

func (b *storageAWS) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	bucket, err := s3.NewBucket(ctx, b.core.cfg.Name, &s3.BucketArgs{
		Bucket:       pulumi.String(b.core.cfg.Name),
		Region:       pulumi.String(b.core.cfg.Location),
		ForceDestroy: pulumi.Bool(b.core.cfg.ForceDestroy),
		Tags:         pulumi.StringMap(b.core.cfg.Labels),
	}, opts...)
	if err != nil {
		return nil, err
	}

	if b.core.cfg.VersioningEnabled {
		_, err = s3.NewBucketVersioning(ctx, b.core.cfg.Name+"-versioning", &s3.BucketVersioningArgs{
			Bucket: bucket.Bucket,
			Region: pulumi.String(b.core.cfg.Location),
			VersioningConfiguration: &s3.BucketVersioningVersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	if b.core.cfg.UniformBucketLevelAccess {
		_, err = s3.NewBucketOwnershipControls(ctx, b.core.cfg.Name+"-ownership", &s3.BucketOwnershipControlsArgs{
			Bucket: bucket.Bucket,
			Region: pulumi.String(b.core.cfg.Location),
			Rule: &s3.BucketOwnershipControlsRuleArgs{
				ObjectOwnership: pulumi.String("BucketOwnerEnforced"),
			},
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	if rules := awsLifecycleRules(b.core.cfg); len(rules) > 0 {
		_, err = s3.NewBucketLifecycleConfiguration(ctx, b.core.cfg.Name+"-lifecycle", &s3.BucketLifecycleConfigurationArgs{
			Bucket: bucket.Bucket,
			Region: pulumi.String(b.core.cfg.Location),
			Rules:  rules,
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	if blocked, ok := awsPublicAccessBlocked(b.core.cfg.PublicAccessPrevention); ok {
		_, err = s3.NewBucketPublicAccessBlock(ctx, b.core.cfg.Name+"-public-access", &s3.BucketPublicAccessBlockArgs{
			Bucket:                bucket.Bucket,
			Region:                pulumi.String(b.core.cfg.Location),
			BlockPublicAcls:       pulumi.Bool(blocked),
			BlockPublicPolicy:     pulumi.Bool(blocked),
			IgnorePublicAcls:      pulumi.Bool(blocked),
			RestrictPublicBuckets: pulumi.Bool(blocked),
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	return bucket, nil
}

func awsLifecycleRules(cfg *Config) s3.BucketLifecycleConfigurationRuleArray {
	rules := make(s3.BucketLifecycleConfigurationRuleArray, 0, len(cfg.LifecycleRules)+1)
	for i, rule := range cfg.LifecycleRules {
		if rule.DeleteAfterDays <= 0 {
			continue
		}
		rules = append(rules, &s3.BucketLifecycleConfigurationRuleArgs{
			Id:     pulumi.Sprintf("delete-%d", i),
			Status: pulumi.String("Enabled"),
			Filter: &s3.BucketLifecycleConfigurationRuleFilterArgs{
				Prefix: pulumi.String(""),
			},
			Expiration: &s3.BucketLifecycleConfigurationRuleExpirationArgs{
				Days: pulumi.Int(rule.DeleteAfterDays),
			},
		})
	}

	if class, days, ok := awsStorageClassTransition(cfg.StorageClass); ok {
		transition := &s3.BucketLifecycleConfigurationRuleTransitionArgs{
			StorageClass: pulumi.String(class),
		}
		if days > 0 {
			transition.Days = pulumi.Int(days)
		}
		rules = append(rules, &s3.BucketLifecycleConfigurationRuleArgs{
			Id:     pulumi.String("storage-class"),
			Status: pulumi.String("Enabled"),
			Filter: &s3.BucketLifecycleConfigurationRuleFilterArgs{
				Prefix: pulumi.String(""),
			},
			Transitions: s3.BucketLifecycleConfigurationRuleTransitionArray{transition},
		})
	}

	return rules
}

func awsStorageClassTransition(class StorageClass) (storageClass string, days int, ok bool) {
	switch class {
	case StorageClassInfrequent:
		return "STANDARD_IA", 30, true
	case StorageClassCold:
		return "GLACIER_IR", 0, true
	case StorageClassArchive:
		return "DEEP_ARCHIVE", 0, true
	default:
		return "", 0, false
	}
}

func awsPublicAccessBlocked(prevention PublicAccessPrevention) (blocked bool, set bool) {
	switch prevention {
	case PublicAccessPreventionEnforced:
		return true, true
	case PublicAccessPreventionOff:
		return false, true
	default:
		return false, false
	}
}
