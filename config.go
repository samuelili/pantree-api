package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application's configuration settings.
// It is designed to be populated from a YAML file.
type Config struct {
	// Secret is a top-level string value.
	Secret string `yaml:"SECRET"`

	// Server holds settings related to the HTTP server.
	Server struct {
		Broadcast string `yaml:"broadcast"`
		Port      string `yaml:"port"`
		SendMail  bool   `yaml:"sendMail"`
	} `yaml:"server"`

	// Database holds connection details for the database.
	Database struct {
		User     string `yaml:"user"`
		DBName   string `yaml:"dbname"`
		Password string `yaml:"password"`
	} `yaml:"database"`

	// Email holds configuration for the email service.
	Email struct {
		From string `yaml:"from"`
		ARN  string `yaml:"arn"`
	} `yaml:"email"`

	// S3 holds configuration for the S3 service.
	S3 struct {
		Bucket string `yaml:"bucket"`
		Region string `yaml:"region"`
	} `yaml:"s3"`

	// AWS holds AWS credentials and configuration.
	AWS struct {
		AccessKeyID     string `yaml:"accessKeyId"`
		SecretAccessKey string `yaml:"secretAccessKey"`
	} `yaml:"aws"`
}

var cfg Config

func readConfig() {
	// Read the YAML file
	data, err := os.ReadFile("properties.yaml")
	if err != nil {
		panic(fmt.Errorf("error while reading config file (properties.yaml), does it exist?: %w", err))
	}

	// Unmarshal the YAML data into our Config struct
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling yaml: %w", err))
	}
}
