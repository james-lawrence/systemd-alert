package main

import (
	"time"

	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications/debug"
	"github.com/james-lawrence/systemd-alert/systemd"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type debugAlert struct {
	uconn, conn *systemd.Conn
	Frequency   time.Duration
}

func (t *debugAlert) configure(cmd *kingpin.CmdClause) {
	cmd.Action(t.execute)
	cmd.Flag("frequency", "frequency to emit events").Default("1s").DurationVar(&t.Frequency)
}

func (t *debugAlert) execute(c *kingpin.ParseContext) error {
	go alerts.Run(t.conn, alerts.AlertNotifiers(debug.NewAlerter()), alerts.AlertFrequency(t.Frequency))
	go alerts.Run(t.uconn, alerts.AlertNotifiers(debug.NewAlerter()), alerts.AlertFrequency(t.Frequency))
	return nil
}
