package telemetry

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
)

const CacheVersion = 14

var summaryFields = []func(*Summary) map[string]int{
	func(summary *Summary) map[string]int { return summary.Skills },
	func(summary *Summary) map[string]int { return summary.Agents },
	func(summary *Summary) map[string]int { return summary.Blocks },
	func(summary *Summary) map[string]int { return summary.Dash },
	func(summary *Summary) map[string]int { return summary.MiddleDot },
	func(summary *Summary) map[string]int { return summary.Lines },
	func(summary *Summary) map[string]int { return summary.Comments },
	func(summary *Summary) map[string]int { return summary.WriteTools },
	func(summary *Summary) map[string]int { return summary.UninspectableWrites },
	func(summary *Summary) map[string]int { return summary.ReadRecords },
	func(summary *Summary) map[string]int { return summary.RecognizedRecords },
	func(summary *Summary) map[string]int { return summary.UnknownRecords },
	func(summary *Summary) map[string]int { return summary.InvalidRecords },
}

type Summary struct {
	Source              string         `json:"source"`
	Days                []string       `json:"days"`
	Skills              map[string]int `json:"skills"`
	Agents              map[string]int `json:"agents"`
	Blocks              map[string]int `json:"blocks"`
	Dash                map[string]int `json:"dash"`
	MiddleDot           map[string]int `json:"middle_dot"`
	Lines               map[string]int `json:"lines"`
	Comments            map[string]int `json:"comments"`
	WriteTools          map[string]int `json:"write_tools"`
	UninspectableWrites map[string]int `json:"uninspectable_writes"`
	ReadRecords         map[string]int `json:"read_records"`
	RecognizedRecords   map[string]int `json:"recognized_records"`
	UnknownRecords      map[string]int `json:"unknown_records"`
	InvalidRecords      map[string]int `json:"invalid_records"`
}

func EmptySummary(source string) Summary {
	return Summary{
		Source: source, Days: []string{}, Skills: map[string]int{}, Agents: map[string]int{},
		Blocks: map[string]int{}, Dash: map[string]int{}, MiddleDot: map[string]int{}, Lines: map[string]int{},
		Comments: map[string]int{}, WriteTools: map[string]int{}, UninspectableWrites: map[string]int{},
		ReadRecords: map[string]int{}, RecognizedRecords: map[string]int{},
		UnknownRecords: map[string]int{}, InvalidRecords: map[string]int{},
	}
}

func SummarizeFile(source, path, since string) (Summary, error) {
	file, err := os.Open(path)
	if err != nil {
		return Summary{}, &ReadInterruptedError{Source: source, Cause: err}
	}
	defer file.Close()
	summary := EmptySummary(source)
	reader := bufio.NewReader(file)
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			summarizeRecord(&summary, source, since, line)
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return Summary{}, &ReadInterruptedError{Source: source, Cause: readErr}
		}
	}
	return summary, nil
}

func summarizeRecord(summary *Summary, source, since string, line []byte) {
	increment(summary.ReadRecords, source, 1)
	var record map[string]any
	if json.Unmarshal(line, &record) != nil || record == nil {
		increment(summary.InvalidRecords, source, 1)
		return
	}
	if !recognizedRecord(source, record) {
		increment(summary.UnknownRecords, source, 1)
		return
	}
	increment(summary.RecognizedRecords, source, 1)
	if timestamp, ok := record["timestamp"].(string); ok && datePrefix(timestamp) {
		summary.Days = append(summary.Days, timestamp[:10])
	}
	for _, event := range normalizedEvents(source, record) {
		summarizeEvent(summary, since, event)
	}
}

func summarizeEvent(summary *Summary, since string, value event) {
	period := era(value.timestamp, since)
	if value.kind == "text" && value.role == "assistant" {
		if text, ok := value.text.(string); ok {
			increment(summary.Blocks, period, 1)
			increment(summary.Dash, period, strings.Count(text, "\u2014"))
			increment(summary.MiddleDot, period, strings.Count(text, "\u00b7"))
		}
		return
	}
	if value.kind != "tool" {
		return
	}
	name := strings.ToLower(stringValue(value.name))
	switch name {
	case "skill":
		increment(summary.Skills, labelled(value.payload, "skill", "name"), 1)
	case "agent", "task", "spawn_agent":
		label := labelled(value.payload, "subagent_type", "agent_type")
		if fallback := sourceLabel(value, name); label == Unknown && fallback != "" {
			label = fallback
		}
		increment(summary.Agents, label, 1)
	}
	if !slices.Contains(explicitWriters, name) {
		if isPotentialWrite(value) {
			increment(summary.WriteTools, name, 1)
			increment(summary.UninspectableWrites, name, 1)
		}
		return
	}
	increment(summary.WriteTools, name, 1)
	lines, inspectable := contentLines(value)
	if !inspectable {
		increment(summary.UninspectableWrites, name, 1)
		return
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		increment(summary.Lines, period, 1)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			increment(summary.Comments, period, 1)
		}
	}
}

func sourceLabel(value event, name string) string {
	if value.source == "codex" && name == "spawn_agent" {
		return "spawn_agent"
	}
	if value.source == "claude" && name == "agent" {
		return "agent"
	}
	return ""
}

func Merge(summaries []Summary) Summary {
	result := EmptySummary("all")
	for _, summary := range summaries {
		result.Days = append(result.Days, summary.Days...)
		for _, resultField := range summaryFields {
			for key, value := range resultField(&summary) {
				increment(resultField(&result), key, value)
			}
		}
	}
	return result
}

func increment(values map[string]int, key string, amount int) {
	values[key] += amount
}
