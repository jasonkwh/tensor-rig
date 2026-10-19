package tailscale

import "text/template"

var cloudConfigTmpl = template.Must(template.New("tailscale").Parse(cloudConfigSrc))

const cloudConfigSrc = `#cloud-config
{{if .CreateLoginUser}}
users:
  - default
  - name: {{ .LoginUser }}
    groups: [sudo, adm]
    shell: /bin/bash
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock_passwd: true
{{end}}
write_files:
  - path: /usr/local/sbin/tensor-rig-tailscale
    owner: root:root
    permissions: "0700"
    content: |
      #!/bin/bash
      set -euo pipefail
      if ! command -v tailscale >/dev/null 2>&1; then
        curl -fsSL https://tailscale.com/install.sh | sh
      fi
      for attempt in 1 2 3 4 5; do
        if tailscale up --auth-key={{ .AuthKey }} --ssh --hostname={{ .Hostname }} --timeout=30s; then
          exit 0
        fi
        echo "tailscale up failed (attempt ${attempt}), retrying" >&2
        sleep 10
      done
      echo "tailscale up failed" >&2
      exit 1
  - path: /etc/systemd/system/tensor-rig-tailscale.service
    owner: root:root
    permissions: "0644"
    content: |
      [Unit]
      Description=Join the tensor-rig tailnet
      After=network-online.target
      Wants=network-online.target

      [Service]
      Type=oneshot
      ExecStart=/usr/local/sbin/tensor-rig-tailscale
      RemainAfterExit=yes

      [Install]
      WantedBy=multi-user.target
runcmd:
  - [systemctl, daemon-reload]
  - [systemctl, enable, --now, tensor-rig-tailscale.service]
`
