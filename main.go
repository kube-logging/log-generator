// Copyright (c) 2021 Cisco All Rights Reserved.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/banzaicloud/log-generator/conf"
	"github.com/banzaicloud/log-generator/formats"
	"github.com/banzaicloud/log-generator/metrics"
	"github.com/banzaicloud/log-generator/stress"
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

type LogLevel struct {
	Level        string    `json:"level"`
	LastModified time.Time `json:"last_modified"`
}

type State struct {
	Memory    stress.Memory              `json:"memory"`
	Cpu       stress.CPU                 `json:"cpu"`
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

	if t.Memory != (stress.Memory{}) {
		s.Memory = t.Memory
		s.Memory.Stress()
	}

	if t.Cpu != (stress.CPU{}) {
		s.Cpu = t.Cpu
		s.Cpu.Stress()
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
		api.GET("state/memory", s.Memory.GetHandler)
		api.PATCH("state/memory", s.Memory.PatchHandler)
		api.GET("state/cpu", s.Cpu.GetHandler)
		api.PATCH("state/cpu", s.Cpu.PatchHandler)
		api.GET("state/log_level", s.logLevelGetHandler)
		api.PATCH("state/log_level", s.logLevelPatchHandler)
		api.GET("state/golang", s.golangGetHandler)
		api.PATCH("state/golang", s.golangPatchHandler)
		api.GET("exceptions/go", exceptionsGoCall)
		api.PATCH("exceptions/go", exceptionsGoCall)

		r.Run(apiAddr)
		s.Memory.Wait()
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
