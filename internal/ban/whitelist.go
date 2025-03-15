// pkg/ban/whitelist.go
package ban

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	whitelist      = make(map[string]struct{})
	whitelistMutex sync.RWMutex
)

// LoadWhitelist loads the whitelist of IP addresses from a file.
func LoadWhitelist(whitelistFile string) error {
	file, err := os.Open(whitelistFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	whitelistMutex.Lock()
	defer whitelistMutex.Unlock()
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			whitelist[ip] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning whitelist file: %v", err)
	}
	return nil
}

// IsWhitelisted checks if an IP address is in the whitelist.
func IsWhitelisted(ip string) bool {
	whitelistMutex.RLock()
	defer whitelistMutex.RUnlock()
	_, exists := whitelist[ip]
	return exists
}
