package native

import (
	"log"

	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications"
	"github.com/james-lawrence/systemd-alert/systemd"
)

func init() {
	notifications.Add("default", func() alerts.Notifier {
		return NewAlerter()
	})
}

// NewAlerter configures the Alerter
func NewAlerter() *Alerter {
	return &Alerter{}
}

// Alerter - sends an alert to a webhook.
type Alerter struct{}

// Alert about the provided units.
func (t Alerter) Alert(units ...*systemd.UnitStatus) {
	for _, unit := range units {
		log.Println("alert", unit)
	}
}
