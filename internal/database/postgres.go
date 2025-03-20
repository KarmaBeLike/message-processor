package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/KarmaBeLike/message-processor/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib" // pgx driver for database/sql
	"github.com/pkg/errors"
)

// ConnectDB establishes a connection to the PostgreSQL database using pgxpool
func ConnectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	// Create a connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse database connection string")
	}

	// Configure the connection pool
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 2 * time.Minute

	// Create the connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	// Ping the database to verify the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "database ping failed")
	}

	slog.Info("Database connection established",
		slog.String("host", cfg.DBHost),
		slog.String("port", cfg.DBPort),
		slog.String("database", cfg.DBName),
	)
	return pool, nil
}

// RunMigrations runs database migrations using the golang-migrate library
func RunMigrations(pool *pgxpool.Pool) error {
	// For migrations, we need to create a separate sql.DB connection
	// since the migrate library works with database/sql
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pool.Config().ConnConfig.User,
		pool.Config().ConnConfig.Password,
		pool.Config().ConnConfig.Host,
		pool.Config().ConnConfig.Port,
		pool.Config().ConnConfig.Database,
	)

	// Open a separate connection using database/sql specifically for migrations
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return errors.Wrap(err, "failed to open database connection for migrations")
	}
	defer db.Close()

	// Set connection pool settings for this migration-specific connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)

	// Create a driver using the sql.DB connection
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to create migration driver")
	}

	// Create a new migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize migration")
	}

	// Run the migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "failed to apply migrations")
	}

	slog.Info("Database migrations applied successfully")
	return nil
}

// CloseDB gracefully closes the database connection pool
func CloseDB(pool *pgxpool.Pool) {
	if pool != nil {
		slog.Info("Closing database connection pool")
		pool.Close()
	}
}
