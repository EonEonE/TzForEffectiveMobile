package models

type Subscription struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`         // Формат "MM-YYYY"
	EndDate     string `json:"end_date,omitempty"` // Формат "MM-YYYY"
}

type SubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Price       int    `json:"price" binding:"required,min=0"`
	StartDate   string `json:"start_date" binding:"required"` // Формат "MM-YYYY"
	EndDate     string `json:"end_date"`                      // Формат "MM-YYYY"
}

type CompositeKey struct {
	UserID      string `uri:"user_id" binding:"required,uuid"`
	ServiceName string `uri:"service_name" binding:"required"`
}

type FilterParams struct {
	UserID      string `form:"user_id"`
	ServiceName string `form:"service_name"`
	StartDate   string `form:"start_date"` // Формат "MM-YYYY"
	EndDate     string `form:"end_date"`   // Формат "MM-YYYY"
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}
