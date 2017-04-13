### systemd-alert

monitors for failures and autorestarts and sends a notification.

### supported systems
- slack
- stderr (debug)
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

```
