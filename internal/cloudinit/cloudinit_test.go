package cloudinit_test

import (
	"os"
	"path/filepath"

	"github.com/jasonkwh/tensor-rig/internal/cloudinit"
)

func (s *cloudInitSuite) TestCloudInitEmptyAuthKey() {
	path := s.writeHeadlessEnv("ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\n")
	_, err := cloudinit.New(cloudinit.WithFilePath(path))
	s.ErrorIs(err, cloudinit.ErrNoAuthKey)
}

func (s *cloudInitSuite) TestCloudInitDefaultFilePathWhenSecretsMissing() {
	s.T().Setenv("HOME", s.T().TempDir())

	ts, err := cloudinit.New()
	s.Require().Error(err)
	s.Equal(cloudinit.DefaultFilePath, ts.FilePath())
}

func (s *cloudInitSuite) TestCloudInitRendersUserData() {
	got := s.renderCloudInitConfig("tskey-auth-example", "jasonkwh")

	for _, want := range []string{
		"#cloud-config",
		"https://tailscale.com/install.sh",
		"--auth-key='tskey-auth-example'",
		"--ssh",
		"tensor-rig-tailscale.service",
		"tensor-rig-node-exporter.service",
		"--web.listen-address=:9100",
		"systemctl",
	} {
		s.Contains(got, want)
	}
}

func (s *cloudInitSuite) TestCloudInitQuotesAuthKey() {
	got := s.renderCloudInitConfig("key'injected", "")
	s.Contains(got, `'key'\''injected'`)
}

func (s *cloudInitSuite) TestCloudInitLoadAuthKeyExpandsHomePath() {
	home := s.T().TempDir()
	s.T().Setenv("HOME", home)

	secretsDir := filepath.Join(home, ".secrets")
	s.Require().NoError(os.Mkdir(secretsDir, 0o700))

	path := filepath.Join(secretsDir, "headless-env")
	err := os.WriteFile(path, []byte("ts_auth_key_rig=tskey-auth-home\n"), 0o600)
	s.Require().NoError(err)

	ts, err := cloudinit.New()
	s.Require().NoError(err)
	s.Equal(cloudinit.DefaultFilePath, ts.FilePath())

	got, err := ts.Execute()
	s.Require().NoError(err)
	s.Contains(got, "tskey-auth-home")
}

func (s *cloudInitSuite) TestCloudInitSkipsLoginUserWhenDefaultUser() {
	got := s.renderCloudInitConfig("tskey-auth-example", cloudinit.DefaultLoginUser)

	s.NotContains(got, "users:")
	s.NotContains(got, "name: "+cloudinit.DefaultLoginUser)
	s.Contains(got, "https://tailscale.com/install.sh")
}

func (s *cloudInitSuite) TestCloudInitCreatesLoginUser() {
	got := s.renderCloudInitConfig("tskey-auth-example", "jasonkwh")

	for _, want := range []string{
		"users:",
		"name: jasonkwh",
		"groups: [sudo, adm]",
		"sudo: ALL=(ALL) NOPASSWD:ALL",
	} {
		s.Contains(got, want)
	}
}
