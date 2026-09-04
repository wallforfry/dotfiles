head_ "Télémétrie du harness"
n=$(grep -c '^    def test_' scripts/test_harness_telemetry.py)
if output=$(PYTHONDONTWRITEBYTECODE=1 python3 -m unittest -q scripts/test_harness_telemetry.py 2>&1); then
  ok "$n/$n tests de normalisation et de cache"
else
  ko "tests de télémétrie rouges"
  printf '%s\n' "$output" | head -5 | sed 's/^/      /'
fi
