package adapter

type CloudInit interface {
	Execute() (string, error)
	FilePath() string
}
