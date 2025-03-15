package ban

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// AlertData holds the data for an alert/ban event.
type AlertData struct {
	IP          string
	Reason      string
	RequestTime string
	RequestID   string
	HTTPHost    string
	LogFile     string // Name of the Nginx log file from which the ban originated.
}

var bannedIPs = struct {
	sync.Mutex
	m map[string]struct{}
}{m: make(map[string]struct{})}

var bannedFilePath string

var notifierFunc func(message string)

// customLabels holds additional custom labels for alerts.
var customLabels map[string]string

// SetCustomLabels sets the custom labels from configuration.
func SetCustomLabels(labels map[string]string) {
	customLabels = labels
}

// BanIP bans an IP based on the provided AlertData.
// It validates the IP, checks against whitelist, writes to the banned file, logs the event,
// and sends a notification if configured.
func BanIP(data AlertData) {
	// Validate IP format.
	if net.ParseIP(data.IP) == nil {
		log.Printf("Invalid IP: %s, skipping", data.IP)
		return
	}

	// Skip private IP addresses.
	if isPrivateIP(data.IP) {
		return
	}

	// Check if the IP is whitelisted.
	if IsWhitelisted(data.IP) {
		return
	}

	bannedIPs.Lock()
	if _, exists := bannedIPs.m[data.IP]; exists {
		bannedIPs.Unlock()
		return // Already banned.
	}
	bannedIPs.m[data.IP] = struct{}{}
	bannedIPs.Unlock()

	// Build a full alert message with additional fields using comma-separated values for logs.
	logMsg := "Banned IP: " + data.IP + ", Reason: " + data.Reason
	if data.RequestTime != "" {
		logMsg += ", Request Time: " + data.RequestTime
	}
	if data.RequestID != "" {
		logMsg += ", Request ID: " + data.RequestID
	}
	if data.HTTPHost != "" {
		logMsg += ", HTTP Host: " + data.HTTPHost
	}
	if data.LogFile != "" {
		logMsg += ", Log File: " + data.LogFile
	}
	if len(customLabels) > 0 {
		logMsg += ", Labels: {"
		first := true
		for k, v := range customLabels {
			if !first {
				logMsg += ", "
			}
			logMsg += k + ": " + v
			first = false
		}
		logMsg += "}"
	}

	// Append the banned IP to the file in the format: "deny <ip>;"
	f, err := os.OpenFile(bannedFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening banned file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString("deny " + data.IP + ";\n"); err != nil {
		log.Printf("Error writing to banned file: %v", err)
		return
	}
	// Log full alert message.
	log.Println(logMsg)

	// Build the alert message for notifications using newline as separator.
	alertMsg := "<b>Banned IP:</b>" + data.IP + "\n" +
		"<b>Reason: </b>" + data.Reason + "\n\n"
	if data.RequestTime != "" {
		alertMsg += "<b>Date: </b>" + strings.Split(data.RequestTime, " ")[0] + "\n"
	}
	if data.HTTPHost != "" {
		alertMsg += "<b>Host: </b>" + data.HTTPHost + "\n"
	}
	if data.LogFile != "" {
		alertMsg += "<b>Log File: </b>" + data.LogFile + "\n\n"
	}
	if len(customLabels) > 0 {
		for k, v := range customLabels {
			alertMsg += "<b>" + k + ": </b>" + v + "\n"
		}
	}

	// Send notification for the new ban.
	notify(alertMsg)
}

// isPrivateIP checks if an IP address is a private or loopback address.
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 127 { // Loopback
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 { // 192.168.0.0/16
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 { // 172.16.0.0/12
			return true
		}
	}
	return false
}

// LoadBannedIPs loads already banned IPs from the banned file.
// Each line is expected to be in the format "deny <ip>;".
func LoadBannedIPs(bannedFile string) error {
	bannedFilePath = bannedFile
	file, err := os.Open(bannedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	bannedIPs.Lock()
	defer bannedIPs.Unlock()
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "deny ") && strings.HasSuffix(line, ";") {
			ip := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "deny "), ";"))
			if ip != "" {
				bannedIPs.m[ip] = struct{}{}
			}
		}
	}
	return scanner.Err()
}

// GetBannedCount returns the current number of banned IPs.
func GetBannedCount() int {
	bannedIPs.Lock()
	defer bannedIPs.Unlock()
	return len(bannedIPs.m)
}

// SetNotifier sets the notifier callback function.
func SetNotifier(fn func(message string)) {
	notifierFunc = fn
}

// notify sends a notification for a banned IP.
func notify(message string) {
	if notifierFunc != nil {
		notifierFunc(message)
	}
}
