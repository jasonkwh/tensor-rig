package storage

type coreOption func(*core)

func WithConfiguration(cfg *Config) coreOption {
	return func(c *core) {
		c.cfg = cfg
	}
}
