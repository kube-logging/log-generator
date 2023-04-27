// Copyright (c) 2023 Cisco All Rights Reserved.

package syslog

import (
	"embed"
	"fmt"
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"
)

//go:embed *.tmpl
var TemplateFS embed.FS

type TemplateData struct {
	Facility int
	severity int
	dateTime time.Time
	Host     string
	AppName  string
	PID      int
	Seq      int
	Msg      string
}

func (t TemplateData) ISODateTime() string {
	return t.dateTime.Format(time.RFC3339)
}

func (t TemplateData) BSDDateTime() string {
	return t.dateTime.Format(time.Stamp)
}

func (t TemplateData) CustomDateTime(format string) string {
	return t.dateTime.Format(format)
}

func (t TemplateData) Pri() string {
	pri := t.Facility*8 + t.severity
	return fmt.Sprintf("<%d>", pri)
}

func (t TemplateData) Severity() string {
	return strconv.FormatInt(int64(t.severity), 10)
}

func SampleData() TemplateData {
	return TemplateData{
		Facility: 20,
		severity: 5,
		dateTime: time.Date(2011, 6, 25, 20, 0, 4, 0, time.UTC),
		Host:     "hostname",
		AppName:  "appname",
		PID:      1143,
		Seq:      1,
		Msg:      "An application event log entry...",
	}
}

func RandomData() TemplateData {
	return TemplateData{
		Facility: randomdata.Number(0, 24),
		severity: randomdata.Number(0, 8),
		dateTime: time.Now().UTC(),
		Host:     fmt.Sprintf("%s-%s", randomdata.Adjective(), randomdata.Noun()),
		AppName:  randomdata.Adjective(),
		PID:      randomdata.Number(1, 10000),
		Seq:      randomdata.Number(1, 10000),
		Msg:      fmt.Sprintf("An application event log entry %s %s", randomdata.Noun(), randomdata.Noun()),
	}
}
