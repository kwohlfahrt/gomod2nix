package main

import "testing"

func TestDepsForPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	expected := map[Dependency]struct{}{{
		Path:    "golang.org/x/tools",
		Version: "v0.0.0-20181004163742-59602fdee893",
	}: struct{}{}}
	// Must run from repo, should be OK for Travis
	deps := DepsForPath(".")

	if len(expected) != len(deps) {
		t.Fail()
	}

	for dep, _ := range deps {
		if _, ok := deps[dep]; !ok {
			t.Fail()
		}
	}
}
