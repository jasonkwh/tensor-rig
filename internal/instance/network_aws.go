package instance

import (
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	awsPublicCIDRBlocks = pulumi.StringArray{
		pulumi.String("0.0.0.0/0"),
	}
	awsPublicIPv6CIDRBlocks = pulumi.StringArray{
		pulumi.String("::/0"),
	}
)

type awsInstanceNetwork struct {
	subnetID      string
	securityGroup *ec2.SecurityGroup
}

func (i *instanceAWS) awsInstanceNetworking(
	ctx *pulumi.Context,
	cfg *Config,
	opts ...pulumi.ResourceOption,
) (awsInstanceNetwork, error) {
	vpcID, subnetID, err := awsNetworkPlacement(ctx, cfg)
	if err != nil {
		return awsInstanceNetwork{}, err
	}

	out := awsInstanceNetwork{subnetID: subnetID}
	if !cfg.RestrictPublicIngress && !cfg.OpenTCP443 {
		return out, nil
	}

	out.securityGroup, err = i.awsRestrictPublicIngress(ctx, cfg, vpcID, opts...)

	return out, err
}

func awsNetworkPlacement(
	ctx *pulumi.Context,
	cfg *Config,
) (string, string, error) {
	if cfg.Network != "" && !strings.EqualFold(cfg.Network, "default") {
		subnet, lookupErr := ec2.LookupSubnet(ctx, &ec2.LookupSubnetArgs{
			Id:     &cfg.Network,
			Region: &cfg.Location,
		})
		if lookupErr != nil {
			return "", "", lookupErr
		}

		return subnet.VpcId, subnet.Id, nil
	}

	vpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{
		Default: pulumi.BoolRef(true),
		Region:  &cfg.Location,
	})
	if err != nil {
		return "", "", err
	}

	zone := cfg.Location + cfg.Zone

	subnet, err := ec2.LookupSubnet(ctx, &ec2.LookupSubnetArgs{
		AvailabilityZone: &zone,
		DefaultForAz:     pulumi.BoolRef(true),
		Region:           &cfg.Location,
		VpcId:            &vpc.Id,
	})
	if err != nil {
		return "", "", err
	}

	return vpc.Id, subnet.Id, nil
}

func (i *instanceAWS) awsRestrictPublicIngress(
	ctx *pulumi.Context,
	cfg *Config,
	vpcID string,
	opts ...pulumi.ResourceOption,
) (*ec2.SecurityGroup, error) {
	name := cfg.Name + "-sg"

	return ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		Name:        pulumi.String(name),
		Description: pulumi.String(awsSecurityGroupDescription(cfg)),
		VpcId:       pulumi.String(vpcID),
		Region:      pulumi.String(cfg.Location),
		Ingress:     awsRestrictedIngress(cfg.OpenTCP443),
		Egress:      awsAllEgress(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(name),
		},
	}, opts...)
}

func awsSecurityGroupDescription(cfg *Config) string {
	if cfg.OpenTCP443 && !cfg.RestrictPublicIngress {
		return "tensor-rig: allow public TCP 443 and Tailscale WireGuard UDP"
	}

	if cfg.OpenTCP443 {
		return "tensor-rig: block public TCP except 443; allow Tailscale WireGuard UDP"
	}

	return "tensor-rig: block public TCP; allow Tailscale WireGuard UDP"
}

func awsRestrictedIngress(openTCP443 bool) ec2.SecurityGroupIngressArray {
	rules := ec2.SecurityGroupIngressArray{
		&ec2.SecurityGroupIngressArgs{
			Protocol:       pulumi.String("udp"),
			FromPort:       pulumi.Int(tailscaleUDPPort),
			ToPort:         pulumi.Int(tailscaleUDPPort),
			CidrBlocks:     awsPublicCIDRBlocks,
			Ipv6CidrBlocks: awsPublicIPv6CIDRBlocks,
		},
	}
	if !openTCP443 {
		return rules
	}

	return append(rules, &ec2.SecurityGroupIngressArgs{
		Protocol:       pulumi.String("tcp"),
		FromPort:       pulumi.Int(publicTCPPort),
		ToPort:         pulumi.Int(publicTCPPort),
		CidrBlocks:     awsPublicCIDRBlocks,
		Ipv6CidrBlocks: awsPublicIPv6CIDRBlocks,
	})
}

func awsAllEgress() ec2.SecurityGroupEgressArray {
	return ec2.SecurityGroupEgressArray{
		&ec2.SecurityGroupEgressArgs{
			Protocol:       pulumi.String("-1"),
			FromPort:       pulumi.Int(0),
			ToPort:         pulumi.Int(0),
			CidrBlocks:     awsPublicCIDRBlocks,
			Ipv6CidrBlocks: awsPublicIPv6CIDRBlocks,
		},
	}
}
