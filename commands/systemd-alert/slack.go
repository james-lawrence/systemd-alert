package main

import (
	"time"

	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications/slack"
	"github.com/james-lawrence/systemd-alert/systemd"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type slackAlert struct {
	Alerter   slack.Alerter
	conn      *systemd.Conn
	Frequency time.Duration
	IgnoreSet []string
}

func (t *slackAlert) configure(cmd *kingpin.CmdClause) {
	t.Alerter = slack.NewAlerter()

	cmd.Action(t.execute)
	cmd.Flag("message", "message to send").Envar("SYSTEMD_ALERT_SLACK_MESSAGE").StringVar(&t.Alerter.Message)
	cmd.Flag("channel", "destination channel of the notification").Envar("SYSTEMD_ALERT_SLACK_MESSAGE").Required().StringVar(&t.Alerter.Channel)
	cmd.Flag("webhook", "url of the webhook").Envar("SYSTEMD_ALERT_SLACK_WEBHOOK_URL").Required().StringVar(&t.Alerter.Webhook)
	cmd.Flag("frequency", "frequency to emit events").Default("5s").DurationVar(&t.Frequency)
	cmd.Flag("ignore", "set of services to ignore").StringsVar(&t.IgnoreSet)
}

func (t *slackAlert) execute(c *kingpin.ParseContext) error {
	go alerts.Run(t.conn, t.Alerter, alerts.AlertFrequency(t.Frequency), alerts.AlertIgnoreServices(t.IgnoreSet...))
	return nil
}
