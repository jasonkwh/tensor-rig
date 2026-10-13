package rig

type Cloud uint8

const (
	CloudGCP Cloud = iota
	CloudAWS
	CloudAzure
)

const (
	CloudDefaultLocationGCP   = "australia-southeast1"
	CloudDefaultLocationAWS   = "ap-southeast-2"
	CloudDefaultLocationAzure = "australiaeast"
)
