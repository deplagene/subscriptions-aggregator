package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestHandlerCreateSubscription(t *testing.T) {
	t.Parallel()

	var gotSub subscriptions.Subscription
	service := &mockSubscriptionsService{
		create: func(_ context.Context, sub subscriptions.Subscription) (uuid.UUID, error) {
			gotSub = sub
			return uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil
		},
	}

	handler := NewHandler(service)
	body := []byte(`{"service_name":" Netflix ","price":400,"user_id":"22222222-2222-2222-2222-222222222222","start_date":"03-2025","end_date":"04-2025"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	newTestRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/subscriptions status = %d, want %d", rec.Code, http.StatusCreated)
	}

	if gotSub.ServiceName != "Netflix" {
		t.Errorf("POST /api/v1/subscriptions service_name = %q, want %q", gotSub.ServiceName, "Netflix")
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal(create response) error = %v, want nil", err)
	}

	if resp["id"] != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("POST /api/v1/subscriptions id = %q, want %q", resp["id"], "11111111-1111-1111-1111-111111111111")
	}
}

func TestHandlerGetSubscriptionByIDNotFound(t *testing.T) {
	t.Parallel()

	service := &mockSubscriptionsService{
		getByID: func(context.Context, uuid.UUID) (*subscriptions.Subscription, error) {
			return nil, subscriptions.ErrSubscriptionNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(service)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/v1/subscriptions/{id} status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerListSubscriptions(t *testing.T) {
	t.Parallel()

	service := &mockSubscriptionsService{
		list: func(context.Context, subscriptions.ListFilter) ([]subscriptions.Subscription, error) {
			endDate := subscriptions.BillingDate{Month: time.April, Year: 2025}
			return []subscriptions.Subscription{
				{
					SubscriptionID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ServiceName:    "Netflix",
					Price:          400,
					UserID:         uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					StartedAt:      subscriptions.BillingDate{Month: time.March, Year: 2025},
					EndedAt:        &endDate,
				},
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(service)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/subscriptions status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandlerCalculateTotal(t *testing.T) {
	t.Parallel()

	service := &mockSubscriptionsService{
		calculateTotal: func(context.Context, subscriptions.TotalFilter) (int64, error) {
			return 1200, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/total?from=03-2025&to=05-2025", nil)
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(service)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/subscriptions/total status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]int64
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal(total response) error = %v, want nil", err)
	}

	if resp["total"] != 1200 {
		t.Errorf("GET /api/v1/subscriptions/total total = %d, want %d", resp["total"], int64(1200))
	}
}

func TestHandlerUpdateSubscription(t *testing.T) {
	t.Parallel()

	var gotSub subscriptions.Subscription
	service := &mockSubscriptionsService{
		update: func(_ context.Context, sub subscriptions.Subscription) error {
			gotSub = sub
			return nil
		},
	}

	body := []byte(`{"service_name":"Spotify","price":500,"user_id":"22222222-2222-2222-2222-222222222222","start_date":"03-2025"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(service)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT /api/v1/subscriptions/{id} status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if gotSub.SubscriptionID != uuid.MustParse("11111111-1111-1111-1111-111111111111") {
		t.Errorf("PUT /api/v1/subscriptions/{id} subscription_id = %s, want %s", gotSub.SubscriptionID, "11111111-1111-1111-1111-111111111111")
	}
}

func TestHandlerDeleteSubscription(t *testing.T) {
	t.Parallel()

	service := &mockSubscriptionsService{
		delete: func(context.Context, uuid.UUID) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(service)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/v1/subscriptions/{id} status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

type mockSubscriptionsService struct {
	calculateTotal func(context.Context, subscriptions.TotalFilter) (int64, error)
	create         func(context.Context, subscriptions.Subscription) (uuid.UUID, error)
	delete         func(context.Context, uuid.UUID) error
	getByID        func(context.Context, uuid.UUID) (*subscriptions.Subscription, error)
	list           func(context.Context, subscriptions.ListFilter) ([]subscriptions.Subscription, error)
	update         func(context.Context, subscriptions.Subscription) error
}

func (m *mockSubscriptionsService) Create(ctx context.Context, sub subscriptions.Subscription) (uuid.UUID, error) {
	if m.create == nil {
		return uuid.Nil, nil
	}

	return m.create(ctx, sub)
}

func (m *mockSubscriptionsService) GetByID(ctx context.Context, id uuid.UUID) (*subscriptions.Subscription, error) {
	if m.getByID == nil {
		return nil, nil
	}

	return m.getByID(ctx, id)
}

func (m *mockSubscriptionsService) List(ctx context.Context, filter subscriptions.ListFilter) ([]subscriptions.Subscription, error) {
	if m.list == nil {
		return nil, nil
	}

	return m.list(ctx, filter)
}

func (m *mockSubscriptionsService) Update(ctx context.Context, sub subscriptions.Subscription) error {
	if m.update == nil {
		return nil
	}

	return m.update(ctx, sub)
}

func (m *mockSubscriptionsService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.delete == nil {
		return nil
	}

	return m.delete(ctx, id)
}

func (m *mockSubscriptionsService) CalculateTotal(ctx context.Context, filter subscriptions.TotalFilter) (int64, error) {
	if m.calculateTotal == nil {
		return 0, nil
	}

	return m.calculateTotal(ctx, filter)
}

var _ subscriptions.ISubscriptionsService = (*mockSubscriptionsService)(nil)

func newTestRouter(handler *Handler) chi.Router {
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func TestHandlerBadSubscriptionID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(&mockSubscriptionsService{})).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/v1/subscriptions/not-a-uuid status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerInternalError(t *testing.T) {
	t.Parallel()

	service := &mockSubscriptionsService{
		delete: func(context.Context, uuid.UUID) error {
			return errors.New("boom")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	newTestRouter(NewHandler(service)).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("DELETE /api/v1/subscriptions/{id} status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
