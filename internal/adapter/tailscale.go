package adapter

type TailscaleCloudConfig interface {
	Execute(
		hostname string,
		loginUser string,
	) (string, error)

	FilePath() string
}
