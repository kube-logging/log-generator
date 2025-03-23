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

	"github.com/gin-gonic/gin"
	"github.com/lthibault/jitterbug"
	logger "github.com/sirupsen/logrus"

	"github.com/kube-logging/log-generator/conf"
	"github.com/kube-logging/log-generator/formats"
	"github.com/kube-logging/log-generator/formats/golang"
	"github.com/kube-logging/log-generator/formats/web"
	"github.com/kube-logging/log-generator/log"
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

	m      sync.Mutex `json:"-"`
	writer LogWriter
}

type LogGenRequest struct {
	Type    string `json:"type"`
	Format  string `json:"format"`
	Count   int    `json:"count"`
	Framing bool   `json:"framing"`
}

func New() *LogGen {
	return &LogGen{
		EventPerSec:    conf.Viper.GetInt("message.event-per-sec"),
		BytePerSec:     conf.Viper.GetInt("message.byte-per-sec"),
		Randomise:      conf.Viper.GetBool("message.randomise"),
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

func (l *LogGen) FormatsGetHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, formats.FormatsByType())
}

func (l *LogGen) GetHandler(ctx *gin.Context) {
	l.m.Lock()
	defer l.m.Unlock()

	ctx.JSON(http.StatusOK, l)
}

func (l *LogGen) PostHandler(ctx *gin.Context) {
	var lr LogGenRequest
	if err := ctx.ShouldBindJSON(&lr); err != nil {
		logger.Error(err.Error())
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
	if _, exists := formats.FormatsByType()[lr.Type]; !exists {
		return fmt.Errorf("type %q does not exist", lr.Type)
	}

	return nil
}

func (lr *LogGenRequest) process(lg *LogGen) log.Log {
	if lr.Count <= 0 {
		return nil
	}

	// TODO configuration management for custom formats?
	if lr.Type == "golang" {
		return formats.NewGolangRandom(lg.GolangLog)
	}

	msg, err := formats.LogFactory(lr.Type, lr.Format, lg.Randomise)

	if err != nil {
		logger.Warnf("Error generating log from request %v, %v", lr, err)
		return nil
	}

	lr.Count--
	return msg
}

func tickerForByte(bandwith int, j jitterbug.Jitter) *jitterbug.Ticker {
	l, _ := formats.NewWeb("nginx", web.TemplateFS)
	_, length := l.String()
	events := float64(1) / (float64(length) / float64(bandwith))
	duration := float64(1000) / float64(events)
	return jitterbug.New(time.Duration(duration)*time.Millisecond, j)

}

func tickerForEvent(events int, j jitterbug.Jitter) *jitterbug.Ticker {
	duration := float64(1000) / float64(events)
	return jitterbug.New(time.Duration(duration)*time.Millisecond, j)
}

func (l *LogGen) GolangGetHandler(c *gin.Context) {
	c.JSON(http.StatusOK, l.GolangLog)
}

func (l *LogGen) GolangPatchHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&l.GolangLog); err != nil {
		logger.Error(err.Error())
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
		l.GolangLog.ErrorWeight = uintPtr(conf.Viper.GetUint("golang.weight.error"))
	}
	if l.GolangLog.WarningWeight == nil {
		l.GolangLog.WarningWeight = uintPtr(conf.Viper.GetUint("golang.weight.warning"))
	}
	if l.GolangLog.InfoWeight == nil {
		l.GolangLog.InfoWeight = uintPtr(conf.Viper.GetUint("golang.weight.info"))
	}
	if l.GolangLog.DebugWeight == nil {
		l.GolangLog.DebugWeight = uintPtr(conf.Viper.GetUint("golang.weight.debug"))
	}

	return nil
}

func (l *LogGen) processRequests() bool {
	l.m.Lock()
	logs := make([]log.Log, 0, l.ActiveRequests.Len())

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

		if request.Framing {
			msg.SetFramed(true)
		}

		logs = append(logs, msg)
		e = e.Next()
	}
	l.m.Unlock()

	if len(logs) == 0 {
		return false
	}

	for _, msg := range logs {
		l.writer.Send(msg)
	}

	return true
}

func (l *LogGen) Run() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan bool, 1)

	go func() {

		// TODO implement main loop for custom formats?

		var counter = 0
		var ticker *jitterbug.Ticker

		// jitter := &jitterbug.Norm{Stdev: time.Millisecond * 300}
		// TODO find a way to set Jitter from params
		jitter := &jitterbug.Norm{}

		if l.EventPerSec > 0 {
			ticker = tickerForEvent(l.EventPerSec, jitter)
		} else if l.BytePerSec > 0 {
			ticker = tickerForByte(l.BytePerSec, jitter)
		}
		count := conf.Viper.GetInt("message.count")

		if len(conf.Viper.GetString("destination.network")) != 0 {
			l.writer = newNetworkWriter(conf.Viper.GetString("destination.network"), conf.Viper.GetString("destination.address"))
		} else {
			l.writer = newStdoutWriter()
		}

		l.golangSet()

		for range ticker.C {
			if conf.Viper.GetBool("nginx.enabled") {
				l.sendIfCount(count, &counter, func() (log.Log, error) {
					if l.Randomise {
						return formats.NewRandomWeb("nginx", web.TemplateFS)
					} else {
						return formats.NewWeb("nginx", web.TemplateFS)
					}
				})
			}
			if conf.Viper.GetBool("apache.enabled") {
				l.sendIfCount(count, &counter, func() (log.Log, error) {
					if l.Randomise {
						return formats.NewRandomWeb("apache", web.TemplateFS)
					} else {
						return formats.NewWeb("apache", web.TemplateFS)
					}
				})
			}
			if conf.Viper.GetBool("golang.enabled") {
				l.sendIfCount(count, &counter, func() (log.Log, error) {
					return formats.NewGolangRandom(l.GolangLog), nil
				})
			}
			pendingRequests := l.processRequests()

			if !pendingRequests && count > 0 && !(counter < count) {
				done <- true
				break
			}
		}

		l.writer.Close()
	}()

	select {
	case <-interrupt:
		logger.Infoln("Recieved interrupt")
		break
	case <-done:
		break
	}
}

func (l *LogGen) sendIfCount(count int, counter *int, f func() (log.Log, error)) {
	if count == -1 || *counter < count {
		n, err := f()
		if err != nil {
			logger.Panic(err)
		}
		l.writer.Send(n)
		*counter++
	}
}
