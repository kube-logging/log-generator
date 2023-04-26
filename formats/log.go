// Copyright (c) 2023 Cisco All Rights Reserved.

package formats

import "github.com/prometheus/client_golang/prometheus"

type Log interface {
	String() (string, float64)
	Labels() prometheus.Labels
}
