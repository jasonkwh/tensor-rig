<h1 align="center">tensor-rig</h1>

<p align="center">
  a personal cloud rig for AI training.
</p>

<p align="center">
  <a href="https://github.com/jasonkwh/tensor-rig/actions/workflows/test.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/tensor-rig/test.yml?label=test" alt="test"></a>
  <a href="https://github.com/jasonkwh/tensor-rig/actions/workflows/lint.yml"><img src="https://img.shields.io/github/actions/workflow/status/jasonkwh/tensor-rig/lint.yml?label=lint" alt="lint"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</p>

---

tensor-rig is a [Pulumi](https://www.pulumi.com/) program that stands up a personal GPU machine for training models. One command creates a private object-storage bucket and a spot virtual machine on Google Cloud or AWS, boots a PyTorch image with NVIDIA drivers, and joins the machine to your Tailscale network so you can SSH in from your tailnet.

The machine is cheap and replaceable. Put datasets and checkpoints in the bucket. The boot disk is removed when the spot instance is reclaimed or the stack is destroyed, and bucket objects expire after seven days unless you change that.

## What it creates

**Bucket.** A GCS or S3 bucket in the configured region. Public access is blocked. Objects are deleted after `storageDeleteAfterDays` (seven by default).

**GPU instance.** A spot virtual machine. The committed GCP stack is an `n1-standard-4` with one NVIDIA T4. The AWS stack is a `g4dn.xlarge`. Both boot a vendor PyTorch image.

**Tailscale SSH.** Cloud-init installs Tailscale and runs `tailscale up --ssh`, using the instance name as the hostname.

**Bucket credentials on the instance.** On GCP the default Compute Engine service account receives `roles/storage.objectAdmin` on the bucket. On AWS the instance profile can read and write S3.

## Prerequisites

- [Go](https://go.dev/dl/) 1.27
- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Tailscale](https://tailscale.com/download) on the machine you connect from
- Cloud credentials, and GPU quota in the target region:
  - GCP: `gcloud auth application-default login`
  - AWS: credentials that can create EC2, IAM, and S3

## Tailscale auth key

Create a reusable auth key in the [Tailscale admin console](https://login.tailscale.com/admin/settings/keys). A reusable key lets a replacement instance join after spot preemption. An ephemeral key removes the node once the machine goes offline. Leave **tags** empty on the key so the node is owned by your user and matches default SSH policies (`autogroup:self`).

Save it in `~/.secrets/headless-env`:

```
ts_auth_key_rig=tskey-auth-...
```

Deploy reads `ts_auth_key_rig` from that file and renders it into the instance cloud-config. Leave the key out of `Pulumi.*.yaml`.

## Deploy

The repo includes two stacks:

| Stack | Region                 | Machine         | GPU          | Boot image                                        |
| ----- | ---------------------- | --------------- | ------------ | ------------------------------------------------- |
| `gcp` | `australia-southeast1` | `n1-standard-4` | 1× NVIDIA T4 | GCP PyTorch 2.9, CUDA 12.9, Ubuntu 24.04          |
| `aws` | `ap-southeast-2`       | `g4dn.xlarge`   | 1× NVIDIA T4 | AWS Deep Learning AMI, PyTorch 2.14, Ubuntu 26.04 |

Log in, select a stack, and apply it:

```bash
pulumi login
pulumi stack select gcp   # or: pulumi stack select aws
pulumi up
```

The first time a stack is missing from your Pulumi backend, create it. The matching `Pulumi.<stack>.yaml` already holds the config:

```bash
pulumi stack init gcp
```

An empty `cloud` value deploys to GCP.

## Connect

After boot, the instance joins your tailnet under `instanceName`. Both committed stacks use `tensor-rig-instance-au`.

First boot adds your **`instanceLoginUser`** login account (default **`jasonkwh`** in `Pulumi.yaml`) with passwordless sudo—the PyTorch images still ship their own OS user, but you SSH as this one. `tailscale ssh` defaults to your **local** username, so keep `instanceLoginUser` in sync with your workstation if you want to connect without a user prefix:

```bash
tailscale ssh tensor-rig-instance-au
```

Copy data to the bucket from your machine or from the instance:

```bash
gcloud storage cp dataset.tar.gz gs://tensor-rig-storage-au/
aws s3 cp dataset.tar.gz s3://tensor-rig-storage-au/ --region ap-southeast-2
```

Spot capacity can disappear. Run `pulumi up` again to request a new machine. Move anything you need to keep into the bucket first: the boot disk is deleted with the instance, and objects older than `storageDeleteAfterDays` are removed.

## Public IP hardening

When `instanceRestrictPublicIngress` is `true` (the default), the stack limits what the public IP accepts. Connect with `tailscale ssh`, not SSH to the cloud address.

**GCP** adds firewall rules on the instance network tag: allow UDP 41641 (Tailscale), deny all other TCP and UDP from the internet. That overrides default rules such as `default-allow-ssh`.

**AWS** attaches a security group with no public TCP ingress and UDP 41641 open for Tailscale direct paths.

Set `instanceRestrictPublicIngress` to `false` if you need inbound TCP on the public IP (for example a temporary Jupyter listener).

## Configuration

Keys and defaults are defined in `Pulumi.yaml`. Stack values live in `Pulumi.gcp.yaml` and `Pulumi.aws.yaml`. Edit the stack file or set a key, then run `pulumi up`:

```bash
pulumi config set instanceMachineType n1-standard-8
pulumi config set storageDeleteAfterDays 30
```

## Tear down

```bash
pulumi destroy
```

This removes the instance. With `storageForceDestroy` left at `false`, a bucket that still contains objects stays in place.

## Development

```bash
make test-unit
make lint
```

`make lint` expects [golangci-lint](https://golangci-lint.run/).
