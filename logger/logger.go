package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func InitLogger(development bool) {
	var config zap.Config

	if development {
		// Для разработки: читабельный вывод в консоль
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// Для продакшена: структурированный JSON
		config = zap.NewProductionConfig()
	}

	// Настройка уровня логирования (можно вынести в переменные окружения)
	config.Level.SetLevel(zapcore.DebugLevel)

	// Строим логгер
	baseLogger, err := config.Build()
	if err != nil {
		panic(err) // Если логгер не создался, падаем сразу
	}
	defer baseLogger.Sync() // Важно: сбрасываем буферизованные логи при выходе

	// Создаем SugaredLogger для удобного логирования в формате printf
	Log = baseLogger.Sugar()

	// Первое информационное сообщение
	Log.Info("Logger initialized successfully")
}
