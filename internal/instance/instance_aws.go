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

	net, err := i.awsInstanceNetworking(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}

	userData, err := i.tsCloudCfg.Execute(cfg.Name, cfg.LoginUser)
	if err != nil {
		return nil, err
	}

	profile, err := awsInstanceProfile(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}

	instance, err := ec2.NewInstance(ctx, cfg.Name, awsInstanceArgs(
		cfg, imageID, net, userData, profile,
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
	net awsInstanceNetwork,
	userData string,
	profile *iam.InstanceProfile,
) *ec2.InstanceArgs {
	args := &ec2.InstanceArgs{
		Ami:                      pulumi.String(imageID),
		InstanceType:             pulumi.String(cfg.MachineType),
		Region:                   pulumi.String(cfg.Location),
		SubnetId:                 pulumi.String(net.subnetID),
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

	if net.securityGroup != nil {
		args.VpcSecurityGroupIds = pulumi.StringArray{net.securityGroup.ID()}
	}

	if cfg.ProvisioningModel.IsSpot() {
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
