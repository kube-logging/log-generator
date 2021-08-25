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
	"github.com/dhoomakethu/stress/utils"
	"github.com/gin-gonic/gin"
	"github.com/lthibault/jitterbug"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	conf.Init()
}

type Memory struct {
	Percentage   int64         `json:"percentage"`
	Active       time.Time     `json:"active"`
	Duration     time.Duration `json:"duration"`
	LastModified time.Time     `json:"last_modified"`
}

type Cpu struct {
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
	Memory   Memory   `json:"memory"`
	Cpu      Cpu      `json:"cpu"`
	LogLevel LogLevel `json:"log_level"`
}

type LogGen interface {
	String() (string, float64)
}

func promHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

var (
	eventEmitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loggen_events_total",
		Help: "The total number of events",
	})
	eventEmittedBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "loggen_events_total_bytes",
		Help: "The total number of events",
	})
)

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
	eventEmitted.Inc()
	eventEmittedBytes.Add(size)
}

func (s *State) cpuGet(c *gin.Context) {
	c.JSON(http.StatusOK, s.Cpu)
}
func (s *State) cpuSet(c *gin.Context) {
	if err := c.ShouldBindJSON(s); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.Cpu.LastModified = time.Now()
	s.Cpu.Active = time.Now().Add(s.Cpu.Duration * time.Second)
	go func() {
		log.Debugf("CPU load test started, duration: %s", s.Cpu.Duration.String())
		sampleInterval := 100 * time.Millisecond
		controller := utils.NewCpuLoadController(sampleInterval, s.Cpu.Load)
		monitor := utils.NewCpuLoadMonitor(s.Cpu.Core, sampleInterval)
		actuator := utils.NewCpuLoadGenerator(controller, monitor, s.Cpu.Duration)
		utils.RunCpuLoader(actuator)
		log.Debugln("CPU load test done.")
	}()
	c.JSON(http.StatusOK, s.Cpu)
}

func (s *State) memoryGet(c *gin.Context) {
	c.JSON(http.StatusOK, s.Memory)
}

func (s *State) memorySet(c *gin.Context) {
	if err := c.ShouldBindJSON(s); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.Memory.LastModified = time.Now()
	s.Memory.Active = time.Now().Add(s.Memory.Duration * time.Second)
	go func() {
		s := make([][]byte, 1000)
		mb := 1024 * 1024
		for i := 0; i < len(s); i++ {
			s[i] = make([]byte, mb)
		}
		log.Debugln("MEM load test done.")
	}()

	c.JSON(http.StatusOK, s.Memory)
}

func (s *State) LogLevelGet(c *gin.Context) {
	s.LogLevel.Level = log.GetLevel().String()
	c.JSON(http.StatusOK, s.LogLevel)
}

func (s *State) LogLevelSet(c *gin.Context) {
	if err := c.ShouldBindJSON(s); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	level, lErr := log.ParseLevel(s.LogLevel.Level)
	if lErr != nil {
		log.Error(lErr.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s valid logLeveles: panic fatal error warn warning info debug trace", lErr.Error())})
		return
	}
	log.SetLevel(level)
	s.LogLevel.LastModified = time.Now()
	c.JSON(http.StatusOK, s.LogLevel)
}

func (s *State) statusGet(c *gin.Context) {
	c.JSON(http.StatusOK, s)
}

func exceptionsGoCall(c *gin.Context) {
	log.Infoln("exceptionsGo")
	c.String(http.StatusOK, "exceptionsGo")
}

func main() {
	metricsAddr := viper.GetString("metrics.addr")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool, 1)

	//ticker := time.NewTicker()
	var s State

	go func() {
		log.Debugf("metrics listen on: %s", metricsAddr)
		r := gin.New()

		r.GET("/", func(c *gin.Context) {
			c.JSON(200, "/ - OK!")
		})
		r.GET(viper.GetString("metrics.path"), promHandler())
		r.GET("memory", s.memoryGet)
		r.PATCH("memory", s.memorySet)
		r.GET("cpu", s.cpuGet)
		r.PATCH("cpu", s.cpuSet)
		r.GET("log_level", s.LogLevelGet)
		r.PATCH("log_level", s.LogLevelSet)
		r.GET("status", s.statusGet)
		r.GET("exceptions/go", exceptionsGoCall)
		r.PATCH("exceptions/go", exceptionsGoCall)

		r.Run(metricsAddr)
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
