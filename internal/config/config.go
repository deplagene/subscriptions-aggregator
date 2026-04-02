package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAPIHost              = "0.0.0.0"
	defaultAPIPort              = 8080
	defaultAPIReadTimeout       = 30 * time.Second
	defaultAPIReadHeaderTimeout = 10 * time.Second
	defaultAPIWriteTimeout      = 30 * time.Second
	defaultAPIShutdownTimeout   = 10 * time.Second
	defaultLogLevel             = "info"
	defaultLogIsJSON            = true
	defaultDBHost               = "postgres"
	defaultDBPort               = 5432
	defaultDBUser               = "postgres"
	defaultDBPassword           = "postgres"
	defaultDBName               = "subaggregator"
	defaultDBSSLMode            = "disable"
	defaultDBMigrationPath      = "/app/migrations"
)

type Config struct {
	API     API
	Logging Logging
	DB      DB
}

type API struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
}

type Logging struct {
	Level  string
	IsJSON bool
}

type DB struct {
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	SSLMode       string
	MigrationPath string
}

func Load() (*Config, error) {
	const op = "internal.config.Load"

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s: load .env: %w", op, err)
	}

	cfg, err := loadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return cfg, nil
}

func (c Config) String() string {
	return fmt.Sprintf("API: {%s}, Logging: {%s}, DB: {%s}", c.API.String(), c.Logging.String(), c.DB.String())
}

func (a API) String() string {
	return fmt.Sprintf(
		"Host: %s, Port: %d, ReadTimeout: %s, ReadHeaderTimeout: %s, WriteTimeout: %s, ShutdownTimeout: %s",
		a.Host,
		a.Port,
		a.ReadTimeout,
		a.ReadHeaderTimeout,
		a.WriteTimeout,
		a.ShutdownTimeout,
	)
}

func (l Logging) String() string {
	return fmt.Sprintf("Level: %s, IsJSON: %t", l.Level, l.IsJSON)
}

func (d DB) String() string {
	return fmt.Sprintf(
		"Host: %s, Port: %d, User: %s, Password: [HIDDEN], DBName: %s, SSLMode: %s, MigrationPath: %s",
		d.Host,
		d.Port,
		d.User,
		d.DBName,
		d.SSLMode,
		d.MigrationPath,
	)
}

func (d DB) DSN() string {
	databaseURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   net.JoinHostPort(d.Host, strconv.Itoa(d.Port)),
		Path:   d.DBName,
	}

	query := databaseURL.Query()
	query.Set("sslmode", d.SSLMode)
	databaseURL.RawQuery = query.Encode()

	return databaseURL.String()
}

func loadFromEnv() (*Config, error) {
	api, err := loadAPIFromEnv()
	if err != nil {
		return nil, err
	}

	logging, err := loadLoggingFromEnv()
	if err != nil {
		return nil, err
	}

	db, err := loadDBFromEnv()
	if err != nil {
		return nil, err
	}

	return &Config{
		API:     api,
		Logging: logging,
		DB:      db,
	}, nil
}

func loadAPIFromEnv() (API, error) {
	port, err := getEnvInt("API_PORT", defaultAPIPort)
	if err != nil {
		return API{}, err
	}

	readTimeout, err := getEnvDuration("API_READ_TIMEOUT", defaultAPIReadTimeout)
	if err != nil {
		return API{}, err
	}

	readHeaderTimeout, err := getEnvDuration("API_READ_HEADER_TIMEOUT", defaultAPIReadHeaderTimeout)
	if err != nil {
		return API{}, err
	}

	writeTimeout, err := getEnvDuration("API_WRITE_TIMEOUT", defaultAPIWriteTimeout)
	if err != nil {
		return API{}, err
	}

	shutdownTimeout, err := getEnvDuration("API_SHUTDOWN_TIMEOUT", defaultAPIShutdownTimeout)
	if err != nil {
		return API{}, err
	}

	return API{
		Host:              getEnvString("API_HOST", defaultAPIHost),
		Port:              port,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		ShutdownTimeout:   shutdownTimeout,
	}, nil
}

func loadLoggingFromEnv() (Logging, error) {
	isJSON, err := getEnvBool("LOG_IS_JSON", defaultLogIsJSON)
	if err != nil {
		return Logging{}, err
	}

	return Logging{
		Level:  getEnvString("LOG_LEVEL", defaultLogLevel),
		IsJSON: isJSON,
	}, nil
}

func loadDBFromEnv() (DB, error) {
	port, err := getEnvInt("DB_PORT", defaultDBPort)
	if err != nil {
		return DB{}, err
	}

	return DB{
		Host:          getEnvString("DB_HOST", defaultDBHost),
		Port:          port,
		User:          getEnvString("DB_USER", defaultDBUser),
		Password:      getEnvString("DB_PASSWORD", defaultDBPassword),
		DBName:        getEnvString("DB_NAME", defaultDBName),
		SSLMode:       getEnvString("DB_SSL_MODE", defaultDBSSLMode),
		MigrationPath: getEnvString("DB_MIGRATION_PATH", defaultDBMigrationPath),
	}, nil
}

func getEnvString(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	rawValue, exists := os.LookupEnv(key)
	if !exists || rawValue == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", key)
	}

	return value, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	rawValue, exists := os.LookupEnv(key)
	if !exists || rawValue == "" {
		return fallback, nil
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean", key)
	}

	return value, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	rawValue, exists := os.LookupEnv(key)
	if !exists || rawValue == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration", key)
	}

	return value, nil
}
