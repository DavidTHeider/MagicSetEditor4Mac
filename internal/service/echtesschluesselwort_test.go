//go:build fremdpakete

package service //nolint:testpackage // NeueWortliste und Einsetzen sind paketintern

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestEinEchtesSchluesselwortTrifftEchtenKartentext ist die Probe hinter dem
// Test — die Zusage, die kein Vorrichtungstest geben kann.
//
// **Sie läuft nicht in der Pipeline** (§3): sie braucht `M15-Magic-Pack` auf
// dem Rechner. Was sie belegt, belegt kein anderer Lauf: dass die
// Trennerbehandlung an den **echten** 15 Parameterarten und den echten
// Schlüsselwörtern greift, nicht nur an den drei, die ein Test hinschreibt.
//
//	go test -tags fremdpakete ./internal/service/ -run TestEinEchtesSchluesselwort
func TestEinEchtesSchluesselwortTrifftEchtenKartentext(t *testing.T) {
	t.Parallel()

	zuhause, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("kein Zuhause: %v", err)
	}

	wurzel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", "M15-Magic-Pack", "data")

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, "magic.mse-game"))
	if err != nil {
		t.Skipf("das Spielpaket unter %s: %v", wurzel, err)
	}

	// **Ohne den Einbinder stehen im Baum nur die `include`-Zeilen.** Die
	// Schlüsselwörter des echten Packs liegen in `keywords`, geholt über
	// `include localized file: keywords` — wer sie im `game` sucht, findet
	// null und hält das für einen Befund über das Programm.
	if befunde := paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)); len(befunde) != 0 {
		t.Logf("%d Befunde beim Einbinden", len(befunde))
	}

	liste, abgeschaltet := NeueWortliste(
		paket.Schluesselwoerter(), paket.Schluesselwortparameterarten())

	t.Logf("%d Schlüsselwörter bereit, %d abgeschaltet", liste.Anzahl(), len(abgeschaltet))

	// **62 von 333 Schlüsselwörtern des echten Packs** tragen einen Trenner,
	// der gefressen werden muss (gezählt am 2026-08-24 über `keywords`).
	// Bei allen 62 blieb der Erinnerungstext aus.
	//
	// **Die Texte stehen so auf echten Karten.** Jeder trägt genau **einen**
	// Trenner zwischen Wort und Parameter; ohne das Fressen verlangte der
	// Ausdruck zwei, und der Erinnerungstext bliebe aus.
	for _, text := range []string{
		"Cumulative upkeep 2",
		"Buyback 3",
		"Echo 1W",
	} {
		gesehen := false

		liste.Einsetzen(text, func(treffer Schluesselworttreffer) string {
			gesehen = true

			return treffer.Text
		})

		if !gesehen {
			t.Errorf("%q trifft kein Schlüsselwort", text)
		}
	}
}
