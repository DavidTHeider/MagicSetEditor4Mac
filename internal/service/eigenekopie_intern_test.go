package service

// **Die Fehlerwege der eigenen Kopie, die nur von innen erreichbar sind** —
// FMT-02
// (#1691).
// Der Bauort ist ein internes Feld; dieselbe Bauform wie beim Anlegen.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kopierablage merkt sich den Katalog — anders als `anlageablage`, die nichts
// zurückgibt: ohne Katalog endete die Kopie schon an »nicht gefunden«, und
// der Weg, um den es geht, bliebe unbetreten.
type kopierablage struct {
	anlageablage

	pakete []domain.Paket
}

func (a *kopierablage) Abgleichen(_ context.Context, pakete []domain.Paket) error {
	a.pakete = pakete

	return nil
}

func (a *kopierablage) Pakete(context.Context) ([]domain.Paket, error) { return a.pakete, nil }

func TestEinUnbrauchbarerBauortHaeltDieKopieAuf(t *testing.T) {
	ort := t.TempDir()

	datei := filepath.Join(ort, "keinordner")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	verwalter := NeuerPaketverwalter(&kopierablage{}, ort, filepath.Join("..", "..", "saat"), ort)
	verwalter.bau = datei

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if _, err := verwalter.Kopieren(t.Context(), "mse-game", "probe", ""); err == nil ||
		!strings.Contains(err.Error(), "die Kopie schreiben") {
		t.Errorf("ein Bauort, der eine Datei ist: %v", err)
	}
}

// **Der Kurzname ersetzt nur seine Zeile** — mit Bytereihenfolgemarke,
// Wagenrücklauf und dort, wo es keinen gibt.
func TestDerKurznameErsetztNurSeineZeile(t *testing.T) {
	for name, fall := range map[string]struct{ vorher, nachher string }{
		"mit Marke und Wagenrücklauf": {
			"\ufeffshort name: Alt\r\ngame: probe\r\n",
			"\ufeffshort name: neu\r\ngame: probe\r\n",
		},
		"eingerückt zählt nicht": {
			"mse version: 2.0.0\n\tshort name: innen\n",
			"mse version: 2.0.0\nshort name: neu\n\tshort name: innen\n",
		},
		"ohne Kurzname": {
			"mse version: 2.0.0\ngame: probe\n",
			"mse version: 2.0.0\nshort name: neu\ngame: probe\n",
		},
	} {
		if neu := string(mitKurzname([]byte(fall.vorher), "neu")); neu != fall.nachher {
			t.Errorf("%s: %q", name, neu)
		}
	}
}

// **Ein Paket, das zwischen Finden und Kopieren verschwand**, meldet sich als
// nicht gefunden — und hinterlässt nichts.
func TestEinVerschwundenesPaketWirdNichtKopiert(t *testing.T) {
	ort := t.TempDir()
	verwalter := NeuerPaketverwalter(&kopierablage{}, ort, ort)

	fort := domain.Paket{Art: "mse-game", Name: "fort", Pfad: filepath.Join(t.TempDir(), "fort.mse-game")}
	if _, err := verwalter.kopieVon(t.Context(), fort, nil, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein verschwundenes Paket: %v", err)
	}

	if eintraege, err := os.ReadDir(ort); err != nil || len(eintraege) != 0 {
		t.Errorf("im verwalteten Ort: %v, %v", eintraege, err)
	}
}
