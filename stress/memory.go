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
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/kube-logging/log-generator/metrics"
)

type Memory struct {
	Megabyte     int64          `json:"megabyte"`
	Active       time.Time      `json:"active"`
	Duration     time.Duration  `json:"duration"`
	LastModified time.Time      `json:"last_modified"`
	mutex        sync.Mutex     `json:"-"`
	wg           sync.WaitGroup `json:"-"`
}

func (m *Memory) CopyFrom(s *Memory) {
	m.Megabyte = s.Megabyte
	m.Active = s.Active
	m.Duration = s.Duration
	m.LastModified = s.LastModified
}

func (m *Memory) GetHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, m)
}

func (m *Memory) PatchHandler(ctx *gin.Context) {
	if err := ctx.ShouldBindJSON(m); err != nil {
		log.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := m.Stress()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return

	}
	ctx.JSON(http.StatusOK, m)
}

func (m *Memory) Stress() error {
	m.LastModified = time.Now()
	m.Active = time.Now().Add(m.Duration * time.Second)
	go m.memoryBallast()
	return nil
}

func (m *Memory) Wait() {
	m.wg.Wait()
}

func (m *Memory) memoryBallast() {
	m.wg.Add(1)
	defer m.wg.Done()
	m.mutex.Lock()
	defer m.mutex.Unlock()

	log.Debugf("MEM load test started. - %s", time.Now().String())
	ballast := make([]byte, m.Megabyte<<20)
	metrics.GeneratedLoad.WithLabelValues("memory").Add(float64(m.Megabyte))
	for i := 0; i < len(ballast); i++ {
		ballast[i] = byte('A')
	}
	<-time.After(m.Duration * time.Second)
	ballast = nil
	runtime.GC()
	metrics.GeneratedLoad.DeleteLabelValues("memory")
	log.Debugf("MEM load test done.- %s", time.Now().String())
}
