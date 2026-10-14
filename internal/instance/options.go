package instance

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

func WithZone(zone string) coreOption {
	return func(c *core) {
		c.cfg.Zone = zone
	}
}

func WithMachineType(machineType string) coreOption {
	return func(c *core) {
		c.cfg.MachineType = machineType
	}
}

func WithBootDiskImage(image string) coreOption {
	return func(c *core) {
		c.cfg.BootDiskImage = image
	}
}

func WithBootDiskSizeGb(sizeGb int) coreOption {
	return func(c *core) {
		c.cfg.BootDiskSizeGb = sizeGb
	}
}

func WithBootDiskType(diskType string) coreOption {
	return func(c *core) {
		c.cfg.BootDiskType = diskType
	}
}

func WithNetwork(network string) coreOption {
	return func(c *core) {
		c.cfg.Network = network
	}
}

func WithNicType(nicType string) coreOption {
	return func(c *core) {
		c.cfg.NicType = nicType
	}
}

func WithTailscaleAuthKey(authKey string) coreOption {
	return func(c *core) {
		c.cfg.TailscaleAuthKey = authKey
	}
}

func WithAssignPublicIP(assign bool) coreOption {
	return func(c *core) {
		c.cfg.AssignPublicIP = assign
	}
}

func WithAcceleratorType(acceleratorType string) coreOption {
	return func(c *core) {
		c.cfg.AcceleratorType = acceleratorType
	}
}

func WithAcceleratorCount(count int) coreOption {
	return func(c *core) {
		c.cfg.AcceleratorCount = count
	}
}

func WithOnHostMaintenance(action string) coreOption {
	return func(c *core) {
		c.cfg.OnHostMaintenance = action
	}
}

func WithProvisioningModel(model string) coreOption {
	return func(c *core) {
		c.cfg.ProvisioningModel = model
	}
}
