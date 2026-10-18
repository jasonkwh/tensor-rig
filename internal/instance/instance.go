package instance

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

func New(
	cloud rig.Cloud,
	tsCloudCfg adapter.TailscaleCloudConfig,
	cfg *Config,
) adapter.Instance {
	switch cloud {
	case rig.CloudAWS:
		return newAWSInstance(tsCloudCfg, cfg)
	case rig.CloudGCP:
		return newGCPInstance(tsCloudCfg, cfg)
	default:
		return newGCPInstance(tsCloudCfg, cfg)
	}
}
