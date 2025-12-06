package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string `mapstructure:"PORT"`
	GinMode            string `mapstructure:"GIN_MODE"`
	RedisAddr          string `mapstructure:"REDIS_ADDR"`
	RedisPassword      string `mapstructure:"REDIS_PASSWORD"`
	RedisDB            int    `mapstructure:"REDIS_DB"`
	DatabaseURL        string `mapstructure:"DATABASE_URL"`
	AWSRegion          string `mapstructure:"AWS_REGION"`
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	S3BucketName       string `mapstructure:"S3_BUCKET_NAME"`
	MaxUploadSize      int64  `mapstructure:"MAX_UPLOAD_SIZE"`
	ThumbnailWidth     int    `mapstructure:"THUMBNAIL_WIDTH"`
	ThumbnailHeight    int    `mapstructure:"THUMBNAIL_HEIGHT"`
	ResizeWidth        int    `mapstructure:"RESIZE_WIDTH"`
	ResizeHeight       int    `mapstructure:"RESIZE_HEIGHT"`
	CompressQuality    int    `mapstructure:"COMPRESS_QUALITY"`
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigFile("../.env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("GIN_MODE", "release")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("AWS_REGION", "us-east-1")
	viper.SetDefault("MAX_UPLOAD_SIZE", 10485760) // 10MB
	viper.SetDefault("THUMBNAIL_WIDTH", 200)
	viper.SetDefault("THUMBNAIL_HEIGHT", 200)
	viper.SetDefault("RESIZE_WIDTH", 1920)
	viper.SetDefault("RESIZE_HEIGHT", 1080)
	viper.SetDefault("COMPRESS_QUALITY", 85)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found, using defaults and environment variables: %v", err)
	} else {
		log.Printf("Environment variables loaded from %s", viper.ConfigFileUsed())
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	log.Println("Configuration loaded successfully")
}
