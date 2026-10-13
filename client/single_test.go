package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vszholobov/tetrisLib/single"
)

func TestRenderSingle(t *testing.T) {
	var output bytes.Buffer
	renderSingle(&output, single.New(1).State())
	lines := strings.Split(output.String(), "\r\n")
	if len(lines) != 24 {
		t.Fatalf("unexpected frame height: %d", len(lines))
	}
	if lines[21] != "############\x1b[K" {
		t.Fatalf("floor missing: %q", lines[21])
	}
	for _, line := range lines[1:21] {
		if len(line) != 15 || line[0] != '#' || line[11] != '#' {
			t.Fatalf("invalid board row: %q", line)
		}
	}
	if !strings.Contains(lines[22], "Score: 0") {
		t.Fatal("score missing")
	}
}

func TestSingleInvalidFlagsWithoutTTY(t *testing.T) {
	if err := runSingle([]string{"--seed", "not-a-number"}); err == nil {
		t.Fatal("invalid seed accepted")
	}
	if err := runSingle([]string{"extra"}); err == nil {
		t.Fatal("unexpected argument accepted")
	}
}
