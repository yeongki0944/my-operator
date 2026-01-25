package devutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/yeongki/my-operator/pkg/kubeutil"
	"github.com/yeongki/my-operator/pkg/slo"
)

// LoadImageToKindClusterWithName loads a local docker image into a kind cluster.
//
// Resolution order for cluster name:
//  1. env KIND_CLUSTER
//  2. default "kind"
//
// logger may be nil (no-op).
// r may be nil (uses kubeutil.DefaultRunner{}).
func LoadImageToKindClusterWithName(ctx context.Context, logger slo.Logger, r kubeutil.CmdRunner, image string) error {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = kubeutil.DefaultRunner{}
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok && v != "" {
		cluster = v
	}

	logger.Logf("loading image into kind cluster=%q image=%q", cluster, image)

	cmd := exec.Command("kind", "load", "docker-image", image, "--name", cluster)
	if _, err := r.Run(ctx, logger, cmd); err != nil {
		return fmt.Errorf("kind load docker-image failed: %w", err)
	}
	return nil
}
