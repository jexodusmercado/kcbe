package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AWS      AWSConfig      `mapstructure:"aws"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Schema   string `mapstructure:"schema"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type AWSConfig struct {
	Region       string `mapstructure:"region"`
	S3BucketName string `mapstructure:"s3_bucket_master"`
}

type AppConfig struct {
	Environment string `mapstructure:"env"`
}

var globalConfig *Config

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables and defaults")
	}

	setDefaults()

	mapEnvVars()

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

func Get() *Config {
	if globalConfig == nil {
		log.Fatal("Configuration not loaded. Call config.Load() first.")
	}
	return globalConfig
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.schema", "public")
	viper.SetDefault("app.env", "local")
	viper.SetDefault("aws.region", "us-east-1")
}

func mapEnvVars() {
	// Server
	viper.BindEnv("server.port", "PORT")

	// Database
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.database", "DATABASE_NAME")
	viper.BindEnv("database.username", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.schema", "DATABASE_SCHEMA")

	// JWT
	viper.BindEnv("jwt.secret", "JWT_SECRET")

	// AWS
	viper.BindEnv("aws.region", "AWS_REGION")
	viper.BindEnv("aws.s3_bucket_master", "S3_BUCKET_MASTER")

	// App
	viper.BindEnv("app.env", "APP_ENV")
}

func validateConfig(config *Config) error {
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if config.Database.Database == "" {
		return fmt.Errorf("BLUEPRINT_DB_DATABASE is required")
	}

	if config.Database.Username == "" {
		return fmt.Errorf("BLUEPRINT_DB_USERNAME is required")
	}

	if config.Database.Password == "" {
		return fmt.Errorf("BLUEPRINT_DB_PASSWORD is required")
	}

	return nil
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=disable",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
		c.Database.Schema,
	)
}
