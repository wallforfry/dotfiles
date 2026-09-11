package verify

import (
	"fmt"
	"strings"
)

func validateBusinessIssueOpening(content string) error {
	lines := nonemptyLines(content)
	if len(lines) < 2 {
		return fmt.Errorf("ouverture incomplète")
	}
	first, hasOutcome := openingValue(lines[0], "Outcome: ")
	if !hasOutcome {
		first, hasOutcome = openingValue(lines[0], "Decision: ")
	}
	if !hasOutcome {
		return fmt.Errorf("le premier élément doit annoncer Outcome ou Decision")
	}
	next, hasNext := openingValue(lines[1], "Next: ")
	if !hasNext {
		return fmt.Errorf("le second élément doit annoncer Next")
	}
	if first == "" {
		return fmt.Errorf("Outcome ou Decision est vide")
	}
	if next == "" {
		return fmt.Errorf("Next est vide")
	}
	return nil
}

func validateBusinessIssueRegressionFixture(content string) error {
	if err := validateBusinessIssueOpening(content); err != nil {
		return err
	}
	for _, heading := range []string{
		"## Contract", "### Acceptance scenarios", "### Completion evidence", "## Progress", "## Errors",
	} {
		if !strings.Contains(content, heading) {
			return fmt.Errorf("section obligatoire absente: %s", heading)
		}
	}
	if err := validateBusinessIssueListBounds(content); err != nil {
		return err
	}
	if err := validateBusinessIssueProgress(content); err != nil {
		return err
	}
	for _, field := range []string{"Location: ", "Cause: ", "Correction: "} {
		if _, found := openingValueAfterHeading(content, "## Errors", field); !found {
			return fmt.Errorf("champ d'erreur obligatoire absent: %s", strings.TrimSuffix(field, ": "))
		}
	}
	return nil
}

func validateBusinessIssueProgress(content string) error {
	lines, found := linesAfterHeading(content, "## Progress")
	if !found || len(lines) == 0 || !strings.HasPrefix(lines[0], "Step ") {
		return fmt.Errorf("Progress doit annoncer l'étape courante")
	}
	if len(lines) < 2 {
		return fmt.Errorf("Progress doit annoncer Next")
	}
	if next, found := openingValue(lines[1], "Next: "); !found || next == "" {
		return fmt.Errorf("Progress doit annoncer Next")
	}
	return nil
}

func validateBusinessIssueListBounds(content string) error {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "- ") || orderedListItem(line) {
			count++
			if count > 5 {
				return fmt.Errorf("liste de plus de cinq éléments")
			}
			continue
		}
		count = 0
	}
	return nil
}

func orderedListItem(line string) bool {
	index := strings.Index(line, ". ")
	if index <= 0 {
		return false
	}
	for _, character := range line[:index] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func openingValueAfterHeading(content, heading, prefix string) (string, bool) {
	lines, found := linesAfterHeading(content, heading)
	if !found {
		return "", false
	}
	for _, line := range lines {
		if value, found := openingValue(line, prefix); found && value != "" {
			return value, true
		}
	}
	return "", false
}

func linesAfterHeading(content, heading string) ([]string, bool) {
	index := strings.Index(content, heading)
	if index < 0 {
		return nil, false
	}
	section := content[index+len(heading):]
	if next := strings.Index(section, "\n## "); next >= 0 {
		section = section[:next]
	}
	lines := []string{}
	for _, line := range strings.Split(section, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines, true
}

func validateBusinessIssueInstruction(content string) error {
	normalized := strings.Join(strings.Fields(content), " ")
	for _, requirement := range []string{
		"Make the draft action-first: begin with the current outcome or decision and one next action",
		"Keep each visible group to five items where possible",
		"acceptance criteria, evidence and unresolved decisions remain complete",
		"State progress explicitly for multi-step work and describe errors by location, cause and correction",
		"Never trade away a material requirement, refusal, prerequisite or evidence obligation for brevity",
	} {
		if !strings.Contains(normalized, requirement) {
			return fmt.Errorf("instruction obligatoire absente: %s", requirement)
		}
	}
	return nil
}

func openingValue(line, prefix string) (string, bool) {
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
}
