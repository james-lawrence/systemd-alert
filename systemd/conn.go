package systemd

import (
	"github.com/godbus/dbus"
	"github.com/pkg/errors"
)

// Conn is a connection to systemd's dbus endpoint.
type Conn struct {
	// sysconn/sysobj are only used to call dbus methods
	sysconn *dbus.Conn
	sysobj  dbus.BusObject

	// sigconn/sigobj are only used to receive dbus signals
	sigconn *dbus.Conn
	sigobj  dbus.BusObject
}

// Close closes an established connection
func (c *Conn) Close() {
	c.sysconn.Close()
	c.sigconn.Close()
}

// Subscribe sets up this connection to subscribe to all systemd dbus events.
// When the connection closes systemd will automatically stop sending signals so
// there is no need to explicitly call Unsubscribe().
func (c *Conn) Subscribe(dst chan *dbus.Signal) error {
	if err := c.sigobj.Call("org.freedesktop.systemd1.Manager.Subscribe", 0).Store(); err != nil {
		return errors.Wrap(err, "failed to subscribe to systemd")
	}

	c.sigconn.Signal(dst)
	return nil
}

// Signals add signals to receive on this connection.
func (c *Conn) Signals(signals ...signal) error {
	for _, signal := range signals {
		if err := signal(c.sigconn); err != nil {
			return errors.Wrap(err, "failed to register signal")
		}
	}
	return nil
}

// Unsubscribe this connection from systemd dbus events.
func (c *Conn) Unsubscribe() error {
	err := c.sigobj.Call("org.freedesktop.systemd1.Manager.Unsubscribe", 0).Store()
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) GetUnitProperty(path dbus.ObjectPath, name string) (result dbus.Variant, err error) {
	err = c.sysconn.Object(c.sysobj.Destination(), path).Call("org.freedesktop.DBus.Properties.Get", 0, "org.freedesktop.systemd1.Unit", name).Store(&result)
	return
}
