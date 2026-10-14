package tailscale_test

import (
	"strings"
	"testing"

	"github.com/jasonkwh/tensor-rig/internal/tailscale"
)

func TestCloudConfigEmptyKey(t *testing.T) {
	t.Parallel()

	got := renderTailscaleCloudConfig(t, "", "box")
	if got != "" {
		t.Fatalf("expected empty cloud-config, got %q", got)
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

	got, err := tailscale.CloudConfig(authKey, hostname)
	if err != nil {
		t.Fatalf("render cloud-config: %v", err)
	}

	return got
}
