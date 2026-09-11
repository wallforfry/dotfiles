package verify

import (
	"fmt"
	"strings"
)

var mergeVerdictPresentationMarkers = []string{
	"Decision: <changes required | approved with reservations | approved>",
	"Next: <one executable action>",
	"Status: <head SHA, merge state, checks, conflicts and open tasks>",
	"Group ledger rows in visible sets of five when more are needed; retain every row.",
	"Blockers: <location, cause, failure mechanism and lift criterion>",
}

var mergeVerdictPresentationOrder = []string{
	"Decision: <changes required | approved with reservations | approved>",
	"Next: <one executable action>",
	"Status: <head SHA, merge state, checks, conflicts and open tasks>",
	"<Anchor sentence - REQUIRED.",
	"<Contract and behaviour ledger - REQUIRED",
	"Group ledger rows in visible sets of five when more are needed; retain every row.",
	"Blockers: <location, cause, failure mechanism and lift criterion>",
}

func validateMergeVerdictPresentation(content string) error {
	for _, marker := range mergeVerdictPresentationMarkers {
		if !strings.Contains(content, marker) {
			return fmt.Errorf("marqueur de présentation absent: %s", marker)
		}
	}
	previous := -1
	for _, marker := range mergeVerdictPresentationOrder {
		position := strings.Index(content, marker)
		if position <= previous {
			return fmt.Errorf("ordre de présentation invalide autour de: %s", marker)
		}
		previous = position
	}
	return nil
}
