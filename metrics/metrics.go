// Copyright © 2023 Kube logging authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License."""

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
