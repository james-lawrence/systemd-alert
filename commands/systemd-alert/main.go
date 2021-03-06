package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/james-lawrence/systemd-alert/systemd"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		pcmd        string
		err         error
		uconn, conn *systemd.Conn
		_, shutdown = context.WithCancel(context.Background())
	)

	if conn, err = systemd.NewSystemConnection(); err != nil {
		log.Fatalln(errors.Wrap(err, "failed to open systemd connection"))
	}

	if uconn, err = systemd.NewUserConnection(); err != nil {
		log.Println(errors.Wrap(err, "failed to open systemd user connection"))
	}

	app := kingpin.New("systemd-alert", "monitoring around systemd")

	cmd := app.Command("slack", "send alerts to slack")
	(&slackAlert{uconn: uconn, conn: conn}).configure(cmd)
	cmd = app.Command("debug", "debug to stderr")
	(&debugAlert{uconn: uconn, conn: conn}).configure(cmd)
	cmd = app.Command("default", "default uses a configuration file to bootstrap notifications").Default()
	(&_default{uconn: uconn, conn: conn}).configure(cmd)

	if pcmd, err = app.Parse(os.Args[1:]); err != nil {
		log.Fatalln(pcmd, errors.Wrap(err, "failed to parse commandline"))
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt, syscall.SIGUSR2)

	for {
		select {
		case s := <-signals:
			switch s {
			case os.Kill, os.Interrupt:
				log.Println("shutdown request received")
				goto done
			}
		}
	}

done:
	shutdown()
}

type agentConfig struct {
	Frequency time.Duration
	Ignore    []string
}

func (t *agentConfig) UnmarshalTOML(decode func(interface{}) error) error {
	type tomlAgent struct {
		Frequency string
		Ignore    []string
	}

	var (
		err  error
		dec  tomlAgent
		freq time.Duration
	)

	if err = decode(&dec); err != nil {
		return err
	}

	if dec.Frequency != "" {
		if freq, err = time.ParseDuration(dec.Frequency); err != nil {
			return errors.Errorf("invalid agent frequency %q: %v", dec.Frequency, err)
		}
	}

	// Assign the decoded value.
	*t = agentConfig{Frequency: freq, Ignore: dec.Ignore}

	return nil
}
