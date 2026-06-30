package util

import (
	"testing"
)

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		id   int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "A"},
		{35, "Z"},
		{36, "a"},
		{61, "z"},
		{62, "10"},
		{100, "1c"},
		{3844, "100"},
		{238328, "1000"},
	}

	for _, tt := range tests {
		got := EncodeBase62(tt.id)
		if got != tt.want {
			t.Errorf("EncodeBase62(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		code string
		want int64
	}{
		{"0", 0},
		{"1", 1},
		{"A", 10},
		{"Z", 35},
		{"a", 36},
		{"z", 61},
		{"10", 62},
		{"1c", 100},
	}

	for _, tt := range tests {
		got := DecodeBase62(tt.code)
		if got != tt.want {
			t.Errorf("DecodeBase62(%q) = %d, want %d", tt.code, got, tt.want)
		}
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	for id := int64(0); id < 100000; id += 1000 {
		code := EncodeBase62(id)
		decoded := DecodeBase62(code)
		if decoded != id {
			t.Errorf("Roundtrip failed: %d -> %q -> %d", id, code, decoded)
		}
	}
}

func TestIsReservedWord(t *testing.T) {
	reserved := []string{"api", "admin", "login", "dashboard", "links", "stats"}
	for _, w := range reserved {
		if !IsReservedWord(w) {
			t.Errorf("IsReservedWord(%q) should be true", w)
		}
	}

	normal := []string{"mycode", "test123", "abc"}
	for _, w := range normal {
		if IsReservedWord(w) {
			t.Errorf("IsReservedWord(%q) should be false", w)
		}
	}
}

func TestIsValidCustomCode(t *testing.T) {
	valid := []string{"abc", "my-link", "test_link", "A3xK9q", "123abc"}
	for _, c := range valid {
		if !IsValidCustomCode(c) {
			t.Errorf("IsValidCustomCode(%q) should be true", c)
		}
	}

	invalid := []string{"", "admin", "too-long-code-that-exceeds-twenty-chars", "code with space", "code@#$"}
	for _, c := range invalid {
		if IsValidCustomCode(c) {
			t.Errorf("IsValidCustomCode(%q) should be false", c)
		}
	}
}
