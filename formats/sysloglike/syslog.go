// Copyright © 2026 Kube logging authors
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

package sysloglike

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/kube-logging/log-generator/log"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

//go:embed *.tmpl
var TemplateFS embed.FS

type Syslog struct {
	Facility    int
	severity    int
	dateTime    time.Time
	Host        string
	AppName     string
	PID         int
	Seq         int
	Msg         string
	statefulSeq *sync.Map
}

func (t Syslog) ISODateTime() string {
	return t.dateTime.Format(time.RFC3339)
}

func (t Syslog) ISODate() string {
	return strings.Split(t.ISODateTime(), "T")[0]
}

func (t Syslog) ISOTime() string {
	return strings.Split(t.ISODateTime(), "T")[1]
}

func (t Syslog) BSDDateTime() string {
	return t.dateTime.Format(time.Stamp)
}

func (t Syslog) CustomDateTime(format string) string {
	return t.dateTime.Format(format)
}

func (t Syslog) UnixTime() string {
	return strconv.FormatInt(t.dateTime.Unix(), 10)
}

func (t Syslog) UnixTimeMicro() string {
	return strconv.FormatInt(t.dateTime.UnixMicro(), 10)
}

func (t Syslog) UnixTimeMicroFraction() string {
	return fmt.Sprintf("%d.%06d", t.dateTime.Unix(), t.dateTime.UnixMicro()%1e6)
}

func (t Syslog) UnixTimeMilli() string {
	return strconv.FormatInt(t.dateTime.UnixMilli(), 10)
}

func (t Syslog) UnixTimeNano() string {
	return strconv.FormatInt(t.dateTime.UnixNano(), 10)
}

func (t Syslog) RFC3339NanoDateTime() string {
	return t.dateTime.Format(time.RFC3339Nano)
}

func (t Syslog) Pri() string {
	pri := t.Facility*8 + t.severity
	return fmt.Sprintf("<%d>", pri)
}

func (t Syslog) Severity() string {
	return strconv.FormatInt(int64(t.severity), 10)
}

func (t Syslog) MonotonSeq(name string, start uint64) uint64 {
	seq, ok := t.statefulSeq.Load(name)
	if !ok {
		t.statefulSeq.Store(name, start)
		return start
	}

	current := cast.ToUint64(seq) + 1
	t.statefulSeq.Store(name, current)
	return current
}

// Similar to MonotonSeq, but simulates a random amount of missing sequences controlled by "gap"
// with a frequency controlled by "rate".
// Example:
//
//	{{ .MonotonSeqGap "inc-gap" 2076326 10 5 }}
//
// The above will generate a sequence starting with 2076326 and
// add a gap between 0-4 randomly after every 10 items in average.
func (t Syslog) MonotonSeqGap(name string, start uint64, rate, gap int) uint64 {
	seq, ok := t.statefulSeq.Load(name)
	if !ok {
		t.statefulSeq.Store(name, start)
		return start
	}

	increment := 1
	if randomdata.Number(rate) == 0 {
		increment = increment + randomdata.Number(gap)
	}
	current := cast.ToUint64(seq) + cast.ToUint64(increment)
	t.statefulSeq.Store(name, current)
	return current
}

type RandomService struct {
	Hosts []string
	Apps  []string
}

func NewRandomService(maxHosts int, maxApps int, seed int64) *RandomService {
	svc := new(RandomService)

	// we might experience a race condition here, but we are dealing with pseudorandom data anyway
	if seed != 0 {
		randomdata.CustomRand(rand.New(rand.NewSource(seed)))
		defer randomdata.CustomRand(rand.New(rand.NewSource(time.Now().UnixNano())))
	}

	svc.Hosts = make([]string, maxHosts)
	for i := 0; i < maxHosts; i++ {
		svc.Hosts[i] = strings.ToLower(randomdata.SillyName())
	}

	svc.Apps = make([]string, maxApps)
	for i := 0; i < maxApps; i++ {
		svc.Apps[i] = strings.ToLower(randomdata.SillyName())
	}
	return svc
}

func (r *RandomService) RandomHost() string {
	if len(r.Hosts) == 0 {
		return "host"
	}
	return r.Hosts[randomdata.Number(0, len(r.Hosts))]
}

func (r *RandomService) RandomApp() string {
	if len(r.Apps) == 0 {
		return "app"
	}
	return r.Apps[randomdata.Number(0, len(r.Apps))]
}

func (r *RandomService) SampleData(sequence *sync.Map) Syslog {
	return Syslog{
		Facility:    20,
		severity:    5,
		dateTime:    time.Now(),
		Host:        viper.GetString("message.host"),
		AppName:     viper.GetString("message.appname"),
		PID:         1143,
		Seq:         1,
		Msg:         "An application event log entry...",
		statefulSeq: sequence,
	}
}

func (r *RandomService) RandomData(sequence *sync.Map) Syslog {
	return Syslog{
		Facility:    randomdata.Number(0, 24),
		severity:    randomdata.Number(0, 8),
		dateTime:    time.Now().UTC(),
		Host:        r.RandomHost(),
		AppName:     r.RandomApp(),
		PID:         randomdata.Number(1, 10000),
		Seq:         randomdata.Number(1, 10000),
		Msg:         fmt.Sprintf("An application event log entry %s %s", randomdata.Noun(), randomdata.Noun()),
		statefulSeq: sequence,
	}
}

var syslogRandomService *RandomService

func syslogRandom() *RandomService {
	if syslogRandomService == nil {
		syslogRandomService = NewRandomService(
			viper.GetInt("message.max-random-hosts"), viper.GetInt("message.max-random-apps"), viper.GetInt64("message.seed"))
	}
	return syslogRandomService
}

func NewSyslog(format string, sequence *sync.Map, templates []fs.FS) (*log.LogTemplate, error) {
	return firstExistingTemplate(format, templates, syslogRandom().SampleData(sequence))
}

func NewRandomSyslog(format string, sequence *sync.Map, templates []fs.FS) (*log.LogTemplate, error) {
	return firstExistingTemplate(format, templates, syslogRandom().RandomData(sequence))
}

func firstExistingTemplate(format string, templates []fs.FS, data log.LogTemplateData) (tpl *log.LogTemplate, err error) {
	var allErr error
	for _, f := range templates {
		tpl, err = log.NewLogTemplate(format, f, data)
		if err == nil {
			return
		} else {
			allErr = errors.Join(allErr, err)
		}
	}
	return tpl, allErr
}
