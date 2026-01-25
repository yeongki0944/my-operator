#!/usr/bin/env bash
set -euo pipefail

# my-operator dev entrypoint (plain kubectl + kind)
#
# - No kubectl wrappers assumed
# - Kind cluster/context sanity checks
# - Optional: install CRDs and wait for CRD established
# - Prints the "evidence collection" loop commands
#
# Usage:
#   ./hack/dev-start-kubectl.sh
#
# Optional env:
#   CLUSTER=my-operator
#   NAMESPACE=default
#   METRICS=http://localhost:8080/metrics
#   CONTROLLER_LABEL=control-plane=controller-manager
#   INSTALL_CRDS=1        (default: 1)
#   APPLY_SAMPLE=0        (default: 0)
#   WAIT_CRDS=1           (default: 1)
#   CRD_WAIT_TIMEOUT=30s  (default: 30s)
#   CRD_NAME=             (optional) wait only this CRD (e.g. joboperators.batch.my.domain). If empty, wait --all.

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

cd "${ROOT}"

log "0) prerequisites 확인"
command -v kind >/dev/null 2>&1 || { log "  ✗ kind not found"; exit 1; }
command -v kubectl >/dev/null 2>&1 || { log "  ✗ kubectl not found"; exit 1; }

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
current_ctx="$(kubectl config view -o jsonpath='{.current-context}' 2>/dev/null || echo '<none>')"
expected_ctx="kind-${CLUSTER}"

if [[ "${current_ctx}" == "${expected_ctx}" ]]; then
  log "  ✓ current-context = ${current_ctx}"
else
  log "  ⚠️ 주의: 현재 컨텍스트가 '${expected_ctx}'가 아닙니다."
  log "     현재: ${current_ctx}"
  log "     → kubectl config use-context ${expected_ctx}"
fi

log "3) 노드 상태 확인 (kubectl get nodes)"
kubectl get nodes

if [[ "${INSTALL_CRDS}" == "1" ]]; then
  log "4) CRD 설치 (make install)"
  make install

  if [[ "${WAIT_CRDS}" == "1" ]]; then
    log "   → CRD 등록 대기(Establish) 중... (timeout=${CRD_WAIT_TIMEOUT})"
    if [[ -n "${CRD_NAME}" ]]; then
      kubectl wait --for=condition=established "crd/${CRD_NAME}" --timeout="${CRD_WAIT_TIMEOUT}" 2>/dev/null || true
    else
      kubectl wait --for=condition=established crd --all --timeout="${CRD_WAIT_TIMEOUT}" 2>/dev/null || true
    fi
  fi
else
  log "4) CRD 설치 스킵 (INSTALL_CRDS=${INSTALL_CRDS})"
fi

if [[ "${APPLY_SAMPLE}" == "1" ]]; then
  if compgen -G "config/samples/*.yaml" >/dev/null; then
    log "5) 샘플 적용 (config/samples/*.yaml)"
    kubectl apply -n "${NAMESPACE}" -f config/samples/
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
echo "       kubectl apply -n ${NAMESPACE} -f config/samples/"
echo
echo "  3) 증거 수집: 메트릭 증가 확인 (로컬 make run 기준: curl localhost)"
echo "     - 컨트롤러를 클러스터 내부(deploy)로 띄웠다면 port-forward가 필요"
echo "       예) kubectl port-forward -n <ns> deploy/<controller-name> 8080:8080"
echo
echo "       METRICS=\${METRICS:-${METRICS}}"
echo "       curl -s \"\$METRICS\" | grep 'e2e_convergence_time_seconds_count' || true"
echo "       curl -s \"\$METRICS\" | grep -E 'e2e_convergence_time_seconds(_bucket|_sum|_count)' | head -n 50 || true"
echo
echo "  4) 증거 수집: observe/skip 로그 확인"
echo "     - make run 터미널에서 직접 확인하거나,"
echo "     - 컨트롤러를 Deployment로 띄운 경우:"
echo "       kubectl logs -n ${NAMESPACE} deploy/<controller-name> | grep -E 'observe success|skip:'"
echo
echo "  5) 컨트롤러 Pod 빠르게 찾기(선택):"
echo "       kubectl get pods -A -l \"${CONTROLLER_LABEL}\""
echo
log "=== dev-start 점검 완료 ==="
echo "  CLUSTER=${CLUSTER}"
echo "  NAMESPACE=${NAMESPACE}"
echo "  METRICS=${METRICS}"
echo "  CONTROLLER_LABEL=${CONTROLLER_LABEL}"
echo "  CRD_NAME=${CRD_NAME:-<empty -> wait --all>}"
