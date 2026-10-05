package storage

import (
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TensorRigStorage interface {
	Create(
		ctx *pulumi.Context,
		opts ...pulumi.ResourceOption,
	) (*storage.Bucket, error)
}
