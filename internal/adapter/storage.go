package adapter

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type TensorRigStorageBucket struct {
	Resource pulumi.Resource
	ID       pulumi.IDOutput
	Name     pulumi.StringOutput
	URI      pulumi.StringOutput
	Region   pulumi.StringOutput
}

type TensorRigStorage interface {
	CreateBucket(ctx *pulumi.Context,
		name string,
		opts ...pulumi.ResourceOption,
	) (*TensorRigStorageBucket, error)
}
