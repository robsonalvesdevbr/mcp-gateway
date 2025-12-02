package eval

import "testing"

func TestIsEmptyOrInvalid(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		// Valores que devem ser considerados inválidos
		{"nil", nil, true},
		{"empty string", "", true},
		{"empty slice", []string{}, true},
		{"empty array", [0]string{}, true},
		{"empty map", map[string]string{}, true},
		{"string []", "[]", true},
		{"string {}", "{}", true},
		{"string <nil>", "<nil>", true},
		{"string null", "null", true},

		// Valores que devem ser considerados válidos
		{"valid path", "/valid/path", false},
		{"valid string", "value", false},
		{"non-empty slice", []string{"item"}, false},
		{"non-empty map", map[string]string{"key": "value"}, false},
		{"number", 123, false},
		{"zero", 0, false}, // Zero é um valor válido
		{"false boolean", false, false}, // False é um valor válido
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEmptyOrInvalid(tt.value)
			if got != tt.want {
				t.Errorf("isEmptyOrInvalid(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestVolumeWithEmptyValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"nil", nil, ""},
		{"empty string", "", ""},
		{"empty slice", []string{}, ""},
		{"empty map", map[string]string{}, ""},
		{"string []", "[]", ""},
		{"string {}", "{}", ""},
		{"string <nil>", "<nil>", ""},
		{"string null", "null", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := volume(tt.value)
			if got != tt.want {
				t.Errorf("volume(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestVolumeWithValidValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "simple path",
			value: "/data",
			want:  "/data:/data",
		},
		{
			name:  "complex path",
			value: "/var/lib/docker/volumes/myvolume",
			want:  "/var/lib/docker/volumes/myvolume:/var/lib/docker/volumes/myvolume",
		},
		{
			name:  "relative path",
			value: "./data",
			want:  "./data:./data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := volume(tt.value)
			if got != tt.want {
				t.Errorf("volume(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestVolumeTargetWithEmptyValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"nil", nil, ""},
		{"empty string", "", ""},
		{"empty slice", []string{}, ""},
		{"string []", "[]", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := volumeTarget(tt.value)
			if got != tt.want {
				t.Errorf("volumeTarget(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestVolumeTargetWithValidValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "simple path",
			value: "/data",
			want:  "/data",
		},
		{
			name:  "complex path",
			value: "/var/lib/docker/volumes/myvolume",
			want:  "/var/lib/docker/volumes/myvolume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := volumeTarget(tt.value)
			if got != tt.want {
				t.Errorf("volumeTarget(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
