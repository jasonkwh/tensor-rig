#!/usr/bin/env bash
set -euo pipefail

readonly DEFAULT_PROJECT="342105122936"
readonly DEFAULT_REPO="jasonkwh/tensor-rig"
readonly DEFAULT_ENVIRONMENT="deploy"
readonly DEFAULT_SA_NAME="pulumi-deploy"
readonly DEFAULT_POOL="github"
readonly DEFAULT_PROVIDER="tensor-rig-v2"

# Least privilege for what this program actually creates: a GCE instance, a
# GCS bucket, and the service account attached to that instance.
readonly SA_ROLES=(roles/compute.admin roles/storage.admin roles/iam.serviceAccountUser)

usage() {
  cat <<EOF
Set up Workload Identity Federation so GitHub Actions can deploy the gcp
stack without a service account key sitting in the repository.

Creates:
  1. a service account that holds the deploy permissions
  2. an IAM policy binding for each role that account needs
  3. a workload identity pool and an OIDC provider that trusts
     token.actions.githubusercontent.com
  4. a binding letting that pool impersonate the service account

The provider's attribute condition only accepts a job running in the given
repository AND in the given GitHub environment, so a fork, a pull request,
or a workflow edited on a branch cannot mint credentials.

Requires the Cloud IAM API, gcloud auth as a user with permission to manage
IAM policies, and about five minutes of propagation time afterwards.

Usage:
  setup-gcp-wif.sh [options]

Options:
  -p, --project PROJECT        GCP project id or number (default: $DEFAULT_PROJECT)
  -r, --repo OWNER/NAME        GitHub repository to trust (default: $DEFAULT_REPO)
  -e, --environment NAME       GitHub environment to trust (default: $DEFAULT_ENVIRONMENT)
  -s, --service-account NAME   service account id to create (default: $DEFAULT_SA_NAME)
  -h, --help                   show this help

Safe to re-run: every step checks whether the resource already exists.
EOF
}

log() {
  printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"
}

die() {
  echo "$*" >&2
  exit 1
}

# A freshly created service account is not visible to IAM policy changes for
# up to a minute, so bindings fail with "does not exist" right after creation.
retry() {
  local max="$1"
  local attempt=1

  shift

  until "$@"; do
    if [[ "${attempt}" -ge "${max}" ]]; then
      return 1
    fi

    log "  retrying in 15s (${attempt}/${max})" >&2
    attempt=$((attempt + 1))
    sleep 15
  done
}

PROJECT="${DEFAULT_PROJECT}"
REPO="${DEFAULT_REPO}"
ENVIRONMENT="${DEFAULT_ENVIRONMENT}"
SA_NAME="${DEFAULT_SA_NAME}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -p|--project)
      PROJECT="$2"
      shift 2
      ;;
    -r|--repo)
      REPO="$2"
      shift 2
      ;;
    -e|--environment)
      ENVIRONMENT="$2"
      shift 2
      ;;
    -s|--service-account)
      SA_NAME="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ "${REPO}" != */* ]]; then
  die "repo must be OWNER/NAME, got: ${REPO}"
fi

command -v gcloud >/dev/null 2>&1 || die "gcloud not found on PATH"

log "resolving project ${PROJECT}"
PROJECT_ID=$(gcloud projects describe "${PROJECT}" --format='value(projectId)')
PROJECT_NUMBER=$(gcloud projects describe "${PROJECT}" --format='value(projectNumber)')

# Every later gcloud call wants the project id; a few reject the number
# outright. The number is only used to build the workload identity paths.
PROJECT="${PROJECT_ID}"

# STS exchanges the GitHub token and IAM Credentials mints the impersonated
# access token; without these the workflow fails at the auth step, not here.
readonly REQUIRED_APIS=(
  iam.googleapis.com
  cloudresourcemanager.googleapis.com
  iamcredentials.googleapis.com
  sts.googleapis.com
)

log "enabling required APIs"
gcloud services enable "${REQUIRED_APIS[@]}" --project="${PROJECT}" --quiet

SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
MEMBER="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${DEFAULT_POOL}/attribute.repository/${REPO}"
PROVIDER_NAME="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${DEFAULT_POOL}/providers/${DEFAULT_PROVIDER}"

# google.subject is used for the condition below; attribute.repository is
# redundant for authorization but shows up in audit logs, which helps when
# you are asking "who deployed this".
ATTRIBUTE_MAPPING="google.subject=assertion.sub,attribute.repository=assertion.repository"
ATTRIBUTE_CONDITION="assertion.repository=='${REPO}' && assertion.environment=='${ENVIRONMENT}'"

log "service account ${SA_EMAIL}"
if gcloud iam service-accounts describe "${SA_EMAIL}" --project="${PROJECT}" >/dev/null 2>&1; then
  log "  exists"
else
  gcloud iam service-accounts create "${SA_NAME}" \
    --project="${PROJECT}" \
    --display-name="tensor-rig Pulumi deploy (GitHub Actions)"
  log "  created"
fi

for role in "${SA_ROLES[@]}"; do
  log "grant ${role}"
  retry 8 gcloud projects add-iam-policy-binding "${PROJECT}" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="${role}" \
    --condition=None \
    --quiet >/dev/null
done

log "workload identity pool ${DEFAULT_POOL}"
if gcloud iam workload-identity-pools describe "${DEFAULT_POOL}" \
  --project="${PROJECT}" --location=global >/dev/null 2>&1; then
  log "  exists"
else
  gcloud iam workload-identity-pools create "${DEFAULT_POOL}" \
    --project="${PROJECT}" \
    --location=global \
    --display-name="GitHub Actions"
  log "  created"
fi

log "OIDC provider ${DEFAULT_PROVIDER}"
# A deleted provider lingers as a DELETED tombstone that describe still
# returns, so presence alone is not enough to decide it is usable.
PROVIDER_STATE=$(gcloud iam workload-identity-pools providers describe "${DEFAULT_PROVIDER}" \
  --project="${PROJECT}" --location=global \
  --workload-identity-pool="${DEFAULT_POOL}" \
  --format='value(state)' 2>/dev/null || true)

if [[ "${PROVIDER_STATE}" == "ACTIVE" ]]; then
  log "  exists"
else
  gcloud iam workload-identity-pools providers create-oidc "${DEFAULT_PROVIDER}" \
    --project="${PROJECT}" \
    --location=global \
    --workload-identity-pool="${DEFAULT_POOL}" \
    --display-name="tensor-rig" \
    --issuer-uri="https://token.actions.githubusercontent.com" \
    --attribute-mapping="${ATTRIBUTE_MAPPING}" \
    --attribute-condition="${ATTRIBUTE_CONDITION}"
  log "  created"
fi

log "allow the pool to impersonate ${SA_NAME}"
retry 8 gcloud iam service-accounts add-iam-policy-binding "${SA_EMAIL}" \
  --project="${PROJECT}" \
  --role="roles/iam.workloadIdentityUser" \
  --member="${MEMBER}" \
  --quiet >/dev/null

cat <<EOF

Done. Values for .github/workflows/deploy-gcp.yml:

  workload_identity_provider: ${PROVIDER_NAME}
  service_account: ${SA_EMAIL}

Or store the account email as the repository variable GCP_SA_EMAIL and leave
the workflow as it is.

Before the first run:
  - create the GitHub environment "${ENVIRONMENT}" and add a required reviewer
  - add the repository secret PULUMI_ACCESS_TOKEN
  - add the repository secret TS_AUTH_KEY_RIG

New pools and IAM bindings can take five minutes to propagate. A 403 right
after this script finishes is usually that, not a typo.
EOF
