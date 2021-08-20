package formats

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/Pallinder/go-randomdata"
	wr "github.com/mroth/weightedrand"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"text/template"

	)


type ApacheLog struct {
	Remote            string
	Host              string
	User              string
	Time              string
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
		Time:              "",
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
		Time:              "",
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
	n.Time = time.Now().Format(viper.GetString("apache.time_format"))

	t,err := template.New("line").Parse(viper.GetString("apache.output_format"))
	if err != nil {
		log.Error(err)
		return "", 0
	}
	output := new(bytes.Buffer)
	err = t.Execute(output, n)
		if err != nil {
		return "", 0
	}
	message := fmt.Sprint(output.String())

	return message, float64(len([]byte(message)))
}
