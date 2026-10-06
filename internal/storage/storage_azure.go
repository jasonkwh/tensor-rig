package storage

import (
	"strings"

	"github.com/jasonkwh/tensor-rig/internal/adapter"
	azurecore "github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var _ adapter.TensorRigStorage = &storageAzure{}

type storageAzure struct {
	core *core
}

func NewAzureStorage(
	opts ...coreOption,
) adapter.TensorRigStorage {
	core := &core{
		cfg: DefaultConfig(),
	}

	for _, opt := range opts {
		opt(core)
	}

	return &storageAzure{
		core: core,
	}
}

func (b *storageAzure) Create(
	ctx *pulumi.Context,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	group, err := azurecore.NewResourceGroup(ctx, b.core.cfg.Name+"-rg", &azurecore.ResourceGroupArgs{
		Location: pulumi.String(b.core.cfg.Location),
		Tags:     pulumi.StringMap(b.core.cfg.Labels),
	}, opts...)
	if err != nil {
		return nil, err
	}

	accountArgs := &storage.AccountArgs{
		Name:                   pulumi.String(azureAccountName(b.core.cfg.Name)),
		ResourceGroupName:      group.Name,
		Location:               group.Location,
		AccountTier:            pulumi.String("Standard"),
		AccountReplicationType: pulumi.String("LRS"),
		AccountKind:            pulumi.String("StorageV2"),
		AccessTier:             pulumi.String(azureAccessTier(b.core.cfg.StorageClass)),
		Tags:                   pulumi.StringMap(b.core.cfg.Labels),
	}

	if allow, ok := azureAllowPublicBlobs(b.core.cfg); ok {
		accountArgs.AllowNestedItemsToBePublic = pulumi.Bool(allow)
	}

	if props := azureBlobProperties(b.core.cfg); props != nil {
		accountArgs.BlobProperties = props
	}

	account, err := storage.NewAccount(ctx, b.core.cfg.Name, accountArgs, opts...)
	if err != nil {
		return nil, err
	}

	_, err = storage.NewContainer(ctx, b.core.cfg.Name+"-container", &storage.ContainerArgs{
		Name:                pulumi.String(b.core.cfg.Name),
		StorageAccountId:    account.ID(),
		ContainerAccessType: pulumi.String("private"),
	}, opts...)
	if err != nil {
		return nil, err
	}

	if rules := azureLifecycleRules(b.core.cfg); len(rules) > 0 {
		_, err = storage.NewManagementPolicy(ctx, b.core.cfg.Name+"-lifecycle", &storage.ManagementPolicyArgs{
			StorageAccountId: account.ID(),
			Rules:            rules,
		}, opts...)
		if err != nil {
			return nil, err
		}
	}

	return account, nil
}

func azureAccountName(name string) string {
	var b strings.Builder

	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}

	accountName := b.String()

	if len(accountName) > 24 {
		accountName = accountName[:24]
	}

	return accountName
}

func azureAccessTier(class StorageClass) string {
	switch class {
	case StorageClassInfrequent:
		return "Cool"
	case StorageClassCold:
		return "Cold"
	default:
		return "Hot"
	}
}

func azureAllowPublicBlobs(cfg *Config) (allow bool, set bool) {
	switch cfg.PublicAccessPrevention {
	case PublicAccessPreventionEnforced:
		return false, true
	case PublicAccessPreventionOff:
		return true, true
	default:
		if cfg.UniformBucketLevelAccess {
			return false, true
		}

		return false, false
	}
}

func azureBlobProperties(cfg *Config) *storage.AccountBlobPropertiesArgs {
	props := &storage.AccountBlobPropertiesArgs{}
	set := false

	if cfg.VersioningEnabled {
		props.VersioningEnabled = pulumi.Bool(true)
		set = true
	}

	if cfg.SoftDeleteEnabled {
		props.DeleteRetentionPolicy = &storage.AccountBlobPropertiesDeleteRetentionPolicyArgs{
			Days: pulumi.Int(7),
		}
		set = true
	}

	if !set {
		return nil
	}

	return props
}

func azureLifecycleRules(cfg *Config) storage.ManagementPolicyRuleArray {
	rules := make(storage.ManagementPolicyRuleArray, 0, len(cfg.LifecycleRules)+1)

	for i, rule := range cfg.LifecycleRules {
		if rule.DeleteAfterDays <= 0 {
			continue
		}
		rules = append(rules, &storage.ManagementPolicyRuleArgs{
			Name:    pulumi.Sprintf("delete-%d", i),
			Enabled: pulumi.Bool(true),
			Filters: &storage.ManagementPolicyRuleFiltersArgs{
				BlobTypes: pulumi.StringArray{pulumi.String("blockBlob")},
			},
			Actions: &storage.ManagementPolicyRuleActionsArgs{
				BaseBlob: &storage.ManagementPolicyRuleActionsBaseBlobArgs{
					DeleteAfterDaysSinceModificationGreaterThan: pulumi.Int(rule.DeleteAfterDays),
				},
			},
		})
	}

	if cfg.StorageClass == StorageClassArchive {
		rules = append(rules, &storage.ManagementPolicyRuleArgs{
			Name:    pulumi.String("storage-class"),
			Enabled: pulumi.Bool(true),
			Filters: &storage.ManagementPolicyRuleFiltersArgs{
				BlobTypes: pulumi.StringArray{pulumi.String("blockBlob")},
			},
			Actions: &storage.ManagementPolicyRuleActionsArgs{
				BaseBlob: &storage.ManagementPolicyRuleActionsBaseBlobArgs{
					TierToArchiveAfterDaysSinceModificationGreaterThan: pulumi.Int(0),
				},
			},
		})
	}

	return rules
}
