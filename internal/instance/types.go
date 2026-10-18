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
	NicType           string // GCP only.
	AssignPublicIP    bool
	AcceleratorType   string // GCP only.
	AcceleratorCount  int    // GCP only.
	OnHostMaintenance string
	ProvisioningModel string
}

func defaultGCPInstanceConfig() *Config {
	return &Config{
		Name: fmt.Sprintf("%s-instance-%s",
			rig.ResourceNamePrefix, "au"),
		Zone:              "a",
		MachineType:       "n1-standard-4",
		BootDiskImage:     gcpDefaultBootDiskImage,
		BootDiskSizeGb:    defaultBootSizeGB,
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

func defaultAWSInstanceConfig() *Config {
	return &Config{
		Name: fmt.Sprintf("%s-instance-%s",
			rig.ResourceNamePrefix, "au"),
		Zone:              "a",
		MachineType:       "g4dn.xlarge",
		BootDiskImage:     awsDefaultBootDiskImage,
		BootDiskSizeGb:    defaultBootSizeGB,
		BootDiskType:      "gp3",
		Network:           "default",
		AssignPublicIP:    true,
		OnHostMaintenance: "terminate",
		ProvisioningModel: "spot",
	}
}
