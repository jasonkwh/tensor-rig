package cloudinit

import (
	"encoding/json"
	"text/template"
)

const ladderRunCmd = `  - [modprobe, tcp_bbr]
  - [sysctl, --system]
  - [systemctl, enable, --now, sing-box.service]`

const ladderFilesSrc = `  - path: /etc/modules-load.d/tensor-rig-bbr.conf
    owner: root:root
    permissions: "0644"
    content: |
      tcp_bbr
  - path: /etc/sysctl.d/99-tensor-rig-bbr.conf
    owner: root:root
    permissions: "0644"
    content: |
      net.core.default_qdisc = fq
      net.ipv4.tcp_congestion_control = bbr
  - path: /usr/local/sbin/tensor-rig-sing-box
    owner: root:root
    permissions: "0700"
    content: |
      #!/bin/bash
      set -euo pipefail
      if [[ -x /usr/local/bin/sing-box ]]; then
        exit 0
      fi
      version={{ .SingBoxVersion }}
      arch=amd64
      tmpdir=$(mktemp -d)
      trap 'rm -rf "${tmpdir}"' EXIT
      base="https://github.com/SagerNet/sing-box/releases/download"
      url="${base}/v${version}/sing-box-${version}-linux-${arch}.tar.gz"
      curl -fsSL "${url}" \
        | tar -xzf - -C "${tmpdir}"
      install -o root -g root -m 0755 \
        "${tmpdir}/sing-box-${version}-linux-${arch}/sing-box" \
        /usr/local/bin/sing-box
  - path: /etc/sing-box/config.json
    owner: root:root
    permissions: "0600"
    content: |
{{ .SingBoxConfig }}
  - path: /etc/systemd/system/sing-box.service
    owner: root:root
    permissions: "0644"
    content: |
      [Unit]
      Description=sing-box
      After=network-online.target
      Wants=network-online.target

      [Service]
      Type=simple
      ExecStartPre=/usr/local/sbin/tensor-rig-sing-box
      ExecStart=/usr/local/bin/sing-box run -c /etc/sing-box/config.json
      Restart=on-failure

      [Install]
      WantedBy=multi-user.target
`

var ladderFilesTmpl = template.Must(template.New("ladder-files").Parse(ladderFilesSrc))

func ladderSingBoxConfig(uuid, privateKey string) (string, error) {
	cfg := map[string]any{
		"log": map[string]any{
			"level":     "warn",
			"timestamp": true,
		},
		"experimental": map[string]any{
			"clash_api": map[string]any{
				"external_controller": "127.0.0.1:9091",
			},
		},
		"inbounds":  []any{ladderInbound(uuid, privateKey)},
		"outbounds": []any{ladderDirectOutbound()},
	}

	encoded, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func ladderInbound(uuid, privateKey string) map[string]any {
	return map[string]any{
		"type":        "vless",
		"tag":         "vless-reality",
		"listen":      "::",
		"listen_port": ladderListenPort,
		"users": []any{
			map[string]any{
				"uuid": uuid,
				"flow": "xtls-rprx-vision",
			},
		},
		"tls": map[string]any{
			"enabled":     true,
			"server_name": ladderServerName,
			"reality": map[string]any{
				"enabled": true,
				"handshake": map[string]any{
					"server":      ladderServerName,
					"server_port": ladderListenPort,
				},
				"private_key": privateKey,
				"short_id":    []string{ladderShortID},
			},
		},
	}
}

func ladderDirectOutbound() map[string]any {
	return map[string]any{
		"type": "direct",
		"tag":  "direct",
	}
}
