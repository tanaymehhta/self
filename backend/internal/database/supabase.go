package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/tanaymehhta/self/backend/pkg/config"
	appLogger "github.com/tanaymehhta/self/backend/pkg/logger"
)

type DB struct {
	*gorm.DB
	logger *appLogger.Logger
}

func NewSupabaseConnection(cfg *config.Config, log *appLogger.Logger) (*DB, error) {
	// Configure GORM logger
	var gormLogger logger.Interface
	if cfg.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Warn)
	}

	// Connect to PostgreSQL (local or Supabase)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(10)                 // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Connected to PostgreSQL database")

	return &DB{
		DB:     db,
		logger: log.WithComponent("database"),
	}, nil
}

func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *DB) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

func (d *DB) GetStats() sql.DBStats {
	sqlDB, _ := d.DB.DB()
	return sqlDB.Stats()
}

// Transaction helper
func (d *DB) WithTransaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// Query helpers with logging
func (d *DB) LogQuery(query string, args ...interface{}) {
	d.logger.Debug("Executing query", "query", query, "args", args)
}