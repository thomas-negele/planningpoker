package game

import (
	"encoding/base32"
	"fmt"
	"io"
)

// idEntropyBytes gives generated IDs 128 bits of randomness.
const idEntropyBytes = 16

// The Crockford alphabet omits I, L, O and U to reduce transcription errors.
const idAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var idEncoding = base32.NewEncoding(idAlphabet).WithPadding(base32.NoPadding)

// newID reads from the supplied entropy source, which must be cryptographically
// secure in production.
func newID(random io.Reader) (string, error) {
	if random == nil {
		return "", fmt.Errorf("%w: no source of randomness was supplied", ErrShortRandomRead)
	}

	buf := make([]byte, idEntropyBytes)
	// Require all entropy bytes before constructing an ID.
	if _, err := io.ReadFull(random, buf); err != nil {
		return "", fmt.Errorf("%w: %v", ErrShortRandomRead, err)
	}

	return idEncoding.EncodeToString(buf), nil
}

var idLength = len(idEncoding.EncodeToString(make([]byte, idEntropyBytes)))

// Room IDs accept 5–64 URL-safe ASCII characters, including user-chosen names.
const (
	MinRoomIDLength = 5
	MaxRoomIDLength = 64
)

// ValidRoomID checks the shared grammar for generated and user-chosen room IDs.
func ValidRoomID(id RoomID) bool {
	if len(id) < MinRoomIDLength || len(id) > MaxRoomIDLength {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}
