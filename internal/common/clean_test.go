package common

import "testing"

func TestCleanJSON_NoMarkers(t *testing.T) {
	input := `{"key": "value"}`
	got := CleanJSON(input)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestCleanJSON_WithJSONBlock(t *testing.T) {
	input := "```json\n{\"key\": \"value\"}\n```"
	want := `{"key": "value"}`
	got := CleanJSON(input)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCleanJSON_WithGenericBlock(t *testing.T) {
	input := "```\n{\"key\": \"value\"}\n```"
	want := `{"key": "value"}`
	got := CleanJSON(input)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCleanJSON_WithWhitespace(t *testing.T) {
	input := "  \n```json\n{\"key\": \"value\"}\n```\n  "
	want := `{"key": "value"}`
	got := CleanJSON(input)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestCleanJSON_Empty(t *testing.T) {
	got := CleanJSON("")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
