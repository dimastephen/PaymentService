package config

import (
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	commandTopic string
	host         string
	port         string
	loglevel     string
	dsn          string
	brokers      []string
	jaegerURL    string
}

func NewConfig() (*config, error) {
	_ = godotenv.Load(".env")

	return &config{
		commandTopic: os.Getenv("KAFKA_COMMAND_TOPIC"),
		host:         os.Getenv("HOST"),
		port:         os.Getenv("PORT"),
		loglevel:     os.Getenv("LOG_LEVEL"),
		dsn:          os.Getenv("POSTGRES_DSN"),
		brokers:      strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		jaegerURL:    os.Getenv("JAEGER_URL"),
	}, nil
}

func (c *config) CommandTopic() string {
	return c.commandTopic
}

func (c *config) Address() string {
	return net.JoinHostPort(c.host, c.port)
}

func (c *config) Level() string {
	return c.loglevel
}

func (c *config) DSN() string {
	return c.dsn
}

func (c *config) Brokers() []string {
	return c.brokers
}

func (c *config) JaegerURL() string { return c.jaegerURL }
