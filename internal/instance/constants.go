package instance

const (
	awsAssumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"Service": "ec2.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}`

	awsS3Policy = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
      "s3:GetBucketLocation"
    ],
    "Resource": "*"
  }]
}`
	// For GCP Deep Learning VM Boot Disk Image, see:
	// https://docs.cloud.google.com/deep-learning-vm/docs/images
	//nolint:lll
	gcpDefaultBootDiskImage = "projects/deeplearning-platform-release/global/images/family/pytorch-2-9-cu129-ubuntu-2404-nvidia-580"

	// AMI ID from the SSM parameter in the instance region:
	// aws ssm get-parameter --region ap-southeast-2 --name
	// /aws/service/deeplearning/ami/x86_64/oss-nvidia-driver-gpu-pytorch-2.14-ubuntu-26.04/latest/ami-id
	// --query Parameter.Value --output text.
	//nolint:lll
	awsDefaultBootDiskImage = "/aws/service/deeplearning/ami/x86_64/oss-nvidia-driver-gpu-pytorch-2.14-ubuntu-26.04/latest/ami-id"

	defaultBootSizeGB = 100
)
