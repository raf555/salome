package hex

import (
	"encoding"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// String is a representation of hex-encoded string.
//
// Empty String is represented by nil or empty slice.
type String []byte

var (
	// ErrInvalidHexStringInput is an error when [String] is malformed.
	// It may occur during Unmarshal functions.
	ErrInvalidHexStringInput = errors.New("hex: invalid string")
)

func (h String) String() string {
	if len(h) == 0 {
		return ""
	}
	return hex.EncodeToString(h)
}

var _ encoding.TextUnmarshaler = (*String)(nil)
var _ encoding.TextMarshaler = String(nil)

// MarshalText implements [encoding.TextMarshaler].
func (h String) MarshalText() ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(dst, h)
	return dst, nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (h *String) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*h = nil
		return nil
	}

	b, err := hex.DecodeString(string(text))
	if err != nil {
		return fmt.Errorf("%w: hex.DecodeString: %w", ErrInvalidHexStringInput, err)
	}

	*h = b
	return nil
}

var _ json.Unmarshaler = (*String)(nil)
var _ json.Marshaler = String(nil)

// MarshalJSON implements [json.Marshaler].
func (h String) MarshalJSON() ([]byte, error) {
	if len(h) == 0 {
		return []byte(`""`), nil
	}

	// MarshalText never returns error
	b, _ := h.MarshalText()

	out := make([]byte, len(b)+2)

	out[0] = '"'
	out[len(out)-1] = '"'
	copy(out[1:], b)

	return out, nil
}

// UnmarshalJSON implements [json.Unmarshaler].
func (h *String) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return ErrInvalidHexStringInput
	}

	err := h.UnmarshalText(b[1 : len(b)-1])
	if err != nil {
		return fmt.Errorf("h.UnmarshalText: %w", err)
	}

	return nil
}
