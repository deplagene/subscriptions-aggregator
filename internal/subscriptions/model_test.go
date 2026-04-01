package subscriptions

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDateSetters(t *testing.T) {
	t.Parallel()

	var date BillingDate

	date.SetMonth(time.July)
	date.SetYear(2025)

	if date.Month != time.July {
		t.Errorf("Date.SetMonth() month = %d, want %d", date.Month, time.July)
	}

	if date.Year != 2025 {
		t.Errorf("Date.SetYear() year = %d, want %d", date.Year, 2025)
	}
}

func TestDateValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		date    BillingDate
		wantErr bool
	}{
		{
			name:    "valid",
			date:    BillingDate{Month: time.July, Year: 2025},
			wantErr: false,
		},
		{
			name:    "invalid month",
			date:    BillingDate{Month: time.Month(13), Year: 2025},
			wantErr: true,
		},
		{
			name:    "invalid year",
			date:    BillingDate{Month: time.July, Year: 0},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.date.Validate()
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("Date.Validate(%+v) error = %v, want error presence = %t", tc.date, err, tc.wantErr)
			}
		})
	}
}

func TestSubscriptionSetters(t *testing.T) {
	t.Parallel()

	subscription := Subscription{}
	startedAt := BillingDate{Month: time.March, Year: 2025}
	endedAt := BillingDate{Month: time.April, Year: 2025}
	userID := uuid.New()

	subscription.SetServiceName("  Yandex Plus  ")
	subscription.SetPrice(400)
	subscription.SetUserID(userID)
	subscription.SetStartedAt(startedAt)
	subscription.SetEndedAt(&endedAt)

	if subscription.ServiceName != "Yandex Plus" {
		t.Errorf("Subscription.SetServiceName() service name = %q, want %q", subscription.ServiceName, "Yandex Plus")
	}

	if subscription.Price != 400 {
		t.Errorf("Subscription.SetPrice() price = %d, want %d", subscription.Price, 400)
	}

	if subscription.UserID != userID {
		t.Errorf("Subscription.SetUserID() user id = %s, want %s", subscription.UserID, userID)
	}

	if subscription.StartedAt != startedAt {
		t.Errorf("Subscription.SetStartedAt() started_at = %+v, want %+v", subscription.StartedAt, startedAt)
	}

	if subscription.EndedAt == nil || *subscription.EndedAt != endedAt {
		t.Errorf("Subscription.SetEndedAt() ended_at = %+v, want %+v", subscription.EndedAt, endedAt)
	}
}

func TestSubscriptionValidate(t *testing.T) {
	t.Parallel()

	valid := Subscription{
		SubscriptionID: uuid.New(),
		ServiceName:    "Netflix",
		Price:          400,
		UserID:         uuid.New(),
		StartedAt:      BillingDate{Month: time.March, Year: 2025},
		EndedAt:        &BillingDate{Month: time.April, Year: 2025},
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("Subscription.Validate(valid) error = %v, want nil", err)
	}

	invalid := Subscription{
		ServiceName: "",
		Price:       0,
		UserID:      uuid.Nil,
		StartedAt:   BillingDate{Month: time.Month(13), Year: 0},
		EndedAt:     &BillingDate{Month: time.January, Year: 2024},
	}

	err := invalid.Validate()
	if err == nil {
		t.Fatal("Subscription.Validate(invalid) error = nil, want non-nil")
	}

	errorText := err.Error()
	for _, want := range []string{
		"service_name must not be empty",
		"price must be greater than zero",
		"user_id must not be nil",
		"started_at:",
	} {
		if !strings.Contains(errorText, want) {
			t.Errorf("Subscription.Validate(invalid) error = %q, want substring %q", errorText, want)
		}
	}
}

func TestSubscriptionValidatePeriod(t *testing.T) {
	t.Parallel()

	subscription := Subscription{
		SubscriptionID: uuid.New(),
		ServiceName:    "Netflix",
		Price:          400,
		UserID:         uuid.New(),
		StartedAt:      BillingDate{Month: time.March, Year: 2025},
		EndedAt:        &BillingDate{Month: time.January, Year: 2025},
	}

	err := subscription.Validate()
	if err == nil {
		t.Fatal("Subscription.Validate(invalid period) error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "ended_at must be after or equal to started_at") {
		t.Errorf(
			"Subscription.Validate(invalid period) error = %q, want substring %q",
			err.Error(),
			"ended_at must be after or equal to started_at",
		)
	}
}
