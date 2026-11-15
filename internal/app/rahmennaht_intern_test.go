//go:build fremdpakete

// Der Wächter über die **Naht**, die #240 zusagt: nicht dass ein Planbauer
// eine Bildebene baut, sondern dass die **laufende Anwendung** eine Karte aus
// einem bezogenen Paket zeichnet.
//
//	MSE_FREMDPAKETE="$HOME/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data" \
//	  go test -tags fremdpakete ./internal/app/ -run RahmenDurchDieNaht -v
package app

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

// **Die Anwendung zeichnet eine Karte aus einem bezogenen Paket.**
//
// # Warum es diesen Wächter zusätzlich gibt
//
// `service.TestDerRahmenEinerRotenKarteZeichnet` ruft `Stilplan` **direkt** auf,
// mit einer von Hand zusammengesetzten `Vorlage`. Das ist eine gute Zusage über
// den Planbauer und **keine** über die Anwendung: Stilspeicher, Kartenplan,
// Composition Root, Bildquelle, Schriftsuche, Rasterer und das DTO liegen
// allesamt daneben. Ein defekter Bildpfad oder eine gelöste Verdrahtung in
// `app.go` ließe ihn grün.
//
// Gefunden vom Zweitblick über #274/#275/#277/#278
// (#292). Die
// Klasse dahinter: **ein Wächter prüft ein Zwischenmodell statt der Zusage
// durch die Schichten.**
//
// # Was er prüft
//
// Er geht über `Vorschau` — dieselbe Methode, die das Fenster ruft — und sieht
// sich das **Bild** an: eine Karte, die gezeichnet hat, ist nicht weiß.
//
// # Was er nicht prüft
//
// **Ob sie richtig aussieht.** Das ist die Aufgabe von `TestDerAbstandZuMSE2`,
// und es ist eine andere Frage: dieser hier fragt, ob überhaupt etwas ankommt.
//
// **Und er endet vor der letzten Naht.** Er ruft die Go-Methode auf und liest
// das Go-DTO; die Wails-Bindung, die JSON-Marken und das Auspacken im Frontend
// liegen dahinter. Eine falsche Struct-Marke — ein `omitempty` auf einem Slice
// etwa — käme hier durch und ließe die Seite leer. Genau diese Lücke benennt
// `AGENTS.md` §3 unter „Zwischen den Prüfebenen", und ihre Antwort ist keine
// weitere Vorrichtung, sondern die **gebaute** Anwendung vor jeder Freigabe.
//
// **Auch die Verzeichnisse sind nicht die echten:** `gestartetMit` setzt Paket-
// und Saatort. Ein Fehler im echten `Paketverzeichnis()` bliebe unsichtbar.
//
// Angestrichen vom Zweitblick über Phase A, der zu Recht darauf bestand, dass
// „durch die Naht" nicht „durch alles" heißt.
func TestDerRahmenDurchDieNaht(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	set := anwendung.SetOeffnen(t.Context(), m15Set(t))
	if set.Fehler != nil {
		t.Fatalf("öffnen: %+v", set.Fehler)
	}

	// **Der Stil steht schon im Set** (`stylesheet: m15`), und zwar in der
	// **Kurzform** — so schreibt sie ein echtes Set (`magic-set-generator`,
	// `src/mse.rs`). Damit prüft dieser Wächter zugleich, dass
	// `<spiel>-<stylesheet>` aufgelöst wird; trüge er `magic-m15`, bliebe
	// genau der Fehler ungeprüft, der #292 aufgedeckt hat.

	karten := anwendung.Karten(t.Context(), set.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if len(karten.Karten) == 0 {
		t.Fatal("das Set führt keine Karte")
	}

	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	bild := bildAus(t, anwendung, vorschau.Vorschau.Bild)

	// **Die Kartengröße kommt aus dem bezogenen Stil**, nicht aus der Saat.
	// Steht hier 80×112, hat der Stilwechsel nicht gegriffen und alles Weitere
	// prüfte die falsche Vorlage.
	if bild.Bounds().Dx() != 375 || bild.Bounds().Dy() != 523 {
		t.Fatalf("die Vorschau ist %dx%d, erwartet 375x523 aus magic-m15",
			bild.Bounds().Dx(), bild.Bounds().Dy())
	}

	anteil := anteilDunkelAmRand(bild)

	t.Logf("%dx%d, %.1f %% des Randes dunkel, %d Befunde",
		bild.Bounds().Dx(), bild.Bounds().Dy(), anteil*100, len(vorschau.Vorschau.Befunde))

	for _, befund := range vorschau.Vorschau.Befunde {
		t.Logf("  %s", befund)
	}

	// **Die Schwelle ist gemessen, nicht geschätzt.** Mit voller Verdrahtung
	// liegt der Anteil bei 100 Prozent; mit einer Bildquelle ohne Wurzeln — der
	// Gegenprobe — bei 0. Neunzig Prozent liegen weit von beidem und lassen
	// Raum für die abgerundeten Übergänge, die der Eckenausschluss nicht ganz
	// erwischt.
	if anteil < 0.9 {
		t.Errorf("nur %.1f %% des Kartenrandes sind dunkel — das Rahmenbild des "+
			"Pakets ist nicht angekommen; Befunde: %v",
			anteil*100, vorschau.Vorschau.Befunde)
	}
}

// anteilDunkelAmRand misst den **äußersten Ring** der Karte.
//
// # Warum nicht die ganze Fläche
//
// Der erste Anlauf zählte dunkle Bildpunkte überall. Das war schon besser als
// „nicht weiß" — aber der Zweitblick über Phase A hat gezeigt, dass es nicht
// hält: die Phasen B bis E bringen **dunklen Text und dunkle Symbole** mit, und
// die wachsen genau in die Zahl hinein, die belegen soll, dass ein *Paketbild*
// angekommen ist. Ein vollständig schwarzes Bild bestünde ohne jedes Paketbild.
//
// **Am äußersten Ring steht nur der Kartenrand.** Kein Textkasten reicht
// dorthin, kein Symbol, kein Wasserzeichen — dort ist entweder das Rahmenbild
// des Pakets oder der weiße Grund, den dieses Programm selbst legt.
//
// # Warum die Ecken ausgenommen sind
//
// Der M15-Rahmen ist abgerundet; in den Ecken liegt kein Rand, sondern nichts.
// Ein Ring, der sie mitzählte, meldete auch bei perfektem Rahmen weniger als
// hundert Prozent und machte die Schwelle zur Glückssache.
//
// **Was er nicht prüft:** *welches* Paketbild angekommen ist. Ein falscher
// Rahmen in der richtigen Farbe käme durch. Das ist die Frage von
// `TestDerAbstandZuMSE2`, und sie wird dort mit einer Zahl beantwortet.
func anteilDunkelAmRand(bild image.Image) float64 {
	const (
		tiefe = 2  // wie weit der Ring nach innen reicht
		ecke  = 30 // was an jeder Ecke als Rundung ausgenommen bleibt
	)

	rahmen := bild.Bounds()
	breite, hoehe := rahmen.Dx(), rahmen.Dy()

	dunkel, gezaehlt := 0, 0

	pruefe := func(x, y int) {
		if (x < ecke || x >= breite-ecke) && (y < ecke || y >= hoehe-ecke) {
			return
		}

		gezaehlt++

		r, g, b, _ := bild.At(rahmen.Min.X+x, rahmen.Min.Y+y).RGBA()
		if r>>8 < 64 && g>>8 < 64 && b>>8 < 64 {
			dunkel++
		}
	}

	for tiefte := range tiefe {
		for x := range breite {
			pruefe(x, tiefte)
			pruefe(x, hoehe-1-tiefte)
		}

		for y := range hoehe {
			pruefe(tiefte, y)
			pruefe(breite-1-tiefte, y)
		}
	}

	if gezaehlt == 0 {
		return 0
	}

	return float64(dunkel) / float64(gezaehlt)
}

// m15Set legt ein Set auf dem **bezogenen** Spiel an.
//
// **Nicht `saatset`**: das steht auf `game: probe`, und ein Stilwechsel auf
// `magic-m15` wäre ein Wechsel zwischen zwei Spielen. Was hier geprüft wird,
// ist die Naht zu einem bezogenen Paket — also muss das Set von Anfang an dort
// zu Hause sein.
func m15Set(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "m15probe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: magic\nstylesheet: m15\n" +
		"set info:\n\ttitle: Aus dem bezogenen Paket\n" +
		"card:\n\tname: Feuerball\n\tcasting cost: 1R\n\tcard color: red\n" +
		"\ttype: Instant\n\ttext: Deal three damage to any target.\n"

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}
