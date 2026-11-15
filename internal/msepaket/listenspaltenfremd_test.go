//go:build fremdpakete

package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was `magic.mse-game` über seine Listenspalten sagt** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// # Warum gegen das echte Paket
//
// Weil die Zahlen die Entscheidung tragen. Ohne sie wäre „die meisten Spalten
// sind nicht von Anfang an zu sehen" eine Behauptung; mit ihnen ist es eine
// Messung — und die Vorgabe `card list visible: false` ist die Antwort darauf.
//
// **Über den Leser dieses Projekts und nicht über `grep`.** Ein `grep` zählt
// Zeilen einer Datei; dieser Lauf zählt, was **ankommt** — samt eingebundenen
// Dateien und samt der Deutung von `0400`. Die Reichweite einer Messung ist
// die Reichweite ihrer Aussage (§10).
func TestDieListenspaltenVonMagic(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if fehler != nil {
		t.Fatalf("magic.mse-game öffnen: %v", fehler)
	}

	// **Die Module werden eingebunden**, bevor gelesen wird — dieselbe Lage
	// und dieselbe Begründung wie in `feldreihenfremd_test.go`.
	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // siehe dort

	spiel, _ := paket.Spielmodell()

	var (
		mitStelle, sichtbar, verboten, breit int
		ausrichtungen                        = map[domain.Listenausrichtung]int{}
		hoechste                             int
	)

	for _, feld := range spiel.Fields {
		if feld.Listenstelle > 0 {
			mitStelle++
		}

		if feld.Listenstelle > hoechste {
			hoechste = feld.Listenstelle
		}

		if feld.Listensichtbar {
			sichtbar++
		}

		if !feld.Listenerlaubt {
			verboten++
		}

		if feld.Listenbreite > 0 {
			breit++
		}

		ausrichtungen[feld.Listenausrichtung]++
	}

	// **Gemessen am 2026-08-27** über `Full-Magic-Pack`, Fassung des Packs auf
	// diesem Rechner. Ändert sich das Paket, ändert sich diese Zahl — und der
	// Lauf sagt es, statt dass eine Behauptung im Kommentar veraltet.
	fuer := func(name string, gezaehlt, erwartet int) {
		t.Helper()

		if gezaehlt != erwartet {
			t.Errorf("%s: %d, erwartet %d", name, gezaehlt, erwartet)
		}
	}

	// **243 Felder mit einer Stelle, 11 davon sichtbar.** Das ist die Zahl,
	// die die Vorgabe trägt: `card list visible:` ohne Angabe ist **falsch**,
	// sonst stünden 243 Spalten nebeneinander.
	fuer("Felder mit einer Stelle", mitStelle, 243)
	fuer("davon sichtbar", sichtbar, 11)
	fuer("verbotene Felder", verboten, 6)
	fuer("Felder mit einer Breite", breit, 243)

	// **Die höchste Stelle ist 9999** — MSEs Zählung ist eine Reihenfolge und
	// keine Spaltennummer; wer sie als Index läse, legte ein Feld mit
	// zehntausend Einträgen an.
	fuer("höchste Stelle", hoechste, 9999)

	// **Und die Ausrichtung ist fast immer links.**
	if ausrichtungen[domain.ListeRechts] == 0 {
		t.Error("kein einziges Feld richtet rechts aus — die Angabe kommt nicht an")
	}

	t.Logf("Stellen %d · sichtbar %d · verboten %d · Breiten %d · höchste %d · %v",
		mitStelle, sichtbar, verboten, breit, hoechste, ausrichtungen)
}
