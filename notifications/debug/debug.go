package debug

import (
	"log"

	"github.com/james-lawrence/systemd-alert/systemd"
)

// NewAlerter configures the Alerter
func NewAlerter() Alerter {
	return Alerter{}
}

// Alerter - sends an alert to a webhook.
type Alerter struct{}

// Alert about the provided units.
func (t Alerter) Alert(units ...*systemd.UnitStatus) {
	for _, unit := range units {
		log.Println("alert", unit)
	}
}
