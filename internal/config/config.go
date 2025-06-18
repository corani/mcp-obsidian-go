package config

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/caarlos0/env"
	dotenv "github.com/joho/godotenv"
)

type Config struct {
	ObsidianAPIKey  string `env:"OBSIDIAN_API_KEY"`
	ObsidianAPIHost string `env:"OBSIDIAN_API_HOST"`
	Logger          *slog.Logger
}

func xdgConfig() string {
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return path.Join(xdgHome, "mcp_obsidian", "config")
	}

	return path.Join(os.Getenv("HOME"), ".config", "mcp_obsidian", "config")
}

func MustLoad(logger *slog.Logger) *Config {
	conf, err := Load(logger)
	if err != nil {
		panic(err)
	}

	return conf
}

func Load(logger *slog.Logger) (*Config, error) {
	conf := new(Config)

	for _, name := range []string{".env", xdgConfig()} {
		if _, err := os.Stat(name); err != nil {
			logger.Debug("config file not found",
				slog.String("name", name))

			continue
		}

		logger.Info("loading config file",
			slog.String("name", name))

		if err := dotenv.Load(name); err != nil {
			return nil, err
		}

		break
	}

	if err := env.Parse(conf); err != nil {
		return nil, err
	}

	conf.Logger = logger
	conf.ObsidianAPIHost = strings.TrimSuffix(conf.ObsidianAPIHost, "/")

	return conf, nil
}
