package httpapi

import (
	"net/http"

	"github.com/deplagene/subaggregator/internal/httpapi/dto"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/go-chi/chi/v5"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	service subscriptions.ISubscriptionsService
}

func NewHandler(service subscriptions.ISubscriptionsService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", h.healthz)
	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", h.createSubscription)
		r.Get("/", h.listSubscriptions)
		r.Get("/total", h.calculateTotal)
		r.Get("/{id}", h.getSubscriptionByID)
		r.Put("/{id}", h.updateSubscription)
		r.Patch("/{id}", h.updateSubscription)
		r.Delete("/{id}", h.deleteSubscription)
	})
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())

	var req dto.CreateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		logger.Error("internal.httpapi.Handler.createSubscription", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := subscriptionFromCreateRequest(req)
	if err != nil {
		logger.Error("internal.httpapi.Handler.createSubscription", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.service.Create(r.Context(), sub)
	if err != nil {
		logger.Error("internal.httpapi.Handler.createSubscription", "error", err)
		writeServiceError(w, err)
		return
	}

	logger.Info("internal.httpapi.Handler.createSubscription", "subscription_id", id)
	writeJSON(w, http.StatusCreated, dto.CreateSubscriptionResponse{ID: id.String()})
}

func (h *Handler) getSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error("internal.httpapi.Handler.getSubscriptionByID", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("internal.httpapi.Handler.getSubscriptionByID", "error", err, "subscription_id", id)
		writeServiceError(w, err)
		return
	}

	logger.Info("internal.httpapi.Handler.getSubscriptionByID", "subscription_id", id)
	writeJSON(w, http.StatusOK, subscriptionResponseFromModel(*sub))
}

func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	filter, err := listFilterFromRequest(r)
	if err != nil {
		logger.Error("internal.httpapi.Handler.listSubscriptions", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.service.List(r.Context(), filter)
	if err != nil {
		logger.Error("internal.httpapi.Handler.listSubscriptions", "error", err)
		writeServiceError(w, err)
		return
	}

	respItems := make([]dto.SubscriptionResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, subscriptionResponseFromModel(item))
	}

	logger.Info("internal.httpapi.Handler.listSubscriptions", "count", len(respItems))
	writeJSON(w, http.StatusOK, dto.ListSubscriptionsResponse{Items: respItems})
}

func (h *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err, "subscription_id", id)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := subscriptionFromUpdateRequest(id, req)
	if err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err, "subscription_id", id)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Update(r.Context(), sub); err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err, "subscription_id", id)
		writeServiceError(w, err)
		return
	}

	logger.Info("internal.httpapi.Handler.updateSubscription", "subscription_id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error("internal.httpapi.Handler.deleteSubscription", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		logger.Error("internal.httpapi.Handler.deleteSubscription", "error", err, "subscription_id", id)
		writeServiceError(w, err)
		return
	}

	logger.Info("internal.httpapi.Handler.deleteSubscription", "subscription_id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) calculateTotal(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	filter, err := totalFilterFromRequest(r)
	if err != nil {
		logger.Error("internal.httpapi.Handler.calculateTotal", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	total, err := h.service.CalculateTotal(r.Context(), filter)
	if err != nil {
		logger.Error("internal.httpapi.Handler.calculateTotal", "error", err)
		writeServiceError(w, err)
		return
	}

	logger.Info("internal.httpapi.Handler.calculateTotal", "total", total)
	writeJSON(w, http.StatusOK, dto.TotalResponse{Total: total})
}
