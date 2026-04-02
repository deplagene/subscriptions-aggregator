package httpapi

import (
	"net/http"

	"github.com/deplagene/subaggregator/internal/httpapi/dto"
	"github.com/deplagene/subaggregator/internal/subscriptions"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	r.Get("/healthz", h.healthz)
	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", h.createSubscription)
		r.Get("/", h.listSubscriptions)
		r.Get("/total", h.calculateTotal)
		r.Get("/{id}", h.getSubscriptionByID)
		r.Put("/{id}", h.putSubscription)
		r.Patch("/{id}", h.patchSubscription)
		r.Delete("/{id}", h.deleteSubscription)
	})
}

// healthz godoc
// @Summary Проверка состояния сервиса
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// createSubscription godoc
// @Summary Создать подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body dto.CreateSubscriptionRequest true "Данные новой подписки"
// @Success 201 {object} dto.CreateSubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())

	var req dto.CreateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		logger.Error("internal.httpapi.Handler.createSubscription", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
		return
	}

	sub, err := subscriptionFromCreateRequest(req)
	if err != nil {
		logger.Error("internal.httpapi.Handler.createSubscription", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
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

// getSubscriptionByID godoc
// @Summary Получить подписку по идентификатору
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *Handler) getSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error("internal.httpapi.Handler.getSubscriptionByID", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
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

// listSubscriptions godoc
// @Summary Получить список подписок
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Лимит"
// @Param offset query int false "Смещение"
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Success 200 {object} dto.ListSubscriptionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	filter, err := listFilterFromRequest(r)
	if err != nil {
		logger.Error("internal.httpapi.Handler.listSubscriptions", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
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

// putSubscription godoc
// @Summary Обновить подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body dto.UpdateSubscriptionRequest true "Новые данные подписки"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *Handler) putSubscription(w http.ResponseWriter, r *http.Request) {
	h.updateSubscription(w, r)
}

// patchSubscription godoc
// @Summary Частично обновить подписку
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body dto.UpdateSubscriptionRequest true "Новые данные подписки"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [patch]
func (h *Handler) patchSubscription(w http.ResponseWriter, r *http.Request) {
	h.updateSubscription(w, r)
}

func (h *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err, "subscription_id", id)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
		return
	}

	sub, err := subscriptionFromUpdateRequest(id, req)
	if err != nil {
		logger.Error("internal.httpapi.Handler.updateSubscription", "error", err, "subscription_id", id)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
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

// deleteSubscription godoc
// @Summary Удалить подписку
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	id, err := parseSubscriptionID(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error("internal.httpapi.Handler.deleteSubscription", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
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

// calculateTotal godoc
// @Summary Рассчитать общую стоимость подписок за период
// @Tags subscriptions
// @Produce json
// @Param from query string true "Начало периода в формате MM-YYYY"
// @Param to query string true "Конец периода в формате MM-YYYY"
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Success 200 {object} dto.TotalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/total [get]
func (h *Handler) calculateTotal(w http.ResponseWriter, r *http.Request) {
	logger := logging.L(r.Context())
	filter, err := totalFilterFromRequest(r)
	if err != nil {
		logger.Error("internal.httpapi.Handler.calculateTotal", "error", err)
		writeError(w, http.StatusBadRequest, userErrorMessage(err))
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
