package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/deplagene/subaggregator/internal/postgres/sqlc"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/theartofdevel/logging"
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
	logger := logging.L(ctx)

	id, err := r.q.CreateSubscription(ctx, toCreateSubscriptionParams(sub))
	if err != nil {
		logger.Error(op, "error", err, "service_name", sub.ServiceName, "user_id", sub.UserID)
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "subscription_id", id)
	return id, nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.GetByID"
	logger := logging.L(ctx)

	row, err := r.q.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error(op, "error", subscriptions.ErrSubscriptionNotFound, "subscription_id", id)
			return nil, fmt.Errorf("%s: %w", op, subscriptions.ErrSubscriptionNotFound)
		}

		logger.Error(op, "error", err, "subscription_id", id)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "subscription_id", id)
	return mapSubscription(row), nil
}

func (r *subscriptionRepository) List(ctx context.Context, filter subscriptions.ListFilter) ([]subscriptions.Subscription, error) {
	const op = "internal.postgres.Repository.List"
	logger := logging.L(ctx)

	rows, err := r.q.ListSubscriptions(ctx, toListSubscriptionsParams(filter))
	if err != nil {
		logger.Error(op, "error", err, "limit", filter.Limit, "offset", filter.Offset, "service_name", filter.ServiceName)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	subs := mapSubscriptions(rows)
	logger.Info(op, "count", len(subs), "limit", filter.Limit, "offset", filter.Offset, "service_name", filter.ServiceName)
	return subs, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub subscriptions.Subscription) error {
	const op = "internal.postgres.Repository.Update"
	logger := logging.L(ctx)

	rowsAffected, err := r.q.UpdateSubscription(ctx, toUpdateSubscriptionParams(sub))
	if err != nil {
		logger.Error(op, "error", err, "subscription_id", sub.SubscriptionID)
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		logger.Error(op, "error", subscriptions.ErrSubscriptionNotFound, "subscription_id", sub.SubscriptionID)
		return fmt.Errorf("%s: %w", op, subscriptions.ErrSubscriptionNotFound)
	}

	logger.Info(op, "subscription_id", sub.SubscriptionID)
	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "internal.postgres.Repository.Delete"
	logger := logging.L(ctx)

	rowsAffected, err := r.q.DeleteSubscription(ctx, id)
	if err != nil {
		logger.Error(op, "error", err, "subscription_id", id)
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		logger.Error(op, "error", subscriptions.ErrSubscriptionNotFound, "subscription_id", id)
		return fmt.Errorf("%s: %w", op, subscriptions.ErrSubscriptionNotFound)
	}

	logger.Info(op, "subscription_id", id)
	return nil
}

func (r *subscriptionRepository) CalculateTotal(ctx context.Context, filter subscriptions.TotalFilter) (int64, error) {
	const op = "internal.postgres.Repository.CalculateTotal"
	logger := logging.L(ctx)

	total, err := r.q.CalculateSubscriptionsTotal(ctx, toCalculateTotalParams(filter))
	if err != nil {
		logger.Error(op, "error", err, "service_name", filter.ServiceName)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "total", total, "service_name", filter.ServiceName)
	return total, nil
}
