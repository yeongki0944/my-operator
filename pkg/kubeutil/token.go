package kubeutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yeongki/my-operator/pkg/slo"
)

type tokenRequest struct {
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}

const tokenRequestBody = `{"apiVersion":"authentication.k8s.io/v1","kind":"TokenRequest"}`

// TODO(kubeutil): When we add TokenRequest options (audiences/expirationSeconds/etc),
// stop using a raw JSON string and marshal a struct instead.
// Rationale: easier to extend safely (optional fields), avoids fragile string edits,
// and produces correct JSON consistently.
// {
//   "apiVersion": "authentication.k8s.io/v1",
//   "kind": "TokenRequest",
//   "spec": {
//     "expirationSeconds": 3600,
//     "audiences": ["api"]
//   }
// }

// ServiceAccountToken requests a token for the given ServiceAccount.
// - Retries until ctx is done.
// - logger may be nil (no-op).
// TODO(refactor): Refactor manual retry loop to use a shared 'Poll' utility.
// Currently, this function implements a custom loop/ticker to handle eventual consistency.
// We should standardize this pattern by implementing a helper (e.g., PollImmediate in pkg/kubeutil/wait.go)
// to improve stability and code reuse across the project.
// context 가 잘 넘어간다, func(ctx context.Context) (string, error => 이것과 동일한 형태이다, Closure 가 선언될 당시의 변수를 캡쳐해서 사용하기 때문에 가능하다.
func ServiceAccountToken(ctx context.Context, logger slo.Logger, r CmdRunner, ns, sa string) (string, error) {
	logger = slo.NewLogger(logger)
	if r == nil {
		r = DefaultRunner{}
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	var lastErr error
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	tryOnce := func() (string, error) {
		cmd := exec.Command("kubectl", "create", "--raw",
			fmt.Sprintf("/api/v1/namespaces/%s/serviceaccounts/%s/token", ns, sa),
			"-f", "-",
		)
		cmd.Stdin = strings.NewReader(tokenRequestBody)
		// ctx 반영, Closure 캡처
		stdout, err := r.Run(ctx, logger, cmd)
		if err != nil {
			return "", fmt.Errorf("token request failed (ns=%s sa=%s): %w", ns, sa, err)
		}

		var tr tokenRequest
		if err := json.Unmarshal([]byte(stdout), &tr); err != nil {
			return "", fmt.Errorf("token response json parse failed: %w (body=%q)", err, stdout)
		}
		if tr.Status.Token == "" {
			return "", fmt.Errorf("token is empty")
		}
		return tr.Status.Token, nil
	}

	if tok, err := tryOnce(); err == nil {
		return tok, nil
	} else {
		lastErr = err
		logger.Logf("token not ready yet: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			if lastErr == nil {
				lastErr = ctx.Err()
			}
			return "", lastErr
		case <-ticker.C:
			tok, err := tryOnce()
			if err == nil {
				return tok, nil
			}
			lastErr = err
			logger.Logf("token not ready yet: %v", err)
		}
	}
}
