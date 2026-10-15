package tailscale_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/tailscale"
	"github.com/stretchr/testify/suite"
)

type tsTestSuite struct {
	suite.Suite
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(tsTestSuite))
}

func (s *tsTestSuite) renderTailscaleCloudConfig(authKey, hostname string) string {
	s.T().Helper()

	path := s.writeHeadlessEnv("ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\nts_auth_key_rig=" + authKey + "\n")
	ts, err := tailscale.New(path)
	s.Require().NoError(err)

	got, err := ts.Execute(hostname)
	s.Require().NoError(err)

	return got
}

func (s *tsTestSuite) writeHeadlessEnv(body string) string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "headless-env")
	err := os.WriteFile(path, []byte(body), 0o600)
	s.Require().NoError(err)

	return path
}
