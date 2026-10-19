package instance

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var gcpPublicIngressSourceRanges = pulumi.StringArray{
	pulumi.String("0.0.0.0/0"),
	pulumi.String("::/0"),
}

type gcpIngressFirewallRule struct {
	suffix   string
	priority int
	allows   compute.FirewallAllowArray
	denies   compute.FirewallDenyArray
}

func (i *instanceGCP) gcpInstanceNetworking(
	ctx *pulumi.Context,
	cfg *Config,
	opts ...pulumi.ResourceOption,
) (*compute.InstanceNetworkInterfaceArgs, pulumi.StringArray, error) {
	tags := gcpInstanceNetworkTags(cfg)

	if cfg.RestrictPublicIngress {
		if err := i.gcpRestrictPublicIngress(ctx, cfg, tags, opts...); err != nil {
			return nil, nil, err
		}
	}

	return gcpNetworkInterface(cfg), tags, nil
}

func gcpNetworkInterface(cfg *Config) *compute.InstanceNetworkInterfaceArgs {
	nic := &compute.InstanceNetworkInterfaceArgs{
		Network: gcpNetwork(cfg),
		NicType: pulumi.String(cfg.NicType),
	}

	if cfg.AssignPublicIP {
		nic.AccessConfigs = compute.InstanceNetworkInterfaceAccessConfigArray{
			&compute.InstanceNetworkInterfaceAccessConfigArgs{},
		}
	}

	return nic
}

func gcpInstanceNetworkTags(cfg *Config) pulumi.StringArray {
	if !cfg.RestrictPublicIngress {
		return pulumi.StringArray{}
	}

	return pulumi.StringArray{pulumi.String(cfg.Name)}
}

func (i *instanceGCP) gcpRestrictPublicIngress(
	ctx *pulumi.Context,
	cfg *Config,
	targetTags pulumi.StringArray,
	opts ...pulumi.ResourceOption,
) error {
	network := gcpNetwork(cfg)
	rules := []gcpIngressFirewallRule{
		{
			suffix:   "allow-tailscale-udp",
			priority: 900, //nolint:mnd
			allows: compute.FirewallAllowArray{
				&compute.FirewallAllowArgs{
					Protocol: pulumi.String("udp"),
					Ports:    pulumi.StringArray{pulumi.String(strconv.Itoa(tailscaleUDPPort))},
				},
			},
		},
		{
			suffix:   "deny-public-tcp",
			priority: 1000, //nolint:mnd
			denies: compute.FirewallDenyArray{
				&compute.FirewallDenyArgs{Protocol: pulumi.String("tcp")},
			},
		},
		{
			suffix:   "deny-public-udp",
			priority: 1100, //nolint:mnd
			denies: compute.FirewallDenyArray{
				&compute.FirewallDenyArgs{Protocol: pulumi.String("udp")},
			},
		},
	}

	for _, rule := range rules {
		if err := gcpNewIngressFirewall(ctx,
			fmt.Sprintf("%s-%s", cfg.Name, rule.suffix),
			network, targetTags, rule, opts...); err != nil {
			return err
		}
	}

	return nil
}

func gcpNewIngressFirewall(
	ctx *pulumi.Context,
	name string,
	network pulumi.StringInput,
	targetTags pulumi.StringArray,
	rule gcpIngressFirewallRule,
	opts ...pulumi.ResourceOption,
) error {
	_, err := compute.NewFirewall(ctx, name, &compute.FirewallArgs{
		Name:         pulumi.String(name),
		Network:      network,
		Direction:    pulumi.String("INGRESS"),
		Priority:     pulumi.Int(rule.priority),
		Allows:       rule.allows,
		Denies:       rule.denies,
		SourceRanges: gcpPublicIngressSourceRanges,
		TargetTags:   targetTags,
	}, opts...)

	return err
}

func gcpNetwork(cfg *Config) pulumi.StringInput {
	if cfg.Network == "" || strings.EqualFold(cfg.Network, "default") {
		return pulumi.String("default")
	}

	return pulumi.String(cfg.Network)
}
