package benchmark

import (
	"encoding/csv"
	"encoding/json"
	"math/big"
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
	if err := decodeJSON(actual, &actualValue); err != nil {
		return false
	}

	var expectedValue any
	if err := decodeJSON(expected, &expectedValue); err != nil {
		return false
	}

	return jsonValuesEqual(actualValue, expectedValue)
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

func decodeJSON(raw string, target *any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(target)
}

func jsonValuesEqual(actual, expected any) bool {
	switch actualTyped := actual.(type) {
	case map[string]any:
		expectedTyped, ok := expected.(map[string]any)
		if !ok || len(actualTyped) != len(expectedTyped) {
			return false
		}
		for key, actualValue := range actualTyped {
			expectedValue, ok := expectedTyped[key]
			if !ok || !jsonValuesEqual(actualValue, expectedValue) {
				return false
			}
		}
		return true
	case []any:
		expectedTyped, ok := expected.([]any)
		if !ok || len(actualTyped) != len(expectedTyped) {
			return false
		}
		for i := range actualTyped {
			if !jsonValuesEqual(actualTyped[i], expectedTyped[i]) {
				return false
			}
		}
		return true
	case json.Number:
		expectedTyped, ok := expected.(json.Number)
		if !ok {
			return false
		}
		return jsonNumbersEqual(actualTyped, expectedTyped)
	case string:
		expectedTyped, ok := expected.(string)
		return ok && actualTyped == expectedTyped
	case bool:
		expectedTyped, ok := expected.(bool)
		return ok && actualTyped == expectedTyped
	case nil:
		return expected == nil
	default:
		return reflect.DeepEqual(actual, expected)
	}
}

func jsonNumbersEqual(actual, expected json.Number) bool {
	actualRat, ok := new(big.Rat).SetString(actual.String())
	if !ok {
		return false
	}

	expectedRat, ok := new(big.Rat).SetString(expected.String())
	if !ok {
		return false
	}

	return actualRat.Cmp(expectedRat) == 0
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
	var headerWidth int
	for i, line := range lines {
		cells, ok := splitMarkdownRow(line)
		if !ok {
			return nil, false
		}
		if i == 1 {
			headerWidth = len(rows[0])
			if !isMarkdownSeparatorRow(cells, headerWidth) {
				return nil, false
			}
			continue
		}
		if i == 0 || i > 1 {
			if i == 0 {
				headerWidth = len(cells)
			} else if len(cells) != headerWidth {
				return nil, false
			}
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
	if len(parts) < 3 {
		return nil, false
	}

	cells := make([]string, 0, len(parts)-2)
	for i := 1; i < len(parts)-1; i++ {
		cells = append(cells, strings.TrimSpace(parts[i]))
	}

	return cells, true
}

func isMarkdownSeparatorRow(cells []string, width int) bool {
	if len(cells) != width || width == 0 {
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
	if strings.Count(actual, "\n") < 1 || strings.Count(expected, "\n") < 1 {
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
