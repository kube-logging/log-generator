// Copyright (c) 2023 Cisco All Rights Reserved.

package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Startup time.Time

var (
	EventEmitted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "loggen_events_total",
		Help: "The total number of events",
	},
		[]string{"type", "severity"})

	EventEmittedBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "loggen_event_bytes_total",
		Help: "The total bytes of events",
	},
		[]string{"type", "severity"})
	GeneratedLoad = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "generated_load",
		Help: "Generated load",
	},
		[]string{"type"})

	Uptime = promauto.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "uptime_seconds",
			Help: "Generator uptime.",
		}, func() float64 {
			return time.Since(Startup).Seconds()
		},
	)
)

func Handler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
