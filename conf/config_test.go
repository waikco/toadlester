package conf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaneDefaults(t *testing.T) {
	tests := map[string]struct {
		want *Config
	}{
		"proper defaults": {
			want: &Config{
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
					Interval: func() *time.Duration {
						t := time.Duration(time.Second * 60)
						return &t
					}(),
				},
				Sleep: func() *time.Duration {
					t := time.Duration(time.Second * 1)
					return &t
				}(),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, SaneDefaults())
		})
	}
}
