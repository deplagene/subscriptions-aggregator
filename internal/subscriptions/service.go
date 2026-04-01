package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
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

	normalizeSubscription(&sub)

	if err := sub.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.repo.Create(ctx, sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *subscriptionsService) GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	const op = "internal.subscriptions.Service.GetByID"

	if id == uuid.Nil {
		return nil, fmt.Errorf("%s: subscription_id must not be nil", op)
	}

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sub, nil
}

func (s *subscriptionsService) List(ctx context.Context, filter ListFilter) ([]Subscription, error) {
	const op = "internal.subscriptions.Service.List"

	normalizeListFilter(&filter)

	if err := validateListFilter(filter); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	subs, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return subs, nil
}

func (s *subscriptionsService) Update(ctx context.Context, sub Subscription) error {
	const op = "internal.subscriptions.Service.Update"

	if sub.SubscriptionID == uuid.Nil {
		return fmt.Errorf("%s: subscription_id must not be nil", op)
	}

	normalizeSubscription(&sub)

	if err := sub.Validate(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *subscriptionsService) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "internal.subscriptions.Service.Delete"

	if id == uuid.Nil {
		return fmt.Errorf("%s: subscription_id must not be nil", op)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *subscriptionsService) CalculateTotal(ctx context.Context, filter TotalFilter) (int64, error) {
	const op = "internal.subscriptions.Service.CalculateTotal"

	normalizeTotalFilter(&filter)

	if err := validateTotalFilter(filter); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	total, err := s.repo.CalculateTotal(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

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
