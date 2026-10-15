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
	gcpDefaultBootDiskImage = "projects/deeplearning-platform-release/global/images/family/pytorch-2-9-cu129-ubuntu-2204-nvidia-580"

	// For AWS Deep Learning AMI, see:
	// https://docs.aws.amazon.com/dlami/latest/devguide/find-dlami-id.html.
	awsDefaultBootDiskImage = "/aws/service/deeplearning/ami/x86_64/oss-nvidia-driver-gpu-pytorch-2.7-ubuntu-22.04/latest/ami-id"

	defaultBootSizeGB = 100
)
