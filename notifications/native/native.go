package native

import (
	"fmt"

	"github.com/0xAX/notificator"
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
	notify := notificator.New(notificator.Options{
		DefaultIcon: "",
		AppName:     "systemd-alert",
	})
	return &Alerter{
		dst: notify,
	}
}

// Alerter - sends an alert to a webhook.
type Alerter struct {
	dst *notificator.Notificator
}

// Alert about the provided units.
func (t Alerter) Alert(units ...*systemd.UnitStatus) {
	for _, unit := range units {
		t.dst.Push(unit.Name, fmt.Sprintf("failed %s - %s", unit.ActiveState, unit.SubState), "", notificator.UR_NORMAL)
	}
}
