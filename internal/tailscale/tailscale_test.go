package tailscale_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/tailscale"
)

func TestCloudConfigEmptyKey(t *testing.T) {
	t.Parallel()

	_, err := tailscale.New("", "box")
	if !errors.Is(err, tailscale.ErrNoAuthKey) {
		t.Fatalf("expected ErrNoAuthKey, got %v", err)
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

	ts, err := tailscale.New(authKey, hostname)
	if err != nil {
		t.Fatalf("new tailscale cloud config: %v", err)
	}

	got, err := ts.Execute()
	if err != nil {
		t.Fatalf("render cloud-config: %v", err)
	}

	return got
}
