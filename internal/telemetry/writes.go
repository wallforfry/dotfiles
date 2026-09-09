package telemetry

import (
	"fmt"
	"slices"
	"strings"
)

func contentLines(value event) ([]string, bool) {
	name := strings.ToLower(stringValue(value.name))
	payload, isObject := object(value.payload)
	switch name {
	case "write":
		return editorLines(payload, isObject, []string{"file_path", "path"}, "content", codeExtensions)
	case "edit":
		return editorLines(payload, isObject, []string{"file_path", "path"}, "new_string", codeExtensions)
	case "multiedit":
		return multiEditLines(payload, isObject), true
	case "notebookedit":
		return editorLines(payload, isObject, []string{"notebook_path", "file_path"}, "new_source", []string{".ipynb"})
	case "filechange":
		if !isObject {
			return nil, false
		}
		return fileChangeLines(payload)
	case "apply_patch":
		text, ok := value.payload.(string)
		if !ok {
			return nil, false
		}
		return patchLines(text), true
	default:
		return nil, false
	}
}

func editorLines(payload map[string]any, ok bool, pathKeys []string, contentKey string, extensions []string) ([]string, bool) {
	if !ok {
		return nil, false
	}
	path := firstTruthyString(payload, pathKeys...)
	if !hasExtension(path, extensions) {
		return []string{}, true
	}
	return strings.Split(stringValueOrEmpty(payload[contentKey]), "\n"), true
}

func multiEditLines(payload map[string]any, ok bool) []string {
	if !ok || !hasExtension(firstTruthyString(payload, "file_path", "path"), codeExtensions) {
		return []string{}
	}
	edits, ok := payload["edits"].([]any)
	if !ok {
		return []string{}
	}
	var lines []string
	for _, rawEdit := range edits {
		if edit, ok := object(rawEdit); ok {
			lines = append(lines, strings.Split(stringValueOrEmpty(edit["new_string"]), "\n")...)
		}
	}
	return lines
}

func fileChangeLines(payload map[string]any) ([]string, bool) {
	changes, ok := object(payload["changes"])
	if !ok {
		return nil, false
	}
	var lines []string
	for path, rawChange := range changes {
		change, ok := object(rawChange)
		if !hasExtension(path, codeExtensions) || !ok {
			continue
		}
		rawContent := change["unified_diff"]
		if rawContent == nil || rawContent == "" {
			rawContent = change["content"]
		}
		content, ok := rawContent.(string)
		if !ok {
			return nil, false
		}
		switch change["type"] {
		case "add":
			lines = append(lines, splitLines(content)...)
		case "update":
			for _, line := range splitLines(content) {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					lines = append(lines, line[1:])
				}
			}
		}
	}
	return lines, true
}

func patchLines(payload string) []string {
	text := strings.ReplaceAll(payload, `\n`, "\n")
	selected := false
	var lines []string
	for _, line := range splitLines(text) {
		switch {
		case strings.HasPrefix(line, "*** Add File: "), strings.HasPrefix(line, "*** Update File: "):
			selected = hasExtension(strings.SplitN(line, ": ", 2)[1], codeExtensions)
		case strings.HasPrefix(line, "*** Delete File: "), strings.HasPrefix(line, "*** End Patch"):
			selected = false
		case strings.HasPrefix(line, "+++ "):
			path := strings.TrimPrefix(line[4:], "b/")
			selected = hasExtension(path, codeExtensions)
		case selected && strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			lines = append(lines, line[1:])
		}
	}
	return lines
}

func splitLines(value string) []string {
	value = strings.TrimSuffix(value, "\n")
	if value == "" {
		return []string{}
	}
	return strings.Split(value, "\n")
}

func hasExtension(path string, extensions []string) bool {
	return slices.ContainsFunc(extensions, func(extension string) bool { return strings.HasSuffix(path, extension) })
}

func firstTruthyString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(values[key]); value != "" {
			return value
		}
	}
	return ""
}

func stringValueOrEmpty(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
