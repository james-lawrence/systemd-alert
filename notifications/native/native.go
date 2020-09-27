package native

import (
	"fmt"
	"log"
	"sync"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	alerts "github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications"
	"github.com/james-lawrence/systemd-alert/systemd"
)

func init() {
	notifications.Add("default", func() alerts.Notifier {
		return NewAlerter(nil)
	})
}

// DefaultAlerter create with default connection.
func DefaultAlerter() *Alerter {
	return NewAlerter(nil)
}

// NewAlerter configures the Alerter
func NewAlerter(conn *dbus.Conn) *Alerter {
	return &Alerter{
		m:       &sync.Mutex{},
		conn:    conn,
		current: make(map[string]uint32),
	}
}

// Alerter - sends an alert to a webhook.
type Alerter struct {
	m       *sync.Mutex
	conn    *dbus.Conn
	current map[string]uint32
}

func (t *Alerter) ensureConn() {
	var (
		err  error
		conn *dbus.Conn
	)

	t.m.Lock()
	defer t.m.Unlock()

	if t.conn == nil {
		if conn, err = dbus.SessionBusPrivate(); err != nil {
			log.Println("unable to connect to dbus - disabling native notifications", err)
			return
		}

		if err = conn.Auth(nil); err != nil {
			log.Println("unable to connect to dbus - disabling native notifications", err)
			return
		}

		if err = conn.Hello(); err != nil {
			log.Println("unable to connect to dbus - disabling native notifications", err)
			return
		}

		t.conn = conn
	}
}

// Alert about the provided units.
func (t *Alerter) Alert(units ...*systemd.UnitStatus) {
	t.ensureConn()

	for _, unit := range units {
		var (
			err error
			id  uint32
		)

		if replace, ok := t.current[unit.Name]; ok {
			id = replace
		}

		n := notify.Notification{
			AppName:    "Systemd Alert",
			ReplacesID: id,
			Summary:    fmt.Sprintf("%s %s - %s", unit.Name, unit.ActiveState, unit.SubState),
		}

		if id, err = notify.SendNotification(t.conn, n); err != nil {
			log.Println("notification failed", err)
			continue
		}

		t.current[unit.Name] = id
	}
}
