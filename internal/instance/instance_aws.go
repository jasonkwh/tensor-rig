package instance

import (
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ = adapter.Instance(&instanceAWS{})

type instanceAWS struct {
	tsCloudCfg adapter.TailscaleCloudConfig
	cfg        *Config
}

func newAWSInstance(tsCloudCfg adapter.TailscaleCloudConfig, cfg *Config) adapter.Instance {
	cfg.OnHostMaintenance = strings.ToLower(cfg.OnHostMaintenance)
	cfg.ProvisioningModel = strings.ToLower(cfg.ProvisioningModel)

	return &instanceAWS{
		tsCloudCfg: tsCloudCfg,
		cfg:        cfg,
	}
}

func (i *instanceAWS) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	cfg := i.cfg

	imageID, err := awsImageID(ctx, cfg)
	if err != nil {
		return nil, err
	}

	subnetID, err := awsSubnetID(ctx, cfg)
	if err != nil {
		return nil, err
	}

	userData, err := i.tsCloudCfg.Execute(cfg.Name)
	if err != nil {
		return nil, err
	}

	profile, err := awsInstanceProfile(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}

	var securityGroup *ec2.SecurityGroup

	if cfg.RestrictPublicIngress {
		vpcID, err := awsVpcID(ctx, cfg)
		if err != nil {
			return nil, err
		}

		securityGroup, err = i.awsRestrictPublicIngress(ctx, cfg, vpcID, opts...)
		if err != nil {
			return nil, err
		}
	}

	instance, err := ec2.NewInstance(ctx, cfg.Name, awsInstanceArgs(
		cfg, imageID, subnetID, userData, profile, securityGroup,
	), opts...)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func awsImageID(ctx *pulumi.Context, cfg *Config) (string, error) {
	if strings.HasPrefix(cfg.BootDiskImage, "ami-") {
		return cfg.BootDiskImage, nil
	}

	param, err := ssm.LookupParameter(ctx, &ssm.LookupParameterArgs{
		Name:   cfg.BootDiskImage,
		Region: &cfg.Location,
	})
	if err != nil {
		return "", err
	}

	return param.Value, nil
}

func awsSubnetID(ctx *pulumi.Context, cfg *Config) (string, error) {
	if cfg.Network != "" && !strings.EqualFold(cfg.Network, "default") {
		return cfg.Network, nil
	}

	vpc, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{
		Default: pulumi.BoolRef(true),
		Region:  &cfg.Location,
	})
	if err != nil {
		return "", err
	}

	zone := cfg.Location + cfg.Zone

	subnet, err := ec2.LookupSubnet(ctx, &ec2.LookupSubnetArgs{
		AvailabilityZone: &zone,
		DefaultForAz:     pulumi.BoolRef(true),
		Region:           &cfg.Location,
		VpcId:            &vpc.Id,
	})
	if err != nil {
		return "", err
	}

	return subnet.Id, nil
}

func awsInstanceProfile(
	ctx *pulumi.Context,
	cfg *Config,
	opts ...pulumi.ResourceOption,
) (*iam.InstanceProfile, error) {
	role, err := iam.NewRole(ctx, cfg.Name+"-role", &iam.RoleArgs{
		Name:             pulumi.String(cfg.Name),
		AssumeRolePolicy: pulumi.String(awsAssumeRolePolicy),
		InlinePolicies: iam.RoleInlinePolicyArray{
			&iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("s3"),
				Policy: pulumi.String(awsS3Policy),
			},
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	profile, err := iam.NewInstanceProfile(ctx, cfg.Name+"-profile", &iam.InstanceProfileArgs{
		Name: pulumi.String(cfg.Name),
		Role: role.Name,
	}, opts...)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func awsInstanceArgs(
	cfg *Config,
	imageID string,
	subnetID string,
	userData string,
	profile *iam.InstanceProfile,
	securityGroup *ec2.SecurityGroup,
) *ec2.InstanceArgs {
	args := &ec2.InstanceArgs{
		Ami:                      pulumi.String(imageID),
		InstanceType:             pulumi.String(cfg.MachineType),
		Region:                   pulumi.String(cfg.Location),
		SubnetId:                 pulumi.String(subnetID),
		AssociatePublicIpAddress: pulumi.Bool(cfg.AssignPublicIP),
		IamInstanceProfile:       profile.Name,
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize:          pulumi.Int(cfg.BootDiskSizeGb),
			VolumeType:          pulumi.String(cfg.BootDiskType),
			DeleteOnTermination: pulumi.Bool(true),
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(cfg.Name),
		},
	}

	if userData != "" {
		args.UserData = pulumi.String(userData)
	}

	if securityGroup != nil {
		args.VpcSecurityGroupIds = pulumi.StringArray{securityGroup.ID()}
	}

	if strings.EqualFold(cfg.ProvisioningModel, "spot") {
		args.InstanceMarketOptions = &ec2.InstanceInstanceMarketOptionsArgs{
			MarketType: pulumi.String("spot"),
			SpotOptions: &ec2.InstanceInstanceMarketOptionsSpotOptionsArgs{
				InstanceInterruptionBehavior: pulumi.String(cfg.OnHostMaintenance),
				SpotInstanceType:             pulumi.String("one-time"),
			},
		}
	}

	return args
}
