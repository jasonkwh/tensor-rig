package adapter

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type Instance interface {
	Create(
		ctx *pulumi.Context,
		opts ...pulumi.ResourceOption,
	) (pulumi.Resource, error)
}
