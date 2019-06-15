package conf

import "time"

// Config is application config
type Config struct {
	Server   *ServerConfig   `json:"server" yaml:"server"`
	Database *DatabaseConfig `json:"database" yaml:"database"`
	Cache    *CacheConfig    `json:"cache" yaml:"cache"`
	Logging  *LoggingConfig  `json:"logging" yaml:"logging"`
	Sleep    *time.Duration  `json:"sleep" yaml:"sleep"`
	Timer    *TimerConfig    `json:"timer" yaml:"timer"`
}

type ServerConfig struct {
	Port string `json:"port" yaml:"port"`
	Cert string `json:"cert" yaml:"cert"`
	Key  string `json:"cert" yaml:"key"`
	TLS  bool   `json:"tls" yaml:"tls"`
}

type DatabaseConfig struct {
	Type         string `json:"type yaml:"type"`
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
	startupSleep := time.Second * 5
	backgroundInterval := time.Second * 60
	var config = &Config{
		Server: &ServerConfig{
			Port: "8080",
			Cert: "certs/cert.crt",
			Key:  "certs/cert.key",
			TLS:  false,
		},
		Database: &DatabaseConfig{
			Host:         "127.0.0.1",
			Port:         5432,
			User:         "user",
			Password:     "password",
			DatabaseName: "test",
			SslMode:      "disable",
			SslFactory:   "org.postgresql.ssl.NonValidatingFactory",
		},
		Cache: &CacheConfig{
			Size: 1000 * 1000,
		},
		Logging: &LoggingConfig{
			Level: "INFO",
		},
		Timer: &TimerConfig{
			Interval: &backgroundInterval,
		},
		Sleep: &startupSleep,
	}

	return config
}
