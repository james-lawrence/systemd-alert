package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/james-lawrence/systemd-alert/systemd"
	"github.com/pkg/errors"
)

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type notification struct {
	Channel string  `json:"channel"`
	Emoji   string  `json:"icon_emoji"`
	Text    string  `json:"text"`
	Fields  []field `json:"fields"`
}

// NewAlerter configures the Alerter
func NewAlerter() Alerter {
	return Alerter{
		client: defaultClient(),
	}
}

func defaultClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Second,
	}
}

// Alerter - sends an alert to a webhook.
type Alerter struct {
	Channel string
	Webhook string
	Message string
	client  *http.Client
}

// Alert about the provided units.
func (t Alerter) Alert(units ...*systemd.UnitStatus) {
	var (
		err  error
		raw  []byte
		resp *http.Response
	)

	if t.client == nil {
		t.client = defaultClient()
	}

	fields := make([]field, 0, len(units))
	for _, unit := range units {
		fields = append(fields, field{Title: unit.Name, Value: fmt.Sprintf("%s - %s", unit.ActiveState, unit.SubState), Short: false})
	}

	msg := os.ExpandEnv(t.Message)

	n := notification{
		Channel: t.Channel,
		Text:    msg,
		Fields:  fields,
	}

	if raw, err = json.Marshal(n); err != nil {
		log.Println(errors.Wrap(err, "failed to encode slack notification"))
		return
	}

	if resp, err = http.Post(t.Webhook, "application/json", bytes.NewReader(raw)); err != nil {
		log.Println(errors.Wrap(err, "failed to post webhook"))
		return
	}

	if resp.StatusCode > 299 {
		log.Println("webhook request failed with status code", resp.StatusCode)
	}
}

func sendSlackNotification(channel, webhook, msg string, units ...*dbus.UnitStatus) {

}
