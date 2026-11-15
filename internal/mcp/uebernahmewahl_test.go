package mcp_test

// **Die Decks eines anderen Profils zur Wahl, über MCP** — UX-08.AK-04
// (#1680).
//
// **Die lesende Hälfte der Übernahme** (§16): wer über ein Modell ein Deck
// holen will, braucht dessen Kennung — und bekam sie bis hierher nirgends.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// mitDeckwahl gibt das Werkzeug über einer Probe.
func mitDeckwahl(t *testing.T, probe *uebernahmeprobe) mcp.Werkzeug {
	t.Helper()

	return werkzeugAus(t, mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(), Deckuebernahme: probe,
	}, "lesen.decks.uebernehmbar")
}

func TestDieDeckwahlNenntDieDecksVonDrueben(t *testing.T) {
	t.Parallel()

	probe := neueUebernahmeprobe()

	aus, fehler := mitDeckwahl(t, probe).Ruf(
		t.Context(), []byte(`{"quellProfil":"2"}`))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.gesehen[0] != "2" || !strings.Contains(aus, `"Nachtlauf"`) {
		t.Errorf("gefragt %v, geantwortet %s", probe.gesehen, aus)
	}

	if mitDeckwahl(t, probe).Art != mcp.Lesend {
		t.Error("die Wahl ändert etwas")
	}
}

func TestDieDeckwahlWeistEinenKrummenRumpfAb(t *testing.T) {
	t.Parallel()

	if _, fehler := mitDeckwahl(t, neueUebernahmeprobe()).Ruf(
		t.Context(), []byte(`{`)); fehler == nil {
		t.Error("ein krummer Rumpf kam durch")
	}
}

func TestEinNeinDerDeckwahlReichtDurch(t *testing.T) {
	t.Parallel()

	probe := neueUebernahmeprobe()
	probe.fehler = errors.New("dafür fehlt dir die Rolle")

	if _, fehler := mitDeckwahl(t, probe).Ruf(
		t.Context(), []byte(`{"quellProfil":"2"}`)); !errors.Is(
		fehler, probe.fehler) {
		t.Errorf("es kam %v", fehler)
	}
}
