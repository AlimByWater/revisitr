package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Module struct{}

func (m *Module) Init() error {
	return godotenv.Load()
}

func GetString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return val
}
