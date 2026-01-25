package kubeutil

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yeongki/my-operator/pkg/slo"
)

// TODO(refactor): Extract the common polling loop logic into a generic 'Poll' function.
// Currently, similar retry loops are duplicated here and in pkg/kubeutil/token.go.
// Implementing a shared 'Poll(ctx, interval, condition)' utility will standardize
// async waiting patterns and reduce code duplication.

// WaitOptions controls polling behavior.
type WaitOptions struct {
	Timeout  time.Duration // overall timeout (0 => default)
	Interval time.Duration // poll interval (0 => default)
}

// withDefaults applies safe defaults.
func (o WaitOptions) withDefaults() WaitOptions {
	if o.Timeout <= 0 {
		o.Timeout = 5 * time.Minute
	}
	if o.Interval <= 0 {
		o.Interval = 5 * time.Second
	}
	return o
}

// WaitControllerManagerReady waits until controller-manager pod is Ready.
// Assumes label selector "control-plane=controller-manager" (kubebuilder default).
func WaitControllerManagerReady(ctx context.Context, logger slo.Logger, r CmdRunner, ns string, opts WaitOptions) error {
	return WaitPodContainerReadyByLabel(
		ctx,
		logger,
		r,
		ns,
		"control-plane=controller-manager",
		0,
		0,
		opts,
	)
}

// WaitPodContainerReadyByLabel waits until the first matching pod's Nth container is ready.
// podIndex/containerIndex default to 0 in most kubebuilder setups.
func WaitPodContainerReadyByLabel(ctx context.Context, logger slo.Logger, r CmdRunner, ns string, labelSelector string, podIndex int, containerIndex int, opts WaitOptions) error {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = DefaultRunner{}
	}
	opts = opts.withDefaults()

	waitCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	jsonpath := fmt.Sprintf(
		"{.items[%d].status.containerStatuses[%d].ready}",
		podIndex,
		containerIndex,
	)

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	tryOnce := func() (bool, error) {
		cmd := exec.Command(
			"kubectl", "get", "pods",
			"-n", ns,
			"-l", labelSelector,
			"-o", "jsonpath="+jsonpath,
		)
		out, err := r.Run(waitCtx, logger, cmd)
		if err != nil {
			return false, err
		}
		return strings.TrimSpace(out) == "true", nil
	}

	if ok, err := tryOnce(); err == nil && ok {
		return nil
	} else if err != nil {
		logger.Logf("wait pod ready: not ready yet: %v", err)
	}

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf(
				"timeout waiting pod ready (ns=%s selector=%q): %w",
				ns,
				labelSelector,
				waitCtx.Err(),
			)

		case <-ticker.C:
			ok, err := tryOnce()
			if err != nil {
				logger.Logf("wait pod ready: not ready yet: %v", err)
				continue
			}
			if ok {
				return nil
			}
		}
	}
}

// WaitServiceHasEndpoints waits until the Endpoints object has at least one address.
func WaitServiceHasEndpoints(ctx context.Context, logger slo.Logger, r CmdRunner, ns string, svc string, opts WaitOptions) error {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = DefaultRunner{}
	}
	opts = opts.withDefaults()

	waitCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	tryOnce := func() (bool, error) {
		cmd := exec.Command(
			"kubectl", "get", "endpoints", svc,
			"-n", ns,
			"-o", "jsonpath={.subsets[0].addresses[0].ip}",
		)
		out, err := r.Run(waitCtx, logger, cmd)
		if err != nil {
			return false, err
		}
		return strings.TrimSpace(out) != "", nil
	}

	if ok, err := tryOnce(); err == nil && ok {
		return nil
	} else if err != nil {
		logger.Logf("wait endpoints: not ready yet: %v", err)
	}

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf(
				"timeout waiting endpoints (ns=%s svc=%s): %w",
				ns,
				svc,
				waitCtx.Err(),
			)

		case <-ticker.C:
			ok, err := tryOnce()
			if err != nil {
				logger.Logf("wait endpoints: not ready yet: %v", err)
				continue
			}
			if ok {
				return nil
			}
		}
	}
}
