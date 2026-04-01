package httpapi

import (
	"net/http"

	"github.com/deplagene/subaggregator/internal/httpapi/dto"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/go-chi/chi/v5"
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
	var req dto.CreateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := subscriptionFromCreateRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.service.Create(r.Context(), sub)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.CreateSubscriptionResponse{ID: id.String()})
}

func (h *Handler) getSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, subscriptionResponseFromModel(*sub))
}

func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	filter, err := listFilterFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.service.List(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	respItems := make([]dto.SubscriptionResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, subscriptionResponseFromModel(item))
	}

	writeJSON(w, http.StatusOK, dto.ListSubscriptionsResponse{Items: respItems})
}

func (h *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := subscriptionFromUpdateRequest(id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Update(r.Context(), sub); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) calculateTotal(w http.ResponseWriter, r *http.Request) {
	filter, err := totalFilterFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	total, err := h.service.CalculateTotal(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.TotalResponse{Total: total})
}
