package notify

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	dbusRemoveMatch            = "org.freedesktop.DBus.RemoveMatch"
	dbusAddMatch               = "org.freedesktop.DBus.AddMatch"
	dbusObjectPath             = "/org/freedesktop/Notifications" // the DBUS object path
	dbusNotificationsInterface = "org.freedesktop.Notifications"  // DBUS Interface
	signalNotificationClosed   = "org.freedesktop.Notifications.NotificationClosed"
	signalActionInvoked        = "org.freedesktop.Notifications.ActionInvoked"
	callGetCapabilities        = "org.freedesktop.Notifications.GetCapabilities"
	callCloseNotification      = "org.freedesktop.Notifications.CloseNotification"
	callNotify                 = "org.freedesktop.Notifications.Notify"
	callGetServerInformation   = "org.freedesktop.Notifications.GetServerInformation"

	channelBufferSize = 10
)

// Notification holds all information needed for creating a notification
type Notification struct {
	AppName string
	// Setting ReplacesID atomically replaces the notification with this ID.
	// Optional.
	ReplacesID uint32
	// See predefined icons here: http://standards.freedesktop.org/icon-naming-spec/icon-naming-spec-latest.html
	// Optional.
	AppIcon string
	Summary string
	Body    string
	// Actions are tuples of (action_key, label), e.g.: []string{"cancel", "Cancel", "open", "Open"}
	Actions []string
	Hints   map[string]dbus.Variant
	// ExpireTimeout: milliseconds to show notification
	ExpireTimeout int32
}

// SendNotification is provided for convenience.
// Use if you only want to deliver a notification and dont care about events.
func SendNotification(conn *dbus.Conn, note Notification) (uint32, error) {
	actions := len(note.Actions)
	if (actions % 2) != 0 {
		return 0, errors.New("actions must be pairs of (key, label)")
	}

	obj := conn.Object(dbusNotificationsInterface, dbusObjectPath)
	call := obj.Call(callNotify, 0,
		note.AppName,
		note.ReplacesID,
		note.AppIcon,
		note.Summary,
		note.Body,
		note.Actions,
		note.Hints,
		note.ExpireTimeout)
	if call.Err != nil {
		return 0, fmt.Errorf("error sending notification: %w", call.Err)
	}
	var ret uint32
	err := call.Store(&ret)
	if err != nil {
		return ret, fmt.Errorf("error getting uint32 ret value: %w", err)
	}
	return ret, nil
}

// ServerInformation is a holder for information returned by
// GetServerInformation call.
type ServerInformation struct {
	Name        string
	Vendor      string
	Version     string
	SpecVersion string
}

// GetServerInformation returns the information on the server.
//
// org.freedesktop.Notifications.GetServerInformation
//
//  GetServerInformation Return Values
//
//		Name		 Type	  Description
//		name		 STRING	  The product name of the server.
//		vendor		 STRING	  The vendor name. For example, "KDE," "GNOME," "freedesktop.org," or "Microsoft."
//		version		 STRING	  The server's version number.
//		spec_version STRING	  The specification version the server is compliant with.
//
func GetServerInformation(conn *dbus.Conn) (ServerInformation, error) {
	obj := conn.Object(dbusNotificationsInterface, dbusObjectPath)
	if obj == nil {
		return ServerInformation{}, errors.New("error creating dbus call object")
	}
	call := obj.Call(callGetServerInformation, 0)
	if call.Err != nil {
		return ServerInformation{}, fmt.Errorf("error calling %v: %v", callGetServerInformation, call.Err)
	}

	ret := ServerInformation{}
	err := call.Store(&ret.Name, &ret.Vendor, &ret.Version, &ret.SpecVersion)
	if err != nil {
		return ret, fmt.Errorf("error reading %v return values: %v", callGetServerInformation, err)
	}
	return ret, nil
}

// GetCapabilities gets the capabilities of the notification server.
// This call takes no parameters.
// It returns an array of strings. Each string describes an optional capability implemented by the server.
//
// See also: https://developer.gnome.org/notification-spec/
// GetCapabilities provide an exported method for this operation
func GetCapabilities(conn *dbus.Conn) ([]string, error) {
	obj := conn.Object(dbusNotificationsInterface, dbusObjectPath)
	call := obj.Call(callGetCapabilities, 0)
	if call.Err != nil {
		return []string{}, call.Err
	}
	var ret []string
	err := call.Store(&ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// Notifier is an interface implementing the operations supported by the
// Freedesktop DBus Notifications object.
//
// New() sets up a Notifier that listens on dbus' signals regarding
// Notifications: NotificationClosed and ActionInvoked.
//
// Signal delivery works by subscribing to all signals regarding Notifications,
// which means you will see signals for Notifications also from other sources,
// not just the latest you sent
//
// Users that only want to send a simple notification, but don't care about
// interacting with signals, can use exported method: SendNotification(conn, Notification)
//
// Caller is responsible for calling Close() before exiting,
// to shut down event loop and cleanup dbus registration.
type Notifier interface {
	SendNotification(n Notification) (uint32, error)
	GetCapabilities() ([]string, error)
	GetServerInformation() (ServerInformation, error)
	CloseNotification(id uint32) (bool, error)
	Close() error
}

// NotificationClosedHandler is called when we receive a NotificationClosed signal
type NotificationClosedHandler func(*NotificationClosedSignal)

// ActionInvokedHandler is called when we receive a signal that one of the action_keys was invoked.
//
// Note that invoking an action often also produces a NotificationClosedSignal,
// so you might receive both a Closed signal and a ActionInvoked signal.
//
// I suspect this detail is implementation specific for the UI interaction,
// and does at least happen on XFCE4.
type ActionInvokedHandler func(*ActionInvokedSignal)

// ActionInvokedSignal holds data from any signal received regarding Actions invoked
type ActionInvokedSignal struct {
	// ID of the Notification the action was invoked for
	ID uint32
	// Key from the tuple (action_key, label)
	ActionKey string
}

// notifier implements Notifier interface
type notifier struct {
	conn     *dbus.Conn
	signal   chan *dbus.Signal
	done     chan bool
	onClosed NotificationClosedHandler
	onAction ActionInvokedHandler
	wg       *sync.WaitGroup
	log      logger
}

type logger interface {
	Printf(format string, v ...interface{})
}

// option overrides certain parts of a Notifier
type option func(*notifier)

// WithLogger sets a new logger func
func WithLogger(logz logger) option {
	return func(n *notifier) {
		n.log = logz
	}
}

// WithOnAction sets ActionInvokedHandler handler
func WithOnAction(h ActionInvokedHandler) option {
	return func(n *notifier) {
		n.onAction = h
	}
}

// WithOnClosed sets NotificationClosed handler
func WithOnClosed(h NotificationClosedHandler) option {
	return func(n *notifier) {
		n.onClosed = h
	}
}

// New creates a new Notifier using conn.
// See also: Notifier
func New(conn *dbus.Conn, opts ...option) (Notifier, error) {
	n := &notifier{
		conn:     conn,
		signal:   make(chan *dbus.Signal, channelBufferSize),
		done:     make(chan bool),
		wg:       &sync.WaitGroup{},
		onClosed: func(s *NotificationClosedSignal) {},
		onAction: func(s *ActionInvokedSignal) {},
		log:      &loggerWrapper{"notify: "},
	}

	for _, val := range opts {
		val(n)
	}

	// add a listener (matcher) in dbus for signals to Notification interface.
	err := n.conn.AddMatchSignal(
		dbus.WithMatchObjectPath(dbusObjectPath),
		dbus.WithMatchInterface(dbusNotificationsInterface),
	)
	if err != nil {
		return nil, fmt.Errorf("error registering for signals in dbus: %w", err)
	}
	// register in dbus for signal delivery
	n.conn.Signal(n.signal)

	// start eventloop
	go n.eventLoop()

	return n, nil
}

func (n notifier) eventLoop() {
	n.wg.Add(1)
	defer n.wg.Done()
	for {
		select {
		case signal := <-n.signal:
			n.handleSignal(signal)
		case <-n.done:
			n.log.Printf("Got Close() signal, shutting down...")
			return
		}
	}
}

// signal handler that translates and sends notifications to channels
func (n notifier) handleSignal(signal *dbus.Signal) {
	switch signal.Name {
	case signalNotificationClosed:
		nc := &NotificationClosedSignal{
			ID:     signal.Body[0].(uint32),
			Reason: Reason(signal.Body[1].(uint32)),
		}
		n.onClosed(nc)
	case signalActionInvoked:
		is := &ActionInvokedSignal{
			ID:        signal.Body[0].(uint32),
			ActionKey: signal.Body[1].(string),
		}
		n.onAction(is)
	default:
		n.log.Printf("Received unknown signal: %+v", signal)
	}
}

func (n *notifier) GetCapabilities() ([]string, error) {
	return GetCapabilities(n.conn)
}
func (n *notifier) GetServerInformation() (ServerInformation, error) {
	return GetServerInformation(n.conn)
}

// SendNotification sends a notification to the notification server and returns the ID or an error.
//
// Implements dbus call:
//
//     UINT32 org.freedesktop.Notifications.Notify (
//	       STRING app_name,
//	       UINT32 replaces_id,
//	       STRING app_icon,
//	       STRING summary,
//	       STRING body,
//	       ARRAY  actions,
//	       DICT   hints,
//	       INT32  expire_timeout
//     );
//
//		Name	    	Type	Description
//		app_name		STRING	The optional name of the application sending the notification. Can be blank.
//		replaces_id	    UINT32	The optional notification ID that this notification replaces. The server must atomically (ie with no flicker or other visual cues) replace the given notification with this one. This allows clients to effectively modify the notification while it's active. A value of value of 0 means that this notification won't replace any existing notifications.
//		app_icon		STRING	The optional program icon of the calling application. Can be an empty string, indicating no icon.
//		summary		    STRING	The summary text briefly describing the notification.
//		body			STRING	The optional detailed body text. Can be empty.
//		actions		    ARRAY	Actions are sent over as a list of pairs. Each even element in the list (starting at index 0) represents the identifier for the action. Each odd element in the list is the localized string that will be displayed to the user.
//		hints	        DICT	Optional hints that can be passed to the server from the client program. Although clients and servers should never assume each other supports any specific hints, they can be used to pass along information, such as the process PID or window ID, that the server may be able to make use of. See Hints. Can be empty.
//      expire_timeout  INT32   The timeout time in milliseconds since the display of the notification at which the notification should automatically close.
//								If -1, the notification's expiration time is dependent on the notification server's settings, and may vary for the type of notification. If 0, never expire.
//
// If replaces_id is 0, the return value is a UINT32 that represent the notification.
// It is unique, and will not be reused unless a MAXINT number of notifications have been generated.
// An acceptable implementation may just use an incrementing counter for the ID.
// The returned ID is always greater than zero. Servers must make sure not to return zero as an ID.
//
// If replaces_id is not 0, the returned value is the same value as replaces_id.
func (n *notifier) SendNotification(note Notification) (uint32, error) {
	return SendNotification(n.conn, note)
}

// CloseNotification causes a notification to be forcefully closed and removed from the user's view.
// It can be used, for example, in the event that what the notification pertains to is no longer relevant,
// or to cancel a notification with no expiration time.
//
// The NotificationClosed (dbus) signal is emitted by this method.
// If the notification no longer exists, an empty D-BUS Error message is sent back.
func (n *notifier) CloseNotification(id uint32) (bool, error) {
	obj := n.conn.Object(dbusNotificationsInterface, dbusObjectPath)
	call := obj.Call(callCloseNotification, 0, id)
	if call.Err != nil {
		return false, call.Err
	}
	return true, nil
}

// NotificationClosedSignal holds data for *Closed callbacks from Notifications Interface.
type NotificationClosedSignal struct {
	// ID of the Notification the signal was invoked for
	ID uint32
	// A reason given if known
	Reason Reason
}

// Reason for the closed notification
type Reason uint32

const (
	// ReasonExpired when a notification expired
	ReasonExpired Reason = 1

	// ReasonDismissedByUser when a notification has been dismissed by a user
	ReasonDismissedByUser Reason = 2

	// ReasonClosedByCall when a notification has been closed by a call to CloseNotification
	ReasonClosedByCall Reason = 3

	// ReasonUnknown when as notification has been closed for an unknown reason
	ReasonUnknown Reason = 4
)

func (r Reason) String() string {
	switch r {
	case ReasonExpired:
		return "Expired"
	case ReasonDismissedByUser:
		return "DismissedByUser"
	case ReasonClosedByCall:
		return "ClosedByCall"
	case ReasonUnknown:
		return "Unknown"
	default:
		return "Other"
	}
}

// Close cleans up and shuts down signal delivery loop
func (n *notifier) Close() error {
	n.done <- true

	// remove signal reception
	n.conn.RemoveSignal(n.signal)

	// unregister in dbus:
	errRemoveMatch := n.conn.RemoveMatchSignal(
		dbus.WithMatchObjectPath(dbusObjectPath),
		dbus.WithMatchInterface(dbusNotificationsInterface),
	)

	close(n.done)

	// wait for eventloop to shut down...
	n.wg.Wait()

	return errRemoveMatch
}

type loggerWrapper struct {
	prefix string
}

func (l *loggerWrapper) Printf(format string, v ...interface{}) {
	log.Printf(l.prefix+format, v...)
}
