// Copyright (c) 2023 Cisco All Rights Reserved.

package loggen

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/banzaicloud/log-generator/formats"
	"github.com/banzaicloud/log-generator/formats/golang"
	"github.com/banzaicloud/log-generator/metrics"
	"github.com/gin-gonic/gin"
	"github.com/lthibault/jitterbug"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type List struct {
	*list.List
}

type LogGen struct {
	EventPerSec    int                       `json:"event_per_sec"`
	BytePerSec     int                       `json:"byte_per_sec"`
	Randomise      bool                      `json:"randomise"`
	ActiveRequests List                      `json:"active_requests"`
	GolangLog      golang.GolangLogIntensity `json:"golang_log"`

	m sync.Mutex `json:"-"`
}

type LogGenRequest struct {
	Type   string `json:"type"`
	Format string `json:"format"`
	Count  int    `json:"count"`
}

func New() *LogGen {
	return &LogGen{
		EventPerSec:    viper.GetInt("message.event-per-sec"),
		BytePerSec:     viper.GetInt("message.byte-per-sec"),
		Randomise:      viper.GetBool("message.randomise"),
		ActiveRequests: List{list.New()},
	}
}

func (l *List) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString("[")

	for e := l.Front(); e != nil; e = e.Next() {
		m, err := json.Marshal(e.Value)

		if err != nil {
			return nil, err
		}

		b.WriteString(string(m))

		if e.Next() != nil {
			b.WriteRune(',')
		}
	}

	b.WriteString("]")
	return b.Bytes(), nil
}

func (l *LogGen) GetHandler(ctx *gin.Context) {
	l.m.Lock()
	defer l.m.Unlock()

	ctx.JSON(http.StatusOK, l)
}

func (l *LogGen) PostHandler(ctx *gin.Context) {
	var lr LogGenRequest
	if err := ctx.ShouldBindJSON(&lr); err != nil {
		log.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := lr.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l.m.Lock()
	defer l.m.Unlock()

	l.ActiveRequests.PushBack(&lr)

	ctx.JSON(http.StatusOK, lr)
}

func (lr *LogGenRequest) Validate() error {
	if _, exists := formats.Types[lr.Type]; !exists {
		return fmt.Errorf("type %q does not exist", lr.Type)
	}

	return nil
}

func (lr *LogGenRequest) process(l *LogGen) formats.Log {
	if lr.Count <= 0 {
		return nil
	}

	// TODO: factory
	var msg formats.Log
	var err error
	switch lr.Type {
	case "syslog":
		if l.Randomise {
			msg, err = formats.NewRandomSyslog(lr.Format)
		} else {
			msg, err = formats.NewSyslog(lr.Format)
		}
	case "web":
		if l.Randomise {
			msg, err = formats.NewRandomWeb(lr.Format)
		} else {
			msg, err = formats.NewWeb(lr.Format)
		}
	case "golang":
		msg = formats.NewGolangRandom(l.GolangLog)
	default:
		log.Panic("invalid LogGenRequest type")
	}

	if err != nil {
		log.Warnf("Error generating log from request %v, %v", lr, err)
		return nil
	}

	lr.Count--
	return msg
}

func tickerForByte(bandwith int, j jitterbug.Jitter) *jitterbug.Ticker {
	l, _ := formats.NewWeb("nginx")
	_, length := l.String()
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

func (l *LogGen) processRequests() bool {
	l.m.Lock()
	logs := make([]formats.Log, 0, l.ActiveRequests.Len())

	e := l.ActiveRequests.Front()
	for e != nil {
		request := e.Value.(*LogGenRequest)

		msg := request.process(l)
		if msg == nil {
			tmp := e
			e = e.Next()
			l.ActiveRequests.Remove(tmp)
			continue
		}

		logs = append(logs, msg)
		e = e.Next()
	}
	l.m.Unlock()

	if len(logs) == 0 {
		return false
	}

	for _, msg := range logs {
		emitMessage(msg)
	}

	return true
}

func (l *LogGen) Run() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool, 1)

	var counter = 0
	var ticker *jitterbug.Ticker

	//jitter := &jitterbug.Norm{Stdev: time.Millisecond * 300}
	// TODO find a way to set Jitter from params
	jitter := &jitterbug.Norm{}

	if l.EventPerSec > 0 {
		ticker = tickerForEvent(l.EventPerSec, jitter)
	} else if l.BytePerSec > 0 {
		ticker = tickerForByte(l.BytePerSec, jitter)
	}
	count := viper.GetInt("message.count")

	l.golangSet()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var n formats.Log
			var err error
			if counter < count && viper.GetBool("nginx.enabled") {
				if l.Randomise {
					n, err = formats.NewRandomWeb("nginx")
				} else {
					n, err = formats.NewWeb("nginx")
				}

				if err != nil {
					log.Panic(err)
				}
				emitMessage(n)
				counter++
			}
			if counter < count && viper.GetBool("apache.enabled") {
				if l.Randomise {
					n, err = formats.NewRandomWeb("apache")
				} else {
					n, err = formats.NewWeb("apache")
				}

				if err != nil {
					log.Panic(err)
				}
				emitMessage(n)
				counter++
			}
			if counter < count && viper.GetBool("golang.enabled") {
				n = formats.NewGolangRandom(l.GolangLog)
				emitMessage(n)
				counter++
			}

			pendingRequests := l.processRequests()

			// If we have count check counter
			if !pendingRequests && count > 0 && !(counter < count) {
				done <- true
			}
		case <-interrupt:
			log.Infoln("Recieved interrupt")
			return

		}
	}
}
