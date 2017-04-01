package systemd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/godbus/dbus"
)

const (
	alpha        = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	num          = `0123456789`
	alphanum     = alpha + num
	signalBuffer = 100
)

// needsEscape checks whether a byte in a potential dbus ObjectPath needs to be escaped
func needsEscape(i int, b byte) bool {
	// Escape everything that is not a-z-A-Z-0-9
	// Also escape 0-9 if it's the first character
	return strings.IndexByte(alphanum, b) == -1 ||
		(i == 0 && strings.IndexByte(num, b) != -1)
}

// PathBusEscape sanitizes a constituent string of a dbus ObjectPath using the
// rules that systemd uses for serializing special characters.
func PathBusEscape(path string) string {
	// Special case the empty string
	if len(path) == 0 {
		return "_"
	}
	n := []byte{}
	for i := 0; i < len(path); i++ {
		c := path[i]
		if needsEscape(i, c) {
			e := fmt.Sprintf("_%x", c)
			n = append(n, []byte(e)...)
		} else {
			n = append(n, c)
		}
	}
	return string(n)
}

// New establishes a connection to any available bus and authenticates.
// Callers should call Close() when done with the connection.
func New() (*Conn, error) {
	conn, err := NewSystemConnection()
	if err != nil && os.Geteuid() == 0 {
		return NewSystemdConnection()
	}
	return conn, err
}

// NewSystemConnection establishes a connection to the system bus and authenticates.
// Callers should call Close() when done with the connection
func NewSystemConnection() (*Conn, error) {
	log.Println("new system connection")
	return NewConnection(func() (*dbus.Conn, error) {
		return dbusAuthHelloConnection(dbus.SystemBusPrivate)
	})
}

// NewUserConnection establishes a connection to the session bus and
// authenticates. This can be used to connect to systemd user instances.
// Callers should call Close() when done with the connection.
func NewUserConnection() (*Conn, error) {
	return NewConnection(func() (*dbus.Conn, error) {
		return dbusAuthHelloConnection(dbus.SessionBusPrivate)
	})
}

// NewSystemdConnection establishes a private, direct connection to systemd.
// This can be used for communicating with systemd without a dbus daemon.
// Callers should call Close() when done with the connection.
func NewSystemdConnection() (*Conn, error) {
	return NewConnection(func() (*dbus.Conn, error) {
		// We skip Hello when talking directly to systemd.
		return dbusAuthConnection(func() (*dbus.Conn, error) {
			return dbus.Dial("unix:path=/run/systemd/private")
		})
	})
}

// NewConnection establishes a connection to a bus using a caller-supplied function.
// This allows connecting to remote buses through a user-supplied mechanism.
// The supplied function may be called multiple times, and should return independent connections.
// The returned connection must be fully initialised: the org.freedesktop.DBus.Hello call must have succeeded,
// and any authentication should be handled by the function.
func NewConnection(dialBus func() (*dbus.Conn, error)) (*Conn, error) {
	sysconn, err := dialBus()
	if err != nil {
		return nil, err
	}

	sigconn, err := dialBus()
	if err != nil {
		sysconn.Close()
		return nil, err
	}

	c := &Conn{
		sysconn: sysconn,
		sysobj:  systemdObject(sysconn),
		sigconn: sigconn,
		sigobj:  systemdObject(sigconn),
	}

	log.Println("connection established")

	return c, nil
}

// GetManagerProperty returns the value of a property on the org.freedesktop.systemd1.Manager
// interface. The value is returned in its string representation, as defined at
// https://developer.gnome.org/glib/unstable/gvariant-text.html
func (c *Conn) GetManagerProperty(prop string) (string, error) {
	variant, err := c.sysobj.GetProperty("org.freedesktop.systemd1.Manager." + prop)
	if err != nil {
		return "", err
	}
	return variant.String(), nil
}

func dbusAuthConnection(createBus func() (*dbus.Conn, error)) (*dbus.Conn, error) {
	conn, err := createBus()
	if err != nil {
		return nil, err
	}

	// Only use EXTERNAL method, and hardcode the uid (not username)
	// to avoid a username lookup (which requires a dynamically linked
	// libc)
	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = conn.Auth(methods)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func dbusAuthHelloConnection(createBus func() (*dbus.Conn, error)) (*dbus.Conn, error) {
	conn, err := dbusAuthConnection(createBus)
	if err != nil {
		return nil, err
	}

	if err = conn.Hello(); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func systemdObject(conn *dbus.Conn) dbus.BusObject {
	return conn.Object("org.freedesktop.systemd1", dbus.ObjectPath("/org/freedesktop/systemd1"))
}
