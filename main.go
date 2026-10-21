package main

import (
	"github.com/jasonkwh/tensor-rig/internal/cloudinit"
	"github.com/jasonkwh/tensor-rig/internal/instance"
	"github.com/jasonkwh/tensor-rig/internal/rig"
	"github.com/jasonkwh/tensor-rig/internal/storage"
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

		bucket, err := store.Create(ctx)
		if err != nil {
			return err
		}

		tsConfig, err := cloudinit.New(
			cloudinit.WithLoginUser(cfg.Get("cloudInitLoginUser")),
			cloudinit.WithHostname(cfg.Get("cloudInitHostname")))
		if err != nil {
			return err
		}

		vm := instance.New(cloud, tsConfig, &instance.Config{
			Name:                  cfg.Get("instanceName"),
			Location:              region,
			Zone:                  cfg.Get("instanceZone"),
			MachineType:           cfg.Get("instanceMachineType"),
			BootDiskImage:         cfg.Get("instanceBootDiskImage"),
			BootDiskSizeGb:        cfg.GetInt("instanceBootDiskSizeGb"),
			BootDiskType:          cfg.Get("instanceBootDiskType"),
			Network:               cfg.Get("instanceNetwork"),
			NicType:               cfg.Get("instanceNicType"),
			AssignPublicIP:        cfg.GetBool("instanceAssignPublicIP"),
			RestrictPublicIngress: cfg.GetBool("instanceRestrictPublicIngress"),
			AcceleratorType:       cfg.Get("instanceAcceleratorType"),
			AcceleratorCount:      cfg.GetInt("instanceAcceleratorCount"),
			OnHostMaintenance:     cfg.Get("instanceOnHostMaintenance"),
			ProvisioningModel:     instance.ProvisioningModel(cfg.Get("instanceProvisioningModel")),
		})

		if _, err = vm.Create(ctx, pulumi.DependsOn([]pulumi.Resource{bucket})); err != nil {
			return err
		}

		return nil
	})
}
