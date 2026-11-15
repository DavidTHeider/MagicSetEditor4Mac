package app

// **Der Zugstand an der Naht** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
//   - Der Zug **kommt durch die Naht** — Nummer, aktiver Sitz, Phase,
//     Schritt und Priorität.
//   - **Leere Listen und kein `nil`**: `gepasst` ist im Fenster sonst `null`,
//     und jede Schleife darüber fällt um.
//   - Leben und Vergiftungsmarken stehen **am Sitz** und für jeden sichtbar.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerZugstandKommtDurchDieNaht(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	zug := angelegt.Partie.Zug

	switch {
	case zug.Nummer != 1:
		t.Errorf("es ist Zug %d", zug.Nummer)
	case zug.Aktiv != "sitz-1":
		t.Errorf("aktiv ist %q", zug.Aktiv)
	case zug.Starter != "sitz-1":
		t.Errorf("angefangen hat %q", zug.Starter)
	case zug.Phase != string(domain.PhaseAnfang):
		t.Errorf("die Phase ist %q", zug.Phase)
	case zug.Schritt != string(domain.SchrittVersorgung):
		t.Errorf("der Schritt ist %q", zug.Schritt)
	case zug.Prioritaet != "sitz-1":
		t.Errorf("Priorität hat %q", zug.Prioritaet)
	case zug.Gepasst == nil:
		t.Error("gepasst ist nil statt leer")
	}

	// **Und jeder Sitz trägt sein Leben** — für jeden sichtbar.
	for _, sitz := range angelegt.Partie.Sitze {
		if sitz.Leben != 20 {
			t.Errorf("der Sitz %q hat %d Leben", sitz.Kennung, sitz.Leben)
		}

		if sitz.Vergiftung != 0 {
			t.Errorf("der Sitz %q ist vergiftet", sitz.Kennung)
		}
	}
}

// **Ein Passen geht über dieselbe Bindung wie jeder andere Befehl** (§16).
//
// **Das ist der Punkt**: `zug.passen` ist ein Spielbefehl und keine eigene
// Bindung — damit steht er an allen vier Flächen zugleich da, ohne dass ein
// Weg viermal gebaut wird.
func TestEinPassenGehtUeberDieBefehlsbindung(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	// **Gepasst wird nach dem Beginn** — vorher nimmt der Sandkasten nur ihn
	// an (CR 103.8, #1851).
	fassung := beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)

	gestellt := gebunden.PartieBefehl(t.Context(), angelegt.Partie.ID, "b1", "sitz-1",
		"zug.passen", nil, fassung)
	if gestellt.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", gestellt.Fehler)
	}

	danach := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1")
	if danach.Fehler != nil {
		t.Fatalf("Partie: %+v", danach.Fehler)
	}

	if danach.Partie.Zug.Prioritaet != "sitz-2" {
		t.Errorf("Priorität hat %q", danach.Partie.Zug.Prioritaet)
	}

	// **Und wer nicht an der Reihe ist, darf nichts stellen, was Priorität
	// braucht** — die Abnahme des Issues.
	for _, art := range danach.Erlaubt {
		if art == "objekt.tappen" {
			t.Error("dem Sitz ohne Priorität ist Tappen erlaubt")
		}
	}
}

// **Der Ausgang kommt durch die Naht** — Werkstück M17 der Regelmaschine
// (#1618).
//
// # Wogegen hier geprüft wird
//
//   - **Eine laufende Partie hat keinen Ausgang** — leerer Sieger, kein
//     Unentschieden, **leere Liste und kein `nil`**.
//   - **Nach dem Aufgeben steht er da**, mit Sieger und mit Grund: `beendet`
//     sagt nicht, warum (CR 104.3).
func TestDerAusgangKommtDurchDieNaht(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	switch {
	case angelegt.Partie.Ausgang.Sieger != "":
		t.Errorf("Sieger ist %q", angelegt.Partie.Ausgang.Sieger)
	case angelegt.Partie.Ausgang.Unentschieden:
		t.Error("die frische Partie gilt als unentschieden")
	case angelegt.Partie.Ausgang.Ausgeschieden == nil:
		t.Error("ausgeschieden ist nil statt leer")
	}

	gestellt := gebunden.PartieBefehl(t.Context(), angelegt.Partie.ID, "b1", "sitz-2",
		"zug.aufgeben", nil, angelegt.Partie.Fassung)
	if gestellt.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", gestellt.Fehler)
	}

	danach := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1")
	if danach.Fehler != nil {
		t.Fatalf("Partie: %+v", danach.Fehler)
	}

	switch {
	case danach.Partie.Ausgang.Sieger != "sitz-1":
		t.Errorf("Sieger ist %q", danach.Partie.Ausgang.Sieger)
	case danach.Partie.Stand != "beendet":
		t.Errorf("der Stand ist %q", danach.Partie.Stand)
	case len(danach.Partie.Ausgang.Ausgeschieden) != 1:
		t.Fatalf("ausgeschieden sind %+v", danach.Partie.Ausgang.Ausgeschieden)
	case danach.Partie.Ausgang.Ausgeschieden[0].Grund != "aufgabe":
		t.Errorf("der Grund ist %q",
			danach.Partie.Ausgang.Ausgeschieden[0].Grund)
	}
}
