package test

import "testing"

// AssertThatLineEqualsTo verifies that an `idx` line contained in the `lines` slice should be equal to the `expected` value
func AssertThatLineEqualsTo(t *testing.T, lines []string, idx int, expected string) {
	if lines[idx] != expected {
		t.Fatalf("expected \"%s\", obtained \"%s\"", expected, lines[idx])
	}
}
