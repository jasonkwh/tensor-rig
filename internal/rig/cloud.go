package rig

import (
	"strings"
)

type Cloud string

const (
	CloudGCP Cloud = "gcp"
	CloudAWS Cloud = "aws"
)

const (
	CloudDefaultLocationGCP = "australia-southeast1"
	CloudDefaultLocationAWS = "ap-southeast-2"
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
