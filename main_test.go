// ABOUTME: Tests for claudia-status formatter.
// ABOUTME: Validates formatting of model, path, tokens, cost, rate limits, and git branch.

package main

import (
	"strings"
	"testing"
	"time"
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

	checks := []string{"Opus", "myapp", "main", "15.2k", "4.5k", "██", "$0.0123"}
	for _, want := range checks {
		if !strings.Contains(result, want) {
			t.Errorf("expected %q in output\ngot: %s", want, result)
		}
	}
}

func TestProgressBar(t *testing.T) {
	cases := []struct {
		pct      float64
		width    int
		expected string
	}{
		{0, 10, "0% ░░░░░░░░░░"},
		{100, 10, "100% ██████████"},
		{50, 10, "50% █████░░░░░"},
		{25, 10, "25% ██░░░░░░░░"},
		{42, 10, "42% ████░░░░░░"},
		{0, 5, "0% ░░░░░"},
		{100, 5, "100% █████"},
	}
	for _, c := range cases {
		got := progressBar(c.pct, c.width)
		if got != c.expected {
			t.Errorf("progressBar(%.0f, %d) = %q, want %q", c.pct, c.width, got, c.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d        time.Duration
		expected string
	}{
		{0, "0m"},
		{-1 * time.Minute, "0m"},
		{30 * time.Minute, "30m"},
		{90 * time.Minute, "1h30m"},
		{2*time.Hour + 5*time.Minute, "2h05m"},
		{10*time.Hour + 0*time.Minute, "10h00m"},
	}
	for _, c := range cases {
		got := formatDuration(c.d)
		if got != c.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", c.d, got, c.expected)
		}
	}
}

func TestFormatWithRateLimits(t *testing.T) {
	resetsAt := time.Now().Add(2*time.Hour + 30*time.Minute).Unix()
	input := Input{
		CWD:   "/Users/jon/projects/myapp",
		Model: Model{DisplayName: "Opus"},
		ContextWindow: ContextWindow{
			UsedPercentage:    25.0,
			TotalInputTokens:  1000,
			TotalOutputTokens: 500,
		},
		Cost: Cost{TotalCostUSD: 0.01},
		RateLimits: RateLimits{
			FiveHour: FiveHourLimit{
				UsedPercentage: 42.5,
				ResetsAt:       resetsAt,
			},
		},
	}

	result := format(input, "main", "/Users/jon")

	if !strings.Contains(result, "5h [") {
		t.Errorf("expected rate limit bar in output, got: %s", result)
	}
	if !strings.Contains(result, "h") {
		t.Errorf("expected hours in reset time, got: %s", result)
	}
}

func TestFormatWithoutRateLimits(t *testing.T) {
	input := Input{
		CWD:   "/Users/jon/projects/myapp",
		Model: Model{DisplayName: "Opus"},
		ContextWindow: ContextWindow{
			UsedPercentage:    25.0,
			TotalInputTokens:  1000,
			TotalOutputTokens: 500,
		},
		Cost: Cost{TotalCostUSD: 0.01},
	}

	result := format(input, "main", "/Users/jon")

	if strings.Contains(result, "5h [") {
		t.Errorf("expected no rate limit section when ResetsAt is 0, got: %s", result)
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

	if strings.Contains(result, "/tmp/scratch [") {
		t.Errorf("expected no branch brackets when branch is empty, got: %s", result)
	}
}
