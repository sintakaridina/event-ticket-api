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
	sslmode := os.Getenv("DB_SSLMODE")

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
		dbname = "payment_service"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	// Create DSN string
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	// Create custom logger for GORM
	newLogger := logger.New(
		&logrusWriter{Logger: logrus.StandardLogger()},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			PrintLog:                  true,
		},
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	maxIdleConns := 10
	maxOpenConns := 100
	connMaxLifetime := time.Hour

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return db, nil
}

// logrusWriter implements io.Writer interface for logrus
type logrusWriter struct {
	Logger *logrus.Logger
}

// Write implements io.Writer interface
func (w *logrusWriter) Printf(format string, args ...interface{}) {
	w.Logger.Infof(format, args...)
}