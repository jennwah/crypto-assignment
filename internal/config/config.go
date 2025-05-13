package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config is a common struct contains other configs.
type Config struct {
	Postgres
	Redis
}

func LoadConfig() (Config, error) {
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(b), "../..")

	_ = godotenv.Load(
		fmt.Sprintf("%s/%s", rootPath, ".env"),
	)

	var c Config

	if err := envconfig.Process("", &c); err != nil {
		return Config{}, fmt.Errorf("unable to decode env into struct: %w", err)
	}

	return c, nil
}
