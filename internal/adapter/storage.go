package adapter

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TensorRigStorage interface {
	Create(
		ctx *pulumi.Context,
		opts ...pulumi.ResourceOption,
	) (pulumi.Resource, error)
}
