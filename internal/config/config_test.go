package config

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	type args struct {
		filePath string
		logger   *log.Logger
	}
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	tests := []struct {
		name string
		args args
		want Config
	}{
		{name: "TestLoadConfig1", args: args{"../../configs/config.example.yaml", logger}, want: Config{Source: "test.ru/food", Log: Log{Level: "info", Format: "json"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadConfig(tt.args.filePath, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
