package pkg

import "testing"

func TestGeneratorShortURL(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "Lenght test",
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(GeneratorShortURL()); got > tt.want {
				t.Errorf("GeneratorShortURL() = %v, want <= %v", got, tt.want)
			}
		})
	}
}
