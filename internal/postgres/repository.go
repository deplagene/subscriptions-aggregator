package postgres

import (
	"context"
	"fmt"

	"github.com/deplagene/subaggregator/internal/postgres/sqlc"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
)

var _ subscriptions.Repository = (*Repository)(nil)

// Repository - PostgreSQL-реализация subscriptions.Repository.
type Repository struct {
	q sqlc.Querier
}

func NewRepository(db sqlc.DBTX) *Repository {
	return &Repository{q: sqlc.New(db)}
}

func (r *Repository) Create(ctx context.Context, sub subscriptions.Subscription) (subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.Create"

	row, err := r.q.CreateSubscription(ctx, toCreateSubscriptionParams(sub))
	if err != nil {
		return subscriptions.Subscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return mapSubscription(row), nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.GetByID"

	row, err := r.q.GetSubscriptionByID(ctx, id)
	if err != nil {
		return subscriptions.Subscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return mapSubscription(row), nil
}

func (r *Repository) List(ctx context.Context, filter subscriptions.ListFilter) ([]subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.List"

	rows, err := r.q.ListSubscriptions(ctx, toListSubscriptionsParams(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mapSubscriptions(rows), nil
}

func (r *Repository) Update(ctx context.Context, sub subscriptions.Subscription) (subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.Update"

	row, err := r.q.UpdateSubscription(ctx, toUpdateSubscriptionParams(sub))
	if err != nil {
		return subscriptions.Subscription{}, fmt.Errorf("%s: %w", op, err)
	}

	return mapSubscription(row), nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "internal.postgres.Repository.Delete"

	rowsAffected, err := r.q.DeleteSubscription(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows affected", op)
	}

	return nil
}

func (r *Repository) CalculateTotal(ctx context.Context, filter subscriptions.TotalFilter) (int64, error) {
	const op = "internal.postgres.Repository.CalculateTotal"

	total, err := r.q.CalculateSubscriptionsTotal(ctx, toCalculateTotalParams(filter))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return total, nil
}
