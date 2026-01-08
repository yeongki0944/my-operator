#!/usr/bin/env bash
set -euo pipefail

# my-operator dev entrypoint (kind + kubebuilder)
#
# Goals:
# - Guardrails: kind cluster/context/nodes sanity checks
# - Optional: install CRDs and wait for CRD established (reduce race)
# - One-liner hints for the "evidence collection" loop:
#   - run controller (make run) locally
#   - apply sample CR
#   - verify /metrics increased (e2e_convergence_time_seconds_count)
#   - verify logs contain observe/skip
#
# Usage:
#   ./hack/dev-start.sh
#
# Optional env:
#   CLUSTER=my-operator
#   NAMESPACE=default
#   METRICS=http://localhost:8080/metrics
#   CONTROLLER_LABEL=control-plane=controller-manager   (or app.kubernetes.io/name=my-operator)
#   INSTALL_CRDS=1    (default: 1)
#   APPLY_SAMPLE=0    (default: 0)  # set 1 to auto-apply config/samples/*
#   WAIT_CRDS=1       (default: 1)  # wait for CRD established after make install
#   CRD_WAIT_TIMEOUT=30s (default: 30s)
#   CRD_NAME=         (optional) wait only this CRD (e.g. joboperators.batch.my.domain). If empty, wait --all.
#   KUBECTL=kubectl   (wrapper supported)

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
log() { echo "[$(date +'%H:%M:%S')] $*"; }

CLUSTER="${CLUSTER:-my-operator}"
NAMESPACE="${NAMESPACE:-default}"
METRICS="${METRICS:-http://localhost:8080/metrics}"
CONTROLLER_LABEL="${CONTROLLER_LABEL:-control-plane=controller-manager}"
INSTALL_CRDS="${INSTALL_CRDS:-1}"
APPLY_SAMPLE="${APPLY_SAMPLE:-0}"

WAIT_CRDS="${WAIT_CRDS:-1}"
CRD_WAIT_TIMEOUT="${CRD_WAIT_TIMEOUT:-30s}"
CRD_NAME="${CRD_NAME:-}"

KUBECTL="${KUBECTL:-kubectl}"

# 권장 실행 위치 안내
if [[ "$(pwd)" == "${ROOT}/hack" ]]; then
  echo "================================================================"
  echo "  [my-operator] 권장 실행 위치는 리포지토리 루트입니다."
  echo
  echo "    예) \$ cd ${ROOT}"
  echo "        \$ ./hack/dev-start.sh"
  echo
  echo "  현재 디렉터리: $(pwd)"
  echo "================================================================"
fi

cd "${ROOT}"

log "0) prerequisites 확인"
command -v kind >/dev/null 2>&1 || { log "  ✗ kind not found in PATH"; exit 1; }
command -v "${KUBECTL}" >/dev/null 2>&1 || { log "  ✗ KUBECTL(${KUBECTL}) not found in PATH"; exit 1; }

log "1) kind 클러스터(${CLUSTER}) 존재 여부 확인"
if ! kind get clusters | grep -q "^${CLUSTER}$"; then
  log "  ✗ kind 클러스터 '${CLUSTER}' 을 찾지 못했습니다."
  log "    → 먼저 아래 명령으로 클러스터를 만들고 다시 실행하세요:"
  echo "      kind create cluster --name ${CLUSTER}"
  exit 1
else
  log "  ✓ kind 클러스터 '${CLUSTER}' 감지"
fi

log "2) kubectl current-context 확인"
# wrapper 출력/포맷 문제를 피하려면 config view jsonpath가 더 안전
current_ctx="$(${KUBECTL} config view -o jsonpath='{.current-context}' 2>/dev/null || echo '<none>')"
expected_ctx="kind-${CLUSTER}"

if [[ "${current_ctx}" == "${expected_ctx}" ]]; then
  log "  ✓ current-context = ${current_ctx}"
else
  log "  ⚠️ 주의: 현재 컨텍스트가 '${expected_ctx}'가 아닙니다."
  log "     현재: ${current_ctx}"
  log "     → ${KUBECTL} config use-context ${expected_ctx}"
fi

log "3) 노드 상태 확인 (${KUBECTL} get nodes)"
${KUBECTL} get nodes

if [[ "${INSTALL_CRDS}" == "1" ]]; then
  log "4) CRD 설치 (make install)"
  make install

  if [[ "${WAIT_CRDS}" == "1" ]]; then
    log "   → CRD 등록 대기(Establish) 중... (timeout=${CRD_WAIT_TIMEOUT})"
    if [[ -n "${CRD_NAME}" ]]; then
      # 특정 CRD만 기다리기 (권장: 더 결정적)
      ${KUBECTL} wait --for=condition=established "crd/${CRD_NAME}" --timeout="${CRD_WAIT_TIMEOUT}" 2>/dev/null || true
    else
      # CRD_NAME을 모르면 전체 기다리기 (간단하지만 범위 큼)
      ${KUBECTL} wait --for=condition=established crd --all --timeout="${CRD_WAIT_TIMEOUT}" 2>/dev/null || true
    fi
  fi
else
  log "4) CRD 설치 스킵 (INSTALL_CRDS=${INSTALL_CRDS})"
fi

if [[ "${APPLY_SAMPLE}" == "1" ]]; then
  if compgen -G "config/samples/*.yaml" >/dev/null; then
    log "5) 샘플 적용 (config/samples/*.yaml)"
    ${KUBECTL} apply -n "${NAMESPACE}" -f config/samples/
  else
    log "5) 샘플 적용 스킵: config/samples/*.yaml 없음"
  fi
else
  log "5) 샘플 적용은 수동 (APPLY_SAMPLE=${APPLY_SAMPLE})"
fi

echo
log "=== 다음 단계(수동) ==="
echo "  1) 새 터미널에서 컨트롤러 실행 (로컬 실행, 로그를 띄워두기):"
echo "       cd ${ROOT}"
echo "       make run"
echo
echo "  2) (선택) 샘플 CR 적용:"
echo "       ${KUBECTL} apply -n ${NAMESPACE} -f config/samples/"
echo
echo "  3) 증거 수집: 메트릭 증가 확인 (로컬 make run 기준: curl localhost)"
echo "     - 컨트롤러를 클러스터 내부(deploy)로 띄웠다면 port-forward가 필요"
echo "       예) ${KUBECTL} port-forward -n <ns> deploy/<controller-name> 8080:8080"
echo
echo "       METRICS=\${METRICS:-${METRICS}}"
echo "       curl -s \"\$METRICS\" | grep 'e2e_convergence_time_seconds_count' || true"
echo "       curl -s \"\$METRICS\" | grep -E 'e2e_convergence_time_seconds(_bucket|_sum|_count)' | head -n 50 || true"
echo
echo "     최소 조건: e2e_convergence_time_seconds_count 가 1 이상(또는 이전 대비 증가)"
echo
echo "  4) 증거 수집: observe/skip 로그 확인"
echo "     - make run 터미널에서 직접 확인하거나,"
echo "     - 컨트롤러를 Deployment로 띄운 경우:"
echo "       ${KUBECTL} logs -n ${NAMESPACE} deploy/<controller-name> | grep -E 'observe success|skip:'"
echo
echo "  5) 컨트롤러 Pod 빠르게 찾기(선택):"
echo "       ${KUBECTL} get pods -A -l \"${CONTROLLER_LABEL}\""
echo
log "=== dev-start 점검 완료 ==="
echo "  CLUSTER=${CLUSTER}"
echo "  NAMESPACE=${NAMESPACE}"
echo "  METRICS=${METRICS}"
echo "  CONTROLLER_LABEL=${CONTROLLER_LABEL}"
echo "  CRD_NAME=${CRD_NAME:-<empty -> wait --all>}"
