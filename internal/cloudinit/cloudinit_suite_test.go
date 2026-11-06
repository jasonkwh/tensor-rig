package cloudinit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/cloudinit"
	"github.com/jasonkwh/tensor-rig/internal/rig"
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
	ts, err := cloudinit.New(cloudinit.WithAuthKeyPath(path), cloudinit.WithLoginUser(loginUser))
	s.Require().NoError(err)

	got, err := ts.Execute()
	s.Require().NoError(err)

	return got
}

func (s *cloudInitSuite) renderLadderConfig(authKey, loginUser, uuid, privateKey string) string {
	s.T().Helper()

	authPath := s.writeHeadlessEnv("ts_auth_key_rig=" + authKey + "\n")
	uuidPath := s.writeSecret("uuid", uuid)
	keyPath := s.writeSecret("private-key", privateKey)
	ts, err := cloudinit.New(
		cloudinit.WithAuthKeyPath(authPath),
		cloudinit.WithLoginUser(loginUser),
		cloudinit.WithRigKind(rig.CloudRigKindLadder),
		cloudinit.WithLadderUUIDPath(uuidPath),
		cloudinit.WithLadderPrivateKeyPath(keyPath),
	)
	s.Require().NoError(err)

	got, err := ts.Execute()
	s.Require().NoError(err)

	return got
}

func (s *cloudInitSuite) writeSecret(name, body string) string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), name)
	err := os.WriteFile(path, []byte(body), 0o600)
	s.Require().NoError(err)

	return path
}

func (s *cloudInitSuite) writeHeadlessEnv(body string) string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "headless-env")
	err := os.WriteFile(path, []byte(body), 0o600)
	s.Require().NoError(err)

	return path
}
