package validation

import (
	"testing"
)

func TestValidateURLs(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr bool
	}{
		{
			name:    "valid single URL",
			input:   []string{"https://example.com"},
			wantErr: false,
		},
		{
			name:    "valid multiple URLs",
			input:   []string{"https://example.com", "http://golang.org"},
			wantErr: false,
		},
		{
			name:    "invalid scheme",
			input:   []string{"ftp://example.com"},
			wantErr: true,
		},
		{
			name:    "missing host",
			input:   []string{"https:///path"},
			wantErr: true,
		},
		{
			name:    "localhost not allowed",
			input:   []string{"http://localhost:8080"},
			wantErr: true,
		},
		{
			name:    "private IP not allowed",
			input:   []string{"http://192.168.1.10"},
			wantErr: true,
		},
		{
			name:    "loopback IP not allowed",
			input:   []string{"https://127.0.0.1"},
			wantErr: true,
		},
		{
			name:    "empty slice (no URLs)",
			input:   []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURLs(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
