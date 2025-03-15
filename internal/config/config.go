// pkg/config/config.go
package config

import (
	"encoding/json"
	"flag"
	"log"
	"time"
)

// Config holds the application configuration parameters.
type Config struct {
	LogDir         string
	BannedFile     string
	WhitelistFile  string
	ReloadInterval time.Duration
	ParseAll       bool
	TelegramToken  string
	TelegramChat   string
	Labels         map[string]string
}

// LoadConfig parses command-line flags and returns a Config instance.
func LoadConfig() *Config {
	logDir := flag.String("logdir", "/var/log/nginx/", "Directory containing Nginx logs")
	bannedFile := flag.String("banned", "/etc/nginx/conf.d/binbanned.conf", "File to write banned IPs")
	whitelistFile := flag.String("whitelist", "/etc/nginx/ip-whitelist", "Whitelist file for IPs that should not be banned")
	reloadInterval := flag.Duration("reload-interval", 10*time.Second, "Interval for checking new bans and reloading nginx")
	parseAll := flag.Bool("parse-all", false, "Parse logs from the beginning")
	telegramToken := flag.String("telegram-token", "", "Telegram Bot token for notifications")
	telegramChat := flag.String("telegram-chat", "", "Telegram Chat ID for notifications")
	labelsStr := flag.String("labels", "", "Custom labels in JSON format (e.g. '{\"server name\": \"promobuilding\", \"foo\":\"bar\"}')")

	flag.Parse()

	var labels map[string]string
	if *labelsStr != "" {
		if err := json.Unmarshal([]byte(*labelsStr), &labels); err != nil {
			log.Fatalf("Error parsing labels JSON: %v", err)
		}
	}

	return &Config{
		LogDir:         *logDir,
		BannedFile:     *bannedFile,
		WhitelistFile:  *whitelistFile,
		ReloadInterval: *reloadInterval,
		ParseAll:       *parseAll,
		TelegramToken:  *telegramToken,
		TelegramChat:   *telegramChat,
		Labels:         labels,
	}
}
