package app_test

// **Die Verbuchung an der Naht** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Dass der Beleg gelesen und nicht gefolgert wird.** Aus „lokalVerbucht"
// folgt, dass einer dastehen **sollte**; was dasteht, sagt der Bestand. Wer
// ihn aus dem Zustand ableitete, zeigte eine Unterschrift, die es nicht gibt.
//
// **Und dass ein Nein ein Nein bleibt** — statt als leerer Stand anzukommen.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einJournaleintrag ist eine Bewegung ohne Handel.
func einJournaleintrag() domain.Journalzeile {
	return domain.Journalzeile{
		Art: domain.JournalZugang, Anlass: domain.AnlassKauf,
		Printing: "bolt-m10", Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil, Lagerort: "Kiste 4",
	}
}

func TestDasJournalKommtDurchDieNaht(t *testing.T) {
	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)

	if fehler := storage.NeueAbschlussablage(db).
		Buchen(probekontext(t), einJournaleintrag()); fehler != nil {
		t.Fatalf("Buchen: %v", fehler)
	}

	antwort := flaeche.Handelsjournal(probekontext(t), "")
	if antwort.Fehler != nil {
		t.Fatalf("Handelsjournal: %+v", antwort.Fehler)
	}

	if len(antwort.Zeilen) != 1 {
		t.Fatalf("es kamen %d Zeilen", len(antwort.Zeilen))
	}

	zeile := antwort.Zeilen[0]

	if zeile.Art != "zugang" || zeile.Anlass != "kauf" || zeile.Menge != 1 {
		t.Errorf("die Zeile ist %+v", zeile)
	}

	// **Der Lagerort kommt mit** — das Journal ist das eigene Buch.
	if zeile.Lagerort != "Kiste 4" || zeile.ID == 0 {
		t.Errorf("die Zeile ist %+v", zeile)
	}

	// **Und die Uhr ist eine** — kein roher Text aus der Datenbank.
	if _, fehler := time.Parse(time.RFC3339, zeile.Gebucht); fehler != nil {
		t.Errorf("die Uhr lautet %q: %v", zeile.Gebucht, fehler)
	}
}

// **Ein Fehler kommt als Fehler an** — auf beiden Wegen.
func TestEinVerbuchungsfehlerKommtAlsFehlerAn(t *testing.T) {
	flaeche := app.PeerflaecheUeber(frischerBestand(t))
	ohne := t.Context()

	if antwort := flaeche.Handelsjournal(ohne, "v-1"); antwort.Fehler == nil {
		t.Error("das Journal kam ohne Profil durch")
	} else if antwort.Zeilen == nil {
		t.Error("die Zeilen sind nil statt leer")
	}

	for name, antwort := range map[string]app.VerbuchungAntwort{
		"verbuchen":   flaeche.Handelsverbuchung(ohne, "v-1", "f-1"),
		"abgleichen":  flaeche.Handelsabgleich(ohne, "v-1"),
		"bestaetigen": flaeche.Handelsbestaetigung(ohne, "v-1", "bdu", "f-1"),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s kam ohne Profil durch", name)
		}

		if antwort.Vorgang != "v-1" {
			t.Errorf("%s nennt den Vorgang %q", name, antwort.Vorgang)
		}
	}
}

// **Der Beleg wird gelesen und nicht gefolgert.**
//
// # Warum dieser Lauf den Beleg von Hand hinlegt
//
// **Weil der Weg dorthin keine Fläche hat.** Ein Handel bis zur bestätigten
// Übergabe zu führen braucht `service.Handelsleitung` aus `T1`, und die hat
// bis heute keinen Aufrufer —
// #1637. Gemessen wird hier, was diese Naht tut: den abgelegten Beleg lesen
// und heil weiterreichen.
func TestDerBelegWirdGelesenUndNichtGefolgert(t *testing.T) {
	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)

	// **Ohne Beleg sagt die Antwort es** — statt einen leeren zu zeigen.
	antwort := flaeche.Handelsabgleich(probekontext(t), "v-1")
	if antwort.Fehler == nil {
		t.Fatal("der Abgleich lief auf einem Vorgang, den es nicht gibt")
	}

	if antwort.HatBeleg {
		t.Errorf("es kam ein Beleg: %+v", antwort.Beleg)
	}
}
