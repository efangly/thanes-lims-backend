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

	// TrustedProxies lists the IPs/CIDR ranges of the load balancers or
	// reverse proxies in front of this service, comma-separated. Only when
	// the direct peer IP is in this list does Fiber trust X-Forwarded-For /
	// the configured ProxyHeader for c.IP() - so the client IP recorded in a
	// Token Family can't be spoofed by an arbitrary caller. Leave empty when
	// the service is exposed directly (c.IP() then uses the real peer IP).
	TrustedProxies []string `env:"TRUSTED_PROXIES" envSeparator:","`
	// ProxyHeader is the header c.IP() reads the client IP from when the
	// request comes through a trusted proxy (e.g. "X-Forwarded-For").
	ProxyHeader string `env:"PROXY_HEADER"`

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
	// OracleEnabled makes the dependency explicit: when true, ORACLE_DSN is
	// required and a bad/missing value fails at boot rather than on the first
	// chatbot request. When false the ADB integration is off regardless of
	// the other ORACLE_* vars.
	OracleEnabled  bool   `env:"ORACLE_ENABLED" envDefault:"false"`
	OracleDSN      string `env:"ORACLE_DSN"`
	OracleTNSAdmin string `env:"ORACLE_TNS_ADMIN"`
}

// validate checks cross-field constraints that the env tags can't express.
func (c *Config) validate() error {
	if c.OracleEnabled && c.OracleDSN == "" {
		return fmt.Errorf("ORACLE_ENABLED=true but ORACLE_DSN is not set")
	}
	return nil
}

// Load reads .env (if present) then binds environment variables onto Config.
// A missing .env file is not an error - cloud/prod environments set env vars directly.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
