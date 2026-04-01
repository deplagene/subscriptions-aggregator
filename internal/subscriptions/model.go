package subscriptions

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Subscription - подписка пользователя.
type Subscription struct {
	SubscriptionID uuid.UUID
	ServiceName    string
	Price          int64
	UserID         uuid.UUID
	StartedAt      BillingDate
	EndedAt        *BillingDate
}

// BillingDate - дата в формате месяц/год.
type BillingDate struct {
	Month time.Month
	Year  int
}

// ListFilter - фильтр списка подписок.
type ListFilter struct {
	Limit       int32
	Offset      int32
	UserID      *uuid.UUID
	ServiceName string
}

// TotalFilter - фильтр подсчета общей стоимости подписок.
type TotalFilter struct {
	From        BillingDate
	To          BillingDate
	UserID      *uuid.UUID
	ServiceName string
}

func (d *BillingDate) SetMonth(month time.Month) {
	d.Month = month
}

func (d *BillingDate) SetYear(year int) {
	d.Year = year
}

func (d BillingDate) Validate() error {
	var errs []error

	if d.Month < time.January || d.Month > time.December {
		errs = append(errs, errors.New("month must be between 1 and 12"))
	}

	if d.Year <= 0 {
		errs = append(errs, errors.New("year must be greater than zero"))
	}

	return errors.Join(errs...)
}

// Before - возвращает true, если d < other.
func (d BillingDate) Before(other BillingDate) bool {
	if d.Year != other.Year {
		return d.Year < other.Year
	}

	return d.Month < other.Month
}

func (s *Subscription) SetServiceName(serviceName string) {
	s.ServiceName = strings.TrimSpace(serviceName)
}

func (s *Subscription) SetPrice(price int64) {
	s.Price = price
}

func (s *Subscription) SetUserID(userID uuid.UUID) {
	s.UserID = userID
}

func (s *Subscription) SetStartedAt(startedAt BillingDate) {
	s.StartedAt = startedAt
}

func (s *Subscription) SetEndedAt(endedAt *BillingDate) {
	if endedAt == nil {
		s.EndedAt = nil
		return
	}

	value := *endedAt
	s.EndedAt = &value
}

func (s Subscription) Validate() error {
	var errs []error

	if strings.TrimSpace(s.ServiceName) == "" {
		errs = append(errs, errors.New("service_name must not be empty"))
	}

	if s.Price <= 0 {
		errs = append(errs, errors.New("price must be greater than zero"))
	}

	if s.UserID == uuid.Nil {
		errs = append(errs, errors.New("user_id must not be nil"))
	}

	if err := s.StartedAt.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("started_at: %w", err))
	}

	if s.EndedAt != nil {
		if err := s.EndedAt.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("ended_at: %w", err))
		} else if s.EndedAt.Before(s.StartedAt) {
			errs = append(errs, errors.New("ended_at must be after or equal to started_at"))
		}
	}

	return errors.Join(errs...)
}
