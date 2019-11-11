package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Pallinder/go-randomdata"
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
		Time:              time.Time{},
		Method:            "GET",
		Path:              "/loggen",
		Code:              200,
		Size:              650,
		Referer:           "-",
		Agent:             "golang/generator",
		HttpXForwardedFor: "-",
	}
}

func NewNginxLogRandom() NginxLog {
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
		Name: "event_emitted_total",
		Help: "The total number of events",
	})
	eventEmittedBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "event_emitted_total_bytes",
		Help: "The total number of events",
	})
	pushErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "push_errors",
		Help: "The total number of Loki push errors",
	})
	pushLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "push_latency",
		Help: "Latency of pushing a batch to Loki",
	})
)

type Entry struct {
	Ts   time.Time
	Line string
}

func send(url string, randomLabels int, entriesPerStream int, host string, debug bool, entries []Entry) {
	type stream struct {
		Labels  string
		Entries []Entry
	}

	payload := struct{ Streams []stream }{Streams: make([]stream, len(entries)/entriesPerStream)}
	for i := range payload.Streams {
		n := rand.Intn(randomLabels)
		payload.Streams[i].Labels = fmt.Sprintf(`{application="my-test-application", type="events", fake="n%d", pod=%q}`, randomLabels*i+n, host)
	}

	for _, entry := range entries {
		stream := rand.Intn(len(payload.Streams))
		payload.Streams[stream].Entries = append(payload.Streams[stream].Entries, entry)
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	resp, err := http.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		fmt.Printf("failed to send %d entries: %v\n", len(entries), err)
		return
	}

	if resp.StatusCode >= 300 {
		debug = true
		pushErrors.Inc()
	} else {
		pushLatency.Observe(time.Since(start).Seconds())
	}

	if debug {
		buf, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("sent %d entries. response %s: %s\n", len(entries), resp.Status, buf)
	}
}

//labels{\"streams\": [{ \"labels\": \"{application=\\\"my-test-application\\\", type=\\\"events\\\"}\", \"entries\": [{ \"ts\": \"${NOW}\", \"line\": \"${LINE}\" }] }]}"

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	minIntervall := flag.String("min-intervall", "100ms", "Minimum interval between log messages")
	maxIntervall := flag.String("max-intervall", "1s", "Maximum interval between log messages")
	count := flag.Int("count", 0, "The amount of log message to emit.")
	randomise := flag.Bool("randomise", true, "Randomise log content")
	eventPerSec := flag.Int("event-per-sec", 10, "The amount of log message to emit/s")
	metricsAddr := flag.String("metrics.addr", ":11000", "Metrics server listen address")
	metrics := flag.Bool("metrics", true, "Provide metrics endpoint")
	lokiURL := flag.String("loki-url", "", "Loki endpoint to send logs to instead of stdout. Example: http://loki:password@localhost:8123/api/prom/push")
	lokiBatch := flag.Int("loki-batch", 100, "Batch size of loki push requests")
	lokiRandomLabel := flag.Int("loki-random", 10, "How much different 'fake' labels should be used as a label")
	debug := flag.Bool("debug", false, "Verbose output")
	entriesPerStream := flag.Int("entries-per-stream", 10, "Average number of entries in each stream within a batch. (Implies number of streams per batch.)")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool)
	messages := make(chan string, 100)
	host, _ := os.Hostname()

	go func() {
		if *lokiURL == "" {
			for msg := range messages {
				fmt.Println(msg)
			}
		} else {
			queue := make([]Entry, 0, *lokiBatch)
			for msg := range messages {
				entry := Entry{
					Ts:   time.Now(),
					Line: msg,
				}
				queue = append(queue, entry)
				if len(queue) >= *lokiBatch {
					send(*lokiURL, *lokiRandomLabel, *entriesPerStream, host, *debug, queue)
					queue = make([]Entry, 0, *lokiBatch)
				}
			}
			if len(queue) > 0 {
				send(*lokiURL, *lokiRandomLabel, *entriesPerStream, host, *debug, queue)
			}
		}
		done <- true
	}()

	go func() {
		i := 0
		for {
			// If we have count check counter
			if *count > 0 && !(i < *count) {
				break
			}

			var n LogGen
			if *randomise {
				n = NewNginxLogRandom()
			} else {
				n = NewNginxLog()
			}

			msg, size := n.String()
			messages <- msg
			eventEmitted.Inc()
			eventEmittedBytes.Add(size)

			var duration time.Duration
			if *eventPerSec > 0 {
				sleepTime := 1000.0 / *eventPerSec
				duration = time.Duration(sleepTime) * time.Millisecond
			} else {
				// Randomise output between min and max
				parsedMaxTime, err := time.ParseDuration(*maxIntervall)
				if err != nil {
					panic(err)
				}
				parsedMinTime, err := time.ParseDuration(*minIntervall)
				if err != nil {
					panic(err)
				}
				duration = time.Duration(rand.Int63n(int64((parsedMaxTime - parsedMinTime) + parsedMinTime)))
			}
			// Sleep before next flush
			time.Sleep(duration)

			// Increment counter
			i++
		}
	}()

	if *metrics {
		go func() {
			fmt.Println("metrics listen on: ", *metricsAddr)
			http.Handle("/metrics", promhttp.Handler())
			http.ListenAndServe(*metricsAddr, nil)
			fmt.Println("main loop started")
		}()
	}

	select {
	case <-done:
		return
	case <-interrupt:
		fmt.Println("Recieved interrupt")
		return
	}

}
