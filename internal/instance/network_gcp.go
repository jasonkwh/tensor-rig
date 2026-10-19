package instance

import (
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (i *instanceGCP) gcpRestrictPublicIngress(
	ctx *pulumi.Context,
	cfg *Config,
	opts ...pulumi.ResourceOption,
) error {
	tag := gcpNetworkTag(cfg)

	_, err := compute.NewFirewall(ctx, cfg.Name+"-allow-tailscale-udp", &compute.FirewallArgs{
		Name:      pulumi.String(cfg.Name + "-allow-tailscale-udp"),
		Network:   gcpNetwork(cfg),
		Direction: pulumi.String("INGRESS"),
		Priority:  pulumi.Int(900),
		Allows: compute.FirewallAllowArray{
			&compute.FirewallAllowArgs{
				Protocol: pulumi.String("udp"),
				Ports:    pulumi.StringArray{pulumi.String(strconv.Itoa(tailscaleUDPPort))},
			},
		},
		SourceRanges: pulumi.StringArray{
			pulumi.String("0.0.0.0/0"),
			pulumi.String("::/0"),
		},
		TargetTags: pulumi.StringArray{pulumi.String(tag)},
	}, opts...)
	if err != nil {
		return err
	}

	_, err = compute.NewFirewall(ctx, cfg.Name+"-deny-public-tcp", &compute.FirewallArgs{
		Name:      pulumi.String(cfg.Name + "-deny-public-tcp"),
		Network:   gcpNetwork(cfg),
		Direction: pulumi.String("INGRESS"),
		Priority:  pulumi.Int(1000),
		Denies: compute.FirewallDenyArray{
			&compute.FirewallDenyArgs{
				Protocol: pulumi.String("tcp"),
			},
		},
		SourceRanges: pulumi.StringArray{
			pulumi.String("0.0.0.0/0"),
			pulumi.String("::/0"),
		},
		TargetTags: pulumi.StringArray{pulumi.String(tag)},
	}, opts...)
	if err != nil {
		return err
	}

	_, err = compute.NewFirewall(ctx, cfg.Name+"-deny-public-udp", &compute.FirewallArgs{
		Name:      pulumi.String(cfg.Name + "-deny-public-udp"),
		Network:   gcpNetwork(cfg),
		Direction: pulumi.String("INGRESS"),
		Priority:  pulumi.Int(1100),
		Denies: compute.FirewallDenyArray{
			&compute.FirewallDenyArgs{
				Protocol: pulumi.String("udp"),
			},
		},
		SourceRanges: pulumi.StringArray{
			pulumi.String("0.0.0.0/0"),
			pulumi.String("::/0"),
		},
		TargetTags: pulumi.StringArray{pulumi.String(tag)},
	}, opts...)

	return err
}

func gcpNetworkTag(cfg *Config) string {
	return cfg.Name
}

func gcpNetwork(cfg *Config) pulumi.StringInput {
	if cfg.Network == "" || strings.EqualFold(cfg.Network, "default") {
		return pulumi.String("default")
	}

	return pulumi.String(cfg.Network)
}
