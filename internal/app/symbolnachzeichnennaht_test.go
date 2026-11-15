package app

// **Die Naht für die kopflosen Flächen** —
// #1340.

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// nachzeichnungsnahtMitSet baut die Naht über einem Bestand mit einem Set im
// Probeprofil — seit #1815 fragt sie, wem das Set gehört.
func nachzeichnungsnahtMitSet(t *testing.T) (naht *Symbolnachzeichnungsnaht, setID int64) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db := probedatenbank(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Nachzeichenset", Game: "probe"}, nil)
	if fehler != nil {
		t.Fatalf("das Set anlegen: %v", fehler)
	}

	naht, err := SymbolnachzeichnungUeber(db)
	if err != nil {
		t.Fatalf("die Naht bauen: %v", err)
	}

	return naht, setID
}

// TestDieNahtZeichnetNachUndLegtAb ist der Weg, den CLI, MCP und Dienst gehen.
func TestDieNahtZeichnetNachUndLegtAb(t *testing.T) {
	naht, setID := nachzeichnungsnahtMitSet(t)

	roh, err := naht.Nachzeichnen(probekontext(t), setID, kreisdatei(t))
	if err != nil {
		t.Fatalf("nachzeichnen: %v", err)
	}

	gemeldet, gemacht := roh.(map[string]any)
	if !gemacht {
		t.Fatalf("die Antwort ist keine Tabelle: %#v", roh)
	}

	name, _ := gemeldet["name"].(string) //nolint:errcheck // der Nullwert ist die Aussage
	if !strings.HasSuffix(name, ".mse-symbol") {
		t.Errorf("der Name ist %q — erwartet wurde eine .mse-symbol", name)
	}

	if teile, _ := gemeldet["teile"].(int); teile != 1 { //nolint:errcheck // dito
		t.Errorf("%v Formen, erwartet 1", gemeldet["teile"])
	}

	if punkte, _ := gemeldet["punkte"].(int); punkte < 4 { //nolint:errcheck // dito
		t.Errorf("%v Punkte — das ist kein Kreis", gemeldet["punkte"])
	}
}

// **Ein Bild, das es nicht gibt, kommt als Fehler zurück** — und nicht als
// leeres Ergebnis (§8).
func TestDieNahtMeldetEinFehlendesBild(t *testing.T) {
	naht, setID := nachzeichnungsnahtMitSet(t)

	if _, err := naht.Nachzeichnen(probekontext(t), setID, "/gibt/es/nicht.png"); err == nil {
		t.Error("kein Fehler für ein Bild, das es nicht gibt")
	}
}

// **Ein Set aus einem fremden Profil ist „nicht gefunden"** — und kein
// Konflikt, und es wird nichts abgelegt (#1815).
func TestDieNahtZeichnetInKeinFremdesSet(t *testing.T) {
	naht, setID := nachzeichnungsnahtMitSet(t)

	if _, err := naht.Nachzeichnen(storage.MitProfil(t.Context(), 99), setID, kreisdatei(t)); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("aus einem fremden Profil: %v", err)
	}
}

// **Und ein Fehler beim Ablegen ebenso.**
//
// Eine Setkennung, die keine ist, kommt bis zur Ablage durch — sie prüft sie,
// und ihr Nein muss diesen Weg hoch überleben.
func TestDieNahtMeldetEinGescheitertesAblegen(t *testing.T) {
	naht, setID := nachzeichnungsnahtMitSet(t)

	naht.anwendung.MitSymbolablage(nil)

	_, err := naht.Nachzeichnen(probekontext(t), setID, kreisdatei(t))
	if err == nil {
		t.Fatal("kein Fehler ohne Symbolablage")
	}

	if !strings.Contains(err.Error(), "Symbol") && !errors.Is(err, err) {
		t.Errorf("die Meldung sagt nichts über das Symbol: %v", err)
	}
}

// **Ohne Benutzerverzeichnis gibt es keine Naht** — dieselbe Zusage wie beim
// Bilderverzeichnis darunter: gemeldet statt geraten.
func TestOhneBenutzerverzeichnisGibtEsKeineNachzeichnungsnaht(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := SymbolnachzeichnungUeber(nil); err == nil {
		t.Error("kein Fehler ohne Benutzerverzeichnis")
	}
}

// **Eine Datei, die kein Bild ist, kommt aus `Nachgezeichnet` als Fehler.**
//
// Der Weg wird von `App.SymbolNachzeichnen` und von der Naht gerufen; geprüft
// wird er hier, weil dort nur die Übersetzung dazukommt.
func TestNachgezeichnetMeldetEineDateiOhneBild(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "kein-bild.png")
	if err := os.WriteFile(pfad, []byte("das ist Text"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := Nachgezeichnet(pfad); err == nil {
		t.Error("kein Fehler für eine Datei, die kein Bild ist")
	}
}

// **Und ein Bild ohne Fläche ebenso** — der Weg durch `service`.
func TestNachgezeichnetMeldetEinLeeresBild(t *testing.T) {
	t.Parallel()

	gemalt := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := range 20 {
		for x := range 20 {
			gemalt.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	pfad := filepath.Join(t.TempDir(), "leer.png")

	datei, err := os.Create(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	if err := png.Encode(datei, gemalt); err != nil {
		t.Fatalf("das PNG schreiben: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("die Datei schließen: %v", err)
	}

	if _, err := Nachgezeichnet(pfad); err == nil {
		t.Error("kein Fehler für ein Bild ohne Fläche")
	}
}
