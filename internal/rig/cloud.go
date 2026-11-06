package rig

import (
	"strings"
)

type Cloud string

const (
	CloudGCP Cloud = "gcp"
	CloudAWS Cloud = "aws"
)

func ParseCloud(name string) (Cloud, error) {
	switch Cloud(strings.ToLower(strings.TrimSpace(name))) {
	case "", CloudGCP:
		return CloudGCP, nil
	case CloudAWS:
		return CloudAWS, nil
	default:
		return CloudGCP, ErrUnknownCloud(name)
	}
}

type CloudRigKind string

const (
	CloudRigKindTensor CloudRigKind = "tensor"
	CloudRigKindLadder CloudRigKind = "ladder"
)

func ParseCloudRigKind(name string) (CloudRigKind, error) {
	switch CloudRigKind(strings.ToLower(strings.TrimSpace(name))) {
	case "", CloudRigKindTensor:
		return CloudRigKindTensor, nil
	case CloudRigKindLadder:
		return CloudRigKindLadder, nil
	default:
		return CloudRigKindTensor, ErrUnknownCloudRigKind(name)
	}
}
