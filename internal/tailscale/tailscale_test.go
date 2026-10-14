package tailscale_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/tailscale"
)

func TestCloudConfigEmptyKey(t *testing.T) {
	t.Parallel()

	path := writeHeadlessEnv(t, "ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\n")
	ts, err := tailscale.New(path, "box")
	if err != nil {
		t.Fatalf("new tailscale cloud config: %v", err)
	}

	_, err = ts.Execute()
	if !errors.Is(err, tailscale.ErrNoAuthKey) {
		t.Fatalf("expected ErrNoAuthKey, got %v", err)
	}
}

func TestCloudConfigEmptyFilePath(t *testing.T) {
	t.Parallel()

	_, err := tailscale.New("", "box")
	if !errors.Is(err, tailscale.ErrNoFilePath) {
		t.Fatalf("expected ErrNoFilePath, got %v", err)
	}
}

func TestCloudConfigEmptyHostname(t *testing.T) {
	t.Parallel()

	_, err := tailscale.New("tskey-auth-example", "")
	if !errors.Is(err, tailscale.ErrNoHostname) {
		t.Fatalf("expected ErrNoHostname, got %v", err)
	}
}

func TestCloudConfigInstallsClient(t *testing.T) {
	t.Parallel()

	got := renderTailscaleCloudConfig(t, "tskey-auth-example", "tensor-rig-instance-au")

	for _, want := range []string{
		"#cloud-config",
		"https://tailscale.com/install.sh",
		"--auth-key='tskey-auth-example'",
		"--ssh",
		"--hostname='tensor-rig-instance-au'",
		"tensor-rig-tailscale.service",
		"systemctl",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cloud-config missing %q\n%s", want, got)
		}
	}
}

func TestTailscaleCloudConfigQuotesAuthKey(t *testing.T) {
	t.Parallel()

	got := renderTailscaleCloudConfig(t, "key'injected", "box")

	if !strings.Contains(got, `'key'\''injected'`) {
		t.Fatalf("auth key was not shell-quoted: %s", got)
	}
}

func renderTailscaleCloudConfig(t *testing.T, authKey, hostname string) string {
	t.Helper()

	path := writeHeadlessEnv(t, "ssid_home=example\npsk_home=secret\nts_auth_key=tskey-auth-fleet\nts_auth_key_rig="+authKey+"\n")
	ts, err := tailscale.New(path, hostname)
	if err != nil {
		t.Fatalf("new tailscale cloud config: %v", err)
	}

	got, err := ts.Execute()
	if err != nil {
		t.Fatalf("render cloud-config: %v", err)
	}

	return got
}

func writeHeadlessEnv(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "headless-env")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write headless-env: %v", err)
	}

	return path
}
