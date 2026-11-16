package password

import (
	"regexp"
	"testing"
)

func TestRandom(t *testing.T) {
	got := Random(16)
	if len(got) != 16 {
		t.Fatalf("Random(16) length = %d, want 16", len(got))
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(got) {
		t.Errorf("Random(16) = %q, want alphanumeric", got)
	}
}

func TestRandomIsUnique(t *testing.T) {
	a, b := Random(16), Random(16)
	if a == b {
		t.Errorf("two Random calls returned the same value %q", a)
	}
}
