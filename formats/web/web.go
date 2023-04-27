// Copyright (c) 2023 Cisco All Rights Reserved.

package web

import (
	"embed"
	"math/rand"
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"

	wr "github.com/mroth/weightedrand"
)

//go:embed *.tmpl
var TemplateFS embed.FS

type TemplateData struct {
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

func SampleData() TemplateData {
	return TemplateData{
		Remote:            "127.0.0.1",
		Host:              "-",
		User:              "-",
		Time:              time.Date(2011, 6, 25, 20, 0, 4, 0, time.UTC),
		Method:            "GET",
		Path:              "/loggen/loggen/loggen/loggen/loggen/loggen/loggen",
		Code:              200,
		Size:              650,
		Referer:           "-",
		Agent:             "golang/generator PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP",
		HttpXForwardedFor: "-",
	}
}

func (t TemplateData) WebServerDateTime() string {
	return t.Time.Format("02/Jan/2006:15:04:05 -0700")
}

func RandomData() TemplateData {
	rand.Seed(time.Now().UTC().UnixNano())

	c, _ := wr.NewChooser(
		wr.Choice{Item: 200, Weight: 7},
		wr.Choice{Item: 404, Weight: 3},
		wr.Choice{Item: 503, Weight: 1},
		wr.Choice{Item: 302, Weight: 2},
		wr.Choice{Item: 403, Weight: 2},
	)

	return TemplateData{
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

func (t TemplateData) Severity() string {
	return strconv.FormatInt(int64(t.Code), 10)
}
