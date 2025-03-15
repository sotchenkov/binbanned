# binbanned

[Readme на русском](https://github.com/sotchenkov/binbanned/blob/main/README-RU.md)

**binbanned** — is a utility for real-time monitoring of Nginx logs and automatic IP blocking at the Nginx level for suspicious requests. It supports both JSON logs and the Common Log Format.

Features

- **Real-time Log Monitoring:** Reads logs using tail mode with support for file rotation.
- **Support for Multiple Log Formats:** Handles both JSON logs and standard logs.
- **Suspicious Request Filtering:** Blocks IP addresses if the URI contains hidden files (e.g., /.env, /.git/config) or if suspicious patterns are detected in the user-agent. This pattern covers most malicious web bots.
- **Whitelist:** Ability to exclude specified IP addresses from being blocked using a separate file.
- **Nginx Configuration Reload Signal:** When an IP is added to the blocklist, the utility sends a signal to Nginx to reload its configuration.
- **Telegram Notifications:** Sends notifications about new blocks via a Telegram bot.
- **Custom Labels:** Allows adding custom labels to alerts and logs using the --labels flag (for example, `--labels '{"server name": "my-server", "region":"us"}`).

### Installation and Build

You can download a precompiled binary or build it yourself:

* **Downloading the Binary**

Visit the releases section, select the version you need, and download the binary:
```bash
wget https://github.com/sotchenkov/binbanned/releases/ (required file)
```

* **Build:**
```bash
git clone git@github.com:sotchenkov/binbanned.git
cd binbanned
go build -o binbanned ./cmd/binbanned/main.go
```
Or for a static build:
```bash
CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o binbanned ./cmd/binbanned/main.go
```

### Usage

1. **Create the necessary files and set permissions:**
```bash
sudo touch /etc/nginx/conf.d/binbanned.conf
sudo touch /etc/nginx/ip-whitelist
sudo mkdir /var/log/binbanned
sudo chmod +x binbanned
sudo mv binbanned /usr/bin/
```

2. **Create a systemd service and specify the necessary parameters:**
```bash
sudo vim /etc/systemd/system/binbanned.service
```

```ini
[Unit]
Description=Binbanned service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/bin/binbanned -telegram-token 'YOUR_TELEGRAM_BOT_TOKEN' -telegram-chat 'YOUR_TELEGRAM_CHAT_ID' --labels '{"server name": "my-server", "region":"ru"}'
Restart=on-failure
RestartSec=20
StandardOutput=append:/var/log/binbanned/binbanned.log
StandardError=append:/var/log/binbanned/binbanned.log

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl start binbanned
sudo systemctl status binbanned
```

### Command-line Flags

    --logdir
    Directory containing Nginx logs (default: /var/log/nginx/).

    --banned
    File to write blocked IPs (default: /etc/nginx/conf.d/binbanned.conf).

    --whitelist
    File with the whitelist of IPs (default: /etc/nginx/ip-whitelist).

    --reload-interval
    Interval for checking new blocks and reloading Nginx (default: 10s).

    --parse-all
    Parse logs from the beginning of files (if specified).

    --telegram-token
    Telegram Bot token for sending notifications.

    --telegram-chat
    Telegram Chat ID for notifications.

    --labels
    Custom labels for alerts/logs in JSON format.

### Nginx Configuration

Ensure that the main Nginx configuration file (/etc/nginx/nginx.conf) includes the directive:
```
include /etc/nginx/conf.d/*.conf;
```
This guarantees that Nginx will apply the settings from the file containing the blocked IPs.