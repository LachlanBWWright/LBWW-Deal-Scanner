package notifications

import (
	"context"
	"log"
)

type AppNotification struct {
	Kind     string             `json:"kind"` // "deal" or "error"
	Source   string             `json:"source"`
	Title    string             `json:"title,omitempty"`
	Url      string             `json:"url,omitempty"`
	Price    *float64           `json:"price,omitempty"`
	ImageUrl *string            `json:"imageUrl,omitempty"`
	Message  string             `json:"message,omitempty"`
	Stack    string             `json:"stack,omitempty"`
	Query    *NotificationQuery `json:"query,omitempty"`
	Tags     []string           `json:"tags,omitempty"`
}

type NotificationQuery struct {
	Type   string `json:"type"`
	Id     string `json:"id"`
	DmOnly bool   `json:"dmOnly,omitempty"`
}

type NotificationProvider interface {
	Name() string
	IsEnabled() bool
	Send(ctx context.Context, notification AppNotification) error
}

type NotificationService struct {
	providers []NotificationProvider
}

func NewNotificationService(providers []NotificationProvider) *NotificationService {
	return &NotificationService{providers: providers}
}

func (s *NotificationService) GetProviderNames() []string {
	names := make([]string, 0, len(s.providers))
	for _, p := range s.providers {
		names = append(names, p.Name())
	}
	return names
}

func (s *NotificationService) Publish(ctx context.Context, notification AppNotification) {
	for _, provider := range s.providers {
		if provider.IsEnabled() {
			go func(p NotificationProvider) {
				if err := p.Send(ctx, notification); err != nil {
					log.Printf("Notification provider %q failed: %v", p.Name(), err)
				}
			}(provider)
		}
	}
}
