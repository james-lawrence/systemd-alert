package systemd

import (
	"github.com/godbus/dbus"
)

type signal func(*dbus.Conn) error

// UnitNewSignal registers to receive new unit signals on the provided connection.
func UnitNewSignal(conn *dbus.Conn) error {
	return conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.systemd1.Manager',member='UnitNew'").Err
}

// UnitRemovedSignal registers to receive new unit signals on the provided connection.
func UnitRemovedSignal(conn *dbus.Conn) error {
	return conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.systemd1.Manager',member='UnitRemoved'").Err
}

// UnitPropertiesChangedSignal registers to receive property changes.
func UnitPropertiesChangedSignal(conn *dbus.Conn) error {
	return conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged'").Err
}

// JobRemovedSignal registers to receive signals when a job is removed.
func JobRemovedSignal(conn *dbus.Conn) error {
	return conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal', interface='org.freedesktop.systemd1.Manager', member='JobRemoved'").Err
}
