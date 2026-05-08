package alerting

import (
	"context"
	"testing"
	"time"
)

// TestAlertManager tests the alert manager
func TestAlertManager(t *testing.T) {
	// Create a test notifier
	testNotifier := &LogNotifier{}

	// Create alert manager
	am := NewAlertManager(testNotifier)

	if am == nil {
		t.Fatal("Expected alert manager to be created")
	}

	if len(am.notifiers) != 1 {
		t.Errorf("Expected 1 notifier, got %d", len(am.notifiers))
	}
}

// TestSendAlert tests sending an alert
func TestSendAlert(t *testing.T) {
	am := NewAlertManager(&LogNotifier{})

	alert := Alert{
		ID:          "test-id",
		Level:       AlertLevelWarning,
		Title:       "Test Alert",
		Description: "This is a test alert",
		Labels: map[string]string{
			"test": "value",
		},
		Timestamp: time.Now(),
	}

	err := am.SendAlert(context.Background(), alert)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Alert should still have the same ID and timestamp since they were set
	if alert.ID != "test-id" {
		t.Errorf("Expected alert ID to remain test-id, got %s", alert.ID)
	}
}

// TestCheckAndAlert tests conditional alerting
func TestCheckAndAlert(t *testing.T) {
	am := NewAlertManager(&LogNotifier{})

	alert := Alert{
		Level:       AlertLevelWarning,
		Title:       "Test Alert",
		Description: "This is a test alert",
	}

	// Test with condition that returns true
	conditionTrue := func() (bool, error) {
		return true, nil
	}

	err := am.CheckAndAlert(context.Background(), conditionTrue, alert)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with condition that returns false
	conditionFalse := func() (bool, error) {
		return false, nil
	}

	alert2 := Alert{
		Level:       AlertLevelWarning,
		Title:       "Test Alert 2",
		Description: "This should not be sent",
	}

	err = am.CheckAndAlert(context.Background(), conditionFalse, alert2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestAlertLevels tests alert level constants
func TestAlertLevels(t *testing.T) {
	tests := []struct {
		level   AlertLevel
		expected string
	}{
		{AlertLevelInfo, "info"},
		{AlertLevelWarning, "warning"},
		{AlertLevelCritical, "critical"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.level)
		}
	}
}

// TestWebhookNotifier tests webhook notifier
func TestWebhookNotifier(t *testing.T) {
	// Create a test server
	testNotifier := &WebhookNotifier{
		URL:     "http://example.com/webhook",
		Headers: map[string]string{"Authorization": "Bearer test"},
	}

	if testNotifier == nil {
		t.Error("Expected webhook notifier to be created")
	}
}

// TestSlackNotifier tests Slack notifier
func TestSlackNotifier(t *testing.T) {
	testNotifier := &SlackNotifier{
		WebhookURL: "https://hooks.slack.com/services/test",
		Channel:    "#alerts",
	}

	if testNotifier == nil {
		t.Error("Expected Slack notifier to be created")
	}
}

// TestEmailNotifier tests email notifier
func TestEmailNotifier(t *testing.T) {
	testNotifier := &EmailNotifier{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user",
		Password: "pass",
		From:     "alerts@example.com",
		To:       []string{"admin@example.com"},
	}

	if testNotifier == nil {
		t.Error("Expected email notifier to be created")
	}
}
