package service

// **Die Fehlerwege des Anlegens** —
// #1350.
//
// **Ein interner Test, weil der Bauort ein internes Feld ist.** Er zu setzen
// ist die einzige Art, die beiden Schreibfehler zu fahren, ohne die Platte des
// Rechners in einen Zustand zu bringen, den ein Test nicht hinterlassen darf —
// dieselbe Bauform wie `schluesselbund.Bund.ort` (§3).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Bauort, der kein Verzeichnis ist, wird gemeldet.**
func TestEinUnbrauchbarerBauortWirdGemeldet(t *testing.T) {
	ort := t.TempDir()

	// **Eine Datei statt eines Verzeichnisses** — `MkdirAll` scheitert daran.
	datei := filepath.Join(ort, "keinordner")
	if err := os.WriteFile(datei, []byte("x"), 0o600); err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	verwalter := NeuerPaketverwalter(&anlageablage{}, ort, ort)
	verwalter.bau = datei

	if _, err := verwalter.Anlegen(t.Context(), Neuanlage{
		Art: "mse-game", Name: "eigenes", Vollname: "Eigenes",
	}); err == nil {
		t.Error("ein Bauort, der eine Datei ist, kam durch")
	}
}

// **Und eine Hauptdatei, die schon ein Verzeichnis ist, ebenso.**
//
// Der zweite Schreibfehler, und er steht neben dem ersten: die beiden liegen
// an verschiedenen Stellen, und ein Test für einen sagte über den anderen
// nichts.
func TestEineUnbeschreibbareHauptdateiWirdGemeldet(t *testing.T) {
	ort := t.TempDir()
	bau := t.TempDir()

	// **Über den Griff und nicht über einen Riegel im Pfad** — hier lag ein
	// Verzeichnis `game` am festen Wegwerfort; seit jeder Bau seinen eigenen
	// Ort hat, gibt es den festen nicht mehr (`wegwerfpaket`).
	verwalter := NeuerPaketverwalter(&anlageablage{}, ort, ort)
	verwalter.bau = bau
	verwalter.ablegen = func(string, []byte) error { return errors.New("die Platte klemmt") }

	if _, err := verwalter.Anlegen(t.Context(), Neuanlage{
		Art: "mse-game", Name: "eigenes", Vollname: "Eigenes",
	}); err == nil || !strings.Contains(err.Error(), "das Paket schreiben") {
		t.Errorf("eine unbeschreibbare Hauptdatei kam durch: %v", err)
	}
}

// **Ein Stil auf einem Spiel, das es nicht gibt, wird am Feld gemeldet.**
func TestEinStilAufEinemFehlendenSpielWirdGemeldet(t *testing.T) {
	ort := t.TempDir()

	_, err := NeuerPaketverwalter(&anlageablage{}, ort, ort).Anlegen(t.Context(), Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "S", Spiel: "gibtsnicht",
	})

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Field != "spiel" {
		t.Fatalf("Fehler ist %v", err)
	}
}

// **Ein Spiel, das sich nicht zerteilen lässt, ebenso.**
func TestEinUnlesbaresSpielWirdGemeldet(t *testing.T) {
	ort := t.TempDir()
	// **Ein Schlüssel, der nur aus einem Doppelpunkt besteht** — der
	// Zerteiler meldet „der Schlüssel ist leer".
	spielMit(t, ort, "krumm", "mse version: 2.0.0\n: ohne Schlüssel\n")

	_, err := NeuerPaketverwalter(&anlageablage{}, ort, ort).Anlegen(t.Context(), Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "S", Spiel: "krumm",
	})

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Field != "spiel" {
		t.Fatalf("Fehler ist %v", err)
	}
}

// **Und ein Spiel ohne Kartenfeld.**
//
// Ein Stil darauf legte nichts aus — und ein Paar aus Spiel und Stil, das eine
// leere Karte malt, ist keine Vorlage.
func TestEinSpielOhneKartenfeldWirdGemeldet(t *testing.T) {
	ort := t.TempDir()
	spielMit(t, ort, "leer", "mse version: 2.0.0\nshort name: leer\n")

	_, err := NeuerPaketverwalter(&anlageablage{}, ort, ort).Anlegen(t.Context(), Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "S", Spiel: "leer",
	})

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Field != "spiel" {
		t.Fatalf("Fehler ist %v", err)
	}
}

// **Ein Feld, das die Saat nicht kennt, bekommt eine Zeilenhöhe.**
//
// Ein Stil auf einem **bezogenen** Spiel ist der Normalfall, und dort heißen
// die Felder anders. Ein Feld, von dem niemand weiß, was darin steht, als
// Absatz auszulegen wäre eine Behauptung.
func TestEinFremdesFeldBekommtEineZeilenhoehe(t *testing.T) {
	ort := t.TempDir()
	spielMit(t, ort, "fremd",
		"mse version: 2.0.0\nshort name: fremd\ncard field:\n\ttype: text\n\tname: rarity\n")

	verwalter := NeuerPaketverwalter(&anlageablage{}, ort, ort)

	if _, err := verwalter.Anlegen(t.Context(), Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "S", Spiel: "fremd",
	}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	stil, lesefehler := os.ReadFile( //nolint:gosec // Pfad aus t.TempDir()
		filepath.Join(ort, "fremd-schlicht.mse-style", "style"))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !strings.Contains(string(stil), "height: 28") {
		t.Errorf("das fremde Feld bekam keine Zeilenhöhe:\n%s", stil)
	}
}

// spielMit legt ein Spielpaket mit diesem Inhalt an.
func spielMit(t *testing.T, ort, name, inhalt string) {
	t.Helper()

	verzeichnis := filepath.Join(ort, name+".mse-game")
	if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
		t.Fatalf("das Spiel anlegen: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(verzeichnis, "game"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("das Spiel schreiben: %v", err)
	}
}

// anlageablage ist ein Katalog, der alles hinnimmt.
type anlageablage struct{}

func (anlageablage) Abgleichen(context.Context, []domain.Paket) error { return nil }

func (anlageablage) Pakete(context.Context) ([]domain.Paket, error) { return nil, nil }

func (anlageablage) PaketEntfernen(context.Context, string, string) error { return nil }

func (anlageablage) PaketeNachMuster(context.Context, string) ([]domain.Paket, error) {
	return nil, nil
}
