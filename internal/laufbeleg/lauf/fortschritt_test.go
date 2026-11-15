package main

import (
	"bytes"
	"strings"
	"testing"
)

// **Der Fortschritt kommt Zeile für Zeile** — auch wenn der Strom mitten in
// einer Zeile abreißt, und ohne dass ein Unterfall oder ein Paket eine Zeile
// bekommt.
func TestDerFortschrittFolgtDemStrom(t *testing.T) {
	t.Parallel()

	var aus bytes.Buffer

	anzeige := fortschritt{ziel: &aus, modul: "m.test"}

	strom := `{"Action":"run","Package":"m.test/internal/app","Test":"TestA"}` + "\n" +
		`{"Action":"pass","Package":"m.test/internal/app","Test":"TestA/unter"}` + "\n" +
		`{"Action":"pass","Package":"m.test/internal/app","Test":"TestA","Elapsed":1.25}` + "\n" +
		`{"Action":"fail","Package":"m.test/internal/app","Test":"TestB","Elapsed":2}` + "\n" +
		`{"Action":"skip","Package":"m.test/internal/app","Test":"TestC"}` + "\n" +
		`{"Action":"fail","Package":"m.test/internal/app"}` + "\n" +
		"keine Zeile von go test\n" +
		`{"Action":"pass","Package":"anderes.test/x","Test":"TestD"}`

	// **In Stücken, die nicht an Zeilen enden.**
	for anfang := 0; anfang < len(strom); anfang += 7 {
		ende := min(anfang+7, len(strom))
		if geschrieben, fehler := anzeige.Write([]byte(strom[anfang:ende])); fehler != nil ||
			geschrieben != ende-anfang {
			t.Fatalf("Write gab %d, %v", geschrieben, fehler)
		}
	}

	anzeige.Schliessen()

	erwartet := []string{
		"  ✓ internal/app.TestA  1,2 s",
		"  ✗ internal/app.TestB  2,0 s",
		"  – internal/app.TestC  übersprungen",
		"  ✓ anderes.test/x.TestD  0,0 s",
	}

	if zeilen := strings.Split(strings.TrimRight(aus.String(), "\n"), "\n"); strings.Join(zeilen, "\n") !=
		strings.Join(erwartet, "\n") {
		t.Errorf("Fortschritt:\n%s\nerwartet:\n%s", aus.String(), strings.Join(erwartet, "\n"))
	}
}
