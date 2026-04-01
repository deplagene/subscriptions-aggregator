package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/deplagene/subaggregator/internal/postgres/sqlc"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var _ subscriptions.ISubscriptionRepository = (*subscriptionRepository)(nil)

// subscriptionRepository - PostgreSQL-реализация subscriptions.subscriptionRepository.
type subscriptionRepository struct {
	q sqlc.Querier
}

func NewRepository(db sqlc.DBTX) *subscriptionRepository {
	return &subscriptionRepository{q: sqlc.New(db)}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub subscriptions.Subscription) (uuid.UUID, error) {
	const op = "internal.postgres.Repository.Create"

	id, err := r.q.CreateSubscription(ctx, toCreateSubscriptionParams(sub))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.GetByID"

	row, err := r.q.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, subscriptions.ErrSubscriptionNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mapSubscription(row), nil
}

func (r *subscriptionRepository) List(ctx context.Context, filter subscriptions.ListFilter) ([]subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.List"

	rows, err := r.q.ListSubscriptions(ctx, toListSubscriptionsParams(filter))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mapSubscriptions(rows), nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub subscriptions.Subscription) error {
	const op = "internal.postgres.Repository.Update"

	rowsAffected, err := r.q.UpdateSubscription(ctx, toUpdateSubscriptionParams(sub))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, subscriptions.ErrSubscriptionNotFound)
	}

	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "internal.postgres.Repository.Delete"

	rowsAffected, err := r.q.DeleteSubscription(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, subscriptions.ErrSubscriptionNotFound)
	}

	return nil
}

func (r *subscriptionRepository) CalculateTotal(ctx context.Context, filter subscriptions.TotalFilter) (int64, error) {
	const op = "internal.postgres.Repository.CalculateTotal"

	total, err := r.q.CalculateSubscriptionsTotal(ctx, toCalculateTotalParams(filter))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return total, nil
}
