package instance

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

func New(
	cloud rig.Cloud,
	opts ...coreOption,
) adapter.Instance {
	switch cloud {
	case rig.CloudAWS:
		return newAWSInstance(opts...)
	case rig.CloudAzure:
		return newAzureInstance(opts...)
	case rig.CloudGCP:
		return newGCPInstance(opts...)
	default:
		return newGCPInstance(opts...)
	}
}
