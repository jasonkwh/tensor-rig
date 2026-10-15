package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

func New(
	cloud rig.Cloud,
	opts ...coreOption,
) adapter.Storage {
	switch cloud {
	case rig.CloudAWS:
		return newAWSStorage(opts...)
	case rig.CloudGCP:
		return newGCPStorage(opts...)
	default:
		return newGCPStorage(opts...)
	}
}
