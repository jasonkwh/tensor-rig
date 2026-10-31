#!/usr/bin/env bash
set -euo pipefail

readonly DEFAULT_ROOT="${ARTIFACT_ROOT:-${HOME}/DeepMario-SMW-Vision}"
ARTIFACT_DIRS=(mario_checkpoints mario_best_model mario_logs)

usage() {
  cat <<'EOF'
Pull training artifacts from the rig bucket back onto the instance.

Usage:
  pull-artifacts.sh [options] SRC

SRC must start with gs:// (GCS, uses gcloud storage) or s3:// (S3, uses
aws s3). Remote prefixes that do not exist yet are skipped, so this is safe
on a fresh box. Nothing is deleted locally.

Options:
  -r, --root DIR   training checkout to restore into (default:
                   $HOME/DeepMario-SMW-Vision, or $ARTIFACT_ROOT)
  -v, --videos     also pull mario_videos (large, off by default)
  -h, --help       show this help

Example:
  pull-artifacts.sh gs://tensor-rig-storage-au/deepmario/rig
  ENCODER=cnn make train    # resume once mario_checkpoints is back

Keep ENCODER and FRAME_STACK identical to the run that produced the
checkpoints: they decide MODEL_NAME, so a mismatch globs nothing and every
observation shape is wrong.
EOF
}

ROOT="${DEFAULT_ROOT}"
SRC=""
WITH_VIDEOS=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    -r|--root)
      ROOT="$2"
      shift 2
      ;;
    -v|--videos)
      WITH_VIDEOS=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      SRC="$1"
      shift
      ;;
  esac
done

if [[ -z "${SRC}" ]]; then
  usage >&2
  exit 2
fi

case "${SRC}" in
  gs://*) CLI=gcloud ;;
  s3://*) CLI=aws ;;
  *)
    echo "source must start with gs:// or s3://: ${SRC}" >&2
    exit 2
    ;;
esac

if ! command -v "${CLI}" >/dev/null 2>&1; then
  echo "${CLI} not found on PATH (needed for ${SRC%%://*}://)" >&2
  exit 1
fi

if [[ ! -d "${ROOT}" ]]; then
  echo "training checkout not found: ${ROOT}" >&2
  exit 1
fi

log() {
  printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"
}

has_remote() {
  local out

  case "${CLI}" in
    gcloud) out=$(gcloud storage ls "${SRC}/${1}" 2>/dev/null || true) ;;
    aws) out=$(aws s3 ls "${SRC}/${1}/" 2>/dev/null || true) ;;
  esac

  [[ -n "${out}" ]]
}

pull_once() {
  local dir dst

  for dir in "${ARTIFACT_DIRS[@]}"; do
    if ! has_remote "${dir}"; then
      log "skip ${dir} (not in bucket)"
      continue
    fi

    dst="${ROOT}/${dir}"
    mkdir -p "${dst}"

    case "${CLI}" in
      gcloud) gcloud storage rsync -r "${SRC}/${dir}" "${dst}" ;;
      aws) aws s3 sync "${SRC}/${dir}" "${dst}" ;;
    esac

    log "pulled ${dir} -> ${dst}"
  done
}

if [[ "${WITH_VIDEOS}" -eq 1 ]]; then
  ARTIFACT_DIRS+=(mario_videos)
fi

pull_once
log "done; latest checkpoint is in ${ROOT}/mario_checkpoints"
