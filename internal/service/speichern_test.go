package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// herkunftspaket legt ein Paketverzeichnis an, wie es auf der Platte liegt:
// die Hauptdatei **und** ein Bild daneben.
func herkunftspaket(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for name, inhalt := range map[string]string{
		"set":    probenquelle,
		"image1": "\x89PNG hier steht ein Bild",
	} {
		if err := os.WriteFile(filepath.Join(ort, name), []byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	return ort
}

// mitHerkunft baut einen Bestand, dessen Set aus einem echten Paket stammt.
func mitHerkunft(t *testing.T, herkunft string) quellenDouble {
	t.Helper()

	bestand := mitQuelle()
	bestand.quelle.Herkunft = herkunft

	return bestand
}

// **Beim Speichern kommt mit, was nicht in der Datenbank steht.**
//
// # Woher diese Zusage kommt
//
// [0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)
// führt die **Hauptdatei** mit, nicht die Bilder — die wögen Megabyte und
// gehören nicht in eine Datenbank. Wer ein Set anderswohin speichert und nur
// die Hauptdatei bekommt, hat ein Set ohne Bilder: es öffnet, und die Karten
// sind leer.
func TestBeimSpeichernKommenDieBilderMit(t *testing.T) {
	herkunft := herkunftspaket(t)
	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")

	ergebnis, err := service.NeuerSetschreiber(mitHerkunft(t, herkunft), &ortsschreiber{}).
		Schreibe(t.Context(), 7, ziel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if len(ergebnis.Befunde) != 0 {
		t.Errorf("Befunde: %v, erwartet keine", ergebnis.Befunde)
	}

	bild, err := os.ReadFile(filepath.Join(ziel, "image1")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das Bild fehlt: %v", err)
	}

	if !strings.HasPrefix(string(bild), "\x89PNG") {
		t.Errorf("das Bild ist nicht dasselbe: %q", bild)
	}
}

// **Ein Set über seine Herkunft zu speichern nimmt die Bilder nicht weg.**
//
// Der häufigste Fall: ⌘S auf ein Set, das aus einer Datei kam. Wer dabei das
// Verzeichnis leerte und nur die Hauptdatei schriebe, verlöre alles daneben.
func TestSpeichernUeberDieHerkunftBehaeltDieBilder(t *testing.T) {
	herkunft := herkunftspaket(t)

	if _, err := service.NeuerSetschreiber(mitHerkunft(t, herkunft), &ortsschreiber{}).
		Schreibe(t.Context(), 7, herkunft); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := os.Stat(filepath.Join(herkunft, "image1")); err != nil {
		t.Errorf("das Bild ist fort: %v", err)
	}
}

// **Eine Herkunft, die es nicht mehr gibt, wird benannt — nicht verschwiegen.**
//
// Die Hauptdatei steht in der Datenbank und kommt heraus; die Bilder nicht. Ein
// Set ohne Bilder ist brauchbarer als gar keines, aber niemand darf das
// erfahren, indem er es öffnet (§8).
func TestEineFehlendeHerkunftWirdBenannt(t *testing.T) {
	bestand := mitHerkunft(t, filepath.Join(t.TempDir(), "gibtsnicht.mse-set"))
	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")

	ergebnis, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ziel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if len(ergebnis.Befunde) == 0 {
		t.Fatal("kein Befund — dass die Bilder fehlen, bliebe unsichtbar")
	}

	// Die Hauptdatei ist trotzdem da.
	if _, err := os.Stat(filepath.Join(ziel, "set")); err != nil {
		t.Errorf("die Hauptdatei fehlt: %v", err)
	}
}

// Ein Set ohne Herkunft ist kein Befund.
//
// Es ist in der Anwendung entstanden; es gibt nichts, was mitkommen könnte.
func TestEinSetOhneHerkunftGibtKeinenBefund(t *testing.T) {
	bestand := mitQuelle()
	ziel := filepath.Join(t.TempDir(), "neu.mse-set")

	ergebnis, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ziel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if len(ergebnis.Befunde) != 0 {
		t.Errorf("Befunde: %v, erwartet keine", ergebnis.Befunde)
	}
}

// **Ohne Ziel geht das Speichern an die Herkunft.**
//
// Das ist ⌘S. Der Ort steht in der Quelle, nicht im Aufruf: die Oberfläche
// soll ihn nicht führen müssen, und ein Ort, den zwei Seiten führen, läuft
// auseinander.
func TestOhneZielGehtEsAnDieHerkunft(t *testing.T) {
	herkunft := herkunftspaket(t)

	ergebnis, err := service.NeuerSetschreiber(mitHerkunft(t, herkunft), &ortsschreiber{}).
		Schreibe(t.Context(), 7, "")
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if ergebnis.Pfad != herkunft {
		t.Errorf("geschrieben nach %q, erwartet %q", ergebnis.Pfad, herkunft)
	}

	if _, err := os.Stat(filepath.Join(herkunft, "set")); err != nil {
		t.Errorf("die Hauptdatei fehlt: %v", err)
	}
}

// **Ohne Ziel und ohne Herkunft ist es ein Fehler, kein Befund.**
//
// Ein Set, das in der Anwendung entstanden ist, hat keinen Ort. Wo eine Datei
// hinsoll, kann niemand raten — die Oberfläche muss fragen.
func TestOhneZielUndOhneHerkunftIstEinFehler(t *testing.T) {
	_, err := service.NeuerSetschreiber(mitQuelle(), &ortsschreiber{}).Schreibe(t.Context(), 7, "")
	if err == nil {
		t.Fatal("kein Fehler — wohin wäre geschrieben worden?")
	}

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler: %v, erwartet ErrNotFound", err)
	}
}
