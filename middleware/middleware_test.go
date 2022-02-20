package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_requestID(t *testing.T) {
	req := httptest.NewRequest("POST", "/save?url=https://tech.ozon.ru", nil)
	req.Header.Set("Request-Id", "777")
	type args struct {
		header string
		r      *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Get an existing number",
			args: args{
				header: "Request-Id",
				r:      req,
			},
			want: "777",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestID(tt.args.header, tt.args.r); got != tt.want {
				t.Errorf("requestID() = %v, want %v", got, tt.want)
			}
		})
	}
}
