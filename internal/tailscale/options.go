package tailscale

type tsOptions func(*tsCloudConfig)

func WithFilePath(filePath string) tsOptions {
	return func(c *tsCloudConfig) {
		c.filePath = filePath
	}
}

func WithHostname(hostname string) tsOptions {
	return func(c *tsCloudConfig) {
		c.hostname = hostname
	}
}

func WithLoginUser(loginUser string) tsOptions {
	return func(c *tsCloudConfig) {
		c.loginUser = loginUser
	}
}
