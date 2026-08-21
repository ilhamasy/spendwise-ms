package config

import (
	"log"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Host       string
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	RedisAddr  string
	JWTSecret  string
}

func Load() *Config {
	cfg := &Config{
		Host:               getEnv("HOST", ""),
		Port:               getEnv("PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "retnotc"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "spendwise_main_db"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "spendwise-secret-key"),
	}

	if gin.Mode() == gin.ReleaseMode && cfg.JWTSecret == "spendwise-secret-key" {
		log.Println("[SECURITY WARNING] OWASP A02: Using default JWT_SECRET in Release mode is insecure. Set JWT_SECRET in environment variables.")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var allowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:3003",
	"https://spendwise.vercel.app",
	"https://spendwise-web.vercel.app",
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			isAllowed := slices.Contains(allowedOrigins, origin) ||
				strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") ||
				strings.HasPrefix(origin, "https://localhost:") ||
				strings.HasPrefix(origin, "https://127.0.0.1:")
			if isAllowed {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
			}
		} else {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
