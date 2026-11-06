package cloudinit

import (
	"strings"
	"text/template"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

var _ = adapter.CloudInit(&tsCloudConfig{})

type tsCloudConfig struct {
	filePath             string
	authKey              string
	hostname             string
	loginUser            string
	rigKind              rig.CloudRigKind
	ladderUUIDPath       string
	ladderPrivateKeyPath string
	ladderUUID           string
	ladderPrivateKey     string
}

type cloudConfigView struct {
	AuthKey         string
	Hostname        string
	LoginUser       string
	CreateLoginUser bool
	ExtraFiles      string
	ExtraRunCmd     string
}

func New(
	opts ...tsOptions,
) (adapter.CloudInit, error) {
	tsc := &tsCloudConfig{
		filePath:             DefaultAuthKeyPath,
		hostname:             defaultHostname,
		loginUser:            DefaultLoginUser,
		rigKind:              rig.CloudRigKindTensor,
		ladderUUIDPath:       DefaultLadderUUIDPath,
		ladderPrivateKeyPath: DefaultLadderPrivateKeyPath,
	}

	for _, opt := range opts {
		opt(tsc)
	}

	var err error

	tsc.authKey, err = loadTailscaleAuthKey(tsc.filePath)
	if err != nil {
		return tsc, err
	}

	if tsc.rigKind != rig.CloudRigKindLadder {
		return tsc, nil
	}

	tsc.ladderUUID, err = loadTrimmedFile(tsc.ladderUUIDPath, ErrLoadLadderUUID, ErrNoLadderUUID)
	if err != nil {
		return tsc, err
	}

	tsc.ladderPrivateKey, err = loadTrimmedFile(
		tsc.ladderPrivateKeyPath,
		ErrLoadLadderPrivateKey,
		ErrNoLadderPrivateKey,
	)

	return tsc, err
}

func (t *tsCloudConfig) Execute() (string, error) {
	if t.rigKind == rig.CloudRigKindLadder {
		return t.executeLadder()
	}

	return t.executeTensor()
}

func (t *tsCloudConfig) FilePath() string {
	return t.filePath
}

func (t *tsCloudConfig) executeTensor() (string, error) {
	return renderCloudConfig(cloudConfigTmpl, t.cloudConfigView("", ""))
}

func (t *tsCloudConfig) executeLadder() (string, error) {
	config, err := ladderSingBoxConfig(t.ladderUUID, t.ladderPrivateKey)
	if err != nil {
		return "", ErrRenderCloudConfig(err)
	}

	files, err := renderCloudConfig(ladderFilesTmpl, struct {
		SingBoxConfig  string
		SingBoxVersion string
	}{
		SingBoxConfig:  indentLines(config, "      "),
		SingBoxVersion: singBoxVersion,
	})
	if err != nil {
		return "", err
	}

	return renderCloudConfig(cloudConfigTmpl, t.cloudConfigView(
		strings.TrimRight(files, "\n")+"\n",
		ladderRunCmd+"\n",
	))
}

func (t *tsCloudConfig) cloudConfigView(extraFiles, extraRunCmd string) cloudConfigView {
	return cloudConfigView{
		AuthKey:         shellSingleQuote(t.authKey),
		Hostname:        shellSingleQuote(t.hostname),
		LoginUser:       t.loginUser,
		CreateLoginUser: t.loginUser != DefaultLoginUser,
		ExtraFiles:      extraFiles,
		ExtraRunCmd:     extraRunCmd,
	}
}

func renderCloudConfig(tmpl *template.Template, data any) (string, error) {
	var body strings.Builder

	if err := tmpl.Execute(&body, data); err != nil {
		return "", ErrRenderCloudConfig(err)
	}

	return body.String(), nil
}
