// Application configuration.
package config

import "os"

type Config struct {
	ServerPort string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string

	LogLevel string
}

func Load() Config {
	return Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "tictactoe"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "tictactoe"),
		PostgresDatabase: getEnv("POSTGRES_DATABASE", "tictactoe"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func (c Config) PostgresDSN() string {
	return "postgres://" + c.PostgresUser + ":" + c.PostgresPassword +
		"@" + c.PostgresHost + ":" + c.PostgresPort + "/" + c.PostgresDatabase +
		"?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
