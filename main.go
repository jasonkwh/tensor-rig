package main

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/instance"
	"github.com/jasonkwh/tensor-rig/internal/rig"
	"github.com/jasonkwh/tensor-rig/internal/storage"
	"github.com/jasonkwh/tensor-rig/internal/tailscale"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		cloud, err := rig.ParseCloud(cfg.Get("cloud"))
		if err != nil {
			return err
		}

		region := cfg.Get("location")

		store := newStorage(cfg, cloud, region)
		if _, err = store.Create(ctx); err != nil {
			return err
		}

		tsConfig, err := tailscale.New()
		if err != nil {
			return err
		}

		vm := instance.New(cloud, tsConfig, instance.WithLocation(region))
		if _, err = vm.Create(ctx); err != nil {
			return err
		}

		return nil
	})
}

func newStorage(cfg *config.Config, cloud rig.Cloud, region string) adapter.Storage {
	return storage.New(cloud,
		storage.WithName(cfg.Get("storageName")),
		storage.WithLocation(region),
		storage.WithForceDestroy(cfg.GetBool("storageForceDestroy")),
		storage.WithVersioning(cfg.GetBool("storageVersioning")),
		storage.WithUniformBucketLevelAccess(cfg.GetBool("storageUniformAccess")),
		storage.WithPublicAccessPrevention(storage.PublicAccessPrevention(cfg.Get("storagePublicAccess"))),
		storage.WithStorageClass(storage.StorageClass(cfg.Get("storageClass"))),
		storage.WithSoftDelete(cfg.GetBool("storageSoftDelete")),
		storage.WithLifecycleRules([]storage.LifecycleRule{{
			DeleteAfterDays: cfg.GetInt("storageDeleteAfterDays"),
		}}),
	)
}
