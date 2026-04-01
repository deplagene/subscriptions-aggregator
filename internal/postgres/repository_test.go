package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/deplagene/subaggregator/internal/postgres/sqlc"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestRepositoryCreate(t *testing.T) {
	t.Parallel()

	sub := subscriptions.Subscription{
		SubscriptionID: uuid.New(),
		ServiceName:    "Netflix",
		Price:          400,
		UserID:         uuid.New(),
		StartedAt:      subscriptions.BillingDate{Month: time.March, Year: 2025},
		EndedAt:        &subscriptions.BillingDate{Month: time.April, Year: 2025},
	}

	mock := &mockQuerier{
		createSubscription: func(_ context.Context, arg sqlc.CreateSubscriptionParams) (sqlc.Subscription, error) {
			if arg.ID != sub.SubscriptionID {
				t.Fatalf("CreateSubscription(id) = %s, want %s", arg.ID, sub.SubscriptionID)
			}

			if arg.Price != int32(sub.Price) {
				t.Fatalf("CreateSubscription(price) = %d, want %d", arg.Price, sub.Price)
			}

			return sqlc.Subscription{
				ID:          arg.ID,
				ServiceName: arg.ServiceName,
				Price:       arg.Price,
				UserID:      arg.UserID,
				StartedAt:   arg.StartedAt,
				EndedAt:     arg.EndedAt,
			}, nil
		},
	}

	repo := &Repository{q: mock}

	got, err := repo.Create(context.Background(), sub)
	if err != nil {
		t.Fatalf("Repository.Create() error = %v, want nil", err)
	}

	if got.SubscriptionID != sub.SubscriptionID {
		t.Errorf("Repository.Create().SubscriptionID = %s, want %s", got.SubscriptionID, sub.SubscriptionID)
	}

	if got.Price != sub.Price {
		t.Errorf("Repository.Create().Price = %d, want %d", got.Price, sub.Price)
	}
}

func TestRepositoryList(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	serviceName := "Netflix"
	mock := &mockQuerier{
		listSubscriptions: func(_ context.Context, arg sqlc.ListSubscriptionsParams) ([]sqlc.Subscription, error) {
			if arg.Limit != 10 {
				t.Fatalf("ListSubscriptions(limit) = %d, want 10", arg.Limit)
			}

			if arg.Offset != 20 {
				t.Fatalf("ListSubscriptions(offset) = %d, want 20", arg.Offset)
			}

			if !arg.UserID.Valid {
				t.Fatal("ListSubscriptions(user_id).Valid = false, want true")
			}

			if arg.ServiceName == nil || *arg.ServiceName != serviceName {
				t.Fatalf("ListSubscriptions(service_name) = %v, want %q", arg.ServiceName, serviceName)
			}

			return []sqlc.Subscription{
				{
					ID:          uuid.New(),
					ServiceName: serviceName,
					Price:       400,
					UserID:      userID,
					StartedAt: pgtype.Date{
						Time:  time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
						Valid: true,
					},
				},
			}, nil
		},
	}

	repo := &Repository{q: mock}

	items, err := repo.List(context.Background(), subscriptions.ListFilter{
		Limit:       10,
		Offset:      20,
		UserID:      &userID,
		ServiceName: serviceName,
	})
	if err != nil {
		t.Fatalf("Repository.List() error = %v, want nil", err)
	}

	if len(items) != 1 {
		t.Fatalf("Repository.List() len = %d, want 1", len(items))
	}

	if items[0].ServiceName != serviceName {
		t.Errorf("Repository.List()[0].ServiceName = %q, want %q", items[0].ServiceName, serviceName)
	}
}

func TestRepositoryDelete(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &Repository{
			q: &mockQuerier{
				deleteSubscription: func(_ context.Context, gotID uuid.UUID) (int64, error) {
					if gotID != id {
						t.Fatalf("DeleteSubscription(id) = %s, want %s", gotID, id)
					}

					return 1, nil
				},
			},
		}

		if err := repo.Delete(context.Background(), id); err != nil {
			t.Fatalf("Repository.Delete() error = %v, want nil", err)
		}
	})

	t.Run("no rows", func(t *testing.T) {
		t.Parallel()

		repo := &Repository{
			q: &mockQuerier{
				deleteSubscription: func(_ context.Context, _ uuid.UUID) (int64, error) {
					return 0, nil
				},
			},
		}

		err := repo.Delete(context.Background(), id)
		if err == nil {
			t.Fatal("Repository.Delete() error = nil, want non-nil")
		}
	})
}

func TestRepositoryCalculateTotal(t *testing.T) {
	t.Parallel()

	filter := subscriptions.TotalFilter{
		From: subscriptions.BillingDate{Month: time.January, Year: 2025},
		To:   subscriptions.BillingDate{Month: time.March, Year: 2025},
	}

	mock := &mockQuerier{
		calculateSubscriptionsTotal: func(_ context.Context, arg sqlc.CalculateSubscriptionsTotalParams) (int64, error) {
			if !arg.From.Valid || arg.From.Time.Month() != time.January {
				t.Fatalf("CalculateSubscriptionsTotal(from) = %+v, want January 2025", arg.From)
			}

			if !arg.To.Valid || arg.To.Time.Month() != time.March {
				t.Fatalf("CalculateSubscriptionsTotal(to) = %+v, want March 2025", arg.To)
			}

			return 1200, nil
		},
	}

	repo := &Repository{q: mock}

	total, err := repo.CalculateTotal(context.Background(), filter)
	if err != nil {
		t.Fatalf("Repository.CalculateTotal() error = %v, want nil", err)
	}

	if total != 1200 {
		t.Errorf("Repository.CalculateTotal() = %d, want 1200", total)
	}
}

func TestRepositoryWrapsErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	repo := &Repository{
		q: &mockQuerier{
			getSubscriptionByID: func(_ context.Context, _ uuid.UUID) (sqlc.Subscription, error) {
				return sqlc.Subscription{}, wantErr
			},
		},
	}

	_, err := repo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Repository.GetByID() error = %v, want wrapped %v", err, wantErr)
	}

	if !strings.Contains(err.Error(), "internal.postgres.Repository.GetByID") {
		t.Fatalf("Repository.GetByID() error = %q, want operation prefix", err.Error())
	}
}

func TestRepositoryDeleteNoRowsHasOp(t *testing.T) {
	t.Parallel()

	repo := &Repository{
		q: &mockQuerier{
			deleteSubscription: func(_ context.Context, _ uuid.UUID) (int64, error) {
				return 0, nil
			},
		},
	}

	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("Repository.Delete() error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "internal.postgres.Repository.Delete") {
		t.Fatalf("Repository.Delete() error = %q, want operation prefix", err.Error())
	}
}

type mockQuerier struct {
	calculateSubscriptionsTotal func(context.Context, sqlc.CalculateSubscriptionsTotalParams) (int64, error)
	createSubscription          func(context.Context, sqlc.CreateSubscriptionParams) (sqlc.Subscription, error)
	deleteSubscription          func(context.Context, uuid.UUID) (int64, error)
	getSubscriptionByID         func(context.Context, uuid.UUID) (sqlc.Subscription, error)
	listSubscriptions           func(context.Context, sqlc.ListSubscriptionsParams) ([]sqlc.Subscription, error)
	updateSubscription          func(context.Context, sqlc.UpdateSubscriptionParams) (sqlc.Subscription, error)
}

func (m *mockQuerier) CalculateSubscriptionsTotal(ctx context.Context, arg sqlc.CalculateSubscriptionsTotalParams) (int64, error) {
	if m.calculateSubscriptionsTotal == nil {
		return 0, nil
	}

	return m.calculateSubscriptionsTotal(ctx, arg)
}

func (m *mockQuerier) CreateSubscription(ctx context.Context, arg sqlc.CreateSubscriptionParams) (sqlc.Subscription, error) {
	if m.createSubscription == nil {
		return sqlc.Subscription{}, nil
	}

	return m.createSubscription(ctx, arg)
}

func (m *mockQuerier) DeleteSubscription(ctx context.Context, id uuid.UUID) (int64, error) {
	if m.deleteSubscription == nil {
		return 0, nil
	}

	return m.deleteSubscription(ctx, id)
}

func (m *mockQuerier) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (sqlc.Subscription, error) {
	if m.getSubscriptionByID == nil {
		return sqlc.Subscription{}, nil
	}

	return m.getSubscriptionByID(ctx, id)
}

func (m *mockQuerier) ListSubscriptions(ctx context.Context, arg sqlc.ListSubscriptionsParams) ([]sqlc.Subscription, error) {
	if m.listSubscriptions == nil {
		return nil, nil
	}

	return m.listSubscriptions(ctx, arg)
}

func (m *mockQuerier) UpdateSubscription(ctx context.Context, arg sqlc.UpdateSubscriptionParams) (sqlc.Subscription, error) {
	if m.updateSubscription == nil {
		return sqlc.Subscription{}, nil
	}

	return m.updateSubscription(ctx, arg)
}
