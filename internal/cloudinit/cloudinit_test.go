package cloudinit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/cloudinit"
	"github.com/jasonkwh/tensor-rig/internal/rig"
)

func (s *cloudInitSuite) TestCloudInitEmptyAuthKey() {
	path := s.writeHeadlessEnv("ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\n")
	_, err := cloudinit.New(cloudinit.WithAuthKeyPath(path))
	s.ErrorIs(err, cloudinit.ErrNoAuthKey)
}

func (s *cloudInitSuite) TestCloudInitDefaultFilePathWhenSecretsMissing() {
	s.T().Setenv("HOME", s.T().TempDir())

	ts, err := cloudinit.New()
	s.Require().Error(err)
	s.Equal(cloudinit.DefaultAuthKeyPath, ts.FilePath())
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

	s.NotContains(got, "sing-box.service")
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
	s.Equal(cloudinit.DefaultAuthKeyPath, ts.FilePath())

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

func (s *cloudInitSuite) TestCloudInitLadderRendersUserData() {
	const (
		uuid       = "11111111-1111-1111-1111-111111111111"
		privateKey = "ladder-private-key"
	)

	got := s.renderLadderConfig("tskey-auth-example", "jasonkwh", uuid+"\n", privateKey+"\n")

	for _, want := range []string{
		"#cloud-config",
		"name: jasonkwh",
		"- make",
		"- python3-venv",
		"tensor-rig-tailscale.service",
		"sing-box.service",
		"net.ipv4.tcp_congestion_control = bbr",
		"net.core.default_qdisc = fq",
		"modprobe, tcp_bbr",
		"tensor-rig-node-exporter.service",
		"--web.listen-address=:9100",
		"https://github.com/SagerNet/sing-box/releases/download",
		"--auth-key='tskey-auth-example'",
	} {
		s.Contains(got, want)
	}

	var cfg map[string]any
	s.Require().NoError(json.Unmarshal([]byte(s.singBoxConfig(got)), &cfg))

	inbounds, ok := cfg["inbounds"].([]any)
	s.Require().True(ok)
	s.Require().Len(inbounds, 1)

	inbound, ok := inbounds[0].(map[string]any)
	s.Require().True(ok)
	s.Equal("vless", inbound["type"])
	s.Equal("::", inbound["listen"])
	s.Equal(float64(443), inbound["listen_port"])

	users, ok := inbound["users"].([]any)
	s.Require().True(ok)
	s.Require().Len(users, 1)

	user, ok := users[0].(map[string]any)
	s.Require().True(ok)
	s.Equal(uuid, user["uuid"])
	s.Equal("xtls-rprx-vision", user["flow"])

	tls, ok := inbound["tls"].(map[string]any)
	s.Require().True(ok)
	s.Equal("www.cloudflare.com", tls["server_name"])

	reality, ok := tls["reality"].(map[string]any)
	s.Require().True(ok)
	s.Equal(privateKey, reality["private_key"])
	s.Equal([]any{"d8424154"}, reality["short_id"])
}

func (s *cloudInitSuite) TestCloudInitLadderQuotesSecrets() {
	got := s.renderLadderConfig("key'injected", cloudinit.DefaultLoginUser, "uuid", "key\"injected")

	s.Contains(got, `'key'\''injected'`)
	s.Contains(got, `key\"injected`)
	s.NotContains(got, "users:")
}

func (s *cloudInitSuite) TestCloudInitLadderEmptyUUID() {
	authPath := s.writeHeadlessEnv("ts_auth_key_rig=tskey-auth-example\n")
	uuidPath := s.writeSecret("uuid", " \n")
	keyPath := s.writeSecret("private-key", "ladder-private-key\n")

	_, err := cloudinit.New(
		cloudinit.WithAuthKeyPath(authPath),
		cloudinit.WithRigKind(rig.CloudRigKindLadder),
		cloudinit.WithLadderUUIDPath(uuidPath),
		cloudinit.WithLadderPrivateKeyPath(keyPath),
	)
	s.ErrorIs(err, cloudinit.ErrNoLadderUUID)
}

func (s *cloudInitSuite) TestCloudInitLadderMissingPrivateKey() {
	authPath := s.writeHeadlessEnv("ts_auth_key_rig=tskey-auth-example\n")
	uuidPath := s.writeSecret("uuid", "11111111-1111-1111-1111-111111111111\n")

	_, err := cloudinit.New(
		cloudinit.WithAuthKeyPath(authPath),
		cloudinit.WithRigKind(rig.CloudRigKindLadder),
		cloudinit.WithLadderUUIDPath(uuidPath),
		cloudinit.WithLadderPrivateKeyPath(filepath.Join(s.T().TempDir(), "missing-private-key")),
	)
	s.Error(err)
	s.ErrorContains(err, "load ladder private key")
}

func (s *cloudInitSuite) TestCloudInitLadderLoadSecretsExpandsHomePath() {
	home := s.T().TempDir()
	s.T().Setenv("HOME", home)

	secretsDir := filepath.Join(home, ".secrets")
	proxyDir := filepath.Join(secretsDir, "proxy-env")
	s.Require().NoError(os.MkdirAll(proxyDir, 0o700))
	s.Require().NoError(os.WriteFile(
		filepath.Join(secretsDir, "headless-env"),
		[]byte("ts_auth_key_rig=tskey-auth-home\n"),
		0o600,
	))
	s.Require().NoError(os.WriteFile(
		filepath.Join(proxyDir, "uuid"),
		[]byte("22222222-2222-2222-2222-222222222222"),
		0o600,
	))
	s.Require().NoError(os.WriteFile(
		filepath.Join(proxyDir, "private-key"),
		[]byte("home-private-key\n"),
		0o600,
	))

	ts, err := cloudinit.New(cloudinit.WithRigKind(rig.CloudRigKindLadder))
	s.Require().NoError(err)

	got, err := ts.Execute()
	s.Require().NoError(err)
	s.Contains(got, "tskey-auth-home")
	s.Contains(got, "22222222-2222-2222-2222-222222222222")
	s.Contains(got, "home-private-key")
	s.Contains(got, "sing-box.service")
	s.Contains(got, "- python3-venv")
}

func (s *cloudInitSuite) TestCloudInitLadderAddsOnlySingBox() {
	const (
		authKey = "tskey-auth-example"
		user    = "jasonkwh"
	)

	tensor := s.renderCloudInitConfig(authKey, user)
	ladder := s.renderLadderConfig(authKey, user, "uuid\n", "private-key\n")

	start := strings.Index(ladder, "  - path: /etc/modules-load.d/tensor-rig-bbr.conf\n")
	end := strings.Index(ladder, "runcmd:\n")
	s.Require().GreaterOrEqual(start, 0)
	s.Require().Greater(end, start)

	stripped := ladder[:start] + ladder[end:]
	for _, line := range []string{
		"  - [modprobe, tcp_bbr]\n",
		"  - [sysctl, --system]\n",
		"  - [systemctl, enable, --now, sing-box.service]\n",
	} {
		stripped = strings.Replace(stripped, line, "", 1)
	}
	s.Equal(tensor, stripped)
}

func (s *cloudInitSuite) singBoxConfig(userData string) string {
	s.T().Helper()

	const header = "path: /etc/sing-box/config.json"
	headerAt := strings.Index(userData, header)
	s.Require().GreaterOrEqual(headerAt, 0)

	rest := userData[headerAt:]
	const marker = "content: |\n"
	markerAt := strings.Index(rest, marker)
	s.Require().GreaterOrEqual(markerAt, 0)

	body := rest[markerAt+len(marker):]
	end := strings.Index(body, "\n  - path:")
	s.Require().GreaterOrEqual(end, 0)

	lines := strings.Split(body[:end], "\n")
	for i, line := range lines {
		s.Require().GreaterOrEqual(len(line), 6)
		lines[i] = line[6:]
	}

	return strings.Join(lines, "\n")
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
