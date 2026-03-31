package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	MongoURI    string `mapstructure:"MONGO_URI"`
	RedisAddr   string `mapstructure:"REDIS_ADDR"`
	RabbitMQURL string `mapstructure:"RABBITMQ_URL"`
	Port        string `mapstructure:"PORT"`
}

func Load() (*Config, error) {
	viper.SetDefault("PORT", "8080")

	viper.SetConfigFile(envFile())
	viper.SetConfigType("env")
	// Missing .env is not fatal — real env vars take precedence anyway.
	_ = viper.ReadInConfig()

	viper.BindEnv("MONGO_URI")
	viper.BindEnv("REDIS_ADDR")
	viper.BindEnv("RABBITMQ_URL")
	viper.BindEnv("PORT")

	viper.AutomaticEnv()

	cfg := &Config{
		MongoURI:    viper.GetString("MONGO_URI"),
		RedisAddr:   viper.GetString("REDIS_ADDR"),
		RabbitMQURL: viper.GetString("RABBITMQ_URL"),
		Port:        viper.GetString("PORT"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// envFile returns the path to the .env file: honours ENV_FILE env var,
// falls back to .env next to the binary's working directory.
func envFile() string {
	if p := os.Getenv("ENV_FILE"); p != "" {
		return p
	}
	return ".env"
}

func (c *Config) validate() error {
	if c.MongoURI == "" {
		return fmt.Errorf("MONGO_URI is required")
	}
	if c.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}
	if c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}
	return nil
}
