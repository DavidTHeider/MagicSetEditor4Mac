package mcp_test

import (
	"fmt"
	"io/fs"
	"strings"
	"syscall"
	"testing"
)

// **Eine volle Platte heißt so, und ihr Pfad bleibt draußen** —
// #1812.
//
// Das Modell bekam `write /Users/…: no space left on device` — einen Pfad,
// den §8 nicht hinauslässt, und keinen Rat.
func TestEineVollePlatteNenntDenRatUndKeinenPfad(t *testing.T) {
	bestand := probenbestand()
	bestand.fehler = fmt.Errorf("die Sets lesen: %w",
		&fs.PathError{Op: "write", Path: "/Users/jemand/geheim", Err: syscall.ENOSPC})

	zeilen := sprich(t, bestand,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.sets","arguments":{}}}`)

	ergebnis := abbildung(t, entpacke(t, zeilen[0])["result"])
	if ergebnis["isError"] != true {
		t.Fatalf("Ergebnis: %v — erwartet war isError", ergebnis)
	}

	text := fmt.Sprint(ergebnis["content"])
	if strings.Contains(text, "/Users/jemand") || !strings.Contains(text, "kein Platz") {
		t.Errorf("die Antwort: %s", text)
	}
}
