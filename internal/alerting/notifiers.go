package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	From         string
	To           []string
}

// Send sends an alert via email
func (n *EmailNotifier) Send(ctx context.Context, alert Alert) error {
	subject := fmt.Sprintf("[%s] %s", alert.Level, alert.Title)
	body := fmt.Sprintf("Description: %s\nLabels: %v\nTime: %s", alert.Description, alert.Labels, alert.Timestamp.Format(time.RFC3339))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		n.From,
		fmt.Sprintf("%v", n.To),
		subject,
		body,
	)

	auth := smtp.PlainAuth("", n.Username, n.Password, n.SMTPHost)
	return smtp.SendMail(
		fmt.Sprintf("%s:%d", n.SMTPHost, n.SMTPPort),
		auth,
		n.From,
		n.To,
		[]byte(msg),
	)
}

// WebhookNotifier Webhook 通知器
type WebhookNotifier struct {
	URL     string
	Headers map[string]string
}

// Send sends an alert via webhook
func (n *WebhookNotifier) Send(ctx context.Context, alert Alert) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.URL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SlackNotifier Slack 通知器
type SlackNotifier struct {
	WebhookURL string
	Channel    string
}

// Send sends an alert via Slack webhook
func (n *SlackNotifier) Send(ctx context.Context, alert Alert) error {
	color := map[AlertLevel]string{
		AlertLevelInfo:     "good",
		AlertLevelWarning:  "warning",
		AlertLevelCritical: "danger",
	}[alert.Level]

	payload := map[string]interface{}{
		"channel": n.Channel,
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  alert.Title,
				"text":   alert.Description,
				"fields": []map[string]interface{}{
					map[string]interface{}{"title": "Level", "value": string(alert.Level)},
					map[string]interface{}{"title": "Time", "value": alert.Timestamp.Format(time.RFC3339)},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// LogNotifier 日志通知器（用于开发/测试）
type LogNotifier struct{}

// Send sends an alert to logs
func (n *LogNotifier) Send(ctx context.Context, alert Alert) error {
	// This is a simple implementation that just logs the alert
	// In production, you might want to use the logging package
	fmt.Printf("[ALERT] %s: %s - %s\n", alert.Level, alert.Title, alert.Description)
	return nil
}
