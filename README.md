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

**GPU instance.** A spot virtual machine. The committed GCP stack is an `n1-standard-8` with one NVIDIA T4. The AWS stack is a `g4dn.2xlarge`. Both boot a vendor PyTorch image.

**Cloud-init.** First boot user-data joins the machine to Tailscale (`tailscale up --ssh` with hostname **`cloudInitHostname`**, default **`jasonkwh-tensor-rig`** in `Pulumi.yaml`) and starts Prometheus **node_exporter** on port **9100** (reachable on the tailnet, not on the public IP). **`instanceName`** is only the cloud resource name (EC2/GCE), not the tailnet name.

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
| `gcp` | `australia-southeast1` | `n1-standard-8` | 1× NVIDIA T4 | GCP PyTorch 2.9, CUDA 12.9, Ubuntu 24.04          |
| `aws` | `ap-southeast-2`       | `g4dn.2xlarge`  | 1× NVIDIA T4 | AWS Deep Learning AMI, PyTorch 2.14, Ubuntu 26.04 |

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

### Deploy from GitHub Actions

The `deploy-gcp` workflow applies the `gcp` stack from a runner, so nothing has to be compiled on a workstation. It authenticates with Workload Identity Federation: no service account key is stored anywhere, and credentials are only minted for jobs in the `deploy` environment.

Run `scripts/setup-gcp-wif.sh` once with credentials that can manage IAM. It creates the service account, the pool, the provider, and the impersonation binding, then prints the values the workflow needs:

```bash
scripts/setup-gcp-wif.sh            # --help for project, repo, environment
```

Before the first run:

- create the GitHub environment `deploy` with a required reviewer
- add repository secret `PULUMI_ACCESS_TOKEN`
- add repository secret `TS_AUTH_KEY_RIG`
- add repository variable `GCP_SA_EMAIL` (the script prints it)

The provider's attribute condition accepts only jobs from this repository running in the `deploy` environment, so a fork, a pull request, or a workflow edited on a branch cannot mint credentials. Pools and IAM bindings can take five minutes to propagate.

## Connect

After boot, the node appears on your tailnet as **`cloudInitHostname`**. Both committed stacks still use **`instanceName`** `tensor-rig-instance-au` in the cloud console.

First boot adds your **`cloudInitLoginUser`** login account (default **`jasonkwh`** in `Pulumi.yaml`) with passwordless sudo—the PyTorch images still ship their own OS user, but you SSH as this one. `tailscale ssh` defaults to your **local** username, so keep `cloudInitLoginUser` in sync with your workstation if you want to connect without a user prefix:

Prometheus on your tailnet can scrape `http://<cloudInitHostname>:9100/metrics` (or the node’s Tailscale IP).

```bash
tailscale ssh jasonkwh-tensor-rig
```

Copy data to the bucket from your machine or from the instance:

```bash
gcloud storage cp dataset.tar.gz gs://tensor-rig-storage-au/
aws s3 cp dataset.tar.gz s3://tensor-rig-storage-au/ --region ap-southeast-2
```

Spot capacity can disappear. Run `pulumi up` again to request a new machine. Move anything you need to keep into the bucket first: the boot disk is deleted with the instance, and objects older than `storageDeleteAfterDays` are removed.

## Training artifacts

Training output lives on the boot disk, which is deleted with the spot instance. Push it to the bucket while the job runs, not after it finishes:

```bash
scripts/sync-artifacts.sh gs://tensor-rig-storage-au/deepmario/rig
scripts/sync-artifacts.sh s3://tensor-rig-storage-au/deepmario/rig
```

One script covers both clouds: the destination scheme selects `gcloud storage rsync` or `aws s3 sync`. It pushes `mario_checkpoints`, `mario_best_model`, `mario_logs` and `mario_videos` from the training checkout every five minutes (`-i`), skipping directories that do not exist yet. The instance service account or instance profile already has write access, so there is no credential setup.

Restore on a replacement instance, then resume:

```bash
scripts/pull-artifacts.sh gs://tensor-rig-storage-au/deepmario/rig
ENCODER=cnn make train
```

`pull-artifacts.sh` skips remote prefixes that are not in the bucket yet, and takes `-v` to include videos. Keep `ENCODER` and `FRAME_STACK` identical to the run that produced the checkpoints: they decide the model name, so a mismatch resumes from nothing.

Neither script passes a delete flag, so `make clean` on the instance cannot wipe the bucket copy. Bucket objects still expire after `storageDeleteAfterDays`.

`sync-artifacts.sh` only makes sense on the instance, where the training output lives. `pull-artifacts.sh` runs there when restoring before a resume, and also works from your workstation if you want the artifacts locally: pass `-r` to point at that checkout, and the copy authenticates as you rather than as the instance role. Both stay in this repo so the training repo does not need to know the bucket exists.

## Public IP hardening

When `instanceRestrictPublicIngress` is `true` (the default), the stack limits what the public IP accepts. Connect with `tailscale ssh`, not SSH to the cloud address.

**GCP** adds firewall rules on the instance network tag: allow UDP 41641 (Tailscale), deny all other TCP and UDP from the internet. That overrides default rules such as `default-allow-ssh`.

**AWS** attaches a security group with no public TCP ingress and UDP 41641 open for Tailscale direct paths.

Set `instanceRestrictPublicIngress` to `false` if you need inbound TCP on the public IP (for example a temporary Jupyter listener).

## Configuration

Keys and defaults are defined in `Pulumi.yaml`. Stack values live in `Pulumi.gcp.yaml` and `Pulumi.aws.yaml`. Edit the stack file or set a key, then run `pulumi up`:

```bash
pulumi config set cloudInitHostname my-gpu-rig
pulumi config set cloudInitLoginUser jasonkwh
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
