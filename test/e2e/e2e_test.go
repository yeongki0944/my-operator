package e2e

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/yeongki/my-operator/pkg/devutil"
	"github.com/yeongki/my-operator/pkg/kubeutil"
	"github.com/yeongki/my-operator/test/e2e/curlmetrics"
	"github.com/yeongki/my-operator/test/e2e/harness"
	e2eenv "github.com/yeongki/my-operator/test/e2e/internal/env"
)

const namespace = "my-operator-system"
const serviceAccountName = "my-operator-controller-manager"
const metricsServiceName = "my-operator-controller-manager-metrics-service"

var _ = Describe("Manager", Ordered, func() {
	var (
		cfg     e2eenv.Options
		token   string
		rootDir string

		cm *curlmetrics.Client
	)

	BeforeAll(func() {
		cfg = e2eenv.LoadOptions()
		By(fmt.Sprintf("ArtifactsDir=%q RunID=%q Enabled=%v", cfg.ArtifactsDir, cfg.RunID, cfg.Enabled))

		var err error
		rootDir, err = devutil.GetProjectDir()
		Expect(err).NotTo(HaveOccurred())

		cm = curlmetrics.New(logger, runner)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		run := func(cmd *exec.Cmd, msg string) string {
			cmd.Dir = rootDir
			out, err := runner.Run(ctx, logger, cmd)
			Expect(err).NotTo(HaveOccurred(), msg)
			return out
		}

		By("creating manager namespace (idempotent via apply)")
		nsManifest := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace)
		cmd := exec.Command("kubectl", "apply", "-f", "-")
		cmd.Dir = rootDir
		cmd.Stdin = strings.NewReader(nsManifest)
		run(cmd, "Failed to apply namespace")

		By("labeling the namespace to enforce the security policy")
		cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace, "pod-security.kubernetes.io/enforce=baseline")
		cmd.Dir = rootDir
		run(cmd, "Failed to label namespace with security policy")

		By("installing CRDs")
		run(exec.Command("make", "install"), "Failed to install CRDs")

		By("deploying the controller-manager")
		run(exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectImage)), "Failed to deploy the controller-manager")

		By("ensuring metrics reader RBAC for controller-manager SA (idempotent)")
		Expect(kubeutil.ApplyClusterRoleBinding(
			ctx, logger, runner,
			"my-operator-e2e-metrics-reader",
			"my-operator-metrics-reader",
			namespace,
			serviceAccountName,
		)).To(Succeed())
	})

	AfterAll(func() {
		if cfg.SkipCleanup {
			By("E2E_SKIP_CLEANUP enabled: skipping cleanup")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		By("best-effort: cleaning up curl-metrics pods")
		_ = cm.CleanupByLabel(ctx, namespace)

		By("undeploying the controller-manager (best-effort)")
		cmd := exec.Command("make", "undeploy")
		cmd.Dir = rootDir
		_, _ = runner.Run(ctx, logger, cmd)

		By("uninstalling CRDs (best-effort)")
		cmd = exec.Command("make", "uninstall")
		cmd.Dir = rootDir
		_, _ = runner.Run(ctx, logger, cmd)

		By("removing manager namespace (best-effort)")
		cmd = exec.Command("kubectl", "delete", "ns", namespace, "--ignore-not-found=true")
		cmd.Dir = rootDir
		_, _ = runner.Run(ctx, logger, cmd)
	})
	// TODO opts *WaitOptions 로 할지 고민 중
	BeforeEach(func() {
		waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer waitCancel()

		opts := kubeutil.WaitOptions{}

		By("waiting controller-manager ready")
		Expect(
			kubeutil.WaitControllerManagerReady(waitCtx, logger, runner, namespace, opts),
		).To(Succeed())

		By("waiting metrics service endpoints ready")
		Expect(
			kubeutil.WaitServiceHasEndpoints(waitCtx, logger, runner, namespace, metricsServiceName, opts),
		).To(Succeed())

		tokCtx, tokCancel := context.WithTimeout(context.Background(), cfg.TokenRequestTimeout)
		defer tokCancel()

		By("requesting service account token")
		t, err := kubeutil.ServiceAccountToken(tokCtx, logger, runner, namespace, serviceAccountName)
		Expect(err).NotTo(HaveOccurred())
		Expect(t).NotTo(BeEmpty())
		token = t
	})

	harness.Attach(
		func() harness.HarnessDeps {
			return harness.HarnessDeps{
				ArtifactsDir: cfg.ArtifactsDir,
				Suite:        "e2e",
				TestCase:     "",
				RunID:        cfg.RunID,
				Enabled:      cfg.Enabled,
			}
		},
		func() harness.FetchDeps {
			return harness.FetchDeps{
				Namespace:          namespace,
				Token:              token,
				MetricsServiceName: metricsServiceName,
				ServiceAccountName: serviceAccountName,
			}
		},
		harness.DefaultV3Specs,
		harness.CurlPodFns{
			// harness가 기존 함수 타입을 기대한다면, 여기서 얇게 어댑트만 유지
			RunCurlMetricsOnce: func(ns, token, metricsSvcName, sa string) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				return cm.RunOnce(ctx, ns, token, metricsSvcName, sa)
			},
			WaitCurlMetricsDone: func(ns, podName string) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				Expect(cm.WaitDone(ctx, ns, podName, 2*time.Second)).To(Succeed())
			},
			CurlMetricsLogs: func(ns, podName string) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				return cm.Logs(ctx, ns, podName)
			},
			DeletePodNoWait: func(ns, podName string) error {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				return cm.DeletePodNoWait(ctx, ns, podName)
			},
		},
	)

	It("should ensure the metrics endpoint is serving metrics", func() {
		By("scraping /metrics via curl pod")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		podName, err := cm.RunOnce(ctx, namespace, token, metricsServiceName, serviceAccountName)
		Expect(err).NotTo(HaveOccurred())

		defer func() { _ = cm.DeletePodNoWait(context.Background(), namespace, podName) }()

		waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer waitCancel()
		Expect(cm.WaitDone(waitCtx, namespace, podName, 2*time.Second)).To(Succeed())

		text, err := cm.Logs(ctx, namespace, podName)
		Expect(err).NotTo(HaveOccurred())

		if !strings.Contains(text, "controller_runtime_reconcile_total") {
			head := text
			if len(head) > 800 {
				head = head[:800]
			}
			logger.Logf("metrics text head:\n%s", head)
		}

		Expect(text).To(ContainSubstring("controller_runtime_reconcile_total"))
		By(fmt.Sprintf("done (timeout=%s)", 2*time.Minute))
	})
})
