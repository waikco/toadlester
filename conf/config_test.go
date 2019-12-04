package conf

import (
	"reflect"
	"testing"
	"time"
)

func TestSaneDefaults(t *testing.T) {
	tests := map[string]struct {
		want *Config
	}{
		"proper defaults": {
			want: &Config{
				Database: &DatabaseConfig{
					Host:         "127.0.0.1",
					Port:         5432,
					User:         "user",
					Password:     "password",
					DatabaseName: "test",
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
			if got := SaneDefaults(); !reflect.DeepEqual(got, test.want) {
				t.Errorf("SaneDefaults() = %v, want %v", got, test.want)
			}
		})
	}
}
