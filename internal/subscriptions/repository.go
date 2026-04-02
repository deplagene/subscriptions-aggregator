package subscriptions

import (
	"context"

	"github.com/google/uuid"
)

// ISubscriptionRepository - интерфейс репозитория подписок.
type ISubscriptionRepository interface {
	Create(ctx context.Context, sub Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	List(ctx context.Context, filter ListFilter) ([]Subscription, error)
	Update(ctx context.Context, sub Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotal(ctx context.Context, filter TotalFilter) (int64, error)
}
