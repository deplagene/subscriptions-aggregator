package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/deplagene/subaggregator/internal/httpapi/dto"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/google/uuid"
)

const defaultListLimit int32 = 50

func subscriptionFromCreateRequest(req dto.CreateSubscriptionRequest) (subscriptions.Subscription, error) {
	userID, err := parseRequiredUUID(req.UserID, "user_id")
	if err != nil {
		return subscriptions.Subscription{}, err
	}

	startDate, err := parseBillingDate(req.StartDate, "start_date")
	if err != nil {
		return subscriptions.Subscription{}, err
	}

	endDate, err := parseOptionalBillingDate(req.EndDate, "end_date")
	if err != nil {
		return subscriptions.Subscription{}, err
	}

	sub := subscriptions.Subscription{
		ServiceName: strings.TrimSpace(req.ServiceName),
		Price:       req.Price,
		UserID:      userID,
		StartedAt:   startDate,
		EndedAt:     endDate,
	}

	if err := sub.Validate(); err != nil {
		return subscriptions.Subscription{}, err
	}

	return sub, nil
}

func subscriptionFromUpdateRequest(id uuid.UUID, req dto.UpdateSubscriptionRequest) (subscriptions.Subscription, error) {
	sub, err := subscriptionFromCreateRequest(dto.CreateSubscriptionRequest{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		return subscriptions.Subscription{}, err
	}

	sub.SubscriptionID = id
	return sub, nil
}

func listFilterFromRequest(r *http.Request) (subscriptions.ListFilter, error) {
	values := r.URL.Query()

	limit := defaultListLimit
	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		parsedLimit, err := strconv.ParseInt(rawLimit, 10, 32)
		if err != nil {
			return subscriptions.ListFilter{}, fmt.Errorf("limit must be a valid integer")
		}

		limit = int32(parsedLimit)
	}

	offset := int32(0)
	if rawOffset := strings.TrimSpace(values.Get("offset")); rawOffset != "" {
		parsedOffset, err := strconv.ParseInt(rawOffset, 10, 32)
		if err != nil {
			return subscriptions.ListFilter{}, fmt.Errorf("offset must be a valid integer")
		}

		offset = int32(parsedOffset)
	}

	if limit < 0 {
		return subscriptions.ListFilter{}, fmt.Errorf("limit must be greater than or equal to zero")
	}

	if offset < 0 {
		return subscriptions.ListFilter{}, fmt.Errorf("offset must be greater than or equal to zero")
	}

	var userID *uuid.UUID
	if rawUserID := strings.TrimSpace(values.Get("user_id")); rawUserID != "" {
		parsedUserID, err := parseRequiredUUID(rawUserID, "user_id")
		if err != nil {
			return subscriptions.ListFilter{}, err
		}

		userID = &parsedUserID
	}

	return subscriptions.ListFilter{
		Limit:       limit,
		Offset:      offset,
		UserID:      userID,
		ServiceName: strings.TrimSpace(values.Get("service_name")),
	}, nil
}

func totalFilterFromRequest(r *http.Request) (subscriptions.TotalFilter, error) {
	values := r.URL.Query()

	from, err := parseBillingDate(values.Get("from"), "from")
	if err != nil {
		return subscriptions.TotalFilter{}, err
	}

	to, err := parseBillingDate(values.Get("to"), "to")
	if err != nil {
		return subscriptions.TotalFilter{}, err
	}

	if to.Before(from) {
		return subscriptions.TotalFilter{}, fmt.Errorf("to must be after or equal to from")
	}

	var userID *uuid.UUID
	if rawUserID := strings.TrimSpace(values.Get("user_id")); rawUserID != "" {
		parsedUserID, err := parseRequiredUUID(rawUserID, "user_id")
		if err != nil {
			return subscriptions.TotalFilter{}, err
		}

		userID = &parsedUserID
	}

	return subscriptions.TotalFilter{
		From:        from,
		To:          to,
		UserID:      userID,
		ServiceName: strings.TrimSpace(values.Get("service_name")),
	}, nil
}

func subscriptionResponseFromModel(sub subscriptions.Subscription) dto.SubscriptionResponse {
	resp := dto.SubscriptionResponse{
		ID:          sub.SubscriptionID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   formatBillingDate(sub.StartedAt),
	}

	if sub.EndedAt != nil {
		endDate := formatBillingDate(*sub.EndedAt)
		resp.EndDate = &endDate
	}

	return resp
}

func parseSubscriptionID(raw string) (uuid.UUID, error) {
	return parseRequiredUUID(raw, "subscription_id")
}

func parseRequiredUUID(raw string, field string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return uuid.Nil, fmt.Errorf("%s must not be empty", field)
	}

	id, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s must be a valid uuid", field)
	}

	return id, nil
}

func parseBillingDate(raw string, field string) (subscriptions.BillingDate, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return subscriptions.BillingDate{}, fmt.Errorf("%s must not be empty", field)
	}

	parsedTime, err := time.Parse("01-2006", trimmed)
	if err != nil {
		return subscriptions.BillingDate{}, fmt.Errorf("%s must be in MM-YYYY format", field)
	}

	return subscriptions.BillingDate{
		Month: parsedTime.Month(),
		Year:  parsedTime.Year(),
	}, nil
}

func parseOptionalBillingDate(raw *string, field string) (*subscriptions.BillingDate, error) {
	if raw == nil {
		return nil, nil
	}

	date, err := parseBillingDate(*raw, field)
	if err != nil {
		return nil, err
	}

	return &date, nil
}

func formatBillingDate(date subscriptions.BillingDate) string {
	return fmt.Sprintf("%02d-%04d", int(date.Month), date.Year)
}
