package cloudinit

import "github.com/jasonkwh/tensor-rig/internal/rig"

type tsOptions func(*tsCloudConfig)

func WithAuthKeyPath(filePath string) tsOptions {
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

func WithLadderUUIDPath(filePath string) tsOptions {
	return func(c *tsCloudConfig) {
		c.ladderUUIDPath = filePath
	}
}

func WithLadderPrivateKeyPath(filePath string) tsOptions {
	return func(c *tsCloudConfig) {
		c.ladderPrivateKeyPath = filePath
	}
}

func WithRigKind(rigKind rig.CloudRigKind) tsOptions {
	return func(c *tsCloudConfig) {
		c.rigKind = rigKind
	}
}
