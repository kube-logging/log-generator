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

type LogGen interface {
	String() (string, float64)
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

func main() {
	metricsAddr := viper.GetString("metrics.addr")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool, 1)

	//ticker := time.NewTicker()

	go func() {
		log.Debugf("metrics listen on: %s", metricsAddr)
		http.Handle(viper.GetString("metrics.path"), promhttp.Handler())
		http.ListenAndServe( metricsAddr, nil)
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
