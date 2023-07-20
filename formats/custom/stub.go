package custom

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Log interface {
	String() (string, float64)
	Labels() prometheus.Labels
}

// Formats returns supported template formats for a given type
func Formats() map[string][]string {
	return map[string][]string{}
}

// LogFactory creates log events for a log format and optionally randomises it
func LogFactory(logType string, format string, randomise bool) (Log, error) {
	return nil, nil
}
