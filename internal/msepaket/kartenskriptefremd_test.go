//go:build fremdpakete

package msepaket_test

// **Die Kartenskripte des echten Spielpakets** —
// #1337.
//
// **Hinter `fremdpakete`** (§3, §11): `magic.mse-game` liegt nicht im
// Repositorium, und dieser Lauf zählt nicht zur Abdeckung. Er steht hier,
// weil eine selbst geschriebene Probe nicht sagt, ob der Leser die **echte**
// Datei trifft — und `magic.mse-game` bindet seine Skripte über
// `include file:` ein.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Zwölf Skripte, und ihre Namen tragen kein `&` mehr.**
//
// Gezählt am 2026-08-31: `add_cards_scripts` führt dreizehn Blöcke, einer
// davon auskommentiert — also zwölf.
func TestDieKartenskripteDesEchtenSpielpakets(t *testing.T) {
	paket, fehler := msepaket.Oeffne(
		"/Users/david/full-magic-pack/data/magic.mse-game")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	// **Erst einbinden.** `magic.mse-game` legt seine Skripte in
	// `add_cards_scripts` und holt die Datei über `include file:`. Wer den
	// Kopf ohne Einbinden liest, findet keinen einzigen Block — genau das war
	// der erste Lauf, und er war rot mit `0 Skripte`.
	// **Mit der Paketwurzel.** Der Einbinder sucht eine eingebundene Datei
	// über die Wurzeln, nicht im geöffneten Paket — ohne sie meldet er acht
	// Befunde der Art „die Datei ist in keinem Paket zu finden", und der
	// Kopf bleibt leer. Das war der zweite rote Lauf.
	for _, befund := range paket.ModuleEinbinden(
		msepaket.NeuerPaketleser("/Users/david/full-magic-pack/data")) {
		t.Logf("Einbindebefund: %s", befund.String())
	}

	skripte := paket.Kartenskripte()

	if len(skripte) != 12 {
		namen := make([]string, 0, len(skripte))
		for _, skript := range skripte {
			namen = append(namen, skript.Name)
		}

		t.Fatalf("%d Skripte, erwartet waren zwölf: %v", len(skripte), namen)
	}

	for _, skript := range skripte {
		if strings.Contains(skript.Name, "&") {
			t.Errorf("der Name %q trägt noch ein Kaufmanns-Und", skript.Name)
		}

		if skript.Name == "" {
			t.Error("ein Skript ohne Namen")
		}

		// **Jedes trägt ein Skript.** Ein Eintrag im Menü, der nichts tut,
		// wäre schlimmer als keiner.
		if strings.TrimSpace(skript.Skript) == "" {
			t.Errorf("das Skript %q ist leer", skript.Name)
		}
	}

	// **Und das erste ist das, was dort steht.** Ohne diese Zusicherung wäre
	// der Lauf auch dann grün, wenn der Leser zwölf leere Hüllen gäbe.
	if skripte[0].Name != "Basic Lands" || skripte[0].Kuerzel != "B" {
		t.Errorf("das erste Skript lautet %+v", skripte[0])
	}

	if !strings.Contains(skripte[0].Skript, "new_card") {
		t.Errorf("das erste Skript ruft kein new_card: %q", skripte[0].Skript)
	}
}
