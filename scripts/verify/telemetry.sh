head_ "Télémétrie du harness"
n=$(grep -h '^    def test_' scripts/test_harness_telemetry*.py | wc -l | tr -d ' ')
if output=$(PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_harness_telemetry*.py' -q 2>&1); then
  ok "$n/$n tests de normalisation et de cache"
else
  ko "tests de télémétrie rouges"
  printf '%s\n' "$output" | head -5 | sed 's/^/      /'
fi
