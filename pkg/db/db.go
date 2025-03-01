package dbpkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// DBConfig holds database configuration values
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	SSLMode  string
}

// Dbconn holds the database connection pool
type Dbconn struct {
	Db     *pgxpool.Pool
	Config DBConfig
}

// loadEnv loads environment variables from the .env file
func loadEnv() DBConfig {
	_ = godotenv.Load()

	return DBConfig{
		User:     getEnv("PG_USER", "postgres"),
		Password: getEnv("PG_PASSWORD", "postgres"),
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnv("PG_PORT", "5432"),
		Name:     getEnv("PG_DB", "random"),
		SSLMode:  getEnv("PG_SSLMODE", "disable"),
	}
}

// InitDB initializes the database connection
func InitDB() (*Dbconn, error) {
	config := loadEnv()

	defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.SSLMode)
	targetDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Name, config.SSLMode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if database exists
	conn, err := pgx.Connect(ctx, defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	defer conn.Close(ctx)

	exists, err := databaseExists(ctx, conn, config.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %v", err)
	}

	if !exists {
		log.Printf("Database %s does not exist. Creating it...", config.Name)
		if err := createDatabase(ctx, conn, config.Name); err != nil {
			return nil, fmt.Errorf("failed to create database: %v", err)
		}
		log.Println("Database created successfully!")
	}

	// Setup connection pool
	poolConfig, err := pgxpool.ParseConfig(targetDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.HealthCheckPeriod = 30 * time.Second
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target database: %v", err)
	}

	log.Println("Connected to PostgreSQL database successfully.")
	return &Dbconn{Db: pool, Config: config}, nil
}

func databaseExists(ctx context.Context, conn *pgx.Conn, name string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", name).Scan(&exists)
	return exists, err
}

func createDatabase(ctx context.Context, conn *pgx.Conn, name string) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", name))
	return err
}

func (dbconn *Dbconn) GetDB() *pgxpool.Pool {
	return dbconn.Db
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}
