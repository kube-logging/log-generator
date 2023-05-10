// Copyright (c) 2023 Cisco All Rights Reserved.

package loggen

import (
	"fmt"
	"net"
	"time"

	"github.com/kube-logging/log-generator/formats"
	"github.com/kube-logging/log-generator/metrics"
	log "github.com/sirupsen/logrus"
)

type LogWriter interface {
	Send(formats.Log)
	Close()
}

type StdoutLogWriter struct{}

func newStdoutWriter() LogWriter {
	return &StdoutLogWriter{}
}

func (*StdoutLogWriter) Send(l formats.Log) {
	msg, size := l.String()
	fmt.Println(msg)

	metrics.EventEmitted.With(l.Labels()).Inc()
	metrics.EventEmittedBytes.With(l.Labels()).Add(size)
}

func (*StdoutLogWriter) Close() {}

type NetworkLogWriter struct {
	network string
	address string
	conn    net.Conn
}

func newNetworkWriter(network string, address string) LogWriter {
	nlw := &NetworkLogWriter{
		network: network,
		address: address,
	}

	nlw.reconnect()
	return nlw
}

func (nlw *NetworkLogWriter) Send(l formats.Log) {
	msg, size := l.String()

	msg += "\n"
	written := 0
	for {
		data := msg[written:]

		n, err := nlw.conn.Write([]byte(data))
		if err != nil {
			log.Errorf("Error sending message (%q), reconnecting...", err.Error())
			nlw.reconnect()
			continue
		}

		written += n

		if written == len(msg) {
			break
		}
	}

	metrics.EventEmitted.With(l.Labels()).Inc()
	metrics.EventEmittedBytes.With(l.Labels()).Add(size)
}

func (nlw *NetworkLogWriter) Close() {
	if nlw.conn != nil {
		nlw.conn.Close()
	}
}

func (nlw *NetworkLogWriter) reconnect() {
	nlw.Close()

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0

	backoff.RetryNotify(func() error {
		log.Infof("Connecting to %s %s...", nlw.network, nlw.address)
		var err error
		nlw.conn, err = net.DialTimeout(nlw.network, nlw.address, 5*time.Second)
		return err
	}, bo, func(err error, delay time.Duration) {
		log.Errorf("Error connecting to server (%q), retrying in %s", err.Error(), delay.String())
	})
}
