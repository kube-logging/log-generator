package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/lthibault/jitterbug"
	wr "github.com/mroth/weightedrand"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const NginxTimeFormat = "02/Jan/2006:15:04:05 -0700"

type LogGen interface {
	String() (string, float64)
}

type NginxLog struct {
	Remote            string
	Host              string
	User              string
	Time              time.Time
	Method            string
	Path              string
	Code              int
	Size              int
	Referer           string
	Agent             string
	HttpXForwardedFor string
}

func NewNginxLog() NginxLog {
	return NginxLog{
		Remote:            "127.0.0.1",
		Host:              "-",
		User:              "-",
		Time:              time.Now(),
		Method:            "GET",
		Path:              "/loggen/loggen/loggen/loggen/loggen/loggen/loggen",
		Code:              200,
		Size:              650,
		Referer:           "-",
		Agent:             "golang/generator PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP",
		HttpXForwardedFor: "-",
	}
}

func NewNginxLogRandom() NginxLog {
	rand.Seed(time.Now().UTC().UnixNano())
	c := wr.NewChooser(
		wr.Choice{Item: 200, Weight: 7},
		wr.Choice{Item: 404, Weight: 3},
		wr.Choice{Item: 503, Weight: 1},
		wr.Choice{Item: 302, Weight: 2},
		wr.Choice{Item: 403, Weight: 2},
	)
	return NginxLog{
		Remote:            randomdata.IpV4Address(),
		Host:              "-",
		User:              "-",
		Time:              time.Now(),
		Method:            randomdata.StringSample("GET", "POST", "PUT"),
		Path:              randomdata.StringSample("/", "/blog", "/index.html", "/products"),
		Code:              c.Pick().(int),
		Size:              rand.Intn(25000-100) + 100,
		Referer:           "-",
		Agent:             randomdata.UserAgentString(),
		HttpXForwardedFor: "-",
	}
}

func (n NginxLog) String() (string, float64) {
	message := fmt.Sprintf("%s %s %s [%s] \"%s %s HTTP/1.1\" %d %d %q %q %q", n.Remote, n.Host, n.User, n.Time.Format(NginxTimeFormat), n.Method, n.Path, n.Code, n.Size, n.Referer, n.Agent, n.HttpXForwardedFor)
	return message, float64(len([]byte(message)))
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
	_, length := NewNginxLog().String()
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
	//minIntervall := flag.String("min-intervall", "100ms", "Minimum interval between log messages")
	//maxIntervall := flag.String("max-intervall", "1s", "Maximum interval between log messages")
	count := flag.Int("count", 10, "The amount of log message to emit.")
	randomise := flag.Bool("randomise", true, "Randomise log content")
	eventPerSec := flag.Int("event-per-sec", 2, "The amount of log message to emit/s")
	bytePerSec := flag.Int("byte-per-sec", 200, "The amount of bytes to emit/s")
	metricsAddr := flag.String("metrics.addr", ":11000", "Metrics server listen address")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool, 1)

	//ticker := time.NewTicker()

	go func() {
		fmt.Println("metrics listen on: ", *metricsAddr)
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(*metricsAddr, nil)
	}()

	var counter = 0
	// Init ticker
	var ticker *jitterbug.Ticker
	var jitter jitterbug.Jitter

	//jitter = &jitterbug.Norm{Stdev: time.Millisecond * 300}
	// TODO find a way to set Jitter from params
	jitter = &jitterbug.Norm{}

	if *eventPerSec > 0 {
		ticker = TickerForEvent(*eventPerSec, jitter)
	} else if *bytePerSec > 0 {
		ticker = TickerForByte(*bytePerSec, jitter)
	}

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var n LogGen
			if *randomise {
				n = NewNginxLogRandom()
			} else {
				n = NewNginxLog()
			}
			emitMessage(n)
			counter++
			// If we have count check counter
			if *count > 0 && !(counter < *count) {
				done <- true
			}
		case <-interrupt:
			fmt.Println("Recieved interrupt")
			return

		}
	}
}
