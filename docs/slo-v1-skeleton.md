# SLO v1 뼈대 구조 및 레거시 코드 정리

이 문서는 v1 SLO 뼈대 코드가 추가된 위치와, 아직 사용하지 않는 레거시 코드 위치를 기록합니다.
레거시 코드는 삭제하지 않고 그대로 보관하며, 이후 어댑터 또는 리팩터링 과정에서 참고하도록 합니다.

## v1 뼈대 코드 위치

- `pkg/slo/spec`: SLI 스펙과 레지스트리 정의
- `pkg/slo/fetch`: 메트릭 스냅샷 Fetcher 인터페이스 및 Prometheus text 파서
- `pkg/slo/summary`: 실행 결과 요약 스키마와 JSON writer
- `pkg/slo/engine`: v1 엔진 및 실행 요청 타입
- `presets/`: controller-runtime 및 my-operator SLI 프리셋
- `test/e2e/harness`: 테스트 시점에 엔진을 호출하는 glue 코드

## 현재 사용되지 않는 레거시 코드 (삭제하지 않음)

아래 경로는 기존 v2 계측/하네스 코드로서, v1 엔진과 직접 연결되지 않습니다.
향후 `engine.Execute` 기반의 어댑터를 추가하거나, v1 구조로 흡수할 때 재검토합니다.

- `pkg/slo/instrumentv2/`: 기존 계측 로직 (레거시)
- `test/e2e/harnessv2/`: v2 하네스 (레거시)
- `test/e2e/instrument/`: v2 계측 테스트 헬퍼 (레거시)
- `test/e2e/helpers_metrics.go`: v2 계측 관련 헬퍼 (레거시)

## 정리 원칙

- 레거시 파일은 삭제하거나 이동하지 않고 그대로 둡니다.
- 신규 v1 코드와 연결될 때까지는 “참고용”으로 남겨둡니다.
- 실제 실행 경로는 v1 엔진(`pkg/slo/engine`)으로 통일합니다.
