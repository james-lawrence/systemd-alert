package alerts

import (
	"log"
	"strings"
	"time"

	"github.com/coreos/go-systemd/dbus"
)

type notifier interface {
	Alert(units ...*dbus.UnitStatus)
}

func isChanged(match filter) func(*dbus.UnitStatus, *dbus.UnitStatus) bool {
	return func(oldu, newu *dbus.UnitStatus) bool {
		// if new state matches then use new unit.
		return match(newu) && *oldu != *newu
	}
}

// Run - runs alerts
func Run(conn *dbus.Conn, a notifier) {
	var (
		err        error
		unitEvents <-chan map[string]*dbus.UnitStatus
		errs       <-chan error
	)
	matcher := or(FilterAutorestart, FilterFailed)
	unitEvents, errs = conn.SubscribeUnitsCustom(time.Second, 0, isChanged(matcher), nil)
	log.Printf("running %T\n", a)
	for {
		select {
		case event := <-unitEvents:
			units := make([]*dbus.UnitStatus, 0, len(event))
			for _, unit := range event {
				if unit != nil && matcher(unit) {
					units = append(units, unit)
				}
			}

			if len(units) > 0 {
				a.Alert(units...)
			}
		case err = <-errs:
			log.Println("errors", err)
		}
	}
}

type filter func(*dbus.UnitStatus) bool

func or(filters ...filter) filter {
	return func(unit *dbus.UnitStatus) bool {
		for _, filter := range filters {
			if filter(unit) {
				return true
			}
		}
		return false
	}
}

func filterByName(name string) filter {
	return func(status *dbus.UnitStatus) bool {
		log.Println("filtering by name", strings.ToLower(name), strings.ToLower(status.Name))
		return strings.ToLower(name) == strings.ToLower(status.Name)
	}
}

// FilterFailed matches units that were failed
func FilterFailed(status *dbus.UnitStatus) bool {
	const (
		failed = "failed"
	)

	return status.SubState == failed
}

// FilterAutorestart matches units that were autorestarted
func FilterAutorestart(status *dbus.UnitStatus) bool {
	const (
		autorestart = "auto-restart"
	)

	return status.SubState == autorestart
}
