package app

import (
	"reflect"
	"testing"

	"github.com/javking07/toadlester/conf"
	"github.com/javking07/toadlester/model"
	"github.com/rs/zerolog"
)

func TestApp_Bootstrap(t *testing.T) {

}

func TestInitDatabase(t *testing.T) {
	type args struct {
		c *conf.Config
	}
	tests := map[string]struct {
		name    string
		args    args
		want    model.Storage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := InitDatabase(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitDatabase() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitLogger(t *testing.T) {
	type args struct {
		c *conf.Config
	}
	tests := map[string]struct {
		name    string
		args    args
		want    *zerolog.Logger
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := InitLogger(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitLogger() got = %v, want %v", got, tt.want)
			}
		})
	}
}
