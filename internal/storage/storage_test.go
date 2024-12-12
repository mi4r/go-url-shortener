package storage

import (
	"testing"
)

func TestGenerateShortID_Length(t *testing.T) {
	id := generateShortID()
	if len(id) != idLength {
		t.Errorf("generateShortID() returned ID with incorrect length: got %d, want %d", len(id), idLength)
	}
}

func TestGenerateShortID_Charset(t *testing.T) {
	id := generateShortID()
	for _, char := range id {
		if !containsChar(charset, char) {
			t.Errorf("generateShortID() returned ID with invalid character: %c", char)
		}
	}
}

func TestGenerateShortID_Uniqueness(t *testing.T) {
	generated := make(map[string]struct{})
	for i := 0; i < 100000; i++ {
		id := generateShortID()
		if _, exists := generated[id]; exists {
			t.Errorf("generateShortID() generated duplicate ID: %s", id)
		}
		generated[id] = struct{}{}
	}
}

// containsChar проверяет, содержит ли строка указанный символ.
func containsChar(str string, char rune) bool {
	for _, c := range str {
		if c == char {
			return true
		}
	}
	return false
}
