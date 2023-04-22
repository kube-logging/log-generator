// Copyright (c) 2021 Cisco All Rights Reserved.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/banzaicloud/log-generator/conf"
	"github.com/banzaicloud/log-generator/formats"
	"github.com/banzaicloud/log-generator/metrics"
	"github.com/dhoomakethu/stress/utils"
	"github.com/gin-gonic/gin"
	"github.com/lthibault/jitterbug"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	conf.Init()
}

type Memory struct {
	Megabyte     int64          `json:"megabyte"`
	Active       time.Time      `json:"active"`
	Duration     time.Duration  `json:"duration"`
	LastModified time.Time      `json:"last_modified"`
	mutex        sync.Mutex     `json:"_"`
	wg           sync.WaitGroup `json:"_"`
}

type CPU struct {
	Load         float64       `json:"load"`
	Duration     time.Duration `json:"duration"`
	Active       time.Time     `json:"active"`
	Core         float64       `json:"core"`
	LastModified time.Time     `json:"last_modified"`
}

type LogLevel struct {
	Level        string    `json:"level"`
	LastModified time.Time `json:"last_modified"`
}

type State struct {
	Memory    Memory                     `json:"memory"`
	Cpu       CPU                        `json:"cpu"`
	LogLevel  LogLevel                   `json:"log_level"`
	GolangLog formats.GolangLogIntensity `json:"golang_log"`
}

type LogGen interface {
	String() (string, float64)
	Labels() prometheus.Labels
}

func TickerForByte(bandwith int, j jitterbug.Jitter) *jitterbug.Ticker {
	_, length := formats.NewNginxLog().String()
	events := float64(1) / (float64(length) / float64(bandwith))
	duration := float64(1000) / float64(events)
	return jitterbug.New(time.Duration(duration)*time.Millisecond, j)

}

func TickerForEvent(events int, j jitterbug.Jitter) *jitterbug.Ticker {
	duration := float64(1000) / float64(events)
	return jitterbug.New(time.Duration(duration)*time.Millisecond, j)
}

func emitMessage(gen LogGen) {
	msg, size := gen.String()
	fmt.Println(msg)
	metrics.EventEmitted.With(gen.Labels()).Inc()
	metrics.EventEmittedBytes.With(gen.Labels()).Add(size)
}

func (s *State) cpuGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.Cpu)
}

func (s *State) cpuPatchHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&s.Cpu); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.cpuSet()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	c.JSON(http.StatusOK, s.Cpu)
}

func (s *State) cpuSet() error {
	s.Cpu.LastModified = time.Now()
	s.Cpu.Active = time.Now().Add(s.Cpu.Duration * time.Second)
	go s.Cpu.cpuLoad()
	return nil
}

func (c CPU) cpuLoad() {
	log.Debugf("CPU load test started, duration: %s", c.Duration.String())
	sampleInterval := 100 * time.Millisecond
	controller := utils.NewCpuLoadController(sampleInterval, c.Load)
	monitor := utils.NewCpuLoadMonitor(c.Core, sampleInterval)
	actuator := utils.NewCpuLoadGenerator(controller, monitor, c.Duration)
	metrics.GeneratedLoad.WithLabelValues("cpu").Add(float64(c.Load))
	utils.RunCpuLoader(actuator)
	metrics.GeneratedLoad.DeleteLabelValues("cpu")
	log.Debugln("CPU load test done.")
}

func (s *State) memoryGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.Memory)
}

func (s *State) memoryPatchHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&s.Memory); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.memorySet()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	c.JSON(http.StatusOK, s.Memory)
}

func (s *State) memorySet() error {
	s.Memory.LastModified = time.Now()
	s.Memory.Active = time.Now().Add(s.Memory.Duration * time.Second)
	go s.Memory.memoryBallast()
	return nil
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

func (s *State) logLevelGetHandler(c *gin.Context) {
	s.LogLevel.Level = log.GetLevel().String()
	c.JSON(http.StatusOK, s.LogLevel)
}

func (s *State) golangGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.GolangLog)
}

func (s *State) golangPatchHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&s.GolangLog); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.golangSet()
	c.JSON(http.StatusOK, s.GolangLog)
}

func uintPtr(u uint) *uint {
	return &u
}

func (s *State) golangSet() error {
	if s.GolangLog.ErrorWeight == nil {
		s.GolangLog.ErrorWeight = uintPtr(viper.GetUint("golang.weight.error"))
	}
	if s.GolangLog.WarningWeight == nil {
		s.GolangLog.WarningWeight = uintPtr(viper.GetUint("golang.weight.warning"))
	}
	if s.GolangLog.InfoWeight == nil {
		s.GolangLog.InfoWeight = uintPtr(viper.GetUint("golang.weight.info"))
	}
	if s.GolangLog.DebugWeight == nil {
		s.GolangLog.DebugWeight = uintPtr(viper.GetUint("golang.weight.debug"))
	}

	return nil
}

func (s *State) logLevelSet() error {
	level, lErr := log.ParseLevel(s.LogLevel.Level)
	if lErr != nil {
		err := fmt.Errorf("%s valid logLeveles: panic fatal error warn warning info debug trace", lErr.Error())
		log.Error(err)
		return err
	}
	log.SetLevel(level)
	log.Infof("New loglevel: %s", level)
	s.LogLevel.LastModified = time.Now()
	return nil
}

func (s *State) logLevelPatchHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&s.LogLevel); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.logLevelSet()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	c.JSON(http.StatusOK, s.LogLevel)
}

func (s *State) stateGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s)
}

func (s *State) statePatchHandler(c *gin.Context) {

	var t State
	if err := c.ShouldBindJSON(&t); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if t.Memory != (Memory{}) {
		s.Memory = t.Memory
		s.memorySet()
	}

	if t.Cpu != (CPU{}) {
		s.Cpu = t.Cpu
		s.cpuSet()
	}

	if t.LogLevel != (LogLevel{}) {
		s.LogLevel = t.LogLevel
		s.logLevelSet()
	}
	c.JSON(http.StatusOK, s)
}

func exceptionsGoCall(c *gin.Context) {
	log.Infoln("exceptionsGo")
	c.String(http.StatusOK, "exceptionsGo")
}

func main() {
	metrics.Startup = time.Now()

	apiAddr := viper.GetString("api.addr")
	apiBasePath := viper.GetString("api.basePath")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool, 1)

	//ticker := time.NewTicker()
	var s State

	go func() {
		log.Debugf("api listen on: %s, basePath: %s", apiAddr, apiBasePath)
		r := gin.New()
		api := r.Group(apiBasePath)
		api.GET("/", func(c *gin.Context) {
			c.JSON(200, "/ - OK!")
		})
		api.GET("metrics", metrics.Handler())
		api.GET("state", s.stateGetHandler)
		api.PATCH("state", s.statePatchHandler)
		api.GET("state/memory", s.memoryGetHandler)
		api.PATCH("state/memory", s.memoryPatchHandler)
		api.GET("state/cpu", s.cpuGetHandler)
		api.PATCH("state/cpu", s.cpuPatchHandler)
		api.GET("state/log", s.logLevelGetHandler)
		api.PATCH("state/log", s.logLevelPatchHandler)
		api.GET("state/golang", s.golangGetHandler)
		api.PATCH("state/golang", s.golangPatchHandler)
		api.GET("exceptions/go", exceptionsGoCall)
		api.PATCH("exceptions/go", exceptionsGoCall)

		r.Run(apiAddr)
		s.Memory.wg.Wait()
	}()

	var counter = 0
	// Init ticker
	var ticker *jitterbug.Ticker
	var jitter jitterbug.Jitter

	//jitter = &jitterbug.Norm{Stdev: time.Millisecond * 300}
	// TODO find a way to set Jitter from params
	jitter = &jitterbug.Norm{}

	eventPerSec := viper.GetInt("message.event-per-sec")
	bytePerSec := viper.GetInt("byte-per-sec")

	if eventPerSec > 0 {
		ticker = TickerForEvent(eventPerSec, jitter)
	} else if bytePerSec > 0 {
		ticker = TickerForByte(bytePerSec, jitter)
	}
	count := viper.GetInt("message.count")
	s.golangSet()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var n LogGen
			if viper.GetBool("nginx.enabled") {
				if viper.GetBool("message.randomise") {
					n = formats.NewNginxLogRandom()
				} else {
					n = formats.NewNginxLog()
				}
				emitMessage(n)
				counter++
			}
			if viper.GetBool("golang.enabled") {
				n = formats.NewGolangLogRandom(s.GolangLog)
				emitMessage(n)
				counter++
			}
			if viper.GetBool("apache.enabled") {
				if viper.GetBool("message.randomise") {
					n = formats.NewApacheLogRandom()
				} else {
					n = formats.NewApacheLog()
				}
				emitMessage(n)
				counter++
			}

			// If we have count check counter
			if count > 0 && !(counter < count) {
				done <- true
			}
		case <-interrupt:
			log.Infoln("Recieved interrupt")
			return

		}
	}
}
