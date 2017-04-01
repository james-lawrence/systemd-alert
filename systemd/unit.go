package systemd

import (
	"github.com/godbus/dbus"
	"github.com/pkg/errors"
)

// UnitStatus - status update about a unit.
type UnitStatus struct {
	Name        string          // The primary unit name as string
	LoadState   string          // The load state (i.e. whether the unit file has been loaded successfully)
	ActiveState string          // The active state (i.e. whether the unit is currently started or not)
	SubState    string          // The sub state (a more fine-grained version of the active state that is specific to the unit type, which the active state is not)
	Path        dbus.ObjectPath // The unit object path
}

type UnitEvent struct {
	Path                            dbus.ObjectPath
	AssertTimestamp                 uint64
	ActiveState                     string
	SubState                        string
	StateChangeTimestamp            uint64
	ActiveEnterTimestampMonotonic   uint64
	ActiveExitTimestamp             uint64
	ActiveExitTimestampMonotonic    uint64
	InactiveExitTimestampMonotonic  uint64
	ActiveEnterTimestamp            uint64
	ConditionResult                 bool
	ConditionTimestamp              uint64
	StateChangeTimestampMonotonic   uint64
	InactiveEnterTimestamp          uint64
	InactiveEnterTimestampMonotonic uint64
	ConditionTimestampMonotonic     uint64
	AssertTimestampMonotonic        uint64
	InactiveExitTimestamp           uint64
	AssertResult                    bool
}

func DecodeUnitEvent(s *dbus.Signal) (UnitEvent, error) {
	var (
		objectType string
		properties map[string]dbus.Variant
		ignored    []string
	)

	if s.Name != "org.freedesktop.DBus.Properties.PropertiesChanged" {
		return UnitEvent{}, errors.Errorf("unexpected interface: %s", s.Name)
	}

	if err := dbus.Store(s.Body, &objectType, &properties, &ignored); err != nil {
		return UnitEvent{}, errors.Wrap(err, "failed to decode properties event")
	}

	if objectType != "org.freedesktop.systemd1.Unit" {
		return UnitEvent{}, errors.Errorf("unexpected interface: %s", objectType)
	}

	return UnitEvent{
		Path:                            s.Path,
		AssertTimestamp:                 properties["AssertTimestamp"].Value().(uint64),
		ActiveState:                     properties["ActiveState"].Value().(string),
		SubState:                        properties["SubState"].Value().(string),
		StateChangeTimestamp:            properties["StateChangeTimestamp"].Value().(uint64),
		ActiveEnterTimestampMonotonic:   properties["ActiveEnterTimestampMonotonic"].Value().(uint64),
		ActiveExitTimestamp:             properties["ActiveExitTimestamp"].Value().(uint64),
		ActiveExitTimestampMonotonic:    properties["ActiveExitTimestampMonotonic"].Value().(uint64),
		InactiveExitTimestampMonotonic:  properties["InactiveExitTimestampMonotonic"].Value().(uint64),
		ActiveEnterTimestamp:            properties["ActiveEnterTimestamp"].Value().(uint64),
		ConditionResult:                 properties["ConditionResult"].Value().(bool),
		ConditionTimestamp:              properties["ConditionTimestamp"].Value().(uint64),
		StateChangeTimestampMonotonic:   properties["StateChangeTimestampMonotonic"].Value().(uint64),
		InactiveEnterTimestamp:          properties["InactiveEnterTimestamp"].Value().(uint64),
		InactiveEnterTimestampMonotonic: properties["InactiveEnterTimestampMonotonic"].Value().(uint64),
		ConditionTimestampMonotonic:     properties["ConditionTimestampMonotonic"].Value().(uint64),
		AssertTimestampMonotonic:        properties["AssertTimestampMonotonic"].Value().(uint64),
		InactiveExitTimestamp:           properties["InactiveExitTimestamp"].Value().(uint64),
		AssertResult:                    properties["AssertResult"].Value().(bool),
	}, nil
}
