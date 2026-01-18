package promtext

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseTextToMap parses Prometheus exposition format (text) into a flat map.
// Key format example:
//
//	metric_name{a="b",c="d"}
//
// If no labels:
//
//	metric_name
//
// v1: minimal parser for common cases (counters/gauges).
func ParseTextToMap(r io.Reader) (map[string]float64, error) {
	out := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// split "key value"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := fields[0]
		valStr := fields[1]
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parse float: %q: %w", line, err)
		}
		out[key] = v
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
