package conf

import (
	"time"
)

// Config is application config
type Config struct {
	Database *DatabaseConfig `json:"database" yaml:"database"`
	Logging  *LoggingConfig  `json:"logging" yaml:"logging"`
	Sleep    *time.Duration  `json:"sleep" yaml:"sleep"`
	Timer    *TimerConfig    `json:"timer" yaml:"timer"`
	Tests    []struct {
		Name     string         `json:"name"`
		Duration *time.Duration `json:"duration"` // in seconds
		TPS      int            `json:"tps"`
		Target   string         `json:"target"`
	} `json:"tests" yaml:"tests"`
}

type DatabaseConfig struct {
	Type         string `json:"type" yaml:"type"`
	Host         string `json:"host" yaml:"host"`
	Port         int    `json:"port" yaml:"port"`
	User         string `json:"user" yaml:"user"`
	Password     string `json:"password" yaml:"password"`
	DatabaseName string `json:"databaseName" yaml:"databaseName"`
	SslMode      string `json:"sslMode" yaml:"sslMode"`
	SslFactory   string `json:"sslFactory" yaml:"sslFactory"`
}

type CacheConfig struct {
	Size int // config size in bytes
}

type TimerConfig struct {
	Interval *time.Duration `json:"interval" yaml:"interval"`
}

type LoggingConfig struct {
	Level string `json:"level" yaml:"level"`
}

// SaneDefaults provides base config for testing
func SaneDefaults() *Config {
	startupSleep := time.Second * 1
	backgroundInterval := time.Second * 60
	var config = &Config{
		Database: &DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Password:     "postgres",
			DatabaseName: "postgres",
			SslMode:      "disable",
			SslFactory:   "org.postgresql.ssl.NonValidatingFactory",
		},
		Logging: &LoggingConfig{
			Level: "debug",
		},
		Timer: &TimerConfig{
			Interval: &backgroundInterval,
		},
		Sleep: &startupSleep,
	}
	return config
}
