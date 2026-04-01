package postgres

import (
	"strings"
	"time"

	"github.com/deplagene/subaggregator/internal/postgres/sqlc"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func mapSubscription(row sqlc.Subscription) *subscriptions.Subscription {
	subscription := subscriptions.Subscription{
		SubscriptionID: row.ID,
		ServiceName:    row.ServiceName,
		Price:          int64(row.Price),
		UserID:         row.UserID,
		StartedAt:      fromPgDate(row.StartedAt),
	}

	if row.EndedAt.Valid {
		endedAt := fromPgDate(row.EndedAt)
		subscription.EndedAt = &endedAt
	}

	return &subscription
}

func mapSubscriptions(rows []sqlc.Subscription) []subscriptions.Subscription {
	items := make([]subscriptions.Subscription, 0, len(rows))
	for _, row := range rows {
		items = append(items, *mapSubscription(row))
	}

	return items
}

func toCreateSubscriptionParams(sub subscriptions.Subscription) sqlc.CreateSubscriptionParams {
	return sqlc.CreateSubscriptionParams{
		ServiceName: sub.ServiceName,
		Price:       int32(sub.Price),
		UserID:      sub.UserID,
		StartedAt:   toPgDate(sub.StartedAt),
		EndedAt:     toPgNullableDate(sub.EndedAt),
	}
}

func toUpdateSubscriptionParams(sub subscriptions.Subscription) sqlc.UpdateSubscriptionParams {
	return sqlc.UpdateSubscriptionParams{
		ID:          sub.SubscriptionID,
		ServiceName: sub.ServiceName,
		Price:       int32(sub.Price),
		UserID:      sub.UserID,
		StartedAt:   toPgDate(sub.StartedAt),
		EndedAt:     toPgNullableDate(sub.EndedAt),
	}
}

func toListSubscriptionsParams(filter subscriptions.ListFilter) sqlc.ListSubscriptionsParams {
	return sqlc.ListSubscriptionsParams{
		Limit:       filter.Limit,
		Offset:      filter.Offset,
		UserID:      toPgUUID(filter.UserID),
		ServiceName: toOptionalString(filter.ServiceName),
	}
}

func toCalculateTotalParams(filter subscriptions.TotalFilter) sqlc.CalculateSubscriptionsTotalParams {
	return sqlc.CalculateSubscriptionsTotalParams{
		From:        toPgDate(filter.From),
		To:          toPgDate(filter.To),
		UserID:      toPgUUID(filter.UserID),
		ServiceName: toOptionalString(filter.ServiceName),
	}
}

func toPgDate(date subscriptions.BillingDate) pgtype.Date {
	return pgtype.Date{
		Time:  time.Date(date.Year, date.Month, 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func toPgNullableDate(date *subscriptions.BillingDate) pgtype.Date {
	if date == nil {
		return pgtype.Date{}
	}

	return toPgDate(*date)
}

func fromPgDate(date pgtype.Date) subscriptions.BillingDate {
	return subscriptions.BillingDate{
		Month: date.Time.Month(),
		Year:  date.Time.Year(),
	}
}

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{
		Bytes: [16]byte(*id),
		Valid: true,
	}
}

func toOptionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
