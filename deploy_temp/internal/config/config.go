package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	JWTSecret            string
	JWTExpiryHr          string
	JWTRefreshExpiryDays string
	SMTPHost             string
	SMTPPort             string
	SMTPUser             string
	SMTPPassword         string
	SMTPFrom             string
	TurnURL              string
	TurnUsername         string
	TurnCredential       string
	// FirebaseCredentialsFile is the path to a Firebase service-account JSON
	// key (Firebase Console -> Project Settings -> Service Accounts ->
	// Generate new private key). Left empty by default — push notifications
	// are silently disabled (no-op) until this is configured; nothing else
	// in the app depends on it.
	FirebaseCredentialsFile string
	// CORSAllowedOrigins is a comma-separated list of extra allowed origins
	// (e.g. "https://api.yourdomain.com,https://app.yourdomain.com") for
	// production. Empty by default — dev already works via the built-in
	// localhost/LAN/ngrok allowance in middleware.CORSMiddleware without
	// needing this set.
	CORSAllowedOrigins string
}

func Load() *Config {
	cwd, _ := os.Getwd()
	envPath := filepath.Join(cwd, ".env")
	if err := godotenv.Load(); err != nil {
		log.Printf("[config] No .env file found at %s (reading from OS environment variables instead): %v", envPath, err)
	} else {
		log.Printf("[config] Loaded .env from %s", envPath)
	}

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "legaltech_db"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		JWTSecret:            getEnv("JWT_SECRET", "change_this_secret_in_env"),
		JWTExpiryHr:          getEnv("JWT_EXPIRY_HOURS", "1"),
		JWTRefreshExpiryDays: getEnv("JWT_REFRESH_EXPIRY_DAYS", "30"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnv("SMTP_PORT", "587"),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:             getEnv("SMTP_FROM", "no-reply@legaltech.ai"),
		// Optional TURN relay for WebRTC when direct P2P/STUN fails (e.g. behind
		// symmetric NAT). Left empty by default — STUN-only still works for most
		// networks. Configure via .env to enable.
		TurnURL:                 getEnv("TURN_URL", ""),
		TurnUsername:            getEnv("TURN_USERNAME", ""),
		TurnCredential:          getEnv("TURN_CREDENTIAL", ""),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		CORSAllowedOrigins:      getEnv("CORS_ALLOWED_ORIGINS", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
