package adapter

type TailscaleCloudConfig interface {
	Execute() (string, error)
}
