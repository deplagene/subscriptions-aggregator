package subscriptions

import (
	"context"

	"github.com/google/uuid"
)

// Repository - интерфейс репозитория подписок.
type Repository interface {
	Create(ctx context.Context, sub Subscription) (Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (Subscription, error)
	List(ctx context.Context, filter ListFilter) ([]Subscription, error)
	Update(ctx context.Context, sub Subscription) (Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotal(ctx context.Context, filter TotalFilter) (int64, error)
}
