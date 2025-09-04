package config

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB() (*gorm.DB, error) {
	// Get database configuration from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	// Set default values if not provided
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if dbname == "" {
		dbname = "notification_service"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	// Create DSN string
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	// Configure GORM logger
	gormLogger := logger.New(
		&logrusWriter{Logger: logrus.StandardLogger()},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			PrintLog:                  true,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set connection pool settings
	maxConnections := 100
	maxIdleConnections := 10
	connectionMaxLifetime := time.Hour

	sqlDB.SetMaxOpenConns(maxConnections)
	sqlDB.SetMaxIdleConns(maxIdleConnections)
	sqlDB.SetConnMaxLifetime(connectionMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("Connected to PostgreSQL database")
	return db, nil
}

// logrusWriter adapts logrus to GORM's logger interface
type logrusWriter struct {
	Logger *logrus.Logger
}

// Printf implements the logger.Writer interface
func (w *logrusWriter) Printf(format string, args ...interface{}) {
	w.Logger.Tracef(format, args...)
}