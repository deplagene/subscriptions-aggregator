package subscriptions

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSubscriptionsServiceCreate(t *testing.T) {
	t.Parallel()

	var gotSub Subscription
	repo := &mockSubscriptionRepository{
		create: func(_ context.Context, sub Subscription) (uuid.UUID, error) {
			gotSub = sub
			return uuid.New(), nil
		},
	}

	service := NewSubscriptionsService(repo)
	sub := Subscription{
		ServiceName: "  Netflix  ",
		Price:       400,
		UserID:      uuid.New(),
		StartedAt:   BillingDate{Month: time.March, Year: 2025},
	}

	createdID, err := service.Create(context.Background(), sub)
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	if gotSub.ServiceName != "Netflix" {
		t.Errorf("Create(service_name) = %q, want %q", gotSub.ServiceName, "Netflix")
	}

	if createdID == uuid.Nil {
		t.Error("Create() = uuid.Nil, want generated id")
	}
}

func TestSubscriptionsServiceGetByIDNotFound(t *testing.T) {
	t.Parallel()

	service := NewSubscriptionsService(&mockSubscriptionRepository{
		getByID: func(context.Context, uuid.UUID) (*Subscription, error) {
			return nil, ErrSubscriptionNotFound
		},
	})

	_, err := service.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, ErrSubscriptionNotFound)
	}
}

func TestSubscriptionsServiceListValidatesFilter(t *testing.T) {
	t.Parallel()

	service := NewSubscriptionsService(&mockSubscriptionRepository{})

	_, err := service.List(context.Background(), ListFilter{
		Limit:  -1,
		Offset: -2,
	})
	if err == nil {
		t.Fatal("List() error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "internal.subscriptions.Service.List") {
		t.Fatalf("List() error = %q, want operation prefix", err.Error())
	}
}

func TestSubscriptionsServiceDeleteNotFound(t *testing.T) {
	t.Parallel()

	service := NewSubscriptionsService(&mockSubscriptionRepository{
		delete: func(context.Context, uuid.UUID) error {
			return ErrSubscriptionNotFound
		},
	})

	err := service.Delete(context.Background(), uuid.New())
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, ErrSubscriptionNotFound)
	}
}

func TestSubscriptionsServiceCalculateTotalValidatesPeriod(t *testing.T) {
	t.Parallel()

	service := NewSubscriptionsService(&mockSubscriptionRepository{})

	_, err := service.CalculateTotal(context.Background(), TotalFilter{
		From: BillingDate{Month: time.April, Year: 2025},
		To:   BillingDate{Month: time.March, Year: 2025},
	})
	if err == nil {
		t.Fatal("CalculateTotal() error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "internal.subscriptions.Service.CalculateTotal") {
		t.Fatalf("CalculateTotal() error = %q, want operation prefix", err.Error())
	}
}

type mockSubscriptionRepository struct {
	calculateTotal func(context.Context, TotalFilter) (int64, error)
	create         func(context.Context, Subscription) (uuid.UUID, error)
	delete         func(context.Context, uuid.UUID) error
	getByID        func(context.Context, uuid.UUID) (*Subscription, error)
	list           func(context.Context, ListFilter) ([]Subscription, error)
	update         func(context.Context, Subscription) error
}

func (m *mockSubscriptionRepository) Create(ctx context.Context, sub Subscription) (uuid.UUID, error) {
	if m.create == nil {
		return uuid.Nil, nil
	}

	return m.create(ctx, sub)
}

func (m *mockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	if m.getByID == nil {
		return nil, nil
	}

	return m.getByID(ctx, id)
}

func (m *mockSubscriptionRepository) List(ctx context.Context, filter ListFilter) ([]Subscription, error) {
	if m.list == nil {
		return nil, nil
	}

	return m.list(ctx, filter)
}

func (m *mockSubscriptionRepository) Update(ctx context.Context, sub Subscription) error {
	if m.update == nil {
		return nil
	}

	return m.update(ctx, sub)
}

func (m *mockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.delete == nil {
		return nil
	}

	return m.delete(ctx, id)
}

func (m *mockSubscriptionRepository) CalculateTotal(ctx context.Context, filter TotalFilter) (int64, error) {
	if m.calculateTotal == nil {
		return 0, nil
	}

	return m.calculateTotal(ctx, filter)
}
