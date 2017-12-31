### systemd-alert

monitors for failures and autorestarts and sends a notification.

### supported notifications
- stderr (debug)
- slack
- influxdb / telegraf
- linux send-notify

### example configuration
```
[agent]
	frequency = "10s"
	ignore = [
		"dnf-makecache.service",
		"openvpn@server.service",
		"${USER}.service",
	]

[[notifications.default]]

[[notifications.debug]]

[[notifications.slack]]
	message = "No place like ${HOME}"
	channel = "#engineering"
	webhook = "http://example.com"

[[notifications.influxdb]]
	address  = "unix:///run/telegraf-ops/telegraf.sock"
	metric   = "systemd"
	database = "ops"
```
