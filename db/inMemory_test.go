package db

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestInMemory_Set(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value string
	}

	tests := []struct {
		name    string
		d       InMemory
		args    args
		wantErr bool
	}{
		{
			name: "Error, already exists",
			d: InMemory{
				"shortUrl": "https://tech.ozon.ru",
			},
			args: args{
				ctx:   httptest.NewRequest("POST", "/save?url=https://tech.ozon.ru", nil).Context(),
				key:   "shortUrl",
				value: "https://tech.ozon.ru",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Set(tt.args.ctx, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("InMemory.Set() error = %v, wantErr 'Ключ '%v' уже существует'", err, tt.args.key)
			}
		})
	}
}

func TestInMemory_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		d       InMemory
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Error, key does not exist",
			d:    InMemory{},
			args: args{
				ctx: httptest.NewRequest("POST", "/save?url=https://tech.ozon.ru", nil).Context(),
				key: "shortUrl",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Get a URL",
			d: InMemory{
				"shortUrl": "https://tech.ozon.ru",
			},
			args: args{
				ctx: httptest.NewRequest("POST", "/save?url=https://tech.ozon.ru", nil).Context(),
				key: "shortUrl",
			},
			want:    "https://tech.ozon.ru",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Get(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("InMemory.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InMemory.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
