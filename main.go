// Copyright (c) 2021 Cisco All Rights Reserved.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/banzaicloud/log-generator/conf"
	"github.com/banzaicloud/log-generator/loggen"
	"github.com/banzaicloud/log-generator/metrics"
	"github.com/banzaicloud/log-generator/stress"
	"github.com/gin-gonic/gin"
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
	Memory   stress.Memory  `json:"memory"`
	Cpu      stress.CPU     `json:"cpu"`
	LogLevel LogLevel       `json:"log_level"`
	Loggen   *loggen.LogGen `json:"loggen"`
}

func (s *State) logLevelGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.LogLevel)
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
		return

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
		s.Memory.CopyFrom(&t.Memory)
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

	var s State
	s.Loggen = loggen.New()
	s.LogLevel.Level = log.GetLevel().String()

	go func() {
		log.Debugf("api listen on: %s, basePath: %s", apiAddr, apiBasePath)
		r := gin.New()
		api := r.Group(apiBasePath)
		api.GET("metrics", metrics.Handler())
		api.GET("/", s.stateGetHandler)
		api.PATCH("/", s.statePatchHandler)
		api.GET("/loggen", s.Loggen.GetHandler)
		api.POST("/loggen", s.Loggen.PostHandler)
		api.GET("/loggen/formats", s.Loggen.FormatsGetHandler)
		api.GET("/memory", s.Memory.GetHandler)
		api.PATCH("/memory", s.Memory.PatchHandler)
		api.GET("/cpu", s.Cpu.GetHandler)
		api.PATCH("/cpu", s.Cpu.PatchHandler)
		api.GET("/log_level", s.logLevelGetHandler)
		api.PATCH("/log_level", s.logLevelPatchHandler)
		api.GET("/golang", s.Loggen.GolangGetHandler)
		api.PATCH("/golang", s.Loggen.GolangPatchHandler)
		api.GET("exceptions/go", exceptionsGoCall)
		api.PATCH("exceptions/go", exceptionsGoCall)

		r.Run(apiAddr)
		s.Memory.Wait()
	}()

	s.Loggen.Run()
}
