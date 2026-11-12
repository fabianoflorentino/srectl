package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	if got := String(); !strings.Contains(got, "srectl") {
		t.Fatalf("String() = %q, want to contain %q", got, "srectl")
	}
}
