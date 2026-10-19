package instance

import (
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (i *instanceAWS) awsRestrictPublicIngress(
	ctx *pulumi.Context,
	cfg *Config,
	vpcID string,
	opts ...pulumi.ResourceOption,
) (*ec2.SecurityGroup, error) {
	ingress := ec2.SecurityGroupIngressArray{
		&ec2.SecurityGroupIngressArgs{
			Protocol:   pulumi.String("udp"),
			FromPort:   pulumi.Int(tailscaleUDPPort),
			ToPort:     pulumi.Int(tailscaleUDPPort),
			CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			Ipv6CidrBlocks: pulumi.StringArray{
				pulumi.String("::/0"),
			},
		},
	}

	return ec2.NewSecurityGroup(ctx, cfg.Name+"-sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.String(cfg.Name + "-sg"),
		Description: pulumi.String("tensor-rig: block public TCP; allow Tailscale WireGuard UDP"),
		VpcId:       pulumi.String(vpcID),
		Region:      pulumi.String(cfg.Location),
		Ingress:     ingress,
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Ipv6CidrBlocks: pulumi.StringArray{
					pulumi.String("::/0"),
				},
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(cfg.Name + "-sg"),
		},
	}, opts...)
}

func awsVpcID(ctx *pulumi.Context, cfg *Config) (string, error) {
	if cfg.Network != "" && !strings.EqualFold(cfg.Network, "default") {
		subnet, err := ec2.LookupSubnet(ctx, &ec2.LookupSubnetArgs{
			Id:     &cfg.Network,
			Region: &cfg.Location,
		})
		if err != nil {
			return "", err
		}

		return subnet.VpcId, nil
	}

	vpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{
		Default: pulumi.BoolRef(true),
		Region:  &cfg.Location,
	})
	if err != nil {
		return "", err
	}

	return vpc.Id, nil
}
