package mcp_test

// **Nichts in diesem Paket macht selbst einen Port auf** — Werkstück A14 von
// Phase 5.3
// (#1453),
// nach [Vermerk 0106](../../docs/entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md).
//
// # Warum das ein Lauf ist und kein Satz
//
// **Weil der Satz seit 0010 dasteht und niemand ihn nachgezählt hat.** Er
// lautet: „Über stdio gibt es keinen Port, keine Adresse und nichts, was auf
// etwas wartet." Eine Zusage ohne Wächter ist die Gestalt aus §10 — und
// gerade diese ließe sich mit **einer** Zeile brechen.
//
// # Und warum an der Quelle und nicht am Verhalten
//
// **Weil ein Lauscher, den niemand startet, kein Verhalten zeigt.** Ein Test,
// der prüft, dass kein Port offen ist, beweist etwas über **diesen** Lauf;
// eine Zeile `http.ListenAndServe` hinter einem Schalter käme durch. Gelesen
// wird deshalb der Quelltext.
//
// # Was sich am 2026-09-15 geändert hat
//
// **`cmd/mse-mcp` darf einen Port aufmachen** — Werkstück N6 von Phase 5.4
// (#1572),
// nach [0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md).
// Der Satz von 0010 — „über stdio gibt es keinen Port" — gilt weiter **für
// stdio**; daneben steht seit N6 ein zweiter Transport.
//
// **Die Zusage wird enger statt aufgehoben**: er darf keinen Port
// **selbst** aufmachen. Jeder Lauscher geht durch `rest.LauschenMit`, und
// dort steht an **einer** Stelle, dass jenseits der Schleife TLS Pflicht ist.
// Ein `net.Listen` daneben wäre genau der zweite Weg, den niemand geprüft
// hat.
//
// **`internal/mcp` bleibt vollständig verboten.** Es baut einen
// `http.Handler` und bindet nichts — ein Blatt, das lauscht, wäre ein Blatt,
// das einen Lebenszyklus hat.
//
// # Was er **nicht** hält
//
// **Nicht, dass der gereichte Lauscher wirklich von `rest.LauschenMit`
// kommt.** Er sieht den Aufruf und nicht seinen Ursprung; ein Lauscher aus
// einer dritten Quelle käme durch. Dagegen steht die Schichtenregel: außer
// `rest` gibt es in `cmd/mse-mcp` nichts, was einen liefert.
//
// **Er sieht `internal/mcp` und `cmd/mse-mcp`, nicht das ganze Programm.**
// `mse-rest` lauscht, und wer dort etwas ändert, findet seinen Wächter in
// `internal/rest`.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lauscherzeichen sind die Gestalten, mit denen ein Prozess auf Verbindungen
// wartet.
//
// **Benannt und nicht geraten**: was fehlt, fällt beim nächsten Umbau auf —
// eine Regex über „listen" fände auch einen Kommentar.
var lauscherzeichen = []string{
	// **Mit Klammer und nicht ohne.** Ohne sie trifft der Vergleich auch
	// `net.Listener` — den **Typ**, den jede Funktion nennt, die einen
	// Lauscher entgegennimmt. Gemessen am 2026-09-15: der erste Anlauf von
	// N6 wurde genau daran rot, und der Befund war keiner.
	"net.Listen(",
	"http.ListenAndServe",
	"httptest.NewServer",
}

// selbstgebaut sind zusätzlich die Gestalten, die `internal/mcp` nicht haben
// darf.
//
// **Ein `http.Server` gehört einem Composition Root.** In einem Blatt wäre er
// ein Lebenszyklus, und den hat ein Blatt nicht.
var selbstgebaut = []string{"http.Server{"}

// **Weder `internal/mcp` noch `cmd/mse-mcp` wartet auf eine Verbindung.**
func TestNichtsAmMCPLauscht(t *testing.T) {
	t.Parallel()

	for verzeichnis, streng := range map[string]bool{
		".": true, "../../cmd/mse-mcp": false,
	} {
		t.Run(verzeichnis, func(t *testing.T) {
			t.Parallel()

			zeichen := lauscherzeichen
			if streng {
				zeichen = append(append([]string{}, lauscherzeichen...),
					selbstgebaut...)
			}

			pruefeOhneLauscher(t, verzeichnis, zeichen)
		})
	}
}

// pruefeOhneLauscher liest jede Quelldatei eines Verzeichnisses.
//
// **Ohne Tests**: dieser Lauf steht selbst in einem, und `httptest` gehört in
// einen Lauf über einen fremden Dienst — nur nicht in den Server.
func pruefeOhneLauscher(t *testing.T, verzeichnis string, zeichen []string) {
	t.Helper()

	eintraege, fehler := os.ReadDir(verzeichnis)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", verzeichnis, fehler)
	}

	gelesen := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		gelesen++

		pruefeDateiOhneLauscher(t, filepath.Join(verzeichnis, name), zeichen)
	}

	// **Und dass überhaupt etwas gelesen wurde** (§10): ein Wächter über null
	// Dateien ist grün und sagt nichts.
	if gelesen == 0 {
		t.Fatalf("in %s wurde keine Quelldatei gelesen", verzeichnis)
	}
}

// pruefeDateiOhneLauscher sieht eine Datei an.
func pruefeDateiOhneLauscher(t *testing.T, pfad string, zeichen []string) {
	t.Helper()

	roh, fehler := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus ReadDir im Repositorium
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	inhalt := string(roh)

	for _, gesucht := range zeichen {
		if !strings.Contains(inhalt, gesucht) {
			continue
		}

		t.Errorf("%s enthält %q — damit macht dieser Prozess selbst einen "+
			"Port auf. Seit N6 (#1572) darf `cmd/mse-mcp` lauschen, aber "+
			"**nur durch `rest.LauschenMit`**: dort steht an einer Stelle, "+
			"dass jenseits der Schleife TLS Pflicht ist "+
			"([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)). "+
			"Ein zweiter Weg wäre einer, den niemand geprüft hat", pfad, gesucht)
	}
}
