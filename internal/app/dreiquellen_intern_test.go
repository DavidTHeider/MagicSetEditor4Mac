//go:build fremdpakete

// **Dieselbe Karte aus drei Quellen** —
// #1029,
// Werkstück Ω.7.
//
// # Die Frage
//
// Eine Karte kann auf drei Wegen in den Bestand kommen: aus einem MSE-Paket,
// aus einem Scryfall-Abgleich und von Hand. **Ergeben die drei dasselbe
// Bild?**
//
// Sie sollten. Alle drei schreiben in dieselben Felder derselben Karte
// derselben Vorlage; was danach kommt — Planbau und Rasterung — ist für alle
// dasselbe. Wo sie auseinanderlaufen, liegt es an der **Naht**, über die die
// Werte hereinkommen, und genau die soll dieser Lauf benennen.
//
// # Was er **nicht** prüft
//
// **Nicht, ob die Werte richtig sind.** Ob Scryfalls `oracle_text` in dieses
// Feld gehört, ist eine Frage an die Zuordnung in `internal/scryfall` und
// steht dort. Hier wird gefragt, ob **dieselben Werte** auf drei Wegen
// dasselbe Bild ergeben.
//
// **Und nicht die Kartennummer.** Drei Karten in einem Set haben drei
// Nummern; eine Vorlage, die sie zeichnet, zeichnet drei verschiedene Bilder,
// und das wäre kein Befund. Die Nummer wird deshalb angeglichen.
package app

import (
	"context"
	"image"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// # Was der Lauf am 2026-08-25 gefunden hat
//
// **Aus dem Paket und von Hand: dasselbe Bild.** 18 Felder, 375 × 523
// Bildpunkte, Punkt für Punkt gleich. Die Einfuhr und der Schreibweg des
// Fensters führen zum selben Ergebnis.
//
// **Aus Scryfall: ein anderes**, aus drei benannten Gründen —
//
//   - `rule text` trägt den **heutigen** Orakeltext, nicht den gedruckten
//     („This creature enters tapped." gegen „…enters the battlefield
//     tapped."). Wizards schreibt Regeltexte um; das ist keine Naht, die sich
//     schließen ließe.
//   - `stylesheet`, `stylesheet version`, `styling data`, `has styling` und
//     `border color` setzt der Abgleich nicht — die Karte fällt auf die
//     Vorgaben des Sets zurück. Für eine Karte, die **in** ein Set kommt, ist
//     das richtig.
//   - `notes` und `image` bringt er nicht mit.
//
// **Und einer, der keiner ist**: `casting cost` steht als `B` statt `[B]`.
// Das sah nach einem Fehler in der Zuordnung aus und ist keiner — gemessen
// am 2026-08-25: **beide ergeben dasselbe Bild**, denn das Feld führt
// `script: mana_filter(value)` (`magic.mse-game/card_fields`, Zeile 496), und
// der Filter nimmt beide Schreibweisen. Die gespeicherten Werte
// unterscheiden sich, die gezeichnete Karte nicht.
//
// TestDieselbeKarteAusDreiQuellen baut sie dreimal und vergleicht.
func TestDieselbeKarteAusDreiQuellen(t *testing.T) {
	anwendung := gestartetMit(t, func() string { return fremdpaketwurzelOderSkip(t) })

	geoeffnet := anwendung.SetOeffnen(context.Background(), kanonsatzpfad(t, "Black"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		t.Skip("kein Rasterer")
	}

	// **Die eingeführte Karte ist die Vorlage für die beiden anderen.**
	ausPaket := ersteKarteMitText(t, anwendung, geoeffnet.SetID)
	werte := feldwerteVon(t, anwendung, ausPaket)

	vonHand := ausDenselbenWerten(t, anwendung, geoeffnet.SetID, werte)
	ausScryfall := ueberDenAbgleich(t, anwendung, geoeffnet.SetID, werte)

	bilder := map[string]*image.NRGBA{
		"aus dem Paket": gezeichnet(t, anwendung, rasterer, ausPaket),
		"von Hand":      gezeichnet(t, anwendung, rasterer, vonHand),
		"aus Scryfall":  gezeichnet(t, anwendung, rasterer, ausScryfall),
	}

	vergleicheDrei(t, bilder, werte, anwendung, map[string]string{
		"von Hand": vonHand, "aus Scryfall": ausScryfall, "aus dem Paket": ausPaket,
	})
}

// ersteKarteMitText gibt eine Karte, die Name **und** Regeltext trägt.
//
// **Nicht einfach die erste.** Eine Karte ohne Regeltext ließe den Weg, auf
// dem sich die drei Quellen am ehesten unterscheiden, ungeprüft — und der Lauf
// sähe genauso grün aus.
func ersteKarteMitText(t *testing.T, anwendung *App, setID string) string {
	t.Helper()

	zeilen := anwendung.Kartenzeilen(context.Background(), setID)
	if zeilen.Fehler != nil {
		t.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	for _, zeile := range zeilen.Zeilen {
		werte := feldwerteVon(t, anwendung, zeile.ID)
		// **`rule text` und `super type`, nicht `text` und `type`.** Die
		// Feldnamen einer Karte sind die des Pakets; im **Zeichenplan** heißt
		// die Textebene `text`, und das ist etwas anderes. Der erste Anlauf
		// suchte nach den Plannamen und fand nie eine Karte.
		if werte["name"] != "" && werte["rule text"] != "" && werte["super type"] != "" {
			return zeile.ID
		}
	}

	t.Fatal("keine Karte mit Name, Obertyp und Regeltext")

	return ""
}

// feldwerteVon liest alle Feldwerte einer Karte.
func feldwerteVon(t *testing.T, anwendung *App, karteID string) map[string]string {
	t.Helper()

	antwort := anwendung.Karte(context.Background(), karteID)
	if antwort.Fehler != nil {
		t.Fatalf("Karte %s: %+v", karteID, antwort.Fehler)
	}

	werte := map[string]string{}
	for _, wert := range antwort.Karte.Werte {
		werte[wert.Feld] = wert.Wert
	}

	return werte
}

// vergleicheDrei hält die drei Bilder gegeneinander.
//
// # Zwei verschiedene Zusagen
//
//  1. **Aus dem Paket und von Hand muss dasselbe Bild werden.** Beide
//     schreiben dieselben Werte in dieselben Felder; wird daraus ein anderes
//     Bild, liegt es an der Einfuhr oder am Schreibweg, und das ist ein
//     Fehler. Diese Zusage macht den Lauf rot.
//  2. **Aus Scryfall darf ein anderes werden**, und der Lauf sagt, woran es
//     liegt. Scryfall liefert den **heutigen** Regeltext, nicht den
//     gedruckten, und es setzt Stilbogen, Stildaten und Randfarbe nicht — die
//     Karte fällt auf die Vorgaben des Sets zurück. Beides ist richtig so und
//     wäre als Fehlschlag nur ein Wächter, den jemand abschaltet.
func vergleicheDrei(
	t *testing.T, bilder map[string]*image.NRGBA, werte map[string]string,
	anwendung *App, karten map[string]string,
) {
	t.Helper()

	const (
		ausPaket = "aus dem Paket"
		vonHand  = "von Hand"
	)

	if !gleicheBildpunkte(bilder[ausPaket], bilder[vonHand]) {
		t.Errorf("%q und %q ergeben verschiedene Bilder — "+
			"dieselben Werte auf zwei Wegen sollten dasselbe ergeben",
			ausPaket, vonHand)
		meldeAbweichendeFelder(t, anwendung, karten, ausPaket, vonHand)
	} else {
		t.Logf("%q und %q: dasselbe Bild, %d Felder, %d × %d Bildpunkte",
			ausPaket, vonHand, len(werte),
			bilder[ausPaket].Bounds().Dx(), bilder[ausPaket].Bounds().Dy())
	}

	// **Und die dritte Tür wird beschrieben, nicht bewertet.**
	namen := make([]string, 0, len(bilder))
	for name := range bilder {
		if name != "aus Scryfall" {
			namen = append(namen, name)
		}
	}

	sort.Strings(namen)

	for _, anderer := range namen {
		if gleicheBildpunkte(bilder["aus Scryfall"], bilder[anderer]) {
			t.Logf("%q und %q: dasselbe Bild", "aus Scryfall", anderer)

			continue
		}

		t.Logf("%q und %q laufen auseinander — die Felder dazu:",
			"aus Scryfall", anderer)
		meldeAbweichendeFelder(t, anwendung, karten, "aus Scryfall", anderer)
	}
}

// meldeAbweichendeFelder sagt, welche Feldwerte sich unterscheiden.
//
// **Das ist der eigentliche Ertrag.** „Zwei Bilder sind verschieden" schickt
// jemanden auf die Suche; „das Feld `text` steht bei der einen anders" nennt
// die Naht.
func meldeAbweichendeFelder(
	t *testing.T, anwendung *App, karten map[string]string, einer, anderer string,
) {
	t.Helper()

	dieser := feldwerteVon(t, anwendung, karten[einer])
	jener := feldwerteVon(t, anwendung, karten[anderer])

	felder := map[string]bool{}
	for feld := range dieser {
		felder[feld] = true
	}

	for feld := range jener {
		felder[feld] = true
	}

	namen := make([]string, 0, len(felder))
	for feld := range felder {
		namen = append(namen, feld)
	}

	sort.Strings(namen)

	unterschieden := 0

	for _, feld := range namen {
		if dieser[feld] == jener[feld] {
			continue
		}

		unterschieden++

		t.Logf("  Feld %q: %q ↔ %q", feld,
			gekuerzt(dieser[feld]), gekuerzt(jener[feld]))
	}

	if unterschieden == 0 {
		t.Logf("  **kein Feldwert unterscheidet sich** — der Unterschied " +
			"entsteht also nicht an der Einfuhr, sondern danach")
	}
}

// gekuerzt gibt einen Wert, der in eine Zeile passt.
func gekuerzt(wert string) string {
	const grenze = 48

	einzeilig := strings.ReplaceAll(wert, "\n", "⏎")
	if len(einzeilig) <= grenze {
		return einzeilig
	}

	return einzeilig[:grenze] + "…"
}

// typzeileAus baut Scryfalls `type_line` aus Ober- und Untertyp.
//
// **Mit dem langen Gedankenstrich**, wie Scryfall ihn schreibt: die Zuordnung
// in `internal/scryfall` trennt daran, und ein Bindestrich ergäbe einen
// Obertyp, der den ganzen Satz trägt.
func typzeileAus(werte map[string]string) string {
	if werte["sub type"] == "" {
		return werte["super type"]
	}

	return werte["super type"] + " — " + werte["sub type"]
}
