package dto

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Netflix"`
	Price       int64   `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"11111111-1111-1111-1111-111111111111"`
	StartDate   string  `json:"start_date" example:"03-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"05-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Spotify"`
	Price       int64   `json:"price" example:"500"`
	UserID      string  `json:"user_id" example:"11111111-1111-1111-1111-111111111111"`
	StartDate   string  `json:"start_date" example:"04-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"06-2025"`
}

type CreateSubscriptionResponse struct {
	ID string `json:"id" example:"22222222-2222-2222-2222-222222222222"`
}

type SubscriptionResponse struct {
	ID          string  `json:"id" example:"22222222-2222-2222-2222-222222222222"`
	ServiceName string  `json:"service_name" example:"Netflix"`
	Price       int64   `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"11111111-1111-1111-1111-111111111111"`
	StartDate   string  `json:"start_date" example:"03-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"05-2025"`
}

type ListSubscriptionsResponse struct {
	Items []SubscriptionResponse `json:"items"`
}
