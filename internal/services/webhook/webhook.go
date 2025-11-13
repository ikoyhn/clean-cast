package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	WebhookTypeDiscord = "discord"
	WebhookTypeSlack   = "slack"
	WebhookTypeGeneric = "generic"

	EventNewEpisode       = "new_episode"
	EventDownloadComplete = "download_complete"
	EventError            = "error"

	StatusPending  = "pending"
	StatusSent     = "sent"
	StatusFailed   = "failed"
	StatusRetrying = "retrying"

	MaxRetries        = 3
	InitialRetryDelay = 5 * time.Second
	MaxRetryDelay     = 5 * time.Minute
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// SendWebhook sends a webhook notification for a given event
func SendWebhook(event string, data map[string]interface{}) {
	// Get all enabled webhook configs
	configs := database.GetWebhookConfigsByEvent(event)
	if len(configs) == 0 {
		logger.Logger.Debug().
			Str("event", event).
			Msg("No webhooks configured for event")
		return
	}

	// Create webhook event payload
	webhookEvent := models.WebhookEvent{
		Event:     event,
		Timestamp: time.Now(),
		Data:      data,
	}

	payloadJSON, err := json.Marshal(webhookEvent)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("event", event).
			Msg("Failed to marshal webhook payload")
		return
	}

	// Send to all configured webhooks
	for _, config := range configs {
		delivery := &models.WebhookDelivery{
			WebhookConfigId: config.Id,
			Event:           event,
			Payload:         string(payloadJSON),
			Status:          StatusPending,
			Attempts:        0,
			MaxRetries:      MaxRetries,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		database.SaveWebhookDelivery(delivery)

		// Send webhook asynchronously
		go sendWebhookDelivery(config, delivery, webhookEvent)
	}
}

// sendWebhookDelivery sends a webhook delivery attempt
func sendWebhookDelivery(config models.WebhookConfig, delivery *models.WebhookDelivery, event models.WebhookEvent) {
	var payload interface{}
	var contentType string

	// Format payload based on webhook type
	switch config.Type {
	case WebhookTypeDiscord:
		payload = formatDiscordPayload(event)
		contentType = "application/json"
	case WebhookTypeSlack:
		payload = formatSlackPayload(event)
		contentType = "application/json"
	case WebhookTypeGeneric:
		payload = event
		contentType = "application/json"
	default:
		payload = event
		contentType = "application/json"
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Int32("delivery_id", delivery.Id).
			Msg("Failed to marshal webhook payload")
		updateDeliveryStatus(delivery, StatusFailed, 0, "", err.Error())
		return
	}

	// Send HTTP request
	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Int32("delivery_id", delivery.Id).
			Msg("Failed to create webhook request")
		updateDeliveryStatus(delivery, StatusFailed, 0, "", err.Error())
		return
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "CleanCast-Webhook/1.0")

	delivery.Attempts++
	resp, err := httpClient.Do(req)

	if err != nil {
		logger.Logger.Error().
			Err(err).
			Int32("delivery_id", delivery.Id).
			Int("attempt", delivery.Attempts).
			Msg("Failed to send webhook")

		// Retry logic with exponential backoff
		if delivery.Attempts < delivery.MaxRetries {
			scheduleRetry(delivery, config, event)
		} else {
			updateDeliveryStatus(delivery, StatusFailed, 0, "", err.Error())
		}
		return
	}
	defer resp.Body.Close()

	// Read response body
	responseBody := make([]byte, 1024)
	n, _ := resp.Body.Read(responseBody)
	responseBodyStr := string(responseBody[:n])

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Logger.Info().
			Int32("delivery_id", delivery.Id).
			Int("status_code", resp.StatusCode).
			Str("webhook_name", config.Name).
			Msg("Webhook sent successfully")
		updateDeliveryStatus(delivery, StatusSent, resp.StatusCode, responseBodyStr, "")
	} else {
		logger.Logger.Warn().
			Int32("delivery_id", delivery.Id).
			Int("status_code", resp.StatusCode).
			Int("attempt", delivery.Attempts).
			Str("response", responseBodyStr).
			Msg("Webhook returned non-success status")

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && delivery.Attempts < delivery.MaxRetries {
			scheduleRetry(delivery, config, event)
		} else {
			errorMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, responseBodyStr)
			updateDeliveryStatus(delivery, StatusFailed, resp.StatusCode, responseBodyStr, errorMsg)
		}
	}
}

// scheduleRetry schedules a webhook retry with exponential backoff
func scheduleRetry(delivery *models.WebhookDelivery, config models.WebhookConfig, event models.WebhookEvent) {
	// Calculate exponential backoff delay
	delay := time.Duration(math.Pow(2, float64(delivery.Attempts-1))) * InitialRetryDelay
	if delay > MaxRetryDelay {
		delay = MaxRetryDelay
	}

	nextRetry := time.Now().Add(delay)
	delivery.NextRetryAt = &nextRetry
	delivery.Status = StatusRetrying
	delivery.UpdatedAt = time.Now()

	database.UpdateWebhookDelivery(delivery)

	logger.Logger.Info().
		Int32("delivery_id", delivery.Id).
		Int("attempt", delivery.Attempts).
		Dur("delay", delay).
		Time("next_retry", nextRetry).
		Msg("Scheduling webhook retry")

	// Schedule retry
	time.AfterFunc(delay, func() {
		sendWebhookDelivery(config, delivery, event)
	})
}

// updateDeliveryStatus updates the webhook delivery status
func updateDeliveryStatus(delivery *models.WebhookDelivery, status string, responseCode int, responseBody string, errorMessage string) {
	delivery.Status = status
	delivery.ResponseCode = responseCode
	delivery.ResponseBody = responseBody
	delivery.ErrorMessage = errorMessage
	delivery.UpdatedAt = time.Now()

	database.UpdateWebhookDelivery(delivery)
}

// formatDiscordPayload formats the webhook payload for Discord
func formatDiscordPayload(event models.WebhookEvent) map[string]interface{} {
	embed := map[string]interface{}{
		"title":       getDiscordTitle(event.Event),
		"description": getDiscordDescription(event),
		"color":       getDiscordColor(event.Event),
		"timestamp":   event.Timestamp.Format(time.RFC3339),
		"fields":      getDiscordFields(event),
	}

	return map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}
}

// formatSlackPayload formats the webhook payload for Slack
func formatSlackPayload(event models.WebhookEvent) map[string]interface{} {
	return map[string]interface{}{
		"text": getSlackText(event),
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": getSlackTitle(event.Event),
				},
			},
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": getSlackDescription(event),
				},
			},
		},
	}
}

// Discord helper functions
func getDiscordTitle(eventType string) string {
	switch eventType {
	case EventNewEpisode:
		return "🎙️ New Episode Available"
	case EventDownloadComplete:
		return "✅ Download Complete"
	case EventError:
		return "❌ Error Occurred"
	default:
		return "📢 Notification"
	}
}

func getDiscordDescription(event models.WebhookEvent) string {
	switch event.Event {
	case EventNewEpisode:
		if title, ok := event.Data["title"].(string); ok {
			return fmt.Sprintf("A new episode is available: **%s**", title)
		}
		return "A new episode is available"
	case EventDownloadComplete:
		if videoId, ok := event.Data["video_id"].(string); ok {
			return fmt.Sprintf("Download completed for video: `%s`", videoId)
		}
		return "Download completed successfully"
	case EventError:
		if msg, ok := event.Data["error"].(string); ok {
			return fmt.Sprintf("An error occurred: %s", msg)
		}
		return "An error occurred"
	default:
		return fmt.Sprintf("Event: %s", event.Event)
	}
}

func getDiscordColor(eventType string) int {
	switch eventType {
	case EventNewEpisode:
		return 0x5865F2 // Blue
	case EventDownloadComplete:
		return 0x57F287 // Green
	case EventError:
		return 0xED4245 // Red
	default:
		return 0x99AAB5 // Gray
	}
}

func getDiscordFields(event models.WebhookEvent) []map[string]interface{} {
	fields := []map[string]interface{}{}

	for key, value := range event.Data {
		if key == "title" || key == "error" || key == "video_id" {
			continue // Skip fields already in description
		}

		fields = append(fields, map[string]interface{}{
			"name":   strings.Title(strings.ReplaceAll(key, "_", " ")),
			"value":  fmt.Sprintf("%v", value),
			"inline": true,
		})
	}

	return fields
}

// Slack helper functions
func getSlackTitle(eventType string) string {
	switch eventType {
	case EventNewEpisode:
		return "🎙️ New Episode Available"
	case EventDownloadComplete:
		return "✅ Download Complete"
	case EventError:
		return "❌ Error Occurred"
	default:
		return "📢 Notification"
	}
}

func getSlackText(event models.WebhookEvent) string {
	return getSlackTitle(event.Event)
}

func getSlackDescription(event models.WebhookEvent) string {
	var parts []string

	switch event.Event {
	case EventNewEpisode:
		if title, ok := event.Data["title"].(string); ok {
			parts = append(parts, fmt.Sprintf("*Episode:* %s", title))
		}
		if podcastName, ok := event.Data["podcast_name"].(string); ok {
			parts = append(parts, fmt.Sprintf("*Podcast:* %s", podcastName))
		}
	case EventDownloadComplete:
		if videoId, ok := event.Data["video_id"].(string); ok {
			parts = append(parts, fmt.Sprintf("*Video ID:* `%s`", videoId))
		}
		if duration, ok := event.Data["duration"].(string); ok {
			parts = append(parts, fmt.Sprintf("*Duration:* %s", duration))
		}
	case EventError:
		if msg, ok := event.Data["error"].(string); ok {
			parts = append(parts, fmt.Sprintf("*Error:* %s", msg))
		}
		if context, ok := event.Data["context"].(string); ok {
			parts = append(parts, fmt.Sprintf("*Context:* %s", context))
		}
	}

	// Add any remaining fields
	for key, value := range event.Data {
		if key == "title" || key == "podcast_name" || key == "video_id" ||
			key == "duration" || key == "error" || key == "context" {
			continue
		}
		parts = append(parts, fmt.Sprintf("*%s:* %v",
			strings.Title(strings.ReplaceAll(key, "_", " ")), value))
	}

	return strings.Join(parts, "\n")
}

// ProcessPendingRetries processes webhook deliveries that are due for retry
func ProcessPendingRetries() {
	deliveries := database.GetPendingWebhookRetries()

	for _, delivery := range deliveries {
		config := database.GetWebhookConfig(delivery.WebhookConfigId)
		if config == nil {
			logger.Logger.Warn().
				Int32("delivery_id", delivery.Id).
				Int32("config_id", delivery.WebhookConfigId).
				Msg("Webhook config not found for delivery")
			updateDeliveryStatus(&delivery, StatusFailed, 0, "", "Webhook config not found")
			continue
		}

		if !config.Enabled {
			logger.Logger.Debug().
				Int32("delivery_id", delivery.Id).
				Msg("Skipping retry for disabled webhook")
			updateDeliveryStatus(&delivery, StatusFailed, 0, "", "Webhook disabled")
			continue
		}

		// Parse the payload back to WebhookEvent
		var event models.WebhookEvent
		if err := json.Unmarshal([]byte(delivery.Payload), &event); err != nil {
			logger.Logger.Error().
				Err(err).
				Int32("delivery_id", delivery.Id).
				Msg("Failed to unmarshal webhook payload for retry")
			updateDeliveryStatus(&delivery, StatusFailed, 0, "", "Invalid payload")
			continue
		}

		go sendWebhookDelivery(*config, &delivery, event)
	}
}
