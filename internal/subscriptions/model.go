package subscriptions

import (
	"time"

	"github.com/google/uuid"
)

// Subscription - подписка пользователя
type Subscription struct {
	SubscriptionID uuid.UUID
	ServiceName    string
	Price          int64 // целое число рублей
	UserID         uuid.UUID
	// дата начала
	StartedAt time.Time
	// дата окончания
	EndedAt *time.Time
}
