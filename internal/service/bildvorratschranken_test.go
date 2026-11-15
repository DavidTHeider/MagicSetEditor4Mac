package service_test

// **Was durch den Bildvorrat passt** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Die vier Schranken aus dem Satz des Issues**: Signatur, Bytezahl,
// Pixelzahl und Quote — dazu die Frist. Die Läufe über Besitz, Kennung und
// Lebenszyklus stehen in `bildvorrat_test.go` daneben.
//
// **Zwei davon sind derselbe Angriff in zwei Gestalten**: eine Antwort, die
// zu viel wiegt, und eine, die zu viel *behaupten* wird. Die erste sieht die
// Bytegrenze, die zweite nur der Kopf.

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Gelesen wird der Kopf der Datei**, nicht ein `Content-Type` und nicht
// eine Endung: beide sagt der Anbieter, und beide kann er falsch sagen.
func TestOhneBildsignaturKeinBild(t *testing.T) {
	t.Parallel()

	for name, inhalt := range map[string][]byte{
		"ein Text":            []byte("das hier ist kein Bild"),
		"leer":                {},
		"ein SVG":             []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script/></svg>`),
		"ein halber PNG-Kopf": []byte("\x89PNG"),
		"ein GIF":             []byte("GIF89a…"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorrat, _ := einVorrat(t, nil)

			_, fehler := vorrat.Legen(einBildbesitz(), inhalt)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", name, fehler)
			}
		})
	}
}

// **Der zweite Fall steht hier, weil der Kopf ihn verspricht.** Er prüfte
// zuerst nur PNG, und ein JPEG wäre still gescheitert: `image.DecodeConfig`
// kennt nur, was ein Paket registriert hat, und `internal/service` hat
// `image/jpeg` bis zu diesem Werkstück nie eingebunden (§10).
func TestPNGUndJPEGKommenDurch(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		inhalt []byte
		typ    string
	}{
		"ein PNG":  {einVorratsbild(t, 12, 9), "image/png"},
		"ein JPEG": {einJPEG(t, 12, 9), "image/jpeg"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorrat, _ := einVorrat(t, nil)

			eintrag, fehler := vorrat.Legen(einBildbesitz(), fall.inhalt)
			if fehler != nil {
				t.Fatalf("%s kam nicht durch: %v", name, fehler)
			}

			if eintrag.Medientyp != fall.typ {
				t.Fatalf("der Medientyp ist %q statt %q", eintrag.Medientyp, fall.typ)
			}

			if eintrag.Breite != 12 || eintrag.Hoehe != 9 {
				t.Fatalf("die Maße sind %d×%d statt 12×9", eintrag.Breite, eintrag.Hoehe)
			}
		})
	}
}

// **Zu viele Bytes werden abgewiesen, bevor irgendetwas ausgepackt wird.**
func TestZuVieleBytesKommenNichtHerein(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, func(g *service.Bildgrenzen) { g.JeBild = 64 })

	_, fehler := vorrat.Legen(einBildbesitz(), einVorratsbild(t, 64, 64))
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das zu große Bild kam durch: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "64") {
		t.Fatalf("die Meldung nennt die Grenze nicht: %s", fehler.Error())
	}
}

// **Und eine Bildbombe an den Pixeln.**
//
// Zweiundsechzig Bytes, die 40.000 × 40.000 behaupten. Die Bytegrenze sieht
// sie nicht — sie ist winzig.
func TestEineBildbombeWirdAnDenPixelnErkannt(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	bombe := einKopfMitMassen(40000, 40000)

	if int64(len(bombe)) >= service.UeblicheBildgrenzen().JeBild {
		t.Fatalf("die Bombe wiegt %d Bytes und liefe schon in die Bytegrenze",
			len(bombe))
	}

	_, fehler := vorrat.Legen(einBildbesitz(), bombe)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Bildbombe kam durch: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "Pixel") {
		t.Fatalf("die Meldung nennt die Pixel nicht: %s", fehler.Error())
	}
}

// **Nach der Frist ist das Bild fort — und seine Datei auch.**
func TestNachDerFristIstDasBildFort(t *testing.T) {
	t.Parallel()

	vorrat, uhr := einVorrat(t, func(g *service.Bildgrenzen) { g.Frist = time.Hour })

	eintrag, fehler := vorrat.Legen(einBildbesitz(), einVorratsbild(t, 8, 8))
	if fehler != nil {
		t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
	}

	*uhr = uhr.Add(time.Hour + time.Second)

	if _, _, danach := vorrat.Lesen(eintrag.Kennung, 7); !errors.Is(
		danach, domain.ErrNotFound) {
		t.Fatalf("das abgelaufene Bild kam noch heraus: %v", danach)
	}

	if weg := vorrat.Aufraeumen(); weg != 1 {
		t.Fatalf("aufgeräumt wurden %d Einträge statt einer", weg)
	}

	uebrig, fehler := os.ReadDir(vorrat.Ort())
	if fehler != nil {
		t.Fatalf("das Verzeichnis ließ sich nicht lesen: %v", fehler)
	}

	if len(uebrig) != 0 {
		t.Fatalf("es liegen noch %d Dateien da", len(uebrig))
	}
}

// **Weil das Alte jemand gleich übernehmen könnte.** Ein Vorrat, der beim
// Anlegen das älteste Bild wegwirft, nimmt einem Menschen den Vorschlag
// weg, den er gerade ansieht — und niemand hat ihm gesagt, dass er weg ist.
func TestDieQuoteHaelt(t *testing.T) {
	t.Parallel()

	bild := einVorratsbild(t, 32, 32)

	vorrat, _ := einVorrat(t, func(g *service.Bildgrenzen) {
		g.Insgesamt = int64(2*len(bild)) + 1
	})

	ersteZwei := make([]string, 0, 2)

	for range 2 {
		eintrag, fehler := vorrat.Legen(einBildbesitz(), bild)
		if fehler != nil {
			t.Fatalf("das Bild ließ sich nicht ablegen: %v", fehler)
		}

		ersteZwei = append(ersteZwei, eintrag.Kennung)
	}

	if _, fehler := vorrat.Legen(einBildbesitz(), bild); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("das dritte Bild kam durch: %v", fehler)
	}

	for _, kennung := range ersteZwei {
		if _, _, fehler := vorrat.Lesen(kennung, 7); fehler != nil {
			t.Fatalf("das Bild %q ist fort: %v", kennung, fehler)
		}
	}

	// **Und wer eines verwirft, macht Platz.**
	if fehler := vorrat.Verwerfen(ersteZwei[0], 7); fehler != nil {
		t.Fatalf("das Verwerfen scheiterte: %v", fehler)
	}

	if _, fehler := vorrat.Legen(einBildbesitz(), bild); fehler != nil {
		t.Fatalf("nach dem Verwerfen ging es immer noch nicht: %v", fehler)
	}
}

// **Ein Kopf, den kein Decoder liest, kommt nicht herein.**
//
// PNG-Signatur, danach Unsinn: die Signaturprüfung sagt ja, `DecodeConfig`
// sagt nein. Ohne diesen Lauf sähe niemand, dass der zweite Schritt greift.
func TestEinKaputterKopfKommtNichtHerein(t *testing.T) {
	t.Parallel()

	vorrat, _ := einVorrat(t, nil)

	if _, fehler := vorrat.Legen(einBildbesitz(),
		[]byte("\x89PNG\r\n\x1a\nund dann Unsinn")); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der kaputte Kopf kam durch: %v", fehler)
	}
}
