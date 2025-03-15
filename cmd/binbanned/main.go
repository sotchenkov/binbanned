package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sotchenkov/binbanned/internal/ban"
	"github.com/sotchenkov/binbanned/internal/config"
	"github.com/sotchenkov/binbanned/internal/nginx"
	"github.com/sotchenkov/binbanned/internal/notifier"
	"github.com/sotchenkov/binbanned/internal/watcher"
)

func main() {
	// Load configuration from command-line flags
	cfg := config.LoadConfig()

	// Initialize Telegram notifier with credentials from config
	notifier.Init(cfg.TelegramToken, cfg.TelegramChat)
	// Set the notifier callback for ban notifications
	ban.SetNotifier(notifier.SendTelegramNotification)
	// Set custom labels for alerts and logs
	ban.SetCustomLabels(cfg.Labels)

	// Load whitelist and banned IPs from files
	if err := ban.LoadWhitelist(cfg.WhitelistFile); err != nil {
		log.Fatalf("Failed to load whitelist: %v", err)
	}
	if err := ban.LoadBannedIPs(cfg.BannedFile); err != nil {
		log.Printf("Error loading banned IPs: %v", err)
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start the log watcher
	go watcher.Start(ctx, cfg)

	// Periodically reload nginx if banned IP count changes
	go func() {
		ticker := time.NewTicker(cfg.ReloadInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				nginx.ReloadIfNeeded()
			}
		}
	}()

	// Enable notifications after an initial scan period
	go func() {
		time.Sleep(60 * time.Second)
		notifier.EnableNotifications()
		log.Println("Telegram notifications enabled for new bans")
	}()

	// Block until context is cancelled
	<-ctx.Done()
	log.Println("Application is shutting down")
}
