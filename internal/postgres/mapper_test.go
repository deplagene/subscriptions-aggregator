package postgres

import (
	"testing"
	"time"

	"github.com/deplagene/subaggregator/internal/postgres/sqlc"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestMapSubscription(t *testing.T) {
	t.Parallel()

	startedAt := pgtype.Date{
		Time:  time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
	endedAt := pgtype.Date{
		Time:  time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}

	row := sqlc.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       400,
		UserID:      uuid.New(),
		StartedAt:   startedAt,
		EndedAt:     endedAt,
	}

	got := mapSubscription(row)

	if got.SubscriptionID != row.ID {
		t.Errorf("mapSubscription(id) = %s, want %s", got.SubscriptionID, row.ID)
	}

	if got.Price != int64(row.Price) {
		t.Errorf("mapSubscription(price) = %d, want %d", got.Price, row.Price)
	}

	if got.StartedAt != (subscriptions.BillingDate{Month: time.March, Year: 2025}) {
		t.Errorf("mapSubscription(started_at) = %+v, want March 2025", got.StartedAt)
	}

	if got.EndedAt == nil || *got.EndedAt != (subscriptions.BillingDate{Month: time.April, Year: 2025}) {
		t.Errorf("mapSubscription(ended_at) = %+v, want April 2025", got.EndedAt)
	}
}

func TestToCreateSubscriptionParams(t *testing.T) {
	t.Parallel()

	end := subscriptions.BillingDate{Month: time.April, Year: 2025}
	sub := subscriptions.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartedAt:   subscriptions.BillingDate{Month: time.March, Year: 2025},
		EndedAt:     &end,
	}

	got := toCreateSubscriptionParams(sub)

	if got.Price != sub.Price {
		t.Errorf("toCreateSubscriptionParams(price) = %d, want %d", got.Price, sub.Price)
	}

	if !got.StartedAt.Valid || got.StartedAt.Time.Year() != 2025 || got.StartedAt.Time.Month() != time.March || got.StartedAt.Time.Day() != 1 {
		t.Errorf("toCreateSubscriptionParams(started_at) = %+v, want 2025-03-01", got.StartedAt)
	}

	if !got.EndedAt.Valid || got.EndedAt.Time.Year() != 2025 || got.EndedAt.Time.Month() != time.April || got.EndedAt.Time.Day() != 1 {
		t.Errorf("toCreateSubscriptionParams(ended_at) = %+v, want 2025-04-01", got.EndedAt)
	}
}

func TestToListSubscriptionsParams(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	filter := subscriptions.ListFilter{
		Limit:       20,
		Offset:      40,
		UserID:      &userID,
		ServiceName: "  Netflix  ",
	}

	got := toListSubscriptionsParams(filter)

	if got.Limit != 20 {
		t.Errorf("toListSubscriptionsParams(limit) = %d, want 20", got.Limit)
	}

	if got.Offset != 40 {
		t.Errorf("toListSubscriptionsParams(offset) = %d, want 40", got.Offset)
	}

	if !got.UserID.Valid {
		t.Error("toListSubscriptionsParams(user_id).Valid = false, want true")
	}

	if got.ServiceName == nil || *got.ServiceName != "Netflix" {
		t.Errorf("toListSubscriptionsParams(service_name) = %v, want Netflix", got.ServiceName)
	}
}

func TestToCalculateTotalParams(t *testing.T) {
	t.Parallel()

	filter := subscriptions.TotalFilter{
		From: subscriptions.BillingDate{Month: time.January, Year: 2025},
		To:   subscriptions.BillingDate{Month: time.March, Year: 2025},
	}

	got := toCalculateTotalParams(filter)

	if !got.From.Valid || got.From.Time.Month() != time.January || got.From.Time.Day() != 1 {
		t.Errorf("toCalculateTotalParams(from) = %+v, want 2025-01-01", got.From)
	}

	if !got.To.Valid || got.To.Time.Month() != time.March || got.To.Time.Day() != 1 {
		t.Errorf("toCalculateTotalParams(to) = %+v, want 2025-03-01", got.To)
	}
}

func TestToPgNullableDateAndOptionalString(t *testing.T) {
	t.Parallel()

	if got := toPgNullableDate(nil); got.Valid {
		t.Errorf("toPgNullableDate(nil).Valid = %t, want false", got.Valid)
	}

	if got := toOptionalString("   "); got != nil {
		t.Errorf("toOptionalString(blank) = %v, want nil", got)
	}
}
