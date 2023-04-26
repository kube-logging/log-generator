// Copyright (c) 2023 Cisco All Rights Reserved.

package stress

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/banzaicloud/log-generator/metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
