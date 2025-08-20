package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"subscription-service/logger"
	"subscription-service/models"
	"time"
)

// Вспомогательные функции для работы с датами
func parseMonthYear(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse("01-2006", dateStr)
}

func formatMonthYear(t time.Time) string {
	return t.Format("01-2006")
}

func toFirstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

type SubscriptionHandler struct {
	DB *sql.DB
}

func NewSubscriptionHandler(db *sql.DB) *SubscriptionHandler {
	return &SubscriptionHandler{DB: db}
}

// CreateSubscription создает новую подписку
// @Summary Create a new subscription
// @Description Create a new subscription for a user. Dates must be in MM-YYYY format.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id path string true "User UUID" Format(uuid)
// @Param service_name path string true "Service Name"
// @Param request body models.SubscriptionRequest true "Subscription Details"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{user_id}/{service_name} [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var key models.CompositeKey
	if err := c.ShouldBindUri(&key); err != nil {
		logger.Log.Warnw("Failed to bind URI parameters", "error", err, "user_id", key.UserID, "service_name", key.ServiceName)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req models.SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("Failed to bind JSON", "error", err, "user_id", key.UserID)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Debugw("Attempting to create subscription", "user_id", key.UserID, "service_name", req.ServiceName, "request", req)

	// Парсим даты
	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		logger.Log.Warnw("Invalid start_date format", "error", err, "start_date", req.StartDate)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}
	startDate = toFirstDayOfMonth(startDate)

	var endDatePtr *time.Time
	if req.EndDate != "" {
		endDate, err := parseMonthYear(req.EndDate)
		if err != nil {
			logger.Log.Warnw("Invalid end_date format", "error", err, "end_date", req.EndDate)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		endDate = toFirstDayOfMonth(endDate)
		endDatePtr = &endDate
	}

	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING service_name, price, user_id, start_date, end_date
	`

	var sub models.Subscription
	var dbEndDate sql.NullTime

	err = h.DB.QueryRow(
		query,
		req.ServiceName,
		req.Price,
		key.UserID,
		startDate,
		endDatePtr,
	).Scan(
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&startDate,
		&dbEndDate,
	)

	if err != nil {
		logger.Log.Errorw("Database error on subscription creation", "error", err, "user_id", key.UserID, "service_name", req.ServiceName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Форматируем даты для ответа
	sub.StartDate = formatMonthYear(startDate)
	if dbEndDate.Valid {
		endDate := formatMonthYear(dbEndDate.Time)
		sub.EndDate = endDate
	}

	logger.Log.Infow("Subscription created successfully", "user_id", sub.UserID, "service_name", sub.ServiceName)
	c.JSON(http.StatusCreated, sub)
}

// UpdateSubscription обновляет существующую подписку
// @Summary Update a subscription
// @Description Update an existing subscription for a user. Dates must be in MM-YYYY format.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id path string true "User UUID" Format(uuid)
// @Param service_name path string true "Service Name"
// @Param request body models.SubscriptionRequest true "Subscription Details"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{user_id}/{service_name} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	var key models.CompositeKey
	if err := c.ShouldBindUri(&key); err != nil {
		logger.Log.Warnw("Failed to bind URI parameters on update", "error", err, "user_id", key.UserID, "service_name", key.ServiceName)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req models.SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("Failed to bind JSON on update", "error", err, "user_id", key.UserID)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Debugw("Attempting to update subscription", "user_id", key.UserID, "service_name", req.ServiceName, "request", req)

	// Парсим даты
	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		logger.Log.Warnw("Invalid start_date format on update", "error", err, "start_date", req.StartDate)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}
	startDate = toFirstDayOfMonth(startDate)

	var endDatePtr *time.Time
	if req.EndDate != "" {
		endDate, err := parseMonthYear(req.EndDate)
		if err != nil {
			logger.Log.Warnw("Invalid end_date format on update", "error", err, "end_date", req.EndDate)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
			return
		}
		endDate = toFirstDayOfMonth(endDate)
		endDatePtr = &endDate
	}

	query := `
		UPDATE subscriptions 
		SET 
			price = $1,
			start_date = $2,
			end_date = $3
		WHERE user_id = $4 AND service_name = $5
		RETURNING service_name, price, user_id, start_date, end_date
	`

	var sub models.Subscription
	var dbEndDate sql.NullTime

	err = h.DB.QueryRow(
		query,
		req.Price,
		startDate,
		endDatePtr,
		key.UserID,
		req.ServiceName,
	).Scan(
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&startDate,
		&dbEndDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Infow("Subscription not found for update", "user_id", key.UserID, "service_name", req.ServiceName)
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		logger.Log.Errorw("Database error on subscription update", "error", err, "user_id", key.UserID, "service_name", req.ServiceName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Форматируем даты для ответа
	sub.StartDate = formatMonthYear(startDate)
	if dbEndDate.Valid {
		sub.EndDate = formatMonthYear(dbEndDate.Time)
	}

	logger.Log.Infow("Subscription updated successfully", "user_id", sub.UserID, "service_name", sub.ServiceName)
	c.JSON(http.StatusOK, sub)
}

// GetSubscription возвращает подписку по составному ключу
// @Summary Get a subscription
// @Description Get a specific subscription by user ID and service name.
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "User UUID" Format(uuid)
// @Param service_name path string true "Service Name"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{user_id}/{service_name} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	var key models.CompositeKey
	if err := c.ShouldBindUri(&key); err != nil {
		logger.Log.Warnw("Failed to bind URI parameters on get", "error", err, "user_id", key.UserID, "service_name", key.ServiceName)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Debugw("Fetching subscription", "user_id", key.UserID, "service_name", key.ServiceName)

	var sub models.Subscription
	query := `
		SELECT service_name, price, user_id, start_date, end_date 
		FROM subscriptions 
		WHERE user_id = $1 AND service_name = $2
	`

	row := h.DB.QueryRow(query, key.UserID, key.ServiceName)

	var startDate time.Time
	var endDate sql.NullTime
	err := row.Scan(
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&startDate,
		&endDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Infow("Subscription not found", "user_id", key.UserID, "service_name", key.ServiceName)
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		logger.Log.Errorw("Database error on subscription fetch", "error", err, "user_id", key.UserID, "service_name", key.ServiceName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	sub.StartDate = formatMonthYear(startDate)
	if endDate.Valid {
		sub.EndDate = formatMonthYear(endDate.Time)
	}

	logger.Log.Debugw("Subscription found", "user_id", sub.UserID, "service_name", sub.ServiceName)
	c.JSON(http.StatusOK, sub)
}

// DeleteSubscription удаляет подписку
// @Summary Delete a subscription
// @Description Delete a specific subscription by user ID and service name.
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "User UUID" Format(uuid)
// @Param service_name path string true "Service Name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{user_id}/{service_name} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	var key models.CompositeKey
	if err := c.ShouldBindUri(&key); err != nil {
		logger.Log.Warnw("Failed to bind URI parameters on delete", "error", err, "user_id", key.UserID, "service_name", key.ServiceName)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Debugw("Deleting subscription", "user_id", key.UserID, "service_name", key.ServiceName)

	query := `
		DELETE FROM subscriptions 
		WHERE user_id = $1 AND service_name = $2
		RETURNING service_name
	`

	var serviceName string
	err := h.DB.QueryRow(query, key.UserID, key.ServiceName).Scan(&serviceName)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Infow("Subscription not found for deletion", "user_id", key.UserID, "service_name", key.ServiceName)
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		logger.Log.Errorw("Database error on subscription deletion", "error", err, "user_id", key.UserID, "service_name", key.ServiceName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	logger.Log.Infow("Subscription deleted successfully", "user_id", key.UserID, "service_name", serviceName)
	c.JSON(http.StatusOK, gin.H{
		"message":      "subscription deleted",
		"service_name": serviceName,
		"user_id":      key.UserID,
	})
}

// ListSubscriptions возвращает список подписок с фильтрацией
// @Summary List subscriptions
// @Description Get a list of subscriptions with optional filtering by user ID and/or service name.
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by User UUID" Format(uuid)
// @Param service_name query string false "Filter by Service Name"
// @Success 200 {array} models.Subscription
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")

	logger.Log.Debugw("Listing subscriptions", "user_id", userID, "service_name", serviceName)

	var rows *sql.Rows
	var err error

	switch {
	case userID != "" && serviceName != "":
		query := `
			SELECT service_name, price, user_id, start_date, end_date
			FROM subscriptions
			WHERE user_id = $1 AND service_name = $2
		`
		rows, err = h.DB.Query(query, userID, serviceName)
	case userID != "":
		query := `
			SELECT service_name, price, user_id, start_date, end_date
			FROM subscriptions
			WHERE user_id = $1
		`
		rows, err = h.DB.Query(query, userID)
	case serviceName != "":
		query := `
			SELECT service_name, price, user_id, start_date, end_date
			FROM subscriptions
			WHERE service_name = $1
		`
		rows, err = h.DB.Query(query, serviceName)
	default:
		query := `
			SELECT service_name, price, user_id, start_date, end_date
			FROM subscriptions
		`
		rows, err = h.DB.Query(query)
	}

	if err != nil {
		logger.Log.Errorw("Database error on listing subscriptions", "error", err, "user_id", userID, "service_name", serviceName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		var startDate time.Time
		var endDate sql.NullTime
		err := rows.Scan(
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&startDate,
			&endDate,
		)
		if err != nil {
			logger.Log.Errorw("Error scanning subscription row", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		sub.StartDate = formatMonthYear(startDate)
		if endDate.Valid {
			sub.EndDate = formatMonthYear(endDate.Time)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Errorw("Error after iterating subscription rows", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	logger.Log.Debugw("Returning subscriptions list", "count", len(subscriptions))
	c.JSON(http.StatusOK, subscriptions)
}

// GetTotalCost вычисляет суммарную стоимость подписок
// @Summary Calculate total cost
// @Description Calculates the total cost of active subscriptions for a given period (inclusive). Optionally filtered by user and service. Dates must be in MM-YYYY format.
// @Tags analytics
// @Produce json
// @Param user_id query string false "Filter by User UUID" Format(uuid)
// @Param service_name query string false "Filter by Service Name"
// @Param start_date query string true "Start Period (MM-YYYY)"
// @Param end_date query string true "End Period (MM-YYYY)"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) GetTotalCost(c *gin.Context) {
	var params models.FilterParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.Log.Warnw("Failed to bind query parameters on total cost", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Debugw("Calculating total cost", "params", params)

	// Парсим даты из формата "MM-YYYY"
	startDate, err := parseMonthYear(params.StartDate)
	if err != nil {
		logger.Log.Warnw("Invalid start_date format on total cost", "error", err, "start_date", params.StartDate)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected MM-YYYY"})
		return
	}
	startDate = toFirstDayOfMonth(startDate)

	endDate, err := parseMonthYear(params.EndDate)
	if err != nil {
		logger.Log.Warnw("Invalid end_date format on total cost", "error", err, "end_date", params.EndDate)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected MM-YYYY"})
		return
	}
	endDate = toFirstDayOfMonth(endDate)

	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE start_date <= $2 AND (end_date >= $1 OR end_date IS NULL)
	`
	args := []interface{}{startDate, endDate}

	if params.UserID != "" {
		query += " AND user_id = $3"
		args = append(args, params.UserID)
		if params.ServiceName != "" {
			query += " AND service_name = $4"
			args = append(args, params.ServiceName)
		}
	} else if params.ServiceName != "" {
		query += " AND service_name = $3"
		args = append(args, params.ServiceName)
	}

	var totalCost int
	err = h.DB.QueryRow(query, args...).Scan(&totalCost)
	if err != nil {
		logger.Log.Errorw("Database error on total cost calculation", "error", err, "query", query, "args", args)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	logger.Log.Infow("Total cost calculated", "total_cost", totalCost, "user_id", params.UserID, "service_name", params.ServiceName)
	c.JSON(http.StatusOK, models.TotalCostResponse{TotalCost: totalCost})
}
