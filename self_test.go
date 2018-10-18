package main

import "testing"

func TestDepsForPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	expected := []Dependency{{
		Path:    "golang.org/x/tools",
		Version: "v0.0.0-20181004163742-59602fdee893",
	}}
	// Must run from repo, should be OK for Travis
	deps := DepsForPath(".")

	if len(expected) != len(deps) {
		t.Fail()
	}

	for i := 0; i < len(deps); i++ {
		if deps[i] != expected[i] {
			t.Fail()
		}
	}
}
