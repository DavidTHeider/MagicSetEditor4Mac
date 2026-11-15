package service

// **Eine geänderte Vorlage gilt, ohne Neustart** — FMT-02, S10.NAT-03
// (#1691).
//
// Gemessen am nativen Fenster (Hüllenprobe `fmt02f`): die Werkstatt sicherte
// einen roten Grund, und die Karte im Set blieb weiß — der Stilspeicher hielt
// die Vorlage seit dem ersten Zeichnen fest. Über seinem Gedächtnis stand, die
// Datei ändere sich nicht, solange die Anwendung läuft; seit der Werkstatt
// stimmt das nicht mehr, und für einen Texteditor daneben hat es nie
// gestimmt.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// stempelkoepfe führt ein Set auf der Kopie von `probe-schlicht`.
type stempelkoepfe struct{}

func (stempelkoepfe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}}, nil
}

func (stempelkoepfe) Setwerte(context.Context, int64) ([]domain.Value, error) { return nil, nil }

// stempelplatz legt Spiel und Stil der Saat in eine eigene Wurzel.
func stempelplatz(t *testing.T) (wurzel string, speicher *Stilspeicher) {
	t.Helper()

	vorher := stempelpause
	stempelpause = 0

	t.Cleanup(func() { stempelpause = vorher })

	wurzel = t.TempDir()
	for _, name := range []string{"probe.mse-game", "probe-schlicht.mse-style"} {
		if err := os.CopyFS(filepath.Join(wurzel, name), os.DirFS(filepath.Join("..", "..", "saat", name))); err != nil {
			t.Fatal(err)
		}
	}

	return wurzel, NeuerStilspeicher(stempelkoepfe{}, wurzel)
}

// zeileErsetzen ersetzt eine Zeile einer Paketdatei.
func zeileErsetzen(t *testing.T, datei, alt, neu string) {
	t.Helper()

	roh, err := os.ReadFile(datei) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(roh), alt) {
		t.Fatalf("%q steht nicht in %s", alt, datei)
	}

	if err := os.WriteFile(datei, []byte(strings.Replace(string(roh), alt, neu, 1)), 0o600); err != nil { //nolint:gosec // aus t.TempDir()
		t.Fatal(err)
	}
}

func TestEinGeaenderterStilGiltOhneNeustart(t *testing.T) {
	wurzel, speicher := stempelplatz(t)

	vorher, _, _ := speicher.StilFuerSet(t.Context(), 7)
	if vorher.Stil.Breite != 375 {
		t.Fatalf("Breite vorher %d", vorher.Stil.Breite)
	}

	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"), "card width: 375", "card width: 200")

	if nachher, _, _ := speicher.StilFuerSet(t.Context(), 7); nachher.Stil.Breite != 200 {
		t.Errorf("Breite nachher %d — der Speicher hielt die alte Vorlage", nachher.Stil.Breite)
	}

	// **Derselbe Weg für den Stil einer Karte**, der nach Namen gemerkt wird.
	karte := &domain.CardRow{SetID: 7, Values: []domain.Value{{Field: Kartenstilfeld, Text: "probe-schlicht"}}}
	if _, _, da := speicher.StilFuerKarte(t.Context(), karte); !da {
		t.Fatal("der Kartenstil fehlt")
	}

	zeileErsetzen(t, filepath.Join(wurzel, "probe-schlicht.mse-style", "style"), "card width: 200", "card width: 150")

	if nachher, _, _ := speicher.StilFuerKarte(t.Context(), karte); nachher.Stil.Breite != 150 {
		t.Errorf("Breite des Kartenstils %d — der Speicher hielt die alte Vorlage", nachher.Stil.Breite)
	}
}

// **Auch das Spiel gehört zur Vorlage** — seine Kartenfelder und sein Skript.
func TestEinGeaendertesSpielGiltOhneNeustart(t *testing.T) {
	wurzel, speicher := stempelplatz(t)

	vorher, _, _ := speicher.StilFuerSet(t.Context(), 7)

	spiel := filepath.Join(wurzel, "probe.mse-game", "game")
	roh, err := os.ReadFile(spiel) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(spiel, append(roh, []byte("\ncard field:\n\ttype: text\n\tname: stempel\n")...), 0o600); err != nil { //nolint:gosec // aus t.TempDir()
		t.Fatal(err)
	}

	if nachher, _, _ := speicher.StilFuerSet(t.Context(), 7); len(nachher.Kartenfelder) != len(vorher.Kartenfelder)+1 {
		t.Errorf("%d Kartenfelder nachher, vorher %d — das Spiel wurde nicht neu gelesen",
			len(nachher.Kartenfelder), len(vorher.Kartenfelder))
	}
}

// **Unverändert bleibt es gemerkt** — geprüft wird, gelesen nicht; die Zeit
// der letzten Prüfung rückt vor.
func TestEinUnveraenderterStilBleibtGemerkt(t *testing.T) {
	_, speicher := stempelplatz(t)

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Fatal("der Stil fehlt")
	}

	erste := speicher.gemerkt[7]

	if _, _, da := speicher.StilFuerSet(t.Context(), 7); !da {
		t.Fatal("der Stil fehlt beim zweiten Mal")
	}

	zweite := speicher.gemerkt[7]
	if zweite.stempel != erste.stempel || !zweite.geprueft.After(erste.geprueft) {
		t.Errorf("Stempel %s → %s, geprüft %v → %v", erste.stempel, zweite.stempel, erste.geprueft, zweite.geprueft)
	}
}
