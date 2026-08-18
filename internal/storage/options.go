package storage

import (
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type coreOption func(*core)

func WithConfiguration(cfg *Config) coreOption {
	return func(c *core) {
		c.cfg = cfg
	}
}

func WithName(name string) coreOption {
	return func(c *core) {
		c.cfg.Name = name
	}
}

func WithLocation(location string) coreOption {
	return func(c *core) {
		c.cfg.Location = location
	}
}

func WithVersioning(enabled bool) coreOption {
	return func(c *core) {
		c.cfg.VersioningEnabled = enabled
	}
}

func WithForceDestroy(force bool) coreOption {
	return func(c *core) {
		c.cfg.ForceDestroy = force
	}
}

func WithLabels(labels map[string]pulumi.StringInput) coreOption {
	return func(c *core) {
		c.cfg.Labels = labels
	}
}

func WithLifecycleRules(rules storage.BucketLifecycleRuleArray) coreOption {
	return func(c *core) {
		c.cfg.LifecycleRules = rules
	}
}

func WithUniformBucketLevelAccess(uniformBucketLevelAccess bool) coreOption {
	return func(c *core) {
		c.cfg.UniformBucketLevelAccess = uniformBucketLevelAccess
	}
}

func WithPublicAccessPrevention(publicAccessPrevention PublicAccessPrevention) coreOption {
	return func(c *core) {
		c.cfg.PublicAccessPrevention = publicAccessPrevention
	}
}

func WithStorageClass(storageClass StorageClass) coreOption {
	return func(c *core) {
		c.cfg.StorageClass = storageClass
	}
}
