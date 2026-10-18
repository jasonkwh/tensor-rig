package storage

import (
	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

func New(
	cloud rig.Cloud,
	cfg *Config,
) adapter.Storage {
	switch cloud {
	case rig.CloudAWS:
		return newAWSStorage(cfg)
	case rig.CloudGCP:
		return newGCPStorage(cfg)
	default:
		return newGCPStorage(cfg)
	}
}
