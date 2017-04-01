package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/james-lawrence/systemd-alert/systemd"
	"github.com/pkg/errors"
)

func main() {
	var (
		pcmd        string
		err         error
		conn        *systemd.Conn
		_, shutdown = context.WithCancel(context.Background())
	)

	if conn, err = systemd.NewSystemConnection(); err != nil {
		log.Fatalln(errors.Wrap(err, "failed to open systemd connection"))
	}

	app := kingpin.New("systemd-alert", "monitoring around systemd")
	cmd := app.Command("slack", "send alerts to slack")
	(&slackAlert{conn: conn}).configure(cmd)
	cmd = app.Command("debug", "debug to stderr")
	(&debugAlert{conn: conn}).configure(cmd)

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
