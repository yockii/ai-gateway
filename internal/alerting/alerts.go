package alerting

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

type Alert struct {
	ID          string            `json:"id"`
	Level       AlertLevel        `json:"level"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time         `json:"timestamp"`
	Resolved    bool              `json:"resolved"`
}

type Notifier interface {
	Send(ctx context.Context, alert Alert) error
}

type AlertManager struct {
	notifiers []Notifier
}

// NewAlertManager creates a new alert manager with the given notifiers
func NewAlertManager(notifiers ...Notifier) *AlertManager {
	return &AlertManager{
		notifiers: notifiers,
	}
}

// SendAlert sends an alert to all registered notifiers
func (am *AlertManager) SendAlert(ctx context.Context, alert Alert) error {
	// Generate ID if not present
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	// Set timestamp if not set
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	// Send to all notifiers
	for _, notifier := range am.notifiers {
		if err := notifier.Send(ctx, alert); err != nil {
			// Log error but continue with other notifiers
			// TODO: Add proper logging here
			continue
		}
	}
	return nil
}

// CheckAndAlert checks a condition and sends an alert if triggered
func (am *AlertManager) CheckAndAlert(
	ctx context.Context,
	condition func() (bool, error),
	alert Alert,
) error {
	triggered, err := condition()
	if err != nil {
		return err
	}

	if triggered {
		alert.Timestamp = time.Now()
		if alert.ID == "" {
			alert.ID = uuid.New().String()
		}
		return am.SendAlert(ctx, alert)
	}

	return nil
}
