package cloudinit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/cloudinit"
	"github.com/stretchr/testify/suite"
)

type cloudInitSuite struct {
	suite.Suite
}

func TestCloudInitSuite(t *testing.T) {
	suite.Run(t, new(cloudInitSuite))
}

func (s *cloudInitSuite) renderCloudInitConfig(authKey, loginUser string) string {
	s.T().Helper()

	path := s.writeHeadlessEnv("ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\nts_auth_key_rig=" + authKey + "\n")
	ts, err := cloudinit.New(cloudinit.WithFilePath(path), cloudinit.WithLoginUser(loginUser))
	s.Require().NoError(err)

	got, err := ts.Execute()
	s.Require().NoError(err)

	return got
}

func (s *cloudInitSuite) writeHeadlessEnv(body string) string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "headless-env")
	err := os.WriteFile(path, []byte(body), 0o600)
	s.Require().NoError(err)

	return path
}
