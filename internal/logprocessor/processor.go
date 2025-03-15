package logprocessor

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/sotchenkov/binbanned/internal/ban"
)

var (
	// hiddenFilePattern checks for segments starting with a dot (e.g., /.env, /.git/config).
	hiddenFilePattern = regexp.MustCompile(`(^|/)\.[^/]+`)
	// commonTimeRegex extracts the timestamp from common log format (inside square brackets).
	commonTimeRegex = regexp.MustCompile(`\[(.*?)\]`)
)

// ProcessLine parses a single log line (JSON or Common Log Format) and processes it.
// The parameter filePath indicates the source log file.
func ProcessLine(line string, filePath string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	var ip, req, referer string
	var requestTime, requestID, httpHost string

	if line[0] == '{' {
		// Process JSON log.
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			log.Printf("Error parsing JSON log: %v", err)
			return
		}
		if v, ok := entry["remote_addr"].(string); ok && v != "" {
			ip = v
		} else if v, ok := entry["real_ip"].(string); ok && v != "" {
			ip = v
		}
		req, _ = entry["request"].(string)
		referer, _ = entry["http_referer"].(string)
		// Extract additional fields if available
		if v, ok := entry["time_local"].(string); ok {
			requestTime = v
		}
		if v, ok := entry["request_id"].(string); ok {
			requestID = v
		}
		if v, ok := entry["http_host"].(string); ok {
			httpHost = v
		}
	} else {
		// Process Common Log Format.
		parts := strings.Split(line, "\"")
		if len(parts) < 3 {
			return
		}
		fields := strings.Fields(parts[0])
		if len(fields) < 1 {
			return
		}
		ip = fields[0]
		req = parts[1]
		referer = strings.TrimSpace(parts[2])

		// Extract request time from common log format using regex.
		if match := commonTimeRegex.FindStringSubmatch(line); len(match) > 1 {
			requestTime = match[1]
		}
		// requestID and httpHost are not available in common log format.
	}

	if req == "" {
		return
	}

	reqParts := strings.Fields(req)
	if len(reqParts) < 2 {
		return
	}

	uri := reqParts[1]

	// Do not ban if the URI starts with "/.well-known".
	if strings.HasPrefix(uri, "/.well-known") {
		return
	}

	if isForbiddenPath(uri) || isForbiddenPath(referer) {
		// Build alert data with additional fields, including the log file name.
		alert := ban.AlertData{
			IP:          ip,
			Reason:      req,
			RequestTime: requestTime,
			RequestID:   requestID,
			HTTPHost:    httpHost,
			LogFile:     filePath, // Pass the source log file name.
		}
		ban.BanIP(alert)
	}
}

// isForbiddenPath checks if a path or referer contains a forbidden pattern.
func isForbiddenPath(path string) bool {
	// Exclude common patterns that should not trigger a ban.
	if strings.Contains(path, ".tmb") || strings.Contains(path, ".php") {
		return false
	}
	return hiddenFilePattern.MatchString(path) || strings.Contains(path, ".env")
}