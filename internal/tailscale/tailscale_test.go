package tailscale_test

import (
	"github.com/jasonkwh/tensor-rig/internal/tailscale"
)

func (s *tsTestSuite) TestCloudConfigEmptyKey() {
	path := s.writeHeadlessEnv("ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\n")
	_, err := tailscale.New(tailscale.WithFilePath(path))
	s.ErrorIs(err, tailscale.ErrNoAuthKey)
}

func (s *tsTestSuite) TestCloudConfigEmptyFilePath() {
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
