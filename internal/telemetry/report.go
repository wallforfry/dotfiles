package telemetry

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
)

func Rate(numerator, denominator int) string {
	if denominator == 0 {
		return "inconnu"
	}
	return fmt.Sprintf("%.3f", float64(numerator)/float64(denominator))
}

func ActivatedBodyCost(summary Summary, localSizes map[string]int64) (int, int64) {
	activations := 0
	var bodyBytes int64
	for name, size := range localSizes {
		count := summary.Skills[name]
		activations += count
		bodyBytes += int64(count) * size
	}
	return activations, bodyBytes
}

func PrintReport(writer io.Writer, sources map[string]int, total Summary, bySource map[string]Summary, hits, elapsed int, localSizes map[string]int64) {
	printSourceOverview(writer, sources, total, hits, elapsed)
	printActivations(writer, sources, total, bySource, localSizes)
	printMeasurements(writer, total)
}

func printSourceOverview(writer io.Writer, sources map[string]int, total Summary, hits, elapsed int) {
	sourceNames := sortedKeys(sources)
	parts := make([]string, 0, len(sourceNames))
	for _, source := range sourceNames {
		parts = append(parts, fmt.Sprintf("%s %d sessions", source, sources[source]))
	}
	fmt.Fprintf(writer, "  sources : %s\n", join(parts))
	if len(total.Days) == 0 {
		fmt.Fprintln(writer, "  période : inconnue")
	} else {
		fmt.Fprintf(writer, "  période : %s au %s\n", slices.Min(total.Days), slices.Max(total.Days))
	}
	fmt.Fprintf(writer, "  cache : %d/%d sessions réutilisées, %d ms\n", hits, sumValues(sources), elapsed)
	fmt.Fprintf(writer, "  formats : %d enregistrements lus, %d enregistrements reconnus, %d inconnus, %d invalides\n",
		sumValues(total.ReadRecords), sumValues(total.RecognizedRecords),
		sumValues(total.UnknownRecords), sumValues(total.InvalidRecords))
}

func printActivations(writer io.Writer, sources map[string]int, total Summary, bySource map[string]Summary, localSizes map[string]int64) {
	fmt.Fprintln(writer, "  signal d'activation des skills :")
	for _, source := range sortedKeys(bySource) {
		activations := sumValues(bySource[source].Skills)
		status := "inconnu, aucun événement explicite"
		if activations > 0 {
			status = fmt.Sprintf("%d événements explicites", activations)
		}
		fmt.Fprintf(writer, "    %-8s %s\n", source, status)
	}
	fmt.Fprintln(writer, "  coût amorti des corps de skills locales :")
	for _, source := range sortedKeys(bySource) {
		activations, bodyBytes := ActivatedBodyCost(bySource[source], localSizes)
		if sumValues(bySource[source].Skills) == 0 {
			fmt.Fprintf(writer, "    %-8s inconnu\n", source)
			continue
		}
		average := 0.0
		if sources[source] != 0 {
			average = float64(bodyBytes) / float64(sources[source])
		}
		fmt.Fprintf(writer, "    %-8s %d octets sur %d activations, %.1f octets/session\n", source, bodyBytes, activations, average)
	}
	localSkills := make(map[string]int, len(localSizes))
	for name := range localSizes {
		localSkills[name] = total.Skills[name]
	}
	printCounter(writer, "skills de ce dépôt", localSkills, 0)
	externalSkills := maps.Clone(total.Skills)
	for name := range localSizes {
		delete(externalSkills, name)
	}
	printCounter(writer, "autres skills activées, top 5", externalSkills, 5)
	printCounter(writer, "subagents", total.Agents, 6)
	fmt.Fprintf(writer, "  qualification absente : inconnue %d, explicitement nulle %d\n",
		total.Skills[Unknown]+total.Agents[Unknown], total.Skills[None]+total.Agents[None])
}

func printMeasurements(writer io.Writer, total Summary) {
	fmt.Fprintln(writer, "  ponctuation interdite par bloc de texte assistant :")
	for _, period := range Eras {
		blocks := total.Blocks[period]
		dash := total.Dash[period]
		dot := total.MiddleDot[period]
		fmt.Fprintf(writer, "    %-8s %6d blocs  cadratin %5d (%s)  point médian %5d (%s)\n",
			period, blocks, dash, Rate(dash, blocks), dot, Rate(dot, blocks))
	}
	fmt.Fprintln(writer, "  lignes de commentaire dans le code écrit :")
	for _, period := range Eras {
		lines := total.Lines[period]
		comments := total.Comments[period]
		fmt.Fprintf(writer, "    %-8s %6d lignes  %6d commentaires  %s\n", period, lines, comments, Rate(comments, lines))
	}
	printCounter(writer, "canaux d'écriture observés", total.WriteTools, 0)
	if len(total.UninspectableWrites) > 0 {
		printCounter(writer, "écritures au contenu inconnu", total.UninspectableWrites, 0)
	}
}

func printCounter(writer io.Writer, title string, values map[string]int, limit int) {
	fmt.Fprintf(writer, "  %s :\n", title)
	type pair struct {
		name  string
		count int
	}
	pairs := make([]pair, 0, len(values))
	for name, count := range values {
		pairs = append(pairs, pair{name: name, count: count})
	}
	sort.Slice(pairs, func(left, right int) bool {
		if pairs[left].count != pairs[right].count {
			return pairs[left].count > pairs[right].count
		}
		return pairs[left].name < pairs[right].name
	})
	if limit > 0 && len(pairs) > limit {
		pairs = pairs[:limit]
	}
	for _, entry := range pairs {
		fmt.Fprintf(writer, "    %5d  %s\n", entry.count, entry.name)
	}
}

func sortedKeys[T any](values map[string]T) []string {
	keys := slices.Collect(maps.Keys(values))
	slices.Sort(keys)
	return keys
}

func sumValues(values map[string]int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func join(values []string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += ", "
		}
		result += value
	}
	return result
}
