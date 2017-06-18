package influxdb

import (
	"log"
	"strings"
	"sync"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/notifications"
	"github.com/james-lawrence/systemd-alert/systemd"
	"github.com/pkg/errors"
)

func init() {
	notifications.Add("influxdb", func() alerts.Notifier {
		return NewAlerter()
	})
}

// NewAlerter configures the Alerter
func NewAlerter() *Alerter {
	return &Alerter{
		Address:   "http://localhost:8086",
		Precision: "ns",
		Database:  "influxdb",
		Metric:    "systemd-unit-failures",
		Once:      &sync.Once{},
	}
}

type clientX interface {
	// Write takes a BatchPoints object and writes all Points to InfluxDB.
	Write(bp client.BatchPoints) error
}

// Alerter - sends an alert to a webhook.
type Alerter struct {
	*sync.Once
	Address   string
	Database  string
	Precision string
	Metric    string
	client    clientX
}

// Alert about the provided units.
func (t *Alerter) Alert(units ...*systemd.UnitStatus) {
	var (
		err    error
		points []*client.Point
		batch  client.BatchPoints
	)

	t.Once.Do(func() {
		if strings.HasPrefix(t.Address, "unix") {
			log.Println("connecting to unix", strings.TrimPrefix(t.Address, "unix://"))
			t.client = newUnixSocket(strings.TrimPrefix(t.Address, "unix://"))
			return
		}

		if strings.HasPrefix(t.Address, "http") {
			if t.client, err = client.NewHTTPClient(client.HTTPConfig{}); err != nil {
				log.Println("failed to create http client", err)
			}
			return
		}
	})

	if t.client == nil {
		log.Println("client is nil, skipping")
		return
	}

	pconfig := client.BatchPointsConfig{
		Database:  t.Database,
		Precision: t.Precision,
	}

	if batch, err = client.NewBatchPoints(pconfig); err != nil {
		log.Println("failed to create batch points", err)
		return
	}

	points = make([]*client.Point, 0, len(units))
	for _, unit := range units {
		var p *client.Point
		p, err = client.NewPoint(t.Metric, map[string]string{}, map[string]interface{}{
			"unit":         unit.Name,
			"active_state": unit.ActiveState,
			"sub_state":    unit.SubState,
		})

		if err != nil {
			log.Println("failed to create point", err)
		} else {
			points = append(points, p)
		}
	}
	batch.AddPoints(points)

	if err = t.client.Write(batch); err != nil {
		log.Println(errors.Wrap(err, "failed to write events"))
		return
	}
}
