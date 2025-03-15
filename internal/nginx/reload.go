// pkg/nginx/reload.go
package nginx

import (
	"log"
	"os/exec"

	"github.com/sotchenkov/binbanned/internal/ban"
)

var lastBannedCount int

// ReloadIfNeeded reloads nginx if the number of banned IPs has changed.
func ReloadIfNeeded() {
	currentCount := ban.GetBannedCount()
	if currentCount != lastBannedCount {
		log.Printf("New banned IPs detected (%d -> %d), reloading nginx...", lastBannedCount, currentCount)
		cmd := exec.Command("nginx", "-s", "reload")
		if err := cmd.Run(); err != nil {
			log.Printf("Error reloading nginx: %v", err)
		} else {
			log.Println("Nginx reloaded successfully")
			lastBannedCount = currentCount
		}
	}
}
