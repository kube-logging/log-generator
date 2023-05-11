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

package stress

import (
	"net/http"
	"time"

	"github.com/dhoomakethu/stress/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/kube-logging/log-generator/metrics"
)

type CPU struct {
	Load         float64       `json:"load"`
	Duration     time.Duration `json:"duration"`
	Active       time.Time     `json:"active"`
	Core         float64       `json:"core"`
	LastModified time.Time     `json:"last_modified"`
}

func (c *CPU) GetHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c)
}

func (c *CPU) PatchHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindJSON(c); err != nil {
		log.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := c.Stress()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return

	}
	ctx.JSON(http.StatusOK, c)
}

func (c *CPU) Stress() error {
	c.LastModified = time.Now()
	c.Active = time.Now().Add(c.Duration * time.Second)
	go c.cpuLoad()
	return nil
}

func (c CPU) cpuLoad() {
	log.Debugf("CPU load test started, duration: %s", (c.Duration * time.Second).String())
	sampleInterval := 100 * time.Millisecond
	controller := utils.NewCpuLoadController(sampleInterval, c.Load)
	monitor := utils.NewCpuLoadMonitor(c.Core, sampleInterval)
	actuator := utils.NewCpuLoadGenerator(controller, monitor, c.Duration)
	metrics.GeneratedLoad.WithLabelValues("cpu").Add(float64(c.Load))
	utils.RunCpuLoader(actuator)
	metrics.GeneratedLoad.DeleteLabelValues("cpu")
	log.Debugln("CPU load test done.")
}
