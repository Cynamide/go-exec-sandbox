package benchmark

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"

	"gexec-sandbox/internal/api"
)

// DefaultGrader preserves the existing stdout-based scoring contract used by code tasks.
type DefaultGrader struct{}

func (DefaultGrader) Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome {
	passed := compareExpectedOutput(task, resp.Stdout, tc.ExpectedOutput)

	score := 0.0
	if passed {
		score = 1.0
	}

	return Outcome{
		Passed: passed,
		Score:  score,
	}
}

func compareExpectedOutput(task Task, actual, expected string) bool {
	format := ""
	if task.ArtifactExpectation != nil {
		format = task.ArtifactExpectation.Format
	}

	switch format {
	case "markdown":
		return markdownEquivalent(actual, expected)
	case "csv":
		return csvEquivalent(actual, expected)
	case "json":
		return jsonEquivalent(actual, expected)
	default:
		return strings.TrimSpace(actual) == strings.TrimSpace(expected)
	}
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

	header, ok := splitMarkdownRow(lines[0])
	if !ok || len(header) == 0 {
		return nil, false
	}

	separator, ok := splitMarkdownRow(lines[1])
	if !ok || len(separator) != len(header) || !isMarkdownSeparatorRow(separator) {
		return nil, false
	}

	rows := make([][]string, 0, len(lines)-1)
	rows = append(rows, header)
	for _, line := range lines[2:] {
		cells, ok := splitMarkdownRow(line)
		if !ok || len(cells) != len(header) {
			return nil, false
		}
		rows = append(rows, cells)
	}

	return rows, true
}

func splitMarkdownRow(line string) ([]string, bool) {
	if !strings.Contains(line, "|") {
		return nil, false
	}

	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return nil, false
	}

	if parts[0] == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}

	return cells, len(cells) > 0
}

func isMarkdownSeparatorRow(cells []string) bool {
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if len(cell) < 3 {
			return false
		}
		hyphens := 0
		for _, r := range cell {
			switch r {
			case '-':
				hyphens++
			case ':':
			default:
				return false
			}
		}
		if hyphens < 3 {
			return false
		}
	}
	return true
}

func csvEquivalent(actual, expected string) bool {
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

func jsonEquivalent(actual, expected string) bool {
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

func decodeJSON(raw string, target *any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing json")
	}
	return nil
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
