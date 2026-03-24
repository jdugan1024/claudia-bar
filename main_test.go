// ABOUTME: Tests for the Claude Code status line formatter.
// ABOUTME: Validates formatting of model, path, tokens, cost, and git branch.

package main

import (
	"strings"
	"testing"
)

func TestShortenPath(t *testing.T) {
	cases := []struct {
		input    string
		home     string
		expected string
	}{
		{"/Users/jon/projects/foo", "/Users/jon", "~/projects/foo"},
		{"/Users/jon", "/Users/jon", "~"},
		{"/etc/something", "/Users/jon", "/etc/something"},
	}
	for _, c := range cases {
		got := shortenPath(c.input, c.home)
		if got != c.expected {
			t.Errorf("shortenPath(%q, %q) = %q, want %q", c.input, c.home, got, c.expected)
		}
	}
}

func TestFormatTokens(t *testing.T) {
	cases := []struct {
		n        int
		expected string
	}{
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{15234, "15.2k"},
		{1000000, "1000.0k"},
	}
	for _, c := range cases {
		got := formatTokens(c.n)
		if got != c.expected {
			t.Errorf("formatTokens(%d) = %q, want %q", c.n, got, c.expected)
		}
	}
}

func TestFormat(t *testing.T) {
	input := Input{
		CWD: "/Users/jon/projects/myapp",
		Model: Model{
			DisplayName: "Opus",
		},
		ContextWindow: ContextWindow{
			UsedPercentage:    25.4,
			TotalInputTokens:  15234,
			TotalOutputTokens: 4521,
		},
		Cost: Cost{
			TotalCostUSD: 0.01234,
		},
	}

	result := format(input, "main", "/Users/jon")

	checks := []string{"Opus", "myapp", "main", "15.2k", "4.5k", "25%", "$0.0123"}
	for _, want := range checks {
		if !strings.Contains(result, want) {
			t.Errorf("expected %q in output\ngot: %s", want, result)
		}
	}
}

func TestFormatNoGitBranch(t *testing.T) {
	input := Input{
		CWD:   "/tmp/scratch",
		Model: Model{DisplayName: "Sonnet"},
		ContextWindow: ContextWindow{
			UsedPercentage:    5.0,
			TotalInputTokens:  100,
			TotalOutputTokens: 50,
		},
		Cost: Cost{TotalCostUSD: 0.001},
	}

	result := format(input, "", "/Users/jon")

	if strings.Contains(result, "[") {
		t.Errorf("expected no branch brackets when branch is empty, got: %s", result)
	}
}
