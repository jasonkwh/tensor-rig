package instance

import (
	"fmt"

	"github.com/jasonkwh/tensor-rig/internal/rig"
)

type core struct {
	cfg *Config
}

type Config struct {
	Name              string
	Location          string
	Zone              string
	MachineType       string
	BootDiskImage     string
	BootDiskSizeGb    int
	BootDiskType      string
	Network           string
	NicType           string
	SshPublicKey      string
	AssignPublicIP    bool
	AcceleratorType   string
	AcceleratorCount  int
	OnHostMaintenance string
	ProvisioningModel string
}

func defaultGCPInstanceConfig() *Config {
	return &Config{
		Name: fmt.Sprintf("%s-instance-%s",
			rig.ResourceNamePrefix, "au"),
		Location:          rig.CloudDefaultLocationGCP,
		Zone:              "a",
		MachineType:       "n1-standard-4",
		BootDiskImage:     "projects/deeplearning-platform-release/global/images/family/pytorch-2-9-cu129-ubuntu-2204-nvidia-580",
		BootDiskSizeGb:    100,
		BootDiskType:      "pd-balanced",
		Network:           "default",
		NicType:           "VIRTIO_NET",
		AssignPublicIP:    true,
		AcceleratorType:   "nvidia-tesla-t4",
		AcceleratorCount:  1,
		OnHostMaintenance: "TERMINATE",
		ProvisioningModel: "SPOT",
	}
}

func defaultAzureInstanceConfig() *Config {
	return &Config{
		Name: fmt.Sprintf("%s-instance-%s",
			rig.ResourceNamePrefix, "au"),
		Location:          rig.CloudDefaultLocationAzure,
		Zone:              "1",
		MachineType:       "Standard_NC4as_T4_v3",
		BootDiskImage:     "microsoft-dsvm:ubuntu-hpc:2204:latest",
		BootDiskSizeGb:    100,
		BootDiskType:      "StandardSSD_LRS",
		NicType:           "Accelerated",
		AssignPublicIP:    true,
		ProvisioningModel: "Spot",
	}
}

func defaultAWSInstanceConfig() *Config {
	return &Config{
		Name: fmt.Sprintf("%s-instance-%s",
			rig.ResourceNamePrefix, "au"),
		Location:          rig.CloudDefaultLocationAWS,
		Zone:              "a",
		MachineType:       "g4dn.xlarge",
		BootDiskImage:     "/aws/service/deeplearning/ami/x86_64/oss-nvidia-driver-gpu-pytorch-2.7-ubuntu-22.04/latest/ami-id",
		BootDiskSizeGb:    100,
		BootDiskType:      "gp3",
		Network:           "default",
		NicType:           "ena",
		AssignPublicIP:    true,
		OnHostMaintenance: "terminate",
		ProvisioningModel: "spot",
	}
}
