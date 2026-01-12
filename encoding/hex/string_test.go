package hex

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestString_String(t *testing.T) {
	tests := []struct {
		name     string
		input    String
		expected string
	}{
		{
			name:     "empty hex string",
			input:    String([]byte{}),
			expected: "",
		},
		{
			name:     "simple hex string",
			input:    String([]byte{0x01, 0x02, 0x03}),
			expected: "010203",
		},
		{
			name:     "hex string with all bytes",
			input:    String([]byte{0xff, 0xaa, 0x55}),
			expected: "ffaa55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestString_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    String
		expected string
		wantErr  bool
	}{
		{
			name:     "empty hex string",
			input:    String([]byte{}),
			expected: "",
			wantErr:  false,
		},
		{
			name:     "simple hex encoding",
			input:    String([]byte{0x01, 0x02, 0x03}),
			expected: "010203",
			wantErr:  false,
		},
		{
			name:     "hex with letters",
			input:    String([]byte{0xde, 0xad, 0xbe, 0xef}),
			expected: "deadbeef",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalText() = %v, want %v", string(result), tt.expected)
			}
		})
	}
}

func TestString_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected String
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: String([]byte{}),
			wantErr:  false,
		},
		{
			name:     "valid hex string",
			input:    "010203",
			expected: String([]byte{0x01, 0x02, 0x03}),
			wantErr:  false,
		},
		{
			name:     "valid hex with letters",
			input:    "deadbeef",
			expected: String([]byte{0xde, 0xad, 0xbe, 0xef}),
			wantErr:  false,
		},
		{
			name:     "uppercase hex",
			input:    "DEADBEEF",
			expected: String([]byte{0xde, 0xad, 0xbe, 0xef}),
			wantErr:  false,
		},
		{
			name:     "invalid hex character",
			input:    "ghijkl",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "odd length hex string",
			input:    "abc",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h String
			err := h.UnmarshalText([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(h) != len(tt.expected) {
					t.Errorf("UnmarshalText() length = %v, want %v", len(h), len(tt.expected))
					return
				}
				for i := range h {
					if h[i] != tt.expected[i] {
						t.Errorf("UnmarshalText() = %v, want %v", h, tt.expected)
						return
					}
				}
			}
		})
	}
}

func TestString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    String
		expected string
		wantErr  bool
	}{
		{
			name:     "empty hex string",
			input:    String([]byte{}),
			expected: `""`,
			wantErr:  false,
		},
		{
			name:     "simple hex string",
			input:    String([]byte{0x01, 0x02, 0x03}),
			expected: `"010203"`,
			wantErr:  false,
		},
		{
			name:     "hex with letters",
			input:    String([]byte{0xde, 0xad, 0xbe, 0xef}),
			expected: `"deadbeef"`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalJSON() = %v, want %v", string(result), tt.expected)
			}
		})
	}
}

func TestString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected String
		wantErr  bool
		errType  error
	}{
		{
			name:     "empty json string",
			input:    `""`,
			expected: String([]byte{}),
			wantErr:  false,
		},
		{
			name:     "valid json hex string",
			input:    `"010203"`,
			expected: String([]byte{0x01, 0x02, 0x03}),
			wantErr:  false,
		},
		{
			name:     "valid json hex with letters",
			input:    `"deadbeef"`,
			expected: String([]byte{0xde, 0xad, 0xbe, 0xef}),
			wantErr:  false,
		},
		{
			name:     "missing opening quote",
			input:    `abc"`,
			expected: nil,
			wantErr:  true,
			errType:  ErrInvalidHexStringInput,
		},
		{
			name:     "missing closing quote",
			input:    `"abc`,
			expected: nil,
			wantErr:  true,
			errType:  ErrInvalidHexStringInput,
		},
		{
			name:     "no quotes",
			input:    `abc`,
			expected: nil,
			wantErr:  true,
			errType:  ErrInvalidHexStringInput,
		},
		{
			name:     "empty input",
			input:    ``,
			expected: nil,
			wantErr:  true,
			errType:  ErrInvalidHexStringInput,
		},
		{
			name:     "single quote",
			input:    `"`,
			expected: nil,
			wantErr:  true,
			errType:  ErrInvalidHexStringInput,
		},
		{
			name:     "invalid hex in json",
			input:    `"xyz"`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h String
			err := h.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errType != nil && err != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("UnmarshalJSON() error = %v, want error type %v", err, tt.errType)
				}
			}
			if !tt.wantErr {
				if len(h) != len(tt.expected) {
					t.Errorf("UnmarshalJSON() length = %v, want %v", len(h), len(tt.expected))
					return
				}
				for i := range h {
					if h[i] != tt.expected[i] {
						t.Errorf("UnmarshalJSON() = %v, want %v", h, tt.expected)
						return
					}
				}
			}
		})
	}
}

func TestString_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input String
	}{
		{
			name:  "empty",
			input: String([]byte{}),
		},
		{
			name:  "simple bytes",
			input: String([]byte{0x01, 0x02, 0x03}),
		},
		{
			name:  "complex bytes",
			input: String([]byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe, 0xba, 0xbe}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			marshaled, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal
			var result String
			err = json.Unmarshal(marshaled, &result)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare
			if len(result) != len(tt.input) {
				t.Errorf("Round trip length = %v, want %v", len(result), len(tt.input))
				return
			}
			for i := range result {
				if result[i] != tt.input[i] {
					t.Errorf("Round trip = %v, want %v", result, tt.input)
					return
				}
			}
		})
	}
}

func TestString_TextRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input String
	}{
		{
			name:  "empty",
			input: String([]byte{}),
		},
		{
			name:  "simple bytes",
			input: String([]byte{0x01, 0x02, 0x03}),
		},
		{
			name:  "complex bytes",
			input: String([]byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe, 0xba, 0xbe}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			marshaled, err := tt.input.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() error = %v", err)
			}

			// Unmarshal
			var result String
			err = result.UnmarshalText(marshaled)
			if err != nil {
				t.Fatalf("UnmarshalText() error = %v", err)
			}

			// Compare
			if len(result) != len(tt.input) {
				t.Errorf("Round trip length = %v, want %v", len(result), len(tt.input))
				return
			}
			for i := range result {
				if result[i] != tt.input[i] {
					t.Errorf("Round trip = %v, want %v", result, tt.input)
					return
				}
			}
		})
	}
}
