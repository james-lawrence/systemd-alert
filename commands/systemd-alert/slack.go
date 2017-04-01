package main

import (
	"time"

	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications/slack"
	"github.com/james-lawrence/systemd-alert/systemd"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type slackAlert struct {
	Alerter slack.Alerter
	conn    *systemd.Conn
}

func (t *slackAlert) configure(cmd *kingpin.CmdClause) {
	t.Alerter = slack.NewAlerter()

	cmd.Action(t.execute)
	cmd.Flag("message", "message to send").Envar("SYSTEMD_ALERT_SLACK_MESSAGE").StringVar(&t.Alerter.Message)
	cmd.Flag("channel", "destination channel of the notification").Envar("SYSTEMD_ALERT_SLACK_MESSAGE").Required().StringVar(&t.Alerter.Channel)
	cmd.Flag("webhook", "url of the webhook").Envar("SYSTEMD_ALERT_SLACK_WEBHOOK_URL").Required().StringVar(&t.Alerter.Webhook)
}

func (t *slackAlert) execute(c *kingpin.ParseContext) error {
	go alerts.Run(t.conn, t.Alerter, alerts.AlertFrequency(5*time.Second))
	return nil
}
