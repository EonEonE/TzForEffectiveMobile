package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"subscription-service/logger"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Загружаем переменные из .env файла
	err := godotenv.Load()
	if err != nil {
		logger.Log.Info("No .env file found. Using system environment variables")
	}

	// Читаем конфигурацию из переменных окружения
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "ForTZ")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Log.Fatal("Invalid DB_PORT: ", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	logger.Log.Infow("Connecting to database", "host", host, "port", port, "dbname", dbname)

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Log.Fatal("Failed to open database connection: ", err)
	}

	// Устанавливаем максимальное количество открытых соединений
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Ping()
	if err != nil {
		logger.Log.Fatal("Failed to ping database: ", err)
	}

	logger.Log.Info("Database connection established successfully")
}

// Вспомогательная функция для получения переменных окружения
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func RunMigrations(db *sql.DB) error {
	migrationPath := filepath.Join("db", "migrations", "001_init.up.sql")
	logger.Log.Infow("Applying migration", "file", migrationPath)

	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		logger.Log.Errorw("Could not read migration file", "error", err, "path", migrationPath)
		return fmt.Errorf("could not read migration file: %v", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		logger.Log.Errorw("Migration execution failed", "error", err)
		return fmt.Errorf("migration failed: %v", err)
	}

	logger.Log.Info("Migration applied successfully!")
	return nil
}

func RunSeeds(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&count)
	if err != nil {
		logger.Log.Errorw("Could not check data count", "error", err)
		return fmt.Errorf("could not check data count: %v", err)
	}

	if count == 0 {
		seedsPath := filepath.Join("db", "seeds", "001_initial_data.sql")
		logger.Log.Infow("Applying seeds", "file", seedsPath)

		sqlBytes, err := os.ReadFile(seedsPath)
		if err != nil {
			logger.Log.Errorw("Could not read seeds file", "error", err, "path", seedsPath)
			return fmt.Errorf("could not read seeds file: %v", err)
		}

		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			logger.Log.Errorw("Seeds execution failed", "error", err)
			return fmt.Errorf("seeds failed: %v", err)
		}

		logger.Log.Infof("Seeds applied successfully! Added %d records", count)
	} else {
		logger.Log.Infof("Table already contains %d records, skipping seeds", count)
	}

	return nil
}
