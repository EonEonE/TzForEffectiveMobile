// @title Subscription Service API
// @version 1.0
// @description Это микросервис для управления подписками пользователей.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
	"subscription-service/db"
	_ "subscription-service/docs" // ОЧЕНЬ ВАЖНО: добавьте этот импорт! Путь должен совпадать с вашим модулем.
	"subscription-service/handlers"
	"subscription-service/logger"
)

func main() {
	// 1. Инициализация логгера
	logger.InitLogger(true)
	defer logger.Log.Sync()

	// 2. Инициализация базы данных
	logger.Log.Info("Initializing database connection...")
	db.InitDB()
	defer db.DB.Close()

	// 3. Миграции и сиды
	logger.Log.Info("Running database migrations...")
	err := db.RunMigrations(db.DB)
	if err != nil {
		logger.Log.Fatalf("Migration failed: %v", err)
	}

	logger.Log.Info("Running database seeds...")
	err = db.RunSeeds(db.DB)
	if err != nil {
		logger.Log.Warnf("Seeds warning: %v", err)
	}

	// 4. Настройка роутера
	router := gin.Default()

	url := ginSwagger.URL("/swagger/doc.json") // Указываем URL к нашему swagger.json
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Middleware для логирования HTTP-запросов
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Log.Infow("HTTP Request",
			"status", param.StatusCode,
			"method", param.Method,
			"path", param.Path,
			"latency", param.Latency,
			"clientIP", param.ClientIP,
		)
		return ""
	}))
	router.Use(gin.Recovery())

	// Инициализация обработчиков
	subscriptionHandler := handlers.NewSubscriptionHandler(db.DB)

	// Маршруты для CRUDL операций
	router.POST("/subscriptions/:user_id/:service_name", subscriptionHandler.CreateSubscription)
	router.GET("/subscriptions/:user_id/:service_name", subscriptionHandler.GetSubscription)
	router.PUT("/subscriptions/:user_id/:service_name", subscriptionHandler.UpdateSubscription)
	router.DELETE("/subscriptions/:user_id/:service_name", subscriptionHandler.DeleteSubscription)
	router.GET("/subscriptions", subscriptionHandler.ListSubscriptions)
	router.GET("/subscriptions/total", subscriptionHandler.GetTotalCost)

	// Получаем порт из переменных окружения
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080" // значение по умолчанию
	}

	// Запуск сервера
	logger.Log.Infow("Server is starting", "port", port)
	if err := router.Run(":" + port); err != nil {
		logger.Log.Fatalw("Failed to start server", "error", err)
	}
}
