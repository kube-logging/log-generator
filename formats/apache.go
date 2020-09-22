package formats

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Pallinder/go-randomdata"
	wr "github.com/mroth/weightedrand"
)

const ApacheTimeFormat = "02/Jan/2006:15:04:05 -0700"


type ApacheLog struct {
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


func NewApacheLog() ApacheLog {
	return ApacheLog{
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


func NewApacheLogRandom() ApacheLog {
	rand.Seed(time.Now().UTC().UnixNano())
	c := wr.NewChooser(
		wr.Choice{Item: 200, Weight: 7},
		wr.Choice{Item: 404, Weight: 3},
		wr.Choice{Item: 503, Weight: 1},
		wr.Choice{Item: 302, Weight: 2},
		wr.Choice{Item: 403, Weight: 2},
	)
	return ApacheLog{
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


func (n ApacheLog) String() (string, float64) {
	message := fmt.Sprintf("%s %s %s [%s] \"%s %s HTTP/1.1\" %d %d %q %q %q", n.Remote, n.Host, n.User, n.Time.Format(ApacheTimeFormat), n.Method, n.Path, n.Code, n.Size, n.Referer, n.Agent, n.HttpXForwardedFor)
	return message, float64(len([]byte(message)))
}
