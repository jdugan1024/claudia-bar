// ABOUTME: claudia-status — displays model, directory, git branch, token usage, cost, and rate limits.
// ABOUTME: Reads JSON from stdin (Claude Code status hook format) and writes a formatted line to stdout.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Model struct {
	DisplayName string `json:"display_name"`
}

type ContextWindow struct {
	UsedPercentage    float64 `json:"used_percentage"`
	TotalInputTokens  int     `json:"total_input_tokens"`
	TotalOutputTokens int     `json:"total_output_tokens"`
}

type Cost struct {
	TotalCostUSD float64 `json:"total_cost_usd"`
}

type FiveHourLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

type RateLimits struct {
	FiveHour FiveHourLimit `json:"five_hour"`
}

type Input struct {
	CWD           string        `json:"cwd"`
	Model         Model         `json:"model"`
	ContextWindow ContextWindow `json:"context_window"`
	Cost          Cost          `json:"cost"`
	RateLimits    RateLimits    `json:"rate_limits"`
}

func shortenPath(path, home string) string {
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

func formatTokens(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000)
}

func gitBranch(dir string) string {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func progressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return fmt.Sprintf("%d%% %s", int(pct), strings.Repeat("█", filled)+strings.Repeat("░", width-filled))
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func format(input Input, branch, home string) string {
	path := shortenPath(input.CWD, home)

	loc := path
	if branch != "" {
		loc = fmt.Sprintf("%s [%s]", path, branch)
	}

	tokens := fmt.Sprintf("in:%s out:%s [%s]",
		formatTokens(input.ContextWindow.TotalInputTokens),
		formatTokens(input.ContextWindow.TotalOutputTokens),
		progressBar(input.ContextWindow.UsedPercentage, 10),
	)

	cost := fmt.Sprintf("$%.4f", input.Cost.TotalCostUSD)

	parts := []string{input.Model.DisplayName, loc, tokens, cost}

	if rl := input.RateLimits.FiveHour; rl.ResetsAt > 0 {
		until := formatDuration(time.Until(time.Unix(rl.ResetsAt, 0)))
		parts = append(parts, fmt.Sprintf("5h [%s] %s", progressBar(rl.UsedPercentage, 10), until))
	}

	return strings.Join(parts, " | ")
}

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()
	branch := gitBranch(input.CWD)

	fmt.Println(format(input, branch, home))
}
