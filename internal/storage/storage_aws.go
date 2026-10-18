package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ = adapter.Storage(&storageAWS{})

type storageAWS struct {
	cfg *Config
}

func newAWSStorage(cfg *Config) adapter.Storage {
	return &storageAWS{
		cfg: cfg,
	}
}

func (b *storageAWS) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	bucket, err := s3.NewBucket(ctx, b.cfg.Name, &s3.BucketArgs{
		Bucket:       pulumi.String(b.cfg.Name),
		Region:       pulumi.String(b.cfg.Location),
		ForceDestroy: pulumi.Bool(b.cfg.ForceDestroy),
		Tags:         pulumi.StringMap(b.cfg.Labels),
	}, opts...)
	if err != nil {
		return nil, err
	}

	if b.cfg.VersioningEnabled {
		_, err = s3.NewBucketVersioning(ctx, b.cfg.Name+"-versioning", &s3.BucketVersioningArgs{
			Bucket: bucket.Bucket,
			Region: pulumi.String(b.cfg.Location),
			VersioningConfiguration: &s3.BucketVersioningVersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	if b.cfg.UniformBucketLevelAccess {
		_, err = s3.NewBucketOwnershipControls(ctx, b.cfg.Name+"-ownership", &s3.BucketOwnershipControlsArgs{
			Bucket: bucket.Bucket,
			Region: pulumi.String(b.cfg.Location),
			Rule: &s3.BucketOwnershipControlsRuleArgs{
				ObjectOwnership: pulumi.String("BucketOwnerEnforced"),
			},
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	if rules := awsLifecycleRules(b.cfg); len(rules) > 0 {
		_, err = s3.NewBucketLifecycleConfiguration(ctx, b.cfg.Name+"-lifecycle", &s3.BucketLifecycleConfigurationArgs{
			Bucket: bucket.Bucket,
			Region: pulumi.String(b.cfg.Location),
			Rules:  rules,
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	if blocked, ok := awsPublicAccessBlocked(b.cfg.PublicAccessPrevention); ok {
		_, err = s3.NewBucketPublicAccessBlock(ctx, b.cfg.Name+"-public-access", &s3.BucketPublicAccessBlockArgs{
			Bucket:                bucket.Bucket,
			Region:                pulumi.String(b.cfg.Location),
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

	if rule := awsStorageClassRule(cfg.StorageClass); rule != nil {
		rules = append(rules, rule)
	}

	return rules
}

func awsStorageClassRule(class StorageClass) *s3.BucketLifecycleConfigurationRuleArgs {
	var (
		storageClass string
		days         int
	)

	switch class {
	case StorageClassInfrequent:
		storageClass, days = "STANDARD_IA", 30
	case StorageClassCold:
		storageClass = "GLACIER_IR"
	case StorageClassArchive:
		storageClass = "DEEP_ARCHIVE"
	default:
		return nil
	}

	transition := &s3.BucketLifecycleConfigurationRuleTransitionArgs{
		StorageClass: pulumi.String(storageClass),
	}
	if days > 0 {
		transition.Days = pulumi.Int(days)
	}

	return &s3.BucketLifecycleConfigurationRuleArgs{
		Id:     pulumi.String("storage-class"),
		Status: pulumi.String("Enabled"),
		Filter: &s3.BucketLifecycleConfigurationRuleFilterArgs{
			Prefix: pulumi.String(""),
		},
		Transitions: s3.BucketLifecycleConfigurationRuleTransitionArray{transition},
	}
}

func awsPublicAccessBlocked(
	prevention PublicAccessPrevention,
) (bool, bool) {
	switch prevention {
	case PublicAccessPreventionEnforced:
		return true, true
	case PublicAccessPreventionOff:
		return false, true
	default:
		return false, false
	}
}
