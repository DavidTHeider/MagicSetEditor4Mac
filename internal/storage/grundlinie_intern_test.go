package storage

// **Die Grundlinie trägt einen leeren Bestand** — J02
// (#2266).
//
// Am 2026-10-12 ergab sie **denselben** Bestand wie die 97 Wanderungen, die
// sie ablöst — Schema, Startzeilen und Folgezähler, verglichen in
// `TestDieGrundlinieGleichtDenWanderungen` (Commit `ebd92fa37`), bevor sie
// fortfielen (Vermerk 0189). Seither ist sie selbst die Quelle; was hier
// bleibt, sind der Weg für den leeren Bestand und die Vorrichtungen, mit denen
// die übrigen Läufe zwei Bestände vergleichen.

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// fluechtig sind die Spalten, deren Saat aus Zeit oder Zufall kommt.
var fluechtig = map[string]bool{
	"bestand.wert": true, "bestandskennung.wert": true,
	"nutzer.angelegt": true, "profile.angelegt": true,
}

func leererBestand(t *testing.T) *sql.DB {
	t.Helper()

	db, fehler := Oeffnen(t.Context(), ImArbeitsspeicher)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	return db
}

func TestNurEinLeererBestandNimmtDieGrundlinie(t *testing.T) {
	db := leererBestand(t)

	tx, fehler := db.BeginTx(t.Context(), nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // Aufräumen

	if leer, err := bestandIstLeer(t.Context(), tx); err != nil || !leer {
		t.Fatalf("ein frischer Bestand gilt nicht als leer: %v, %v", leer, err)
	}

	if _, err := tx.ExecContext(t.Context(), "CREATE TABLE etwas (x)"); err != nil {
		t.Fatal(err)
	}

	if leer, err := bestandIstLeer(t.Context(), tx); err != nil || leer {
		t.Errorf("ein Bestand mit einer Tabelle gilt als leer: %v, %v", leer, err)
	}
}

func TestDieGrundlinieVerbuchtJedeEnthalteneWanderung(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	verbucht, fehler := AngewendeteMigrationen(t.Context(), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	namen, _, _ := grundlinienumfang(grundlinie) //nolint:errcheck // eingebettet

	// **Und dahinter die Wanderungen danach** (Vermerk 0189, Satz 6; seit M01,
	// #2345, gibt es eine).
	if erwartet := append(slices.Clone(namen), danach(t)...); !slices.Equal(verbucht, erwartet) {
		t.Errorf("verbucht sind %d von %d Wanderungen", len(verbucht), len(erwartet))
	}

	if !slices.Contains(namen, grundlinieBis) {
		t.Errorf("die Grundlinie nennt %s, das es nicht gibt", grundlinieBis)
	}
}

// bestandsabdruck beschreibt Schema, Zeilen und Folgezähler eines Bestands —
// ohne die Buchführung und mit flüchtigen Spalten nur in ihrer Gestalt.
func bestandsabdruck(t *testing.T, db *sql.DB) string {
	t.Helper()

	schema := abfrage(t, db, "SELECT type, name, tbl_name, coalesce(sql, '') FROM sqlite_master WHERE name <> 'migrationen' ORDER BY type, name")
	aus := make([]string, 0, len(schema))

	for _, zeile := range schema {
		aus = append(aus, "schema "+strings.Join(zeile, " · "))
	}

	for _, tabelle := range abfrage(t, db, "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT IN ('migrationen', 'sqlite_sequence') ORDER BY name") {
		spalten := abfrage(t, db, fmt.Sprintf("SELECT name FROM pragma_table_info('%s')", tabelle[0]))
		zeilen := abfrage(t, db, fmt.Sprintf(`SELECT * FROM %q`, tabelle[0]))

		var gesehen []string

		for _, werte := range zeilen {
			for i, spalte := range spalten {
				if i < len(werte) && fluechtig[tabelle[0]+"."+spalte[0]] && werte[i] != "" {
					werte[i] = "<gesetzt>"
				}
			}

			gesehen = append(gesehen, tabelle[0]+" "+strings.Join(werte, " · "))
		}

		slices.Sort(gesehen)
		aus = append(aus, gesehen...)
	}

	for _, zeile := range abfrage(t, db, "SELECT name, seq FROM sqlite_sequence WHERE name <> 'migrationen' ORDER BY name") {
		aus = append(aus, "folge "+strings.Join(zeile, " · "))
	}

	return strings.Join(aus, "\n")
}

func abfrage(t *testing.T, db *sql.DB, text string) [][]string {
	t.Helper()

	zeilen, fehler := db.QueryContext(context.Background(), text)
	if fehler != nil {
		t.Fatalf("%s: %v", text, fehler)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen

	spalten, _ := zeilen.Columns() //nolint:errcheck // nach einer geglückten Abfrage

	var aus [][]string

	for zeilen.Next() {
		werte := make([]any, len(spalten))
		ziele := make([]any, len(spalten))

		for i := range werte {
			ziele[i] = &werte[i]
		}

		if fehler := zeilen.Scan(ziele...); fehler != nil {
			t.Fatal(fehler)
		}

		text := make([]string, len(werte))
		for i, wert := range werte {
			if b, ist := wert.([]byte); ist {
				wert = fmt.Sprintf("%x", b)
			}

			text[i] = fmt.Sprint(wert)
		}

		aus = append(aus, text)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("%s: %v", text, fehler)
	}

	return aus
}

func abdruckunterschied(a, b string) string {
	links, rechts := strings.Split(a, "\n"), strings.Split(b, "\n")

	var aus []string

	for _, zeile := range links {
		if !slices.Contains(rechts, zeile) {
			aus = append(aus, "− "+zeile)
		}
	}

	for _, zeile := range rechts {
		if !slices.Contains(links, zeile) {
			aus = append(aus, "+ "+zeile)
		}
	}

	if len(aus) > 20 {
		aus = append(aus[:20], fmt.Sprintf("… und %d weitere", len(aus)-20))
	}

	return strings.Join(aus, "\n")
}
