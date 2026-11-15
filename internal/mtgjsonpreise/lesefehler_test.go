package mtgjsonpreise_test

// **Was eine fremde Datei alles sein kann** — Werkstück A3 von Phase 5.3
// (#1442).
//
// „Fremde Daten sind der Normalfall, nicht der Ausnahmefall" (§8). Eine
// Preisdatei kommt von einem fremden Server, sie ist ein Gigabyte groß, und
// niemand liest sie vorher durch.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
)

// **Was kaputt ist, wird abgewiesen — und sagt, woran es lag.**
func TestEineKaputtePreisdateiWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"gar kein JSON":             `das ist keine Datei`,
		"kein Objekt":               `[1, 2, 3]`,
		"ohne data":                 `{"meta": {"date": "2026-09-11"}}`,
		"data ist kein Objekt":      `{"data": 42}`,
		"eine abgeschnittene Zeile": `{"data": {"aaaa1111-2222-3333-4444-555555555555": {`,
	}

	for name, text := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := mtgjsonpreise.Lies(
				strings.NewReader(text), gewollt(), mtgjsonpreise.Grenzen{})
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("abgewiesen wurde mit %v", fehler)
			}
		})
	}
}

// **Und was kein gzip ist, ebenso.**
func TestEineUngepacktePreisdateiWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, _, fehler := mtgjsonpreise.LiesGepackt(
		strings.NewReader(dieGestalt), gewollt(), mtgjsonpreise.Grenzen{})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen wurde mit %v", fehler)
	}
}

// **Der Deckel liegt auf dem ausgepackten Strom.**
//
// # Warum das die Grenze ist, die zählt
//
// Zehn Megabyte gzip können ein Terabyte werden. Eine Grenze auf der
// **gepackten** Seite wäre keine — sie ließe genau den Strom durch, gegen den
// sie schützen soll.
func TestDerDeckelGreiftAmAusgepacktenStrom(t *testing.T) {
	t.Parallel()

	_, _, fehler := mtgjsonpreise.LiesGepackt(
		bytes.NewReader(gepackt(t)), gewollt(),
		mtgjsonpreise.Grenzen{HoechstensBytes: 16})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("abgewiesen wurde mit %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "16 Byte") {
		t.Errorf("die Meldung nennt die Grenze nicht: %v", fehler)
	}

	// **Die Gegenprobe**: mit Platz geht derselbe Strom durch.
	if _, _, fehler := mtgjsonpreise.LiesGepackt(
		bytes.NewReader(gepackt(t)), gewollt(),
		mtgjsonpreise.Grenzen{HoechstensBytes: 1 << 20},
	); fehler != nil {
		t.Errorf("mit Platz: %v", fehler)
	}
}

// **Und einer auf der Zahl der Stände.**
func TestDerDeckelGreiftAnDerZahlDerStaende(t *testing.T) {
	t.Parallel()

	_, _, fehler := mtgjsonpreise.Lies(
		strings.NewReader(dieGestalt), gewollt(),
		mtgjsonpreise.Grenzen{HoechstensStaende: 2})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen wurde mit %v", fehler)
	}
}

// **Ein leerer Crosswalk gibt keine Stände — und keinen Fehler.**
//
// **Das ist kein Fehlschlag, sondern eine Auskunft**: die Datei führt jede
// Karte, die es gibt; wer keine davon besitzt, bekommt nichts. Ein Fehler
// hier machte aus einem leeren Bestand einen kaputten.
func TestEinLeererCrosswalkGibtNichts(t *testing.T) {
	t.Parallel()

	staende, befunde, fehler := mtgjsonpreise.Lies(
		strings.NewReader(dieGestalt), nil, mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	if len(staende) != 0 || len(befunde) != 0 {
		t.Errorf("%d Stände, %d Befunde", len(staende), len(befunde))
	}
}

// **Was die Anwendung nicht kennt, ist ein Befund und kein Abbruch.**
//
// §16: „Pflichtfelder streng, Zusatzfelder tolerant." Ein dritter Markt oder
// eine vierte Ausführung bei MTGJSON darf die zwei bekannten nicht ausfallen
// lassen — und **verschluckt** wird er auch nicht.
func TestWasUnbekanntIstWirdGemeldetUndUebergangen(t *testing.T) {
	t.Parallel()

	const fremd = `{"data": {"aaaa1111-2222-3333-4444-555555555555": {
	  "mond": {"haendler": {"currency": "USD", "retail": {"normal": {"2026-09-10": 1.00}}}},
	  "paper": {"cardkingdom": {"currency": "USD", "retail": {
	    "glitzernd": {"2026-09-10": 2.00},
	    "normal":    {"morgen": 3.00, "2026-09-10": 4.00}
	  }}}
	}}}`

	staende, befunde, fehler := mtgjsonpreise.Lies(
		strings.NewReader(fremd), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	// **Der eine Preis, der zu deuten war.**
	if len(staende) != 1 || staende[0].Wert != 400 {
		t.Fatalf("%d Stände: %+v", len(staende), staende)
	}

	// **Drei Befunde**: der Markt, die Ausführung und der Tag.
	if len(befunde) != 3 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}

	for _, befund := range befunde {
		if befund.Art != domain.ArtUmgebung || befund.Datei == "" {
			t.Errorf("der Befund zeigt auf nichts: %+v", befund)
		}
	}
}

// **Ein Preis, der kein Preis ist, wird gemeldet — die übrigen bleiben.**
func TestEinUndeutbarerPreisLaesstDieAnderenStehen(t *testing.T) {
	t.Parallel()

	const krumm = `{"data": {"aaaa1111-2222-3333-4444-555555555555": {
	  "paper": {"cardkingdom": {"currency": "USD", "retail": {"normal": {
	    "2026-09-09": 4.175,
	    "2026-09-10": 4.17
	  }}}}
	}}}`

	staende, befunde, fehler := mtgjsonpreise.Lies(
		strings.NewReader(krumm), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	if len(staende) != 1 || staende[0].Wert != 417 {
		t.Fatalf("%d Stände: %+v", len(staende), staende)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "4.175") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Was beim Lesen einer Karte schiefgeht, kommt heraus.**
//
// **Die Wege, die eine heile Datei nicht hergibt** (§3): eine Kennung, die
// kein Text ist, gibt es in JSON nicht — wohl aber einen Wert, der kein Objekt
// ist, und einen Strom, der mitten in einer Karte endet.
func TestEinFehlerBeimLesenEinerKarteKommtHeraus(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"eine gewollte Karte ist kein Objekt": `{"data": {
		  "aaaa1111-2222-3333-4444-555555555555": 42}}`,
		"eine gewollte Karte bricht ab": `{"data": {
		  "aaaa1111-2222-3333-4444-555555555555": {"paper":`,
		"eine fremde Karte bricht ab": `{"data": {
		  "cccc1111-2222-3333-4444-555555555555": {"paper":`,
		"ein Markt ist kein Objekt": `{"data": {
		  "aaaa1111-2222-3333-4444-555555555555": {"paper": 7}}}`,
	}

	for name, text := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := mtgjsonpreise.Lies(
				strings.NewReader(text), gewollt(), mtgjsonpreise.Grenzen{},
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("abgewiesen wurde mit %v", fehler)
			}
		})
	}
}

// **Ein `meta` davor wird übersprungen, nicht gelesen.**
//
// **Und das ist die Probe auf [bisZumFeld]**: der Weg dorthin muss jeden Wert
// **überspringen** können, und `meta` ist der, den MTGJSON wirklich schickt.
func TestEinMetaDavorStoertNicht(t *testing.T) {
	t.Parallel()

	staende, _, fehler := mtgjsonpreise.Lies(
		strings.NewReader(dieGestalt), gewollt(), mtgjsonpreise.Grenzen{})
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	if len(staende) == 0 {
		t.Error("nach dem meta kam nichts")
	}

	// **Und ein kaputtes `meta` hält an** — der Fehlerweg des Überspringens.
	if _, _, fehler := mtgjsonpreise.Lies(
		strings.NewReader(`{"meta": [1, , 2], "data": {}}`), gewollt(),
		mtgjsonpreise.Grenzen{},
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("das kaputte meta kam durch: %v", fehler)
	}
}

// **Ein Strom, der mitten im Wort endet, sagt es.**
//
// **Drei Stellen, drei Fälle** (§3): der Zerteiler kann vor dem Feldnamen
// abbrechen, hinter ihm und mitten unter den Kennungen. Jede davon ist ein
// eigener Weg, und ein Lauf über einen ließe die anderen unbetreten.
func TestEinAbgeschnittenerStromSagtEs(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"vor dem Feldnamen":    `{"me`,
		"hinter dem Feldnamen": `{"data":`,
		"unter den Kennungen":  `{"data": {"aaa`,
	}

	for name, text := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := mtgjsonpreise.Lies(
				strings.NewReader(text), gewollt(), mtgjsonpreise.Grenzen{},
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("abgewiesen wurde mit %v", fehler)
			}
		})
	}
}
