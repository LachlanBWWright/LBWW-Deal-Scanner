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
	jobs      chan notificationJob
}

type notificationJob struct {
	provider     NotificationProvider
	ctx          context.Context
	notification AppNotification
}

const notificationWorkerCount = 4
const notificationQueueSize = 128

func NewNotificationService(providers []NotificationProvider) *NotificationService {
	service := &NotificationService{
		providers: append([]NotificationProvider(nil), providers...),
		jobs:      make(chan notificationJob, notificationQueueSize),
	}
	for range notificationWorkerCount {
		go service.runWorker()
	}
	return service
}

func (s *NotificationService) runWorker() {
	for job := range s.jobs {
		if err := job.provider.Send(job.ctx, job.notification); err != nil {
			log.Printf("Notification provider %q failed: %v", job.provider.Name(), err)
		}
	}
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
			select {
			case s.jobs <- notificationJob{provider: provider, ctx: ctx, notification: notification}:
			default:
				log.Printf("Notification queue full; dropping notification for provider %q", provider.Name())
			}
		}
	}
}
