package adapter

type TailscaleCloudConfig interface {
	Execute(hostname string) (string, error)
}
