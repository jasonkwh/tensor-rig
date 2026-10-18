package instance

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
