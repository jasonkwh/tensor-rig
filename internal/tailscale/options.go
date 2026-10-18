package tailscale

type tsOptions func(*tsCloudConfig)

func WithFilePath(filePath string) tsOptions {
	return func(c *tsCloudConfig) {
		c.filePath = filePath
	}
}
