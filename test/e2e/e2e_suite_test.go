package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/yeongki/my-operator/pkg/devutil"
	"github.com/yeongki/my-operator/pkg/kubeutil"
	"github.com/yeongki/my-operator/pkg/slo"
	"github.com/yeongki/my-operator/test/e2e/e2eutil"
)

var (
	// Optional Environment Variables:
	// - CERT_MANAGER_INSTALL_SKIP=true: Skips CertManager installation during test setup.
	skipCertManagerInstall = os.Getenv("CERT_MANAGER_INSTALL_SKIP") == "true"

	// isCertManagerAlreadyInstalled will be set true when CertManager CRDs are found on the cluster.
	isCertManagerAlreadyInstalled = false

	// projectImage is the name of the image which will be built and loaded with the code source changes to be tested.
	projectImage = "example.com/my-operator:v0.0.1"

	// logger is the suite logger. It is always safe (nil -> no-op).
	logger = slo.NewLogger(e2eutil.GinkgoLog)

	// runner is used by kubeutil/devutil helpers (context-aware).
	runner kubeutil.CmdRunner = kubeutil.DefaultRunner{}
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.Logf("Starting my-operator integration test suite")
	RunSpecs(t, "e2e suite")
}

var _ = BeforeSuite(func() {
	// A reasonable default guard for setup steps.
	// Individual kubectl commands also have their own timeouts (e.g. kubectl wait --timeout).
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	By("building the manager(Operator) image")
	root, err := devutil.GetProjectDir()
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectImage))
	cmd.Dir = root

	_, err = runner.Run(ctx, logger, cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to build the manager(Operator) image")

	By("loading the manager(Operator) image on Kind")
	Expect(devutil.LoadImageToKindClusterWithName(ctx, logger, runner, projectImage)).
		To(Succeed(), "Failed to load the manager(Operator) image into Kind")

	// Setup CertManager before the suite if not skipped and if not already installed.
	if skipCertManagerInstall {
		logger.Logf("CERT_MANAGER_INSTALL_SKIP=true: skipping cert-manager setup")
		return
	}

	By("checking if cert-manager is installed already")
	isCertManagerAlreadyInstalled = kubeutil.IsCertManagerCRDsInstalled(ctx, logger, runner)
	if isCertManagerAlreadyInstalled {
		logger.Logf("WARNING: cert-manager is already installed; skipping installation")
		return
	}

	By("installing cert-manager")
	Expect(kubeutil.InstallCertManager(ctx, logger, runner)).
		To(Succeed(), "Failed to install cert-manager")
})

var _ = AfterSuite(func() {
	if skipCertManagerInstall || isCertManagerAlreadyInstalled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	By("uninstalling cert-manager (best-effort)")
	if err := kubeutil.UninstallCertManager(ctx, logger, runner); err != nil {
		warnf("failed to uninstall cert-manager: %v", err)
	}
})

func warnf(format string, args ...any) {
	logger.Logf("WARNING: "+format, args...)
}
