package verify

import (
	"fmt"
	"os"
	"regexp"
)

var goTestPattern = regexp.MustCompile(`(?m)^func Test[A-Za-z0-9_]+\(`)

func (v *verifier) checkTelemetry() {
	v.head("Télémétrie du harness")
	tests := 0
	files := glob(v.path("internal", "telemetry", "*_test.go"))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			v.ko(fmt.Sprintf("%s illisible", relative(v.root, file)))
			return
		}
		tests += len(goTestPattern.FindAll(content, -1))
	}
	if len(files) == 0 || tests == 0 {
		v.ko("tests de télémétrie absents")
		return
	}
	output, err := v.command(nil, "go", "test", "./internal/telemetry")
	if err != nil {
		v.ko("tests de télémétrie rouges")
		fmt.Fprint(v.stderr, indent(firstLines(string(output), 5)))
		return
	}
	v.ok(fmt.Sprintf("%d/%d tests de normalisation et de cache", tests, tests))
}
