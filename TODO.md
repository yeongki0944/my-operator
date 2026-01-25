#### TODO
~~- ./hack/dev-start.sh 오탐하는 것 해결해야함. wrapper 나오는 부분 해결했는데 다시 문제남.~~
~~- 네임스페이스 지금 default 로 쓰는데 정하고 통일 해야함.~~
~~- instrument 하고 recorder 좀 겹치는 느낌이 있음.~~
~~- method 로 help 관련 옮기기~~  
~~- a_metrics_text_delta.go 이거 일단 살리는 방향으로..~~


~~### 1) 프로덕션 경로에서 `kubectl shell-out` 제거 (client-go/controller-runtime 전환)~~ -> 테스트 에서는 kubectl 형태로 하는게 더 안정적 이고 멱등성 좋음.
~~- Kubernetes API 직접 호출(client-go / controller-runtime client)로 전환:~~
  ~~- ServiceAccount 토큰 요청: `kubectl create --raw .../token` → `client.CoreV1().ServiceAccounts(ns).CreateToken(...)~~
  ~~- RBAC 생성/적용: `kubectl apply -f -` → `client.RbacV1().ClusterRoleBindings().Create/Update(...)`~~
  ~~- Wait 로직: `kubectl get ... -o jsonpath=...` → Watch/Informer 기반으로 개선(효율/안정성)~~
- 현재 `kubectl` 기반 로직은 **e2e/dev tooling에만** 남기고, 프로덕션 코드에서는 사용 금지

### 2) 매니페스트/인자 Injection 방지 (tooling 포함)
- RBAC YAML을 `fmt.Sprintf`로 조합하는 방식 중단 -> 일단 namespace 만 구현했는데 나머지 것들도 같은 방식으로 구현해야 한다.
  - ~~`rbacv1.ClusterRoleBinding` 등 **타입(구조체) 기반**으로 만들고 `sigs.k8s.io/yaml` 또는 JSON 마샬링 사용~~ -> 이건 그냥 템플릿 방식으로 구현함.
- 외부/비신뢰 입력은 사전 검증
  - DNS-1123 name, namespace 규칙 등 검증 후 명령/매니페스트에 반영

### 3) Runner 에러 처리 구조화
- stderr+stdout 합쳐 문자열로만 내보내는 방식 개선
    - 구조화된 에러 타입 도입 (아래 예시, 꼭 이렇게 안해도 됨.):
      - `type CmdError struct { Cmd string; Stdout string; Stderr string; Err error }`
      - `Error()`는 사람이 보기 좋게 출력하되, 호출자는 `errors.As`로 필드를 분석 가능하게

### 4) 외부 컴포넌트 버전/URL 하드코딩 완화
- cert-manager / prometheus-operator 설치 로직의 버전/URL 하드코딩 개선
  - e2e/dev에서는 env override 제공 (예: `CERT_MANAGER_VERSION`, `PROM_OPERATOR_VERSION`)

### 5) SLO 정책: Counter reset 처리 재검토
- `ComputeDelta`에서 counter reset 감지 시 정책 정리
  - Judge 단계 스킵 대신 `InsufficientData/DataInvalid` 같은 명시 상태 도입 고려
  - 가능하면 Prometheus의 rate/increase 방식으로 reset 보정하는 전략 검토

### 실행할때 (붙여넣기용)
export E2E_SKIP_CLEANUP=1
export ARTIFACTS_DIR=/tmp/slo-artifacts
export SLOLAB_ENABLED=1
export CI_RUN_ID=local-$(date +%s)

make test-e2e


