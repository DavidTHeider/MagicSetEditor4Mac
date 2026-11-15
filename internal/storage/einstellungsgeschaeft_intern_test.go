package storage

// **Der Vergleich liegt im Schreibgeschäft** — Hinweis 12 des Nachlaufs zum
// Zweitblick über Phase 5.2 (#1435).
//
// Der Lauf in `einstellungsaustausch_test.go` belegt, **dass** verglichen wird:
// eine überholte Revision schreibt nicht. Er belegt nicht, **wo** — der
// veraltete Stand steht dort schon vor dem Aufruf, und ein Vergleich in einem
// eigenen, vorher abgeschlossenen Lesegeschäft wäre ebenso grün gewesen.
//
// **Hier wird die Reihenfolge selbst gelesen.**

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Kein `t.Parallel()`**: der Aufzeichner ist eine Paketvariable.
func TestDerRevisionsvergleichLiegtImSchreibgeschaeft(t *testing.T) {
	db, fehler := sql.Open("aufzeichner", ":memory:")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	mitschrift()

	// **Die Revision muss stimmen** — sonst endete der Weg beim Vergleich, und
	// die Mitschrift sagte nichts über das Schreiben dahinter. Der
	// Aufzeichner antwortet mit leeren Tafeln und dem Menschen 1.
	heute := domain.Standmarke(1, 1, map[string]string{}, map[string]string{})

	// **Der Ausgang ist gleichgültig** — geprüft wird der Weg dorthin.
	//nolint:errcheck // der Ausgang ist gleichgültig, die Mitschrift zählt
	_ = NeueEinstellungen(db).EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "en"}, nil, heute)

	geschrieben := mitschrift()

	beginn := stelleVon(geschrieben, "BEGIN")
	abschluss := stelleVon(geschrieben, "COMMIT")
	vergleich := stelleMit(geschrieben, "FROM settings")

	if beginn < 0 || vergleich < 0 {
		t.Fatalf("die Mitschrift ist unvollständig:\n%s",
			strings.Join(geschrieben, "\n"))
	}

	if vergleich < beginn {
		t.Errorf("die Revision wurde vor dem Geschäft gelesen:\n%s",
			strings.Join(geschrieben, "\n"))
	}

	if abschluss >= 0 && vergleich > abschluss {
		t.Errorf("die Revision wurde nach dem Abschluss gelesen:\n%s",
			strings.Join(geschrieben, "\n"))
	}

	// **Ein Geschäft und nicht zwei**: ein zweites `BEGIN` hieße, dass Lesen
	// und Schreiben in verschiedenen Ständen liegen.
	if zahl := zaehleGleiche(geschrieben, "BEGIN"); zahl != 1 {
		t.Errorf("%d Geschäfte statt einem:\n%s", zahl,
			strings.Join(geschrieben, "\n"))
	}
}

// stelleVon gibt die Stelle der ersten wortgleichen Zeile — oder -1.
func stelleVon(mitschrift []string, wort string) int {
	for stelle, zeile := range mitschrift {
		if zeile == wort {
			return stelle
		}
	}

	return -1
}

// stelleMit gibt die Stelle der ersten Zeile, die den Ausschnitt enthält.
func stelleMit(mitschrift []string, teil string) int {
	for stelle, zeile := range mitschrift {
		if strings.Contains(zeile, teil) {
			return stelle
		}
	}

	return -1
}

// zaehleGleiche zählt die wortgleichen Zeilen.
func zaehleGleiche(mitschrift []string, wort string) int {
	var zahl int

	for _, zeile := range mitschrift {
		if zeile == wort {
			zahl++
		}
	}

	return zahl
}
