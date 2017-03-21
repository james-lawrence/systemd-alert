package main

import (
	"github.com/coreos/go-systemd/dbus"
	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications/slack"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type slackAlert struct {
	Alerter slack.Alerter
	conn    *dbus.Conn
}

func (t *slackAlert) configure(cmd *kingpin.CmdClause) {
	t.Alerter = slack.NewAlerter()

	cmd.Action(t.execute)
	cmd.Flag("message", "message to send").Envar("SYSTEMD_ALERT_SLACK_MESSAGE").StringVar(&t.Alerter.Message)
	cmd.Flag("channel", "destination channel of the notification").Envar("SYSTEMD_ALERT_SLACK_MESSAGE").Required().StringVar(&t.Alerter.Channel)
	cmd.Flag("webhook", "url of the webhook").Envar("SYSTEMD_ALERT_SLACK_WEBHOOK_URL").Required().StringVar(&t.Alerter.Webhook)
}

func (t *slackAlert) execute(c *kingpin.ParseContext) error {
	var (
		err   error
		units <-chan map[string]*dbus.UnitStatus
		errs  <-chan error
	)

	if units, errs, err = subscribeToSignals(t.conn); err != nil {
		return errors.Wrap(err, "failed to subscribe to systemd events")
	}

	go alerts.Run(t.Alerter, units, errs)
	return nil
}
