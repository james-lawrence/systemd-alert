package notifications

import "github.com/james-lawrence/systemd-alert"

type creator func() alerts.Notifier

var Plugins = map[string]creator{}

func Add(name string, creator creator) {
	Plugins[name] = creator
}
