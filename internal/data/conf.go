package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type DBConfig struct {
	Machine  string `json:"machine"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type FilesConfig struct {
	SmallFolder   string `json:"smallFolder"`
	RegularFolder string `json:"regularFolder"`
	JobFolder     string `json:"jobFolder"`
	Dimension     int    `json:"dimension"`
}

type ConfigData struct {
	Auth            string      `json:"auth"`
	TokenTimeToLive int         `json:"tokenTimeToLive"`
	Access          string      `json:"access"`
	DB              DBConfig    `json:"db"`
	ServerAddress   string      `json:"serverAddress"`
	Context         string      `json:"context"`
	Files           FilesConfig `json:"files"`
}

// LoadConfig reads the CONFIG_FILE environment variable, loads the configuration file,
// and returns a ConfigData struct with logged values.
func LoadConfig() (*ConfigData, error) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		return nil, fmt.Errorf("CONFIG_FILE environment variable not set")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigData
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Log all config values one by one
	slog.Info("Configuration loaded",
		"auth_url", config.Auth,
		"token_ttl_seconds", config.TokenTimeToLive,
		"access_scope", config.Access,
		"server_address", config.ServerAddress,
		"context_path", config.Context,
	)

	slog.Info("Database configuration",
		"machine", config.DB.Machine,
		"username", config.DB.Username,
		"database", config.DB.Database,
		"password_filled", config.DB.Password != "",
	)

	slog.Info("Files configuration",
		"small_folder", config.Files.SmallFolder,
		"regular_folder", config.Files.RegularFolder,
		"job_folder", config.Files.JobFolder,
		"dimension", config.Files.Dimension,
	)

	return &config, nil
}

// ConnectDB establishes a connection to the database and verifies connectivity.
func ConnectDB(dbConfig DBConfig) (*sql.DB, error) {
	// DSN format for MariaDB: username:password@tcp(host:port)/database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Machine,
		dbConfig.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify connectivity with a simple query
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("Database connection established successfully")
	return db, nil
}
