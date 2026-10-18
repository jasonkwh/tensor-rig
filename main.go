package main

import (
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

		store := storage.New(cloud, &storage.Config{
			Name:                     cfg.Get("storageName"),
			Location:                 region,
			ForceDestroy:             cfg.GetBool("storageForceDestroy"),
			LifecycleRules:           []storage.LifecycleRule{{DeleteAfterDays: cfg.GetInt("storageDeleteAfterDays")}},
			VersioningEnabled:        cfg.GetBool("storageVersioning"),
			UniformBucketLevelAccess: cfg.GetBool("storageUniformAccess"),
			PublicAccessPrevention:   storage.PublicAccessPrevention(cfg.Get("storagePublicAccess")),
			StorageClass:             storage.StorageClass(cfg.Get("storageClass")),
			SoftDeleteEnabled:        cfg.GetBool("storageSoftDelete"),
		})

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
