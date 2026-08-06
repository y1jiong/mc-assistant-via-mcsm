package main

import "testing"

func TestParseOptions(t *testing.T) {
	t.Parallel()

	opts, err := parseOptions([]string{
		"--data", "custom.json",
		"--insecure",
		"--tp-team", "red",
		"--tp-count-per-coordinate", "3",
		"--coordinate-file", "coordinates.txt",
		"--delay", "600",
	})
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}

	if opts.dataFile != "custom.json" || !opts.insecure || opts.tpTeam != "red" {
		t.Fatalf("parseOptions() returned unexpected options: %+v", opts)
	}
	if opts.tpCountPerCoordinate != 3 || opts.coordinateFile != "coordinates.txt" || opts.delayMilliseconds != 600 {
		t.Fatalf("parseOptions() returned unexpected options: %+v", opts)
	}
}

func TestParseOptionsDefaults(t *testing.T) {
	t.Parallel()

	opts, err := parseOptions(nil)
	if err != nil {
		t.Fatalf("parseOptions() error = %v", err)
	}
	if opts.tpCountPerCoordinate != 1 {
		t.Fatalf("tpCountPerCoordinate = %d, want 1", opts.tpCountPerCoordinate)
	}
}
