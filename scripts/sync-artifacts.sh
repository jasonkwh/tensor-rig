#!/usr/bin/env bash
set -euo pipefail

readonly DEFAULT_ROOT="${ARTIFACT_ROOT:-${HOME}/DeepMario-SMW}"
readonly ARTIFACT_DIRS=(mario_checkpoints mario_best_model mario_logs mario_videos)

usage() {
  cat <<'EOF'
Push training artifacts to the rig bucket, on a loop.

Usage:
  sync-artifacts.sh [options] DEST

DEST must start with gs:// (GCS, uses gcloud storage) or s3:// (S3, uses
aws s3). The instance service account / instance profile already has write
access to the bucket, so no credentials are needed on the box.

Options:
  -r, --root DIR       training checkout to push (default: $HOME/DeepMario-SMW,
                       or $ARTIFACT_ROOT)
  -i, --interval SECS  seconds between passes (default: 300; 0 = push once and
                       exit, useful right before tearing the rig down)
  -h, --help           show this help

Dirs pushed: mario_checkpoints, mario_best_model, mario_logs, mario_videos.
Missing dirs are skipped. Nothing is ever deleted at the destination.

Examples:
  sync-artifacts.sh gs://tensor-rig-storage-au/deepmario/rig
  sync-artifacts.sh -i 0 s3://tensor-rig-storage-au/deepmario/rig
EOF
}

ROOT="${DEFAULT_ROOT}"
INTERVAL=300
DEST=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -r|--root)
      ROOT="$2"
      shift 2
      ;;
    -i|--interval)
      INTERVAL="$2"
      shift 2
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
      DEST="$1"
      shift
      ;;
  esac
done

if [[ -z "${DEST}" ]]; then
  usage >&2
  exit 2
fi

if [[ ! -d "${ROOT}" ]]; then
  echo "training checkout not found: ${ROOT}" >&2
  exit 1
fi

case "${DEST}" in
  gs://*) CLI=gcloud ;;
  s3://*) CLI=aws ;;
  *)
    echo "destination must start with gs:// or s3://: ${DEST}" >&2
    exit 2
    ;;
esac

if ! command -v "${CLI}" >/dev/null 2>&1; then
  echo "${CLI} not found on PATH (needed for ${DEST%%://*}://)" >&2
  exit 1
fi

if ! [[ "${INTERVAL}" =~ ^[0-9]+$ ]]; then
  echo "interval must be a non-negative integer: ${INTERVAL}" >&2
  exit 2
fi

log() {
  printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"
}

push_once() {
  local dir src

  for dir in "${ARTIFACT_DIRS[@]}"; do
    src="${ROOT}/${dir}"
    if [[ ! -d "${src}" ]]; then
      continue
    fi

    case "${CLI}" in
      gcloud) gcloud storage rsync -r "${src}" "${DEST}/${dir}" ;;
      aws) aws s3 sync "${src}" "${DEST}/${dir}" ;;
    esac
  done
}

if [[ "${INTERVAL}" -eq 0 ]]; then
  push_once
  log "pushed ${ROOT} -> ${DEST}"
  exit 0
fi

trap 'log "stopping"; push_once || true; exit 0' INT TERM

log "pushing ${ROOT} -> ${DEST} every ${INTERVAL}s via ${CLI}"

while true; do
  if push_once; then
    log "pass ok"
  else
    log "pass failed, retrying in ${INTERVAL}s"
  fi
  sleep "${INTERVAL}"
done
