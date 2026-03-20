package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	postgresDSN string
	brokers     []string
	group       string
	eventTopic  string
	jaegerURL   string
}

func NewConfig() (*Config, error) {
	_ = Load()

	cfg := &Config{
		postgresDSN: os.Getenv("POSTGRES_DSN"),
		brokers:     strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		group:       os.Getenv("KAFKA_PROJECTION_GROUP"),
		eventTopic:  os.Getenv("KAFKA_EVENT_TOPIC"),
		jaegerURL:   os.Getenv("JAEGER_URL"),
	}
	if cfg.postgresDSN == "" {
		return nil, errors.New("postgresDSN is blank")
	}
	return cfg, nil
}

func Load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) PostgresDSN() string {
	return c.postgresDSN
}
func (c *Config) Brokers() []string {
	return c.brokers
}
func (c *Config) Group() string {
	return c.group
}
func (c *Config) EventTopic() string {
	return c.eventTopic
}
func (c *Config) JaegerURL() string { return c.jaegerURL }
