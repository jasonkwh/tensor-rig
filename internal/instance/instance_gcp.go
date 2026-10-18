package instance

import (
	"fmt"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ = adapter.Instance(&instanceGCP{})

type instanceGCP struct {
	tsCloudCfg adapter.TailscaleCloudConfig
	cfg        *Config
}

func newGCPInstance(tsCloudCfg adapter.TailscaleCloudConfig, cfg *Config) adapter.Instance {
	return &instanceGCP{
		tsCloudCfg: tsCloudCfg,
		cfg:        cfg,
	}
}

func (i *instanceGCP) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	cfg := i.cfg
	scheduling := &compute.InstanceSchedulingArgs{
		OnHostMaintenance: pulumi.String(cfg.OnHostMaintenance),
		ProvisioningModel: pulumi.String(cfg.ProvisioningModel),
	}

	if cfg.ProvisioningModel == "SPOT" {
		scheduling.AutomaticRestart = pulumi.Bool(false)
		scheduling.Preemptible = pulumi.Bool(true)
	}

	nic := &compute.InstanceNetworkInterfaceArgs{
		Network: pulumi.String(cfg.Network),
		NicType: pulumi.String(cfg.NicType),
	}

	if cfg.AssignPublicIP {
		nic.AccessConfigs = compute.InstanceNetworkInterfaceAccessConfigArray{
			&compute.InstanceNetworkInterfaceAccessConfigArgs{},
		}
	}

	metadata, err := i.gcpInstanceMetadata(cfg)
	if err != nil {
		return nil, err
	}

	instanceArgs := &compute.InstanceArgs{
		Name:        pulumi.String(cfg.Name),
		MachineType: pulumi.String(cfg.MachineType),
		Zone:        pulumi.String(fmt.Sprintf("%s-%s", cfg.Location, cfg.Zone)),
		BootDisk: &compute.InstanceBootDiskArgs{
			InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
				Image: pulumi.String(cfg.BootDiskImage),
				Size:  pulumi.Int(cfg.BootDiskSizeGb),
				Type:  pulumi.String(cfg.BootDiskType),
			},
		},
		NetworkInterfaces: compute.InstanceNetworkInterfaceArray{nic},
		ServiceAccount: &compute.InstanceServiceAccountArgs{
			Scopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/devstorage.read_write"),
			},
		},
		Scheduling: scheduling,
		Metadata:   metadata,
	}

	if cfg.AcceleratorCount > 0 {
		instanceArgs.GuestAccelerators = compute.InstanceGuestAcceleratorArray{
			&compute.InstanceGuestAcceleratorArgs{
				Count: pulumi.Int(cfg.AcceleratorCount),
				Type:  pulumi.String(cfg.AcceleratorType),
			},
		}
	}

	instance, err := compute.NewInstance(ctx, cfg.Name, instanceArgs, opts...)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (i *instanceGCP) gcpInstanceMetadata(cfg *Config) (pulumi.StringMap, error) {
	metadata := pulumi.StringMap{
		"install-nvidia-driver": pulumi.String("True"),
	}

	userData, err := i.tsCloudCfg.Execute(cfg.Name)
	if err != nil {
		return nil, err
	}

	if userData != "" {
		metadata["user-data"] = pulumi.String(userData)
	}

	return metadata, nil
}
