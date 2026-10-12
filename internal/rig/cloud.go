package rig

type Cloud uint8

const (
	CloudGCP Cloud = iota
	CloudAWS
	CloudAzure
)

const (
	CloudDefaultLocationGCP   string = "australia-southeast2"
	CloudDefaultLocationAWS   string = "ap-southeast-4"
	CloudDefaultLocationAzure string = "australiasoutheast"
)
