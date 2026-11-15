// **Die Flächen ohne Fenster behalten ihr Zeichenwerk** —
// #2005.
//
// Bis dahin baute [app.Zeichennaht] es je Ruf: Stilpaket lesen, Skripte
// zerteilen, Schriften suchen, Rahmen entschlüsseln — für **jede** Karte und
// jeden Ausdruck über `mse-mcp` und `mse-rest`. Der Leistungsvergleich (QA-04)
// maß 1,8 s je Karte gegen 43–61 ms im Fenster.
//
// Behalten heißt: **was sich ändert, muss trotzdem gelten.** Die Tests hier
// prüfen beides — dass das Werk bleibt, und dass ein Stilwechsel, ein neues
// Paket und eine geänderte Paketdatei es trotzdem erreichen.
package app_test

import (
	"bytes"
	"database/sql"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// zeichenlage ist ein Zuhause mit der Saat im Paketverzeichnis und ein
// Bestand mit einem Set darin.
type zeichenlage struct {
	pakete         string
	db             *sql.DB
	setID, karteID int64
}

func zeichenzuhause(t *testing.T) zeichenlage {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	lage := zeichenlage{db: frischerBestand(t)}
	lage.setID, lage.karteID = saatsetImBestand(t, lage.db)

	pakete, err := app.Paketverzeichnis()
	if err != nil {
		t.Fatal(err)
	}

	lage.pakete = pakete

	return lage
}

// naht ruft die Zeichennaht über dem Bestand der Lage.
func (l zeichenlage) naht(t *testing.T) *app.Zeichenwerk {
	t.Helper()

	werk, err := app.Zeichennaht(l.db)
	if err != nil {
		t.Fatalf("Zeichennaht: %v", err)
	}

	return werk
}

// wechsle stellt den Stil des Sets über die Stilnaht um — wie `mse-mcp` und
// `mse-rest` es tun.
func (l zeichenlage) wechsle(t *testing.T, paket string) {
	t.Helper()

	weg, err := app.StilnahtUeber(l.db)
	if err != nil {
		t.Fatal(err)
	}

	if err := weg.Wechseln(probekontext(t), l.setID, paket); err != nil {
		t.Fatalf("Wechseln: %v", err)
	}
}

// gezeichnet zeichnet eine Karte und gibt die Bytes der Datei.
func gezeichnet(t *testing.T, werk *app.Zeichenwerk, karteID int64) []byte {
	t.Helper()

	ziel := filepath.Join(t.TempDir(), "karte.png")
	if _, err := werk.SchreibeBild(probekontext(t), karteID, ziel, render.GueteMittel); err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	return roh
}

func TestDieZeichennahtBehaeltIhrWerkJeBestand(t *testing.T) {
	lage := zeichenzuhause(t)

	if erstes, zweites := lage.naht(t), lage.naht(t); erstes != zweites {
		t.Error("zwei Rufe über demselben Bestand bauten zwei Werke")
	}

	anderes, err := app.Zeichennaht(frischerBestand(t))
	if err != nil {
		t.Fatal(err)
	}

	if anderes == lage.naht(t) {
		t.Error("zwei Bestände teilen ein Werk — die Leser darin gehören dem ersten")
	}
}

// **Ein Paket, das dazukommt, gilt beim nächsten Ruf** — FMT-02 hielt das für
// jede Naht fest, und ein behaltenes Werk mit festen Wurzeln hätte es wieder
// gebrochen.
func TestEinNeuesPaketGiltBeimNaechstenRuf(t *testing.T) {
	lage := zeichenzuhause(t)

	vorher := lage.naht(t)

	// **Ein Paket in einer vorhandenen Wurzel**: die Wurzeln bleiben dieselben.
	if err := os.CopyFS(filepath.Join(lage.pakete, "probe-kopie.mse-style"),
		os.DirFS(filepath.Join(lage.pakete, "probe-schlicht.mse-style"))); err != nil {
		t.Fatal(err)
	}

	if nachher := lage.naht(t); nachher == vorher {
		t.Error("ein neues Paket in der Wurzel ließ das Werk stehen")
	}

	// **Und eine neue Wurzel**, wie ein bezogenes Repositorium sie mitbringt.
	sammlung := filepath.Join(lage.pakete, "sammlung", "data")
	if err := os.CopyFS(filepath.Join(sammlung, "probe-zweit.mse-style"),
		os.DirFS(filepath.Join(lage.pakete, "probe-schlicht.mse-style"))); err != nil {
		t.Fatal(err)
	}

	if nachher := lage.naht(t); !strings.Contains(strings.Join(nachher.Wurzeln, "\n"), sammlung) {
		t.Errorf("die neue Wurzel fehlt: %v", nachher.Wurzeln)
	}
}

// **Der Stilwechsel einer Fläche ohne Fenster wirkt beim nächsten Bild** —
// und das Werk bleibt dabei dasselbe. Ohne die zweite Hälfte bewiese die
// erste nur, dass neu gebaut wurde.
func TestEinStilwechselWirktSofort(t *testing.T) {
	lage := zeichenzuhause(t)

	werk := lage.naht(t)
	vorher := gezeichnet(t, werk, lage.karteID)

	lage.wechsle(t, "probe-bild")

	if lage.naht(t) != werk {
		t.Fatal("der Stilwechsel baute ein neues Werk — dann prüft dieser Test das Behalten nicht")
	}

	if bytes.Equal(gezeichnet(t, lage.naht(t), lage.karteID), vorher) {
		t.Error("nach dem Stilwechsel kam dasselbe Bild heraus")
	}
}

// **Und eine Paketdatei, die sich ändert, wirkt ohne Neustart** — auch ein
// Bild: die Bildquelle hält es entschlüsselt.
//
// **Gewartet wird eine Sekunde**: so lange gilt ein gemerkter Stil ungeprüft
// (`stempelpause` in `internal/service/stilstempel.go`).
func TestEinGeaenderterRahmenWirktOhneNeustart(t *testing.T) {
	lage := zeichenzuhause(t)
	lage.wechsle(t, "probe-bild")

	werk := lage.naht(t)
	vorher := gezeichnet(t, werk, lage.karteID)

	rot := image.NewNRGBA(image.Rect(0, 0, 375, 523))
	for i := range rot.Pix {
		rot.Pix[i] = []uint8{200, 20, 20, 255}[i%4]
	}

	var roh bytes.Buffer
	if err := png.Encode(&roh, rot); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1100 * time.Millisecond)

	if err := os.WriteFile(filepath.Join(lage.pakete, "probe-bild.mse-style", "rahmen.png"), roh.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	if lage.naht(t) != werk {
		t.Fatal("die Änderung baute ein neues Werk — dann prüft dieser Test das Räumen nicht")
	}

	nachher := gezeichnet(t, werk, lage.karteID)
	if bytes.Equal(nachher, vorher) {
		t.Fatal("nach der Änderung am Rahmen kam dasselbe Bild heraus")
	}

	bild, err := png.Decode(bytes.NewReader(nachher))
	if err != nil {
		t.Fatal(err)
	}

	farbe, istNRGBA := color.NRGBAModel.Convert(bild.At(100, 200)).(color.NRGBA)
	if !istNRGBA {
		t.Fatalf("kein NRGBA: %T", farbe)
	}

	if farbe.R < 150 || farbe.G > 80 {
		t.Errorf("Bildpunkt (100,200) = %+v — erwartet der neue rote Rahmen", farbe)
	}
}

// **Ein behaltenes Werk ersetzt die Orte nicht**: fehlt das Zuhause, fehlt
// die Naht — auch über einem Bestand, für den es schon ein Werk gibt.
func TestEinBehaltenesWerkErsetztDasZuhauseNicht(t *testing.T) {
	lage := zeichenzuhause(t)
	lage.naht(t)

	t.Setenv("HOME", "")

	if _, err := app.Zeichennaht(lage.db); err == nil {
		t.Error("ohne Zuhause kam das behaltene Werk zurück")
	}
}
