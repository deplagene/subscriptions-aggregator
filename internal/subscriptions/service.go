package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/theartofdevel/logging"
)

var _ ISubscriptionsService = (*subscriptionsService)(nil)

type ISubscriptionsService interface {
	Create(ctx context.Context, sub Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	List(ctx context.Context, filter ListFilter) ([]Subscription, error)
	Update(ctx context.Context, sub Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotal(ctx context.Context, filter TotalFilter) (int64, error)
}

type subscriptionsService struct {
	repo ISubscriptionRepository
}

func NewSubscriptionsService(repo ISubscriptionRepository) *subscriptionsService {
	return &subscriptionsService{
		repo: repo,
	}
}

func (s *subscriptionsService) Create(ctx context.Context, sub Subscription) (uuid.UUID, error) {
	const op = "internal.subscriptions.Service.Create"
	logger := logging.L(ctx)

	normalizeSubscription(&sub)

	if err := sub.Validate(); err != nil {
		logger.Error(op, "error", err)
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.repo.Create(ctx, sub)
	if err != nil {
		logger.Error(op, "error", err, "service_name", sub.ServiceName, "user_id", sub.UserID)
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "subscription_id", id, "service_name", sub.ServiceName, "user_id", sub.UserID)
	return id, nil
}

func (s *subscriptionsService) GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	const op = "internal.subscriptions.Service.GetByID"
	logger := logging.L(ctx)

	if id == uuid.Nil {
		err := fmt.Errorf("%s: subscription_id must not be nil", op)
		logger.Error(op, "error", err)
		return nil, err
	}

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error(op, "error", err, "subscription_id", id)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "subscription_id", id)
	return sub, nil
}

func (s *subscriptionsService) List(ctx context.Context, filter ListFilter) ([]Subscription, error) {
	const op = "internal.subscriptions.Service.List"
	logger := logging.L(ctx)

	normalizeListFilter(&filter)

	if err := validateListFilter(filter); err != nil {
		logger.Error(op, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		logger.Error(op, "error", err, "limit", filter.Limit, "offset", filter.Offset, "service_name", filter.ServiceName)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "count", len(subs), "limit", filter.Limit, "offset", filter.Offset, "service_name", filter.ServiceName)
	return subs, nil
}

func (s *subscriptionsService) Update(ctx context.Context, sub Subscription) error {
	const op = "internal.subscriptions.Service.Update"
	logger := logging.L(ctx)

	if sub.SubscriptionID == uuid.Nil {
		err := fmt.Errorf("%s: subscription_id must not be nil", op)
		logger.Error(op, "error", err)
		return err
	}

	normalizeSubscription(&sub)

	if err := sub.Validate(); err != nil {
		logger.Error(op, "error", err, "subscription_id", sub.SubscriptionID)
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		logger.Error(op, "error", err, "subscription_id", sub.SubscriptionID)
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "subscription_id", sub.SubscriptionID)
	return nil
}

func (s *subscriptionsService) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "internal.subscriptions.Service.Delete"
	logger := logging.L(ctx)

	if id == uuid.Nil {
		err := fmt.Errorf("%s: subscription_id must not be nil", op)
		logger.Error(op, "error", err)
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error(op, "error", err, "subscription_id", id)
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "subscription_id", id)
	return nil
}

func (s *subscriptionsService) CalculateTotal(ctx context.Context, filter TotalFilter) (int64, error) {
	const op = "internal.subscriptions.Service.CalculateTotal"
	logger := logging.L(ctx)

	normalizeTotalFilter(&filter)

	if err := validateTotalFilter(filter); err != nil {
		logger.Error(op, "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	total, err := s.repo.CalculateTotal(ctx, filter)
	if err != nil {
		logger.Error(op, "error", err, "service_name", filter.ServiceName)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	logger.Info(op, "total", total, "service_name", filter.ServiceName)
	return total, nil
}

func normalizeSubscription(sub *Subscription) {
	sub.SetServiceName(sub.ServiceName)
}

func normalizeListFilter(filter *ListFilter) {
	filter.ServiceName = strings.TrimSpace(filter.ServiceName)
}

func normalizeTotalFilter(filter *TotalFilter) {
	filter.ServiceName = strings.TrimSpace(filter.ServiceName)
}

func validateListFilter(filter ListFilter) error {
	var errs []error

	if filter.Limit < 0 {
		errs = append(errs, errors.New("limit must be greater than or equal to zero"))
	}

	if filter.Offset < 0 {
		errs = append(errs, errors.New("offset must be greater than or equal to zero"))
	}

	if filter.UserID != nil && *filter.UserID == uuid.Nil {
		errs = append(errs, errors.New("user_id must not be nil"))
	}

	return errors.Join(errs...)
}

func validateTotalFilter(filter TotalFilter) error {
	var errs []error

	if err := filter.From.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("from: %w", err))
	}

	if err := filter.To.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("to: %w", err))
	}

	if filter.To.Before(filter.From) {
		errs = append(errs, errors.New("to must be after or equal to from"))
	}

	if filter.UserID != nil && *filter.UserID == uuid.Nil {
		errs = append(errs, errors.New("user_id must not be nil"))
	}

	return errors.Join(errs...)
}
