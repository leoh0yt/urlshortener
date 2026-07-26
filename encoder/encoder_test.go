package encoder

import "testing"

func TestBase62Encode(t *testing.T) {
	encoder := NewBase62Encoder()

	tests := []struct {
		name     string
		val      uint64
		expected string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"ten", 10, "A"},
		{"sixty_four", 64, "12"},
		{"large_value", 123456789, "8M0kX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Encode10(tt.val)
			if result != tt.expected {
				t.Errorf("Encoding10(%d) = %s, expected %s", tt.val, result, tt.expected)
			}
		})
	}
}

func TestBase62Decode(t *testing.T) {
	encoder := NewBase62Encoder()

	tests := []struct {
		name     string
		val      string
		expected uint64
	}{
		{"zero", "0", 0},
		{"zero_with_padding", "0000000000", 0},
		{"one", "1", 1},
		{"ten", "A", 10},
		{"sixty_four", "12", 64},
		{"large_value", "8M0kX", 123456789},
		{"large_value_with_padding", "0000008M0kX", 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := encoder.Decode10(tt.val)
			if result != tt.expected {
				t.Errorf("Encoding10(%s) = %d, expected %d", tt.val, result, tt.expected)
			}
		})
	}
}

func TestBase62DecodeInvalid(t *testing.T) {
	encoder := NewBase62Encoder()

	tests := []struct {
		name string
		val  string
	}{
		{"invalid_start", "---fsdfasd"},
		{"invalid_center", "12fsda---fdas"},
		{"invalid_end", "fsdf312a--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encoder.Decode10(tt.val)
			if err == nil {
				t.Error("Expected error")
			}
		})
	}
}
