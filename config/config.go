package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConfig    DBConfig
	KafkaConfig KafkaConfig
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		DBConfig: DBConfig{
			DBHost:     os.Getenv("DB_HOST"),
			DBPort:     os.Getenv("DB_PORT"),
			DBUser:     os.Getenv("DB_USER"),
			DBPassword: os.Getenv("DB_PASSWORD"),
			DBName:     os.Getenv("DB_NAME"),
		},
		KafkaConfig: KafkaConfig{
			KafkaHost: os.Getenv("KAFKA_HOST"),
			KafkaPort: os.Getenv("KAFKA_PORT"),
			Topic:     os.Getenv("TOPIC"),
			Group:     os.Getenv("GROUP"),
		},
	}
	return config, nil
}
