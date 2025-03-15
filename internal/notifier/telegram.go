// pkg/notifier/telegram.go
package notifier

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

var (
	telegramToken        string
	telegramChat         string
	notificationsEnabled bool
)

// Init sets the Telegram token and chat ID.
func Init(token, chat string) {
	telegramToken = token
	telegramChat = chat
}

// EnableNotifications enables sending Telegram notifications.
func EnableNotifications() {
	notificationsEnabled = true
}

// SendTelegramNotification sends a notification message to the configured Telegram chat.
func SendTelegramNotification(message string) {
	if telegramToken == "" || telegramChat == "" {
		// Notifications are not configured.
		return
	}
	if !notificationsEnabled {
		return
	}
	apiURL := "https://api.telegram.org/bot" + telegramToken + "/sendMessage"
	data := url.Values{}
	data.Set("chat_id", telegramChat)
	data.Set("text", message)
	data.Set("parse_mode", "HTML")

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		log.Printf("Error sending Telegram notification: %v", err)
		return
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
}
