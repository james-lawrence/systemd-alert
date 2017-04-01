package alerts

import (
	"log"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/james-lawrence/systemd-alert/systemd"
)

type notifier interface {
	Alert(units ...*systemd.UnitStatus)
}

func isChanged(match filter) func(*systemd.UnitStatus, *systemd.UnitStatus) bool {
	return func(oldu, newu *systemd.UnitStatus) bool {
		// if new state matches then use new unit status.
		return match(newu) && *oldu != *newu
	}
}

type runOption func(*RunConfig)

// RunConfig run configuration options
type RunConfig struct {
	Frequency time.Duration
}

// AlertFrequency how often to dump the alerts.
func AlertFrequency(d time.Duration) func(*RunConfig) {
	return func(c *RunConfig) {
		c.Frequency = d
	}
}

// Run - runs alerts
func Run(conn *systemd.Conn, a notifier, options ...runOption) {
	config := RunConfig{
		Frequency: 1 * time.Second,
	}

	for _, opt := range options {
		opt(&config)
	}

	events, err := receiveEvents(conn)
	if err != nil {
		conn.Close()
		log.Println(err)
		return
	}

	matcher := or(FilterAutorestart, FilterFailed)
	log.Printf("running %T\n", a)

	batch := make(map[string]*systemd.UnitStatus)
	ticker := time.NewTicker(config.Frequency)
	defer ticker.Stop()
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}

			original := batch[event.Name]
			if original == nil {
				original = &systemd.UnitStatus{}
			}

			if isChanged(matcher)(original, event) {
				batch[event.Name] = event
			}
		case _ = <-ticker.C:
			events := make([]*systemd.UnitStatus, 0, len(batch))
			for _, unit := range batch {
				events = append(events, unit)
			}

			a.Alert(events...)
			batch = make(map[string]*systemd.UnitStatus)
		}
	}
}

func receiveEvents(conn *systemd.Conn) (<-chan *systemd.UnitStatus, error) {
	var (
		err error
	)
	src := make(chan *dbus.Signal)
	dst := make(chan *systemd.UnitStatus)
	if err = conn.Subscribe(src); err != nil {
		return nil, err
	}

	if err = conn.Signals(systemd.UnitPropertiesChangedSignal); err != nil {
		return nil, err
	}

	go func() {
		for s := range src {
			var (
				err           error
				status        systemd.UnitEvent
				unitName      dbus.Variant
				unitLoadState dbus.Variant
			)

			if s.Body[0] != "org.freedesktop.systemd1.Unit" {
				continue
			}

			if status, err = systemd.DecodeUnitEvent(s); err != nil {
				log.Println(err)
				continue
			}

			if unitName, err = conn.GetUnitProperty(status.Path, "Id"); err != nil {
				log.Println("failed to get unit property: Id", err)
				continue
			}

			if unitLoadState, err = conn.GetUnitProperty(status.Path, "LoadState"); err != nil {
				log.Println("failed to get unit property: LoadState", err)
				continue
			}

			dst <- &systemd.UnitStatus{
				Name:        unitName.Value().(string),
				LoadState:   unitLoadState.Value().(string),
				ActiveState: status.ActiveState,
				SubState:    status.SubState,
				Path:        status.Path,
			}
		}
	}()

	return dst, nil
}

type filter func(*systemd.UnitStatus) bool

func or(filters ...filter) filter {
	return func(unit *systemd.UnitStatus) bool {
		for _, filter := range filters {
			if filter(unit) {
				return true
			}
		}
		return false
	}
}

func filterByName(name string) filter {
	return func(status *systemd.UnitStatus) bool {
		log.Println("filtering by name", strings.ToLower(name), strings.ToLower(status.Name))
		return strings.ToLower(name) == strings.ToLower(status.Name)
	}
}

// FilterFailed matches units that were failed
func FilterFailed(status *systemd.UnitStatus) bool {
	const (
		failed = "failed"
	)

	return status.SubState == failed
}

// FilterAutorestart matches units that were autorestarted
func FilterAutorestart(status *systemd.UnitStatus) bool {
	const (
		autorestart = "auto-restart"
	)

	return status.SubState == autorestart
}
