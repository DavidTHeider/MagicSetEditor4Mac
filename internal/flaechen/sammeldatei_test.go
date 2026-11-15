// Eine Sammeldatei, die ein Index lesen kann — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Selbst geschrieben und nicht abgeschrieben
//
// **§11**: fremde Daten kommen nicht ins Repositorium, und eine Zeile aus
// einem echten Scryfall-Bulk wäre eine. Was hier steht, hat dieselbe **Form**
// — JSONL, gzip, die Felder, die `internal/scryfall` liest — und eigenen
// Inhalt.
//
// # Wozu es sie gibt
//
// **Bis A14 gab es keinen wirklich gebauten Index in der Pipeline.** Der Lauf
// in `suchindex_test.go` sagt es in seinem Kopf: „Dafür braucht es eine
// laufende Einbettungslaufzeit und eine Sammeldatei; beides steht hinter
// `netzlauf`." Die Laufzeit stellt `kilaufzeitstoff_test.go` nach, die Datei
// steht hier — und damit werden der **Fortsetzungszeiger** und der
// **laufende Aufbau** prüfbar, die das Issue verlangt.
package flaechen_test

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// diePrüfkarten sind drei selbst verfasste Karten in Scryfalls Form.
//
// **Drei und nicht eine**: eine Trefferliste mit einem Eintrag hat keine
// zweite Seite, und der Fortsetzungszeiger wäre nicht zu stellen.
var diePruefkarten = []string{
	`{"object":"card","id":"00000000-0000-4000-8000-000000000101",` +
		`"oracle_id":"00000000-0000-4000-8000-000000000201",` +
		`"name":"Funkenwurf","lang":"de","layout":"normal",` +
		`"mana_cost":"{R}","cmc":1,"type_line":"Spontanzauber",` +
		`"oracle_text":"Funkenwurf fügt einer Kreatur 2 Schaden zu.",` +
		`"colors":["R"],"color_identity":["R"],"rarity":"common",` +
		`"set":"prf","set_name":"Prüfsatz","collector_number":"1",` +
		`"released_at":"2026-01-01","legalities":{},` +
		`"scryfall_uri":"https://beispiel.invalid/karte/101"}`,
	`{"object":"card","id":"00000000-0000-4000-8000-000000000102",` +
		`"oracle_id":"00000000-0000-4000-8000-000000000202",` +
		`"name":"Steinwall","lang":"de","layout":"normal",` +
		`"mana_cost":"{1}{W}","cmc":2,"type_line":"Kreatur — Mauer",` +
		`"oracle_text":"Verteidiger. Steinwall blockt jede Runde.",` +
		`"colors":["W"],"color_identity":["W"],"rarity":"common",` +
		`"set":"prf","set_name":"Prüfsatz","collector_number":"2",` +
		`"released_at":"2026-01-01","legalities":{},` +
		`"scryfall_uri":"https://beispiel.invalid/karte/102"}`,
	`{"object":"card","id":"00000000-0000-4000-8000-000000000103",` +
		`"oracle_id":"00000000-0000-4000-8000-000000000203",` +
		`"name":"Blockhilfe","lang":"de","layout":"normal",` +
		`"mana_cost":"{G}","cmc":1,"type_line":"Spontanzauber",` +
		`"oracle_text":"Eine blockende Kreatur erhält +2/+2 bis zum Ende des Zuges.",` +
		`"colors":["G"],"color_identity":["G"],"rarity":"common",` +
		`"set":"prf","set_name":"Prüfsatz","collector_number":"3",` +
		`"released_at":"2026-01-01","legalities":{},` +
		`"scryfall_uri":"https://beispiel.invalid/karte/103"}`,
}

// eineSammeldatei schreibt die drei gepackt und gibt den Pfad.
func eineSammeldatei(t *testing.T) string {
	t.Helper()

	return sammeldateiAus(t, diePruefkarten)
}

// eineGrosseSammeldatei gibt viele Karten — für Läufe, die Zeit brauchen.
//
// **Nicht, weil mehr Karten mehr prüfen**, sondern weil ein Aufbau lang genug
// dauern muss, dass ein zweiter ihn **antrifft**. Mit drei Karten ist er nach
// 150 ms vorbei, und der zweite liefe danach — der Lauf prüfte dann eine
// Gleichzeitigkeit, die es nicht gab (§10).
func eineGrosseSammeldatei(t *testing.T, anzahl int) string {
	t.Helper()

	zeilen := make([]string, 0, anzahl)
	for nummer := range anzahl {
		zeilen = append(zeilen, eineWeitereKarte(nummer))
	}

	return sammeldateiAus(t, zeilen)
}

// eineWeitereKarte baut eine Zeile mit eigener Kennung.
func eineWeitereKarte(nummer int) string {
	kennung := fmt.Sprintf("%03d", nummer+10)

	return `{"object":"card","id":"00000000-0000-4000-8000-000000000` + kennung +
		`","oracle_id":"00000000-0000-4000-8001-000000000` + kennung +
		`","name":"Probekarte ` + strconv.Itoa(nummer) + `","lang":"de",` +
		`"layout":"normal","mana_cost":"{1}","cmc":1,"type_line":"Spontanzauber",` +
		`"oracle_text":"Probekarte ` + strconv.Itoa(nummer) +
		` tut beim Blocken nichts Besonderes.",` +
		`"colors":[],"color_identity":[],"rarity":"common",` +
		`"set":"prf","set_name":"Prüfsatz","collector_number":"` +
		strconv.Itoa(nummer+10) + `","released_at":"2026-01-01","legalities":{},` +
		`"scryfall_uri":"https://beispiel.invalid/karte/` + kennung + `"}`
}

// sammeldateiAus schreibt Zeilen gepackt und gibt den Pfad.
func sammeldateiAus(t *testing.T, zeilen []string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "pruefkarten.json.gz")

	datei, fehler := os.Create(pfad) //nolint:gosec // ein Pfad aus t.TempDir()
	if fehler != nil {
		t.Fatalf("die Sammeldatei anlegen: %v", fehler)
	}

	packer := gzip.NewWriter(datei)

	if _, fehler := packer.Write(
		[]byte(strings.Join(zeilen, "\n") + "\n")); fehler != nil {
		t.Fatalf("die Sammeldatei schreiben: %v", fehler)
	}

	for _, schliessen := range []func() error{packer.Close, datei.Close} {
		if fehler := schliessen(); fehler != nil {
			t.Fatalf("die Sammeldatei schließen: %v", fehler)
		}
	}

	return pfad
}
