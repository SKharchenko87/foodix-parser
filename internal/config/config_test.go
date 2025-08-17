package config

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{
			name: "TestLoadConfig1",
			args: args{"../../configs/config.example.yaml"},
			want: Config{
				Log:     Log{Level: "info", Format: "json"},
				Sources: []SourceConfig{{Name: "test.ru", URL: "test.ru/food", Timeout: 10000, ProductPerPage: 80}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
