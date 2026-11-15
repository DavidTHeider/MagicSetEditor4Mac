package mcp_test

// **Was an der Deckübernahme schiefgehen kann** — Werkstück 12 von Phase 4.12
// (#1351).
//
// # Warum das Werkzeug einen eigenen Lauf braucht
//
// **Weil der Idempotenzlauf es nicht ruft.** Er fährt nur die Werkzeuge, die
// Idempotenz **zusagen**, und die Übernahme tut das nicht: ein zweiter Ruf
// legt ein zweites Deck an. Ohne diesen Lauf bliebe ihr Rumpf unbetreten —
// dieselbe Lücke, die bei den Token- und Vorgabenwerkzeugen schon einmal
// gefunden wurde (Commit `27f79352`).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// mitUebernahme gibt das Übernahmewerkzeug über einer Probe.
func mitUebernahme(t *testing.T, probe *uebernahmeprobe) mcp.Werkzeug {
	t.Helper()

	return werkzeugAus(t, mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(), Deckuebernahme: probe,
	}, "schreiben.deck.uebernehmen")
}

// **Der Erfolgsweg gibt das Deck samt dem, was nicht mitkam.**
func TestDieUebernahmeGibtDasDeckUndDieBefunde(t *testing.T) {
	t.Parallel()

	probe := neueUebernahmeprobe()

	aus, fehler := mitUebernahme(t, probe).Ruf(
		t.Context(), []byte(`{"quellProfil":"2","quellDeck":"7"}`))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.gesehen != [2]string{"2", "7"} {
		t.Errorf("gefragt wurde nach %v", probe.gesehen)
	}

	// **Beides in einer Antwort**: die Kennung, damit man das Deck wiederholt,
	// und die Befunde, damit man weiß, was fehlt.
	if !strings.Contains(aus, `"deckID"`) || !strings.Contains(aus, `"befunde"`) {
		t.Errorf("die Antwort lautet %s", aus)
	}
}

// **Ein krummer Rumpf wird abgewiesen.**
func TestEinKrummerRumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := mitUebernahme(t, neueUebernahmeprobe()).Ruf(
		t.Context(), []byte(`{`)); fehler == nil {
		t.Error("ein krummer Rumpf kam durch")
	}
}

// **Und ein Nein der Naht reicht durch.**
//
// **Ohne diesen Lauf wäre ein Werkzeug grün, das jeden Fehlschlag als Erfolg
// meldet** (§10) — und bei einer Übernahme hieße das: „geholt" für ein Deck,
// das nie ankam.
func TestEinNeinDerUebernahmeReichtDurch(t *testing.T) {
	t.Parallel()

	probe := neueUebernahmeprobe()
	probe.fehler = errors.New("dafür fehlt dir die Rolle")

	if _, fehler := mitUebernahme(t, probe).Ruf(
		t.Context(), []byte(`{"quellProfil":"2","quellDeck":"7"}`)); !errors.Is(
		fehler, probe.fehler) {
		t.Errorf("Fehler = %v, erwartet das Nein der Naht", fehler)
	}
}
