package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/Pallinder/go-randomdata"
	wr "github.com/mroth/weightedrand"
)

const NginxTimeFormat = "02/Jan/2006:15:04:05 -0700"

// (?<remote>[^ ]*) (?<host>[^ ]*) (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)"(?:\s+(?<http_x_forwarded_for>[^ ]+))?)?$/
// 10.20.64.0 - - [05/Oct/2019:06:49:04 +0000] "GET / HTTP/1.1" 200 612 "-" "kube-probe/1.15" "-"

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
		Path:              "/",
		Code:              200,
		Size:              650,
		Referer:           "-",
		Agent:             "",
		HttpXForwardedFor: "-",
	}
}

func NewNginxLogRandom() NginxLog {
	rand.Seed(time.Now().UTC().UnixNano())
	c := wr.NewChooser(
		wr.Choice{Item: 200, Weight: 7},
		wr.Choice{Item: 404, Weight: 3},
		wr.Choice{Item: 503, Weight: 1},
		wr.Choice{Item: 302, Weight: 4},
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

func (n NginxLog) String() string {
	return fmt.Sprintf("%s %s %s [%s] \"%s %s HTTP/1.1\" %d %d %q %q %q", n.Remote, n.Host, n.User, n.Time.Format(NginxTimeFormat), n.Method, n.Path, n.Code, n.Size, n.Referer, n.Agent, n.HttpXForwardedFor)
}

func main() {
	minIntervall := flag.Int("min-intervall", 3, "Minimum interval between log messages")
	maxIntervall := flag.Int("max-intervall", 8, "Maximum interval between log messages")
	count := flag.Int("count", 10000, "The amount of log message to emit.")

	flag.Parse()

	for i := 0; i < *count; i++ {

		time.Sleep(time.Duration(rand.Intn((*maxIntervall-*minIntervall)+*minIntervall)) * time.Second)
		n := NewNginxLogRandom()
		fmt.Println(n)
	}
}
