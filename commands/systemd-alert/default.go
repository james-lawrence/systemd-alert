package main

import (
	"log"

	"github.com/james-lawrence/systemd-alert"
	"github.com/james-lawrence/systemd-alert/internal/config"
	"github.com/james-lawrence/systemd-alert/notifications"
	"github.com/james-lawrence/systemd-alert/systemd"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type _default struct {
	conn   *systemd.Conn
	Config string
}

func (t *_default) configure(cmd *kingpin.CmdClause) {
	cmd.Flag("config", "path to the file containing the configuration").Default("example.toml").ExistingFileVar(&t.Config)
	cmd.Action(t.execute)
}

func (t *_default) execute(c *kingpin.ParseContext) error {
	var (
		err error
	)

	a := agentConfig{}
	alerters := []alerts.Notifier{}

	tbl := config.Decode(t.Config)

	if err = toml.UnmarshalTable(tbl.Fields["agent"].(*ast.Table), &a); err != nil {
		return errors.Wrap(err, "failed to parse agent configuration")
	}

	log.Printf("agent config: %#v\n", a)
	for name, configs := range tbl.Fields["notifications"].(*ast.Table).Fields {
		var (
			ok     bool
			plugin func() alerts.Notifier
		)

		if plugin, ok = notifications.Plugins[name]; !ok {
			continue
		}

		log.Println("loading plugin", name)
		for _, config := range configs.([]*ast.Table) {
			x := plugin()
			if err = toml.UnmarshalTable(config, x); err != nil {
				log.Println("failed to load plugin", name, "line:", config.Line, err)
				continue
			}
			alerters = append(alerters, x)
		}
	}

	go alerts.Run(t.conn,
		alerts.AlertNotifiers(alerters...),
		alerts.AlertFrequency(a.Frequency),
		alerts.AlertIgnoreServices(a.Ignore...),
	)
	return nil
}
