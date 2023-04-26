// Copyright (c) 2023 Cisco All Rights Reserved.

package loggen

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/banzaicloud/log-generator/formats"
	"github.com/banzaicloud/log-generator/formats/golang"
	"github.com/banzaicloud/log-generator/formats/web"
	"github.com/banzaicloud/log-generator/metrics"
	"github.com/gin-gonic/gin"
	"github.com/lthibault/jitterbug"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LogGen struct {
	GolangLog golang.GolangLogIntensity `json:"golang_log"`
}

func tickerForByte(bandwith int, j jitterbug.Jitter) *jitterbug.Ticker {
	_, length := web.NewNginxLog().String()
	events := float64(1) / (float64(length) / float64(bandwith))
	duration := float64(1000) / float64(events)
	return jitterbug.New(time.Duration(duration)*time.Millisecond, j)

}

func tickerForEvent(events int, j jitterbug.Jitter) *jitterbug.Ticker {
	duration := float64(1000) / float64(events)
	return jitterbug.New(time.Duration(duration)*time.Millisecond, j)
}

func emitMessage(gen formats.Log) {
	msg, size := gen.String()
	fmt.Println(msg)
	metrics.EventEmitted.With(gen.Labels()).Inc()
	metrics.EventEmittedBytes.With(gen.Labels()).Add(size)
}

func (l *LogGen) GolangGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, l.GolangLog)
}

func (l *LogGen) GolangPatchHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&l.GolangLog); err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l.golangSet()
	c.JSON(http.StatusOK, l.GolangLog)
}

func uintPtr(u uint) *uint {
	return &u
}

func (l *LogGen) golangSet() error {
	if l.GolangLog.ErrorWeight == nil {
		l.GolangLog.ErrorWeight = uintPtr(viper.GetUint("golang.weight.error"))
	}
	if l.GolangLog.WarningWeight == nil {
		l.GolangLog.WarningWeight = uintPtr(viper.GetUint("golang.weight.warning"))
	}
	if l.GolangLog.InfoWeight == nil {
		l.GolangLog.InfoWeight = uintPtr(viper.GetUint("golang.weight.info"))
	}
	if l.GolangLog.DebugWeight == nil {
		l.GolangLog.DebugWeight = uintPtr(viper.GetUint("golang.weight.debug"))
	}

	return nil
}

func (l *LogGen) Run() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool, 1)

	var counter = 0
	// Init ticker
	var ticker *jitterbug.Ticker

	//jitter := &jitterbug.Norm{Stdev: time.Millisecond * 300}
	// TODO find a way to set Jitter from params
	jitter := &jitterbug.Norm{}

	eventPerSec := viper.GetInt("message.event-per-sec")
	bytePerSec := viper.GetInt("byte-per-sec")

	if eventPerSec > 0 {
		ticker = tickerForEvent(eventPerSec, jitter)
	} else if bytePerSec > 0 {
		ticker = tickerForByte(bytePerSec, jitter)
	}
	count := viper.GetInt("message.count")

	l.golangSet()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var n formats.Log
			if viper.GetBool("nginx.enabled") {
				if viper.GetBool("message.randomise") {
					n = web.NewNginxLogRandom()
				} else {
					n = web.NewNginxLog()
				}
				emitMessage(n)
				counter++
			}
			if viper.GetBool("golang.enabled") {
				n = golang.NewGolangLogRandom(l.GolangLog)
				emitMessage(n)
				counter++
			}
			if viper.GetBool("apache.enabled") {
				if viper.GetBool("message.randomise") {
					n = web.NewApacheLogRandom()
				} else {
					n = web.NewApacheLog()
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
