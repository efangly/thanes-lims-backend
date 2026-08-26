package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string `env:"APP_ENV" envDefault:"local"`
	AppPort     string `env:"APP_PORT" envDefault:"8080"`
	AutoMigrate bool   `env:"AUTO_MIGRATE" envDefault:"false"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	DatabaseURL string `env:"DATABASE_URL,required"`

	JWTAccessSecret  string        `env:"JWT_ACCESS_SECRET,required"`
	JWTRefreshSecret string        `env:"JWT_REFRESH_SECRET,required"`
	JWTAccessTTL     time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL    time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`

	// CookieSecure controls the Secure attribute on the Refresh Cookie - set
	// to false only for local http:// development (see ADR 0004).
	CookieSecure bool `env:"COOKIE_SECURE" envDefault:"true"`
	// CORSAllowOrigins lists exact origins allowed to make credentialed
	// cross-origin requests, comma-separated (e.g. when the frontend dev
	// server runs on a different port than the API). Leave empty in
	// same-origin production deployments - CORS doesn't apply to same-origin
	// requests at all, so no config is needed there (see ADR 0004).
	CORSAllowOrigins []string `env:"CORS_ALLOW_ORIGINS" envSeparator:","`

	MinioEndpoint  string `env:"MINIO_ENDPOINT,required"`
	MinioAccessKey string `env:"MINIO_ACCESS_KEY,required"`
	MinioSecretKey string `env:"MINIO_SECRET_KEY,required"`
	MinioBucket    string `env:"MINIO_BUCKET,required"`
	MinioUseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"true"`

	// RedisURL backs the read-through Cache (see
	// docs/adr/0005-redis-cache-for-refresh-tokens-and-location-full-path.md).
	RedisURL string `env:"REDIS_URL,required"`

	AutoReorderEnabled  bool          `env:"AUTO_REORDER_ENABLED" envDefault:"true"`
	AutoReorderInterval time.Duration `env:"AUTO_REORDER_INTERVAL" envDefault:"1h"`

	// Oracle (Select AI chatbot POC) - separate ADB instance, not the Postgres system of record.
	OracleDSN      string `env:"ORACLE_DSN"`
	OracleTNSAdmin string `env:"ORACLE_TNS_ADMIN"`
}

// Load reads .env (if present) then binds environment variables onto Config.
// A missing .env file is not an error - cloud/prod environments set env vars directly.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
