package kubeutil

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yeongki/my-operator/pkg/slo"
)

const (
	PrometheusOperatorVersion = "v0.77.1"

	// Split to satisfy lll (max 120 chars) while keeping identical URL.
	prometheusOperatorURLTmpl = "https://github.com/prometheus-operator/" +
		"prometheus-operator/releases/download/%s/bundle.yaml"
)

func PrometheusOperatorURL() string {
	return fmt.Sprintf(prometheusOperatorURLTmpl, PrometheusOperatorVersion)
}

// InstallPrometheusOperator installs Prometheus Operator bundle.
// - enabled=false이면 설치를 건너뛰고 nil 반환(테스트/운영에서 토글하기 쉬움).
// - logger may be nil (no-op).
// - r may be nil (uses DefaultRunner).
func InstallPrometheusOperator(
	ctx context.Context,
	logger slo.Logger,
	r CmdRunner,
	enabled bool,
) error {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = DefaultRunner{}
	}
	if !enabled {
		logger.Logf("prometheus-operator install skipped (disabled)")
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	url := PrometheusOperatorURL()
	logger.Logf(
		"installing prometheus-operator version=%s",
		PrometheusOperatorVersion,
	)

	cmd := exec.Command("kubectl", "apply", "-f", url) // apply is idempotent
	_, err := r.Run(ctx, logger, cmd)
	return err
}

// UninstallPrometheusOperator uninstalls Prometheus Operator bundle.
// - logger may be nil (no-op).
// - r may be nil (uses DefaultRunner).
func UninstallPrometheusOperator(
	ctx context.Context,
	logger slo.Logger,
	r CmdRunner,
) error {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = DefaultRunner{}
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	url := PrometheusOperatorURL()
	logger.Logf(
		"uninstalling prometheus-operator version=%s",
		PrometheusOperatorVersion,
	)

	cmd := exec.Command(
		"kubectl", "delete", "-f", url,
		"--ignore-not-found=true",
	)
	_, err := r.Run(ctx, logger, cmd)
	return err
}

// IsPrometheusOperatorCRDsInstalled checks if Prometheus Operator CRDs exist.
// - logger may be nil (no-op).
// - r may be nil (uses DefaultRunner).
func IsPrometheusOperatorCRDsInstalled(
	ctx context.Context,
	logger slo.Logger,
	r CmdRunner,
) bool {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = DefaultRunner{}
	}
	if err := ctx.Err(); err != nil {
		logger.Logf("IsPrometheusOperatorCRDsInstalled: ctx error: %v", err)
		return false
	}

	prometheusCRDs := []string{
		"prometheuses.monitoring.coreos.com",
		"prometheusrules.monitoring.coreos.com",
		"prometheusagents.monitoring.coreos.com",
	}

	cmd := exec.Command(
		"kubectl", "get", "crds",
		"-o", "custom-columns=NAME:.metadata.name",
	)
	out, err := r.Run(ctx, logger, cmd)
	if err != nil {
		return false
	}

	for _, line := range strings.Split(out, "\n") {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		for _, crd := range prometheusCRDs {
			if strings.Contains(s, crd) {
				return true
			}
		}
	}
	return false
}
