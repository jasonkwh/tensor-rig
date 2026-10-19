package tailscale_test

import (
	"os"
	"path/filepath"

	"github.com/jasonkwh/tensor-rig/internal/tailscale"
)

func (s *tsTestSuite) TestCloudConfigEmptyKey() {
	path := s.writeHeadlessEnv("ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\n")
	_, err := tailscale.New(tailscale.WithFilePath(path))
	s.ErrorIs(err, tailscale.ErrNoAuthKey)
}

func (s *tsTestSuite) TestCloudConfigEmptyFilePath() {
	s.T().Setenv("HOME", s.T().TempDir())

	ts, err := tailscale.New()
	s.Require().Error(err)
	s.Equal(tailscale.DefaultFilePath, ts.FilePath())
}

func (s *tsTestSuite) TestCloudConfigInstallsClient() {
	got := s.renderTailscaleCloudConfig("tskey-auth-example", "tensor-rig-instance-au")

	for _, want := range []string{
		"#cloud-config",
		"https://tailscale.com/install.sh",
		"--auth-key='tskey-auth-example'",
		"--ssh",
		"--hostname='tensor-rig-instance-au'",
		"tensor-rig-tailscale.service",
		"systemctl",
	} {
		s.Contains(got, want)
	}
}

func (s *tsTestSuite) TestTailscaleCloudConfigQuotesAuthKey() {
	got := s.renderTailscaleCloudConfig("key'injected", "box")
	s.Contains(got, `'key'\''injected'`)
}

func (s *tsTestSuite) TestLoadAuthKeyExpandsHomePath() {
	home := s.T().TempDir()
	s.T().Setenv("HOME", home)

	secretsDir := filepath.Join(home, ".secrets")
	s.Require().NoError(os.Mkdir(secretsDir, 0o700))

	path := filepath.Join(secretsDir, "headless-env")
	err := os.WriteFile(path, []byte("ts_auth_key_rig=tskey-auth-home\n"), 0o600)
	s.Require().NoError(err)

	ts, err := tailscale.New()
	s.Require().NoError(err)
	s.Equal(tailscale.DefaultFilePath, ts.FilePath())

	got, err := ts.Execute("rig")
	s.Require().NoError(err)
	s.Contains(got, "tskey-auth-home")
}
