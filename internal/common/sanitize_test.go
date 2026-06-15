package common

import "testing"

func TestSanitizeInput_Normal(t *testing.T) {
	input := "Go 并发编程"
	got := SanitizeInput(input)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestSanitizeInput_RemovesControlChars(t *testing.T) {
	input := "hello\x00world\x01test"
	want := "helloworldtest"
	got := SanitizeInput(input)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestSanitizeInput_PreservesNewlineAndTab(t *testing.T) {
	input := "line1\n\tline2"
	got := SanitizeInput(input)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestSanitizeInput_TruncatesLong(t *testing.T) {
	input := string(make([]byte, 3000))
	got := SanitizeInput(input)
	if len(got) > 2000 {
		t.Errorf("expected max 2000 chars, got %d", len(got))
	}
}
