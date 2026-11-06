package cloudinit

const (
	DefaultAuthKeyPath = "~/.secrets/headless-env"
	authKeyField       = "ts_auth_key_rig"
	DefaultLoginUser   = "ubuntu"
	defaultHostname    = "tensor-rig-instance-au"

	DefaultLadderUUIDPath       = "~/.secrets/proxy-env/uuid"
	DefaultLadderPrivateKeyPath = "~/.secrets/proxy-env/private-key"
	ladderServerName            = "www.cloudflare.com"
	ladderShortID               = "d8424154"
	ladderListenPort            = 443
	singBoxVersion              = "1.14.4"
)
