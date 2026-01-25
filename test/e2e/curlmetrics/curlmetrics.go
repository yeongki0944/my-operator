package curlmetrics

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yeongki/my-operator/pkg/kubeutil"
	"github.com/yeongki/my-operator/pkg/slo"
)

const PodLabelSelector = "app=curl-metrics"

// Client runs curl-metrics pods and fetches logs.
// It is test-oriented (uses kubectl + curlimages/curl).
type Client struct {
	Logger slo.Logger
	Runner kubeutil.CmdRunner

	// Tunables (optional)
	Image            string
	LabelSelector    string
	PodNamePrefix    string
	ServiceURLFormat string // e.g. "https://%s.%s.svc:8443/metrics"
}

// New creates a client with safe defaults.
// logger may be nil.
func New(logger slo.Logger, r kubeutil.CmdRunner) *Client {
	if r == nil {
		r = kubeutil.DefaultRunner{}
	}
	return &Client{
		Logger:           slo.NewLogger(logger),
		Runner:           r,
		Image:            "curlimages/curl:latest",
		LabelSelector:    PodLabelSelector,
		PodNamePrefix:    "curl-metrics",
		ServiceURLFormat: "https://%s.%s.svc:8443/metrics",
	}
}

// RunOnce creates a short-lived curl pod that scrapes /metrics.
// It returns the created pod name.
// It does NOT wait; call WaitDone then Logs.
func (c *Client) RunOnce(ctx context.Context, ns, token, metricsSvcName, serviceAccountName string) (string, error) {
	c.Logger = slo.NewLogger(c.Logger)
	if c.Runner == nil {
		c.Runner = kubeutil.DefaultRunner{}
	}

	// best-effort cleanup of previous curl-metrics pods
	_ = c.CleanupByLabel(ctx, ns)

	podName := fmt.Sprintf("%s-%d", c.PodNamePrefix, time.Now().UnixNano())
	metricsURL := fmt.Sprintf(c.ServiceURLFormat, metricsSvcName, ns)

	// keep -k for self-signed cert in test env, keep output clean (no -v)
	curlCmd := fmt.Sprintf(`set -euo pipefail;
curl -ksS --fail-with-body -H "Authorization: Bearer %s" "%s";`, token, metricsURL)

	cmd := exec.Command(
		"kubectl", "run", podName,
		"--restart=Never",
		"--namespace", ns,
		"--image", c.Image,
		"--labels", c.LabelSelector,
		"--overrides",
		fmt.Sprintf(`{
  "apiVersion":"v1",
  "kind":"Pod",
  "metadata":{
    "name":"%s",
    "namespace":"%s",
    "labels":{"app":"curl-metrics"}
  },
  "spec":{
    "serviceAccountName":"%s",
    "restartPolicy":"Never",
    "containers":[{
      "name":"curl",
      "image":"%s",
      "command":["/bin/sh","-c",%q],
      "securityContext":{
        "allowPrivilegeEscalation": false,
        "capabilities": { "drop": ["ALL"] },
        "runAsNonRoot": true,
        "runAsUser": 1000,
        "seccompProfile": { "type": "RuntimeDefault" }
      }
    }]
  }
}`, podName, ns, serviceAccountName, c.Image, curlCmd),
	)

	_, err := c.Runner.Run(ctx, c.Logger, cmd)
	return podName, err
}

// WaitDone waits until the curl pod reaches a terminal phase (Succeeded/Failed).
func (c *Client) WaitDone(ctx context.Context, ns, podName string, poll time.Duration) error {
	c.Logger = slo.NewLogger(c.Logger)
	if c.Runner == nil {
		c.Runner = kubeutil.DefaultRunner{}
	}
	if poll <= 0 {
		poll = 2 * time.Second
	}

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	// immediate first check
	if done, err := c.isTerminal(ctx, ns, podName); err != nil {
		return err
	} else if done {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := c.isTerminal(ctx, ns, podName)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}

// Logs returns kubectl logs of the given pod.
func (c *Client) Logs(ctx context.Context, ns, podName string) (string, error) {
	c.Logger = slo.NewLogger(c.Logger)
	if c.Runner == nil {
		c.Runner = kubeutil.DefaultRunner{}
	}

	cmd := exec.Command("kubectl", "logs", podName, "-n", ns)
	return c.Runner.Run(ctx, c.Logger, cmd)
}

// DeletePodNoWait deletes pod best-effort without waiting.
func (c *Client) DeletePodNoWait(ctx context.Context, ns, podName string) error {
	c.Logger = slo.NewLogger(c.Logger)
	if c.Runner == nil {
		c.Runner = kubeutil.DefaultRunner{}
	}

	cmd := exec.Command(
		"kubectl", "delete", "pod", podName,
		"-n", ns,
		"--ignore-not-found=true",
		"--wait=false",
	)
	_, err := c.Runner.Run(ctx, c.Logger, cmd)
	return err
}

// CleanupByLabel deletes all curl-metrics pods by label selector (best-effort, no wait).
func (c *Client) CleanupByLabel(ctx context.Context, ns string) error {
	c.Logger = slo.NewLogger(c.Logger)
	if c.Runner == nil {
		c.Runner = kubeutil.DefaultRunner{}
	}

	cmd := exec.Command(
		"kubectl", "delete", "pod",
		"-n", ns,
		"-l", c.LabelSelector,
		"--ignore-not-found=true",
		"--wait=false",
	)
	_, err := c.Runner.Run(ctx, c.Logger, cmd)
	// best-effort이라 여기서 에러를 hard fail로 만들지 않으려면 호출부에서 무시해도 됨.
	return err
}

func (c *Client) isTerminal(ctx context.Context, ns, podName string) (bool, error) {
	cmd := exec.Command(
		"kubectl", "get", "pod", podName,
		"-n", ns,
		"-o", "jsonpath={.status.phase}",
	)
	out, err := c.Runner.Run(ctx, c.Logger, cmd)
	if err != nil {
		return false, err
	}
	phase := strings.TrimSpace(out)
	return phase == "Succeeded" || phase == "Failed", nil
}
