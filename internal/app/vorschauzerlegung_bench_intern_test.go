//go:build fremdpakete

// **Wo die Zeit einer Vorschau hingeht** —
// #1270.
//
// # Warum diese Messung vor jeder Optimierung steht
//
// Im laufenden Fenster braucht eine Vorschau **0,4–1,0 s**; die Schwelle für
// „sofort" liegt bei 100 ms. `BenchmarkWegRastern` misst für `Black.mse-set`
// **76 ms** — und beantwortet damit nicht, wo die übrigen Hunderte
// Millisekunden liegen: an einem anderen Stil, an der Kodierung oder an der
// Naht.
//
// **Die Reichweite einer Messung ist die Reichweite ihrer Aussage** (§10). Vier
// Zahlen aus einem Lauf sagen mehr als eine Vermutung über vier Stellen.
//
// # Was hier getrennt wird
//
//	Kartenplan   — Paket lesen, Skripte auswerten, Maße rechnen
//	Zeichne      — die Ebenen malen
//	png.Encode   — das Bild in Bytes
//	base64       — die Bytes in eine Data-URI
//
// **Was hier nicht drin ist**: die Wails-Naht. Sie ist getrennt gemessen —
// eine winzige Antwort (`Datenstand`) kostet im Fenster **0,3 ms**, die Naht
// selbst ist also nicht der Grund.
package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// schwelleVorschauM15 gilt für **eine** Karte von `Vanguard.mse-set` mit
// `magic-m15-altered`, in Bildschirmauflösung.
//
// **Gemessen am 2026-08-28** auf einem MacBook mit Apple Silicon: 53,6 ms für
// den ganzen Go-Weg — Kartenplan 28,2, Zeichnen 13,2, PNG 6,5, Base64 0,07.
//
// **80 ms lassen keine Verdopplung durch** (Faktor 1,5) — dieselbe Enge wie
// bei `schwelleRastern` und aus demselben Grund: gefangen werden soll ein
// spürbarer Rückschritt, nicht das Rauschen.
//
// **Und sie liegt über den 53,6 ms und unter den 100 ms**, die #1270 als Ziel
// nennt. Dass der Go-Weg das Ziel heute **hält**, ist die eigentliche Auskunft
// dieser Messung; was im Fenster darüber hinaus vergeht, liegt nicht hier.
const schwelleVorschauM15 = 80 * time.Millisecond

// BenchmarkWegVorschauzerlegung misst die vier Teile einer Vorschau.
//
// **An `Vanguard.mse-set`**, weil das der Stil ist, an dem die 0,4–1,0 s
// gemessen wurden (`magic-m15-altered`) — und nicht `Black.mse-set`, an dem
// die vorhandene Schwelle hängt. Zwei Stile sind zwei Messungen; eine Zahl von
// dem einen sagt über den anderen nichts.
func BenchmarkWegVorschauzerlegung(b *testing.B) {
	anwendung := messanwendung(b)
	setID := eingefahren(b, anwendung, "Vanguard.mse-set")

	zeilen := anwendung.Kartenzeilen(context.Background(), setID)
	if zeilen.Fehler != nil {
		b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	kennung, befund := kennungAus(zeilen.Zeilen[0].ID)
	if befund != nil {
		b.Fatalf("Kennung: %+v", befund)
	}

	// **Im Profil der Anwendung, wie die Vorschau** — ohne las der Planbauer
	// den Setkopf nicht und maß still den Behelfsplan (QA-03, #1656).
	profil := anwendung.imProfil(context.Background())

	karte, lesefehler := anwendung.kartenleser.KarteLesen(profil, kennung)
	if lesefehler != nil {
		b.Fatalf("KarteLesen: %v", lesefehler)
	}

	// **Einmal vorab**, damit Stilpaket, Schriften und Bilder warm sind —
	// gemessen wird, was danach **jeder Tastendruck** kostet.
	erste := anwendung.Vorschau(context.Background(), zeilen.Zeilen[0].ID, 1)
	if erste.Fehler != nil {
		b.Fatalf("Vorschau: %+v", erste.Fehler)
	}

	keineFehlendeSchrift(b, erste.Vorschau.Befunde)

	plan, befunde, planfehler := anwendung.planbauer.Kartenplan(profil, karte)
	if planfehler != nil || len(befunde) > 0 {
		b.Fatalf("Kartenplan: %v %v — gemessen würde der Behelfsplan", planfehler, befunde)
	}

	bild, _, rasterfehler := anwendung.rasterer.Zeichne(
		context.Background(), plan, 1, render.GueteMittel)
	if rasterfehler != nil {
		b.Fatalf("Zeichne: %v", rasterfehler)
	}

	var gepackt bytes.Buffer
	if err := anwendung.packen(&gepackt, bild); err != nil {
		b.Fatalf("packen: %v", err)
	}

	roh := gepackt.Bytes()

	b.Logf("Bild %dx%d, PNG %d KB, Data-URI %d KB",
		bild.Bounds().Dx(), bild.Bounds().Dy(),
		len(roh)/1024, base64.StdEncoding.EncodedLen(len(roh))/1024)

	for _, teil := range []struct {
		name string
		ruf  func()
	}{
		{"Kartenplan", func() {
			//nolint:errcheck // gemessen wird die Zeit, nicht der Fehler
			_, _, _ = anwendung.planbauer.Kartenplan(profil, karte)
		}},
		{"Zeichne", func() {
			//nolint:errcheck // dito
			_, _, _ = anwendung.rasterer.Zeichne(
				context.Background(), plan, 1, render.GueteMittel)
		}},
		{"png.Encode", func() {
			var puffer bytes.Buffer
			//nolint:errcheck // dito
			_ = anwendung.packen(&puffer, bild)
		}},
		{"base64", func() {
			_ = base64.StdEncoding.EncodeToString(roh)
		}},
		{"Vorschau ganz", func() {
			_ = anwendung.Vorschau(context.Background(), zeilen.Zeilen[0].ID, 1)
		}},
	} {
		b.Run(teil.name, func(b *testing.B) {
			for b.Loop() {
				teil.ruf()
			}

			b.Logf("%s: %v je Durchlauf", teil.name,
				(b.Elapsed() / time.Duration(max(b.N, 1))).Round(time.Microsecond))
		})
	}

	// **Und der ganze Weg hält seine Schwelle.**
	//
	// **Am Ende und über den äußeren Lauf**: `b.Elapsed()` zählt hier die Zeit
	// aller Unterläufe zusammen, und deren Zahl steht in `b.N` des äußeren —
	// die einzelnen Teile haben ihre eigene Ausgabe, gefangen wird der Weg.
	for b.Loop() {
		_ = anwendung.Vorschau(context.Background(), zeilen.Zeilen[0].ID, 1)
	}

	haelt(b, "eine Vorschau ganz (m15-altered)", schwelleVorschauM15)
}
