package benchmark

import (
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strings"

	"gexec-sandbox/internal/api"
)

// DefaultGrader preserves the existing stdout-based scoring contract used by code tasks.
type DefaultGrader struct{}

func (DefaultGrader) Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome {
	passed := outputsMatch(resp.Stdout, tc.ExpectedOutput)

	score := 0.0
	if passed {
		score = 1.0
	}

	return Outcome{
		Passed: passed,
		Score:  score,
	}
}

func outputsMatch(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)

	if actual == expected {
		return true
	}

	if jsonEquivalent(actual, expected) {
		return true
	}

	if markdownEquivalent(actual, expected) {
		return true
	}

	if csvEquivalent(actual, expected) {
		return true
	}

	return false
}

func jsonEquivalent(actual, expected string) bool {
	if !(strings.HasPrefix(actual, "{") || strings.HasPrefix(actual, "[")) {
		return false
	}
	if !(strings.HasPrefix(expected, "{") || strings.HasPrefix(expected, "[")) {
		return false
	}

	var actualValue any
	if err := json.Unmarshal([]byte(actual), &actualValue); err != nil {
		return false
	}

	var expectedValue any
	if err := json.Unmarshal([]byte(expected), &expectedValue); err != nil {
		return false
	}

	return reflect.DeepEqual(actualValue, expectedValue)
}

func markdownEquivalent(actual, expected string) bool {
	actualRows, ok := parseMarkdownTable(actual)
	if !ok {
		return false
	}

	expectedRows, ok := parseMarkdownTable(expected)
	if !ok {
		return false
	}

	return reflect.DeepEqual(actualRows, expectedRows)
}

func parseMarkdownTable(raw string) ([][]string, bool) {
	lines := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
	}

	if len(lines) < 2 {
		return nil, false
	}

	rows := make([][]string, 0, len(lines)-1)
	for i, line := range lines {
		cells, ok := splitMarkdownRow(line)
		if !ok {
			return nil, false
		}
		if i == 1 {
			if !isMarkdownSeparatorRow(cells) {
				return nil, false
			}
			continue
		}
		if i == 0 || i > 1 {
			rows = append(rows, cells)
		}
	}

	return rows, true
}

func splitMarkdownRow(line string) ([]string, bool) {
	if !strings.Contains(line, "|") {
		return nil, false
	}

	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cell := strings.TrimSpace(part)
		if cell == "" {
			continue
		}
		cells = append(cells, cell)
	}

	return cells, len(cells) > 0
}

func isMarkdownSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}

	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			return false
		}
		for _, r := range cell {
			if r != '-' && r != ':' {
				return false
			}
		}
	}

	return true
}

func csvEquivalent(actual, expected string) bool {
	if !strings.Contains(expected, ",") || !strings.Contains(actual, ",") {
		return false
	}

	actualRows, err := parseCSV(actual)
	if err != nil {
		return false
	}

	expectedRows, err := parseCSV(expected)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(actualRows, expectedRows)
}

func parseCSV(raw string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(strings.TrimSpace(raw)))
	reader.TrimLeadingSpace = true
	return reader.ReadAll()
}
