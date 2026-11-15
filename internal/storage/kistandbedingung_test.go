package storage_test

// **Der Stand wird in der Schreibtransaktion geprüft** — Befund 2 des
// Zweitblicks über Phase 5.3.
//
// # Die Zusage, um die es hier geht
//
// `service` liest die Revision der Karte, vergleicht sie mit der des Auftrags
// und ruft **danach** den Schreibweg. Zwischen beidem liegt alles: zwei
// Übernahmen desselben Feldes kamen beide durch die Prüfung, und die zweite
// übermalte die erste samt ihrer Herkunft, ohne dass es jemand sieht.
//
// **Prüfen lässt sich das nur hier.** `service` bekommt einen Fehler und
// sieht nicht, wann er entstand; ob die Bedingung mit der Wirkung zusammen
// stattfindet, sagt allein die Datenbank — und der Weg, es zu messen, ist
// eine Karte, die sich **zwischen** dem Lesen des Standes und dem Schreiben
// ändert.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Karte, die sich inzwischen geändert hat, wird nicht überschrieben.**
func TestEineInzwischenGeaenderteKarteWirdNichtUeberschrieben(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)

	// **Der Stand, den ein Auftrag eingefroren hätte.**
	stand := standVon(t, db, karteID)

	// **Und dazwischen schreibt jemand anderes** — genau der zweite
	// Schreibweg, den `pruefeStand` nicht mehr sieht.
	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t),
		karteID, []domain.Value{{Field: "name", Text: "Von Hand"}},
		"", []domain.KIFeldherkunft{eineFeldherkunft("name")}, nil, "",
	); fehler != nil {
		t.Fatalf("dazwischenschreiben: %v", fehler)
	}

	_, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t),
		karteID, []domain.Value{{Field: "name", Text: "Aus dem Vorschlag"}},
		"ki", []domain.KIFeldherkunft{eineFeldherkunft("name")}, nil, stand)

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Übernahme kam durch: %v", fehler)
	}

	// **Und nichts steht da.** Ein abgewiesener Schreibvorgang, der die
	// Hälfte hinterließe, wäre schlimmer als einer, der durchgeht.
	if wert := feldwertVon(t, db, karteID, "name"); wert != "Von Hand" {
		t.Errorf("im Feld steht %q", wert)
	}
}

// **Und auf dem erwarteten Stand geht sie durch.**
//
// **Die Gegenprobe** (§10): eine Bedingung, die immer abwiese, hielte den Test
// darüber grün und machte jede Übernahme unmöglich.
func TestAufDemErwartetenStandGehtDieUebernahmeDurch(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Flammenball"})

	schreiber := storage.NeuerKartenschreiber(db)
	stand := standVon(t, db, karteID)

	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t),
		karteID, []domain.Value{{Field: "name", Text: "Blitzschlag"}},
		"ki", []domain.KIFeldherkunft{eineFeldherkunft("name")}, nil, stand,
	); fehler != nil {
		t.Fatalf("auf dem eigenen Stand: %v", fehler)
	}

	if wert := feldwertVon(t, db, karteID, "name"); wert != "Blitzschlag" {
		t.Errorf("im Feld steht %q", wert)
	}
}

// **Dasselbe für die atomare Gruppe** — und für sie ist es schärfer: schon
// **eine** veränderte Karte muss die ganze Gruppe abweisen, sonst bleibt der
// Teilbestand zurück, den A13.1b ausschließt.
func TestEineVeraenderteKarteHaeltDieGanzeGruppeAn(t *testing.T) {
	db := frischeDatenbank(t)
	ersteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Erste"})
	zweiteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Zweite"})

	schreiber := storage.NeuerKartenschreiber(db)

	stand := standVon(t, db, ersteID)

	// **Nur die zweite hat sich geändert.**
	if _, _, fehler := schreiber.FelderSetzenMitHerkunft(probekontext(t),
		zweiteID, []domain.Value{{Field: "name", Text: "Von Hand"}},
		"", []domain.KIFeldherkunft{eineFeldherkunft("name")}, nil, "",
	); fehler != nil {
		t.Fatalf("dazwischenschreiben: %v", fehler)
	}

	fehler := schreiber.GruppeSetzenMitHerkunft(probekontext(t),
		[]domain.Kartenschritt{
			{
				KarteID: ersteID, Werte: []domain.Value{{Field: "name", Text: "Neu 1"}},
				Herkunft: "ki", Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
				Erwartet: stand,
			},
			{
				KarteID: zweiteID, Werte: []domain.Value{{Field: "name", Text: "Neu 2"}},
				Herkunft: "ki", Herkuenfte: []domain.KIFeldherkunft{eineFeldherkunft("name")},
				Erwartet: "ein-stand-von-gestern",
			},
		})

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Gruppe kam durch: %v", fehler)
	}

	// **Auch die unveränderte Karte bleibt, wie sie war** — das ist der
	// Unterschied zwischen „ganz oder gar nicht" und „so weit es ging".
	if wert := feldwertVon(t, db, ersteID, "name"); wert != "Erste" {
		t.Errorf("die erste Karte trägt %q", wert)
	}
}

// ------------------------------------------------------------------ Hilfsmittel

// standVon gibt den gerechneten Stand einer Karte.
//
// **Über den Leser und nicht über eine eigene Rechnung**: derselbe Weg, den
// die Naht geht, und ein zweiter liefe beim ersten neuen Feld auseinander.
func standVon(t *testing.T, db *sql.DB, karteID int64) string {
	t.Helper()

	zeile, fehler := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("Karte lesen: %v", fehler)
	}

	return zeile.Stand()
}

// feldwertVon gibt, was heute im Feld steht.
func feldwertVon(t *testing.T, db *sql.DB, karteID int64, feld string) string {
	t.Helper()

	zeile, fehler := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("Karte lesen: %v", fehler)
	}

	for _, wert := range zeile.Values {
		if wert.Field == feld {
			return wert.Text
		}
	}

	return ""
}
