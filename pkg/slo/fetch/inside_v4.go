package fetch

import "context"

// InsideSnapshotFetchV4 fetches metrics using the inside CurlPod-only boundary.
func InsideSnapshotFetchV4(ctx context.Context, fetchFunc func(context.Context) (string, error)) (string, []string) {
	body, err := fetchFunc(ctx)
	if err != nil {
		return "", []string{err.Error()}
	}
	return body, nil
}
