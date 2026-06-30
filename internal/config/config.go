package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	ServerHost string
	BaseURL    string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	JWTSecret      string
	JWTExpireHours int

	RateLimitGlobal int
	RateLimitCreate int

	GeoIPDBPath string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "shortener"),
		DBPassword: getEnv("DB_PASSWORD", "shortener_dev"),
		DBName:     getEnv("DB_NAME", "url_shortener"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		JWTSecret:      getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 72),

		RateLimitGlobal: getEnvInt("RATE_LIMIT_GLOBAL", 100),
		RateLimitCreate: getEnvInt("RATE_LIMIT_CREATE", 10),

		GeoIPDBPath: getEnv("GEOIP_DB_PATH", "./data/GeoLite2-City.mmdb"),
	}
}

func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword +
		"@" + c.DBHost + ":" + c.DBPort +
		"/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
