package verify

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type routingCase struct {
	id       string
	kind     string
	expected []string
	excluded []string
	prompt   string
}

func RunRouting(root, corpus string, stdout, stderr io.Writer) int {
	absolute, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(stderr, "skill-routing: racine invalide: %s\n", err)
		return 1
	}
	if corpus == "" {
		corpus = filepath.Join(absolute, "internal", "verify", "testdata", "skill-routing-cases.tsv")
	}
	v := &verifier{root: absolute, stdout: stdout, stderr: stderr}
	result, err := v.validateRouting(corpus)
	if err != nil {
		fmt.Fprintf(stderr, "skill-routing: %s\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "skill-routing: %s\n", result)
	return 0
}

func parseSlugList(value string) []string {
	if value == "-" {
		return nil
	}
	return strings.Split(value, ",")
}

func loadRoutingCases(path string) ([]routingCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("corpus absent: %s", path)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() || scanner.Text() != "id\tkind\texpected\texcluded\tprompt" {
		return nil, fmt.Errorf("corpus TSV invalide")
	}
	idPattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	seen := map[string]bool{}
	var cases []routingCase
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 5 || !idPattern.MatchString(fields[0]) || !validRoutingKind(fields[1]) || fields[4] == "" {
			return nil, fmt.Errorf("corpus TSV invalide")
		}
		if seen[fields[0]] {
			return nil, fmt.Errorf("identifiants dupliqués: %s", fields[0])
		}
		seen[fields[0]] = true
		cases = append(cases, routingCase{fields[0], fields[1], parseSlugList(fields[2]), parseSlugList(fields[3]), fields[4]})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("corpus TSV invalide: %w", err)
	}
	return cases, nil
}

func validRoutingKind(kind string) bool {
	return kind == "positive" || kind == "negative" || kind == "ambiguous"
}

func (v *verifier) routingCoverage(cases []routingCase) (map[string]map[string]bool, int, error) {
	coverage := map[string]map[string]bool{"positive": {}, "negative": {}, "ambiguous": {}}
	ambiguous := 0
	for _, item := range cases {
		for _, slug := range append(append([]string{}, item.expected...), item.excluded...) {
			if _, err := os.Stat(v.path("dot_config", "agent-skills", slug, "SKILL.md")); err != nil {
				return nil, 0, fmt.Errorf("skill inconnue: %s", slug)
			}
		}
		if err := validateRoutingContract(item); err != nil {
			return nil, 0, err
		}
		if item.kind == "ambiguous" {
			ambiguous++
		}
		for _, slug := range append(append([]string{}, item.expected...), item.excluded...) {
			coverage[item.kind][slug] = true
		}
	}
	return coverage, ambiguous, nil
}

func validateRoutingContract(item routingCase) error {
	switch item.kind {
	case "positive":
		if len(item.expected) == 0 || len(item.excluded) != 0 {
			return fmt.Errorf("%s: contrat positif invalide", item.id)
		}
	case "negative":
		if len(item.expected) != 0 || len(item.excluded) == 0 {
			return fmt.Errorf("%s: contrat négatif invalide", item.id)
		}
	case "ambiguous":
		if len(item.excluded) == 0 {
			return fmt.Errorf("%s: scénario ambigu sans exclusion", item.id)
		}
	}
	return nil
}

func (v *verifier) validateRouting(path string) (string, error) {
	cases, err := loadRoutingCases(path)
	if err != nil {
		return "", err
	}
	coverage, ambiguous, err := v.routingCoverage(cases)
	if err != nil {
		return "", err
	}
	dirs := globDirs(v.path("dot_config", "agent-skills", "*"))
	sort.Strings(dirs)
	for _, dir := range dirs {
		slug := filepath.Base(filepath.Clean(dir))
		content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
		if err != nil {
			return "", fmt.Errorf("%s: SKILL.md illisible", slug)
		}
		length := len([]rune(parseFrontmatter(string(content)).description))
		if length >= 400 {
			return "", fmt.Errorf("%s: description de %d caractères, limite locale 399", slug, length)
		}
		if !coverage["positive"][slug] {
			return "", fmt.Errorf("%s: scénario positif absent", slug)
		}
		if !coverage["negative"][slug] {
			return "", fmt.Errorf("%s: scénario négatif absent", slug)
		}
	}
	if ambiguous < 3 {
		return "", fmt.Errorf("moins de trois scénarios ambigus")
	}
	return fmt.Sprintf("%d positives, %d négatives, %d ambiguës, 0 erreur", len(dirs), len(dirs), ambiguous), nil
}

func (v *verifier) checkRouting() {
	v.head("Routage des skills")
	result, err := v.validateRouting(v.path("internal", "verify", "testdata", "skill-routing-cases.tsv"))
	if err != nil {
		v.ko("contrats de routage invalides")
		fmt.Fprint(v.stderr, indent(firstLines("skill-routing: "+err.Error(), 3)))
		return
	}
	v.ok(result)
}
