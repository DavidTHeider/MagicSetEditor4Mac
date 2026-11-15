package app_test

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Goldbildvergleich über den ganzen Weg: Stil → Plan → Bild.**
//
// # Warum er hier steht und nicht in `internal/service`
//
// **Die Schichtenregel.** `service` darf `schriften` nicht kennen — das ist die
// Plattformseite, und der Composition Root verdrahtet sie (§2). Der Vergleich
// braucht beides, also steht er dort, wo beides zusammenkommt.
//
// Der Vergleich in `render` hält den **Rasterer** fest: er nimmt einen von Hand
// geschriebenen Plan. Dieser hier hält den Weg fest, den
// #120 zusagt —
// aus einem **echten Stilpaket** der eigenen Saat, samt ausgewerteten
// Feldskripten.
//
// Was hier rot wird, ist eine Änderung am Zusammenspiel: eine andere Deutung
// des Stils, eine andere Ausrichtung, eine andere Schriftgröße.
func TestDieSaatkarteBleibtWieSieIst(t *testing.T) {
	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(probekontext(t), saatkarte(), service.Vorlage{Stil: saatstil(t)})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Mit der Schriftsuche des Rechners**: die Saat nennt `Helvetica`, und
	// eine mitgelieferte Schriftdatei hätte einen Urheber (§11).
	bild, rasterbefunde, err := render.NeuerRasterer(nil).
		MitSchriften(schriften.Neue(schriften.Orte())).
		Zeichne(probekontext(t), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(rasterbefunde) > 0 {
		t.Fatalf("Rasterbefunde: %v", rasterbefunde)
	}

	vergleicheGoldbild(t, bild, "saatkarte-stil.png")
}

// **Derselbe Weg, aber mit Bild und Maske** — die Zusage aus
// #173.
//
// # Warum ein zweites Goldbild und nicht ein erweitertes erstes
//
// Das erste hält die **Textdeutung** fest. Käme der Rahmen dazu, wäre bei jeder
// Abweichung die erste Frage, welche der beiden Deutungen sie verursacht hat —
// und ein Goldbild, das zwei Fragen beantwortet, beantwortet keine.
//
// Das Bild kommt aus der Saat, gezeichnet von
// `werkzeuge.TestDieSaatgrafikIstSelbstGezeichnet`: zwei Rechtecke und eine
// abgerundete Maske, beide aus einer Rechenvorschrift im Repository. Fremde
// Grafiken kommen hier nicht hinein (§11).
//
// **Die Gegenprobe ist gefahren**: ohne `mask:` im Saatstil weicht das Bild an
// 328 Bildpunkten ab — an den vier Ecken. Ohne Bildquelle bleibt der Rahmen
// ganz weg.
func TestDieSaatkarteMitRahmenBleibtWieSieIst(t *testing.T) {
	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(probekontext(t), saatkarte(), service.Vorlage{Stil: saatstilAus(t, "probe-bild.mse-style")})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Mit einer echten Bildquelle über `saat/`.** Ohne sie meldete jede
	// Bildebene, dass keine eingerichtet ist — und das Goldbild zeigte genau
	// den weißen Grund, den es zu widerlegen gilt.
	bild, rasterbefunde, err := render.
		NeuerRasterer(msepaket.NeueBildquelle(filepath.Join("..", "..", "saat"))).
		MitSchriften(schriften.Neue(schriften.Orte())).
		Zeichne(probekontext(t), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(rasterbefunde) > 0 {
		t.Fatalf("Rasterbefunde: %v", rasterbefunde)
	}

	vergleicheGoldbild(t, bild, "saatkarte-rahmen.png")
}

// vergleicheGoldbild hält ein gerastertes Bild gegen sein Goldbild.
func vergleicheGoldbild(t *testing.T, bild image.Image, name string) {
	t.Helper()

	pfad := filepath.Join("..", "..", "testdaten", "goldbilder", name)

	gold, err := os.Open(pfad) // #nosec G304 -- fester Pfad im Repository
	if err != nil {
		// **Fehlt es, wird es angelegt — und der Lauf ist trotzdem rot.**
		// Niemand soll ein Goldbild aus einem kaputten Lauf erben, ohne es
		// gesehen zu haben.
		datei, fehler := os.Create(pfad) // #nosec G304 -- dito
		if fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		if fehler := png.Encode(datei, bild); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}

		if fehler := datei.Close(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}

		t.Fatalf("das Goldbild fehlte und wurde angelegt: %s — bitte ansehen", pfad)
	}

	defer func() { _ = gold.Close() }() //nolint:errcheck // gelesen ist gelesen

	erwartet, err := png.Decode(gold)
	if err != nil {
		t.Fatalf("das Goldbild ist kaputt: %v", err)
	}

	vergleich, err := render.Vergleiche(erwartet, bild)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !vergleich.Innerhalb() {
		t.Errorf("%s hat sich geändert: %+v", name, vergleich)
	}
}

// saatstil öffnet den schlichten Stil der eigenen Saat.
func saatstil(t *testing.T) msepaket.Stil {
	t.Helper()

	return saatstilAus(t, "probe-schlicht.mse-style")
}

// saatstilAus öffnet ein Stilpaket der eigenen Saat.
//
// **Aus `saat/` und nicht aus einem bezogenen Paket** (§11): der Prüflauf muss
// auf jeder Maschine laufen, und fremde Vorlagen liegen nicht im Repository.
func saatstilAus(t *testing.T, paketname string) msepaket.Stil {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "saat", paketname))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	return stil
}

// saatkarte gibt die Karte, die das Goldbild zeigt.
func saatkarte() *domain.CardRow {
	return &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
		{Field: "text", Text: "Füge drei Schaden zu."},
		{Field: "farbe", Text: "rot"},
	}}
}
