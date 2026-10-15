package instance

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

func New(
	cloud rig.Cloud,
	tsCloudCfg adapter.TailscaleCloudConfig,
	opts ...coreOption,
) adapter.Instance {
	switch cloud {
	case rig.CloudAWS:
		return newAWSInstance(tsCloudCfg, opts...)
	case rig.CloudAzure:
		return newAzureInstance(tsCloudCfg, opts...)
	case rig.CloudGCP:
		return newGCPInstance(tsCloudCfg, opts...)
	default:
		return newGCPInstance(tsCloudCfg, opts...)
	}
}
