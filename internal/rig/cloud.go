package rig

type Cloud uint8

const (
	CloudGCP Cloud = iota
	CloudAWS
)

const (
	CloudDefaultLocationGCP = "australia-southeast1"
	CloudDefaultLocationAWS = "ap-southeast-2"
)
