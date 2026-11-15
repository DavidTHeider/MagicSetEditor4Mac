package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// fortschritt zeigt, während `go test` läuft, jeden Fall, der endet.
//
// **Zeile für Zeile**, und der Strom darf mitten in einer Zeile abreißen:
// was nach dem letzten Zeilenende kommt, wartet auf den Rest.
type fortschritt struct {
	ziel  io.Writer
	modul string
	rest  []byte
}

// Write nimmt ein Stück des Stroms.
func (f *fortschritt) Write(stueck []byte) (int, error) {
	f.rest = append(f.rest, stueck...)

	for {
		ende := bytes.IndexByte(f.rest, '\n')
		if ende < 0 {
			return len(stueck), nil
		}

		f.zeile(f.rest[:ende])
		f.rest = f.rest[ende+1:]
	}
}

// Schliessen zeigt, was nach dem letzten Zeilenende noch stand.
func (f *fortschritt) Schliessen() {
	f.zeile(f.rest)
	f.rest = nil
}

// zeile zeigt eine Zeile, wenn sie das Ende eines obersten Falls meldet.
func (f *fortschritt) zeile(text []byte) {
	var ereignis laufbeleg.Ereignis
	if json.Unmarshal(text, &ereignis) != nil || ereignis.Test == "" ||
		strings.Contains(ereignis.Test, "/") {
		return
	}

	name := strings.TrimPrefix(strings.TrimPrefix(ereignis.Package, f.modul), "/") + "." + ereignis.Test
	dauer := strings.Replace(fmt.Sprintf("%.1f s", ereignis.Elapsed), ".", ",", 1)

	// **Die Anzeige ist Beiwerk**, und ihr Fehler wird verworfen: sie hängt
	// im selben `io.MultiWriter` wie das Rohprotokoll, und ein Fehler hier
	// hielte `go test` an. Was der Lauf sah, tragen Protokoll und Beleg.
	switch ereignis.Action {
	case "pass":
		_, _ = fmt.Fprintf(f.ziel, "  ✓ %s  %s\n", name, dauer) //nolint:errcheck // siehe darüber
	case "fail":
		_, _ = fmt.Fprintf(f.ziel, "  ✗ %s  %s\n", name, dauer) //nolint:errcheck // siehe darüber
	case "skip":
		_, _ = fmt.Fprintf(f.ziel, "  – %s  übersprungen\n", name) //nolint:errcheck // siehe darüber
	}
}
