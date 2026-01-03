#!/bin/bash
set -e

# 색상 정의
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "=== 메트릭 검증 시작 ==="

# 1. 토큰 발급 (CI 환경 가정: serviceaccount 사용)
echo "1. 인증 토큰 발급 중..."
TOKEN=$(kubectl create token default -n default)

# 2. 메트릭 조회 (로컬 포트 8443 가정)
# CI에서는 포트 포워딩을 백그라운드로 띄워놓고 이 스크립트를 실행하게 됩니다.
METRICS_URL="https://localhost:8443/metrics"

echo "2. 메트릭 데이터 가져오는 중... ($METRICS_URL)"
# -k: SSL 무시, -s: silent
RAW_METRICS=$(curl -k -s -H "Authorization: Bearer $TOKEN" $METRICS_URL)

# 메트릭을 못 가져왔으면 실패 처리
if [ -z "$RAW_METRICS" ]; then
    echo -e "${RED}❌ 메트릭을 조회할 수 없습니다.${NC}"
    exit 1
fi

# 3. 데이터 파싱 및 검증

# [검증 1] Reconcile 성공 횟수 확인 (최소 1회 이상이어야 함)
# grep으로 해당 라인을 찾고, awk로 마지막 숫자(값)만 추출
SUCCESS_COUNT=$(echo "$RAW_METRICS" | grep 'joboperator_reconcile_total{.*result="success"}' | awk '{print $2}')
# 값이 없으면 0으로 취급
SUCCESS_COUNT=${SUCCESS_COUNT:-0}

echo -e "▶ Reconcile 성공 횟수: ${SUCCESS_COUNT}"

if [ "$SUCCESS_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✔ 성공 카운트 검증 통과${NC}"
else
    echo -e "${RED}❌ 실패: 성공 횟수가 0입니다. Operator가 동작하지 않았습니다.${NC}"
    exit 1
fi

# [검증 2] Reconcile 에러 횟수 확인 (0이어야 함)
# 에러 메트릭은 에러가 없으면 아예 출력되지 않을 수 있음 -> grep 실패 시 0으로 처리
ERROR_COUNT=$(echo "$RAW_METRICS" | grep 'joboperator_reconcile_total{.*result="error"}' | awk '{print $2}' || echo "0")
ERROR_COUNT=${ERROR_COUNT:-0}

echo -e "▶ Reconcile 에러 횟수: ${ERROR_COUNT}"

if [ "$ERROR_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✔ 에러 카운트 검증 통과${NC}"
else
    echo -e "${RED}❌ 실패: 에러가 발생했습니다. (횟수: $ERROR_COUNT)${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== 모든 메트릭 검증 성공! ===${NC}"
exit 0