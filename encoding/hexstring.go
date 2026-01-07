package encoding

import (
	"encoding"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidHexStringInput = errors.New("encoding: hexstring: invalid string")
)

// HexString is a representation of hex-encoded string.
//
// Empty string is represented by nil or empty slice.
type HexString []byte

func (h HexString) String() string {
	return fmt.Sprintf(`HexString("%s")`, hex.EncodeToString(h))
}

var _ encoding.TextUnmarshaler = (*HexString)(nil)
var _ encoding.TextMarshaler = HexString(nil)

// MarshalText implements [encoding.TextMarshaler].
func (h HexString) MarshalText() ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(dst, h)
	return dst, nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (h *HexString) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*h = nil
		return nil
	}

	b, err := hex.DecodeString(string(text))
	if err != nil {
		return fmt.Errorf("hex.DecodeString: %w", err)
	}

	*h = b
	return nil
}

var _ json.Unmarshaler = (*HexString)(nil)
var _ json.Marshaler = HexString(nil)

// MarshalJSON implements [json.Marshaler].
func (h HexString) MarshalJSON() ([]byte, error) {
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
func (h *HexString) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return ErrInvalidHexStringInput
	}

	err := h.UnmarshalText(b[1 : len(b)-1])
	if err != nil {
		return fmt.Errorf("HexString.UnmarshalText: %w", err)
	}

	return nil
}
