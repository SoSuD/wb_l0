package config

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Logger   Logger
	Server   Server
	Kafka    Kafka
	Postgres Postgres
	Cache    Cache
}

type Cache struct {
	Capacity int
}

type Logger struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}
type Server struct {
	Port string
	Mode string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
	SslMode  string
}

type Kafka struct {
	Orders KafkaConsumer
}

type KafkaConsumer struct {
	Brokers []string
	Topic   string
	GroupId string
}

func LoadConfig(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigFile(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	return v, nil
}

// Parse config file
func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}
