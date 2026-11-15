package flaechen_test

// **Drei am Tisch, und einer geht** — CR 800.4, Werkstück M13
// (#1615).
//
// # Wogegen hier geprüft wird
//
// **Dass die Partie weiterläuft.** Das ist der Unterschied zwischen zwei und
// drei Spielern (CR 800.4), und er ist an einer Fläche zu sehen: nachdem
// einer aufgegeben hat, hat der **übernächste** Priorität und nicht der
// Gegangene.
//
// **Bis zum 2026-09-17 hing die Partie dort.** Die Reihenfolge zählte ihn
// mit, gab ihm Priorität, und niemand konnte sie ihm nehmen.

import (
	"strconv"
	"strings"
	"testing"
)

// **Nach dem Aufgeben überspringt die Priorität den Gegangenen** —
// CR 800.4.
func TestNachDemAufgebenUeberspringtDiePrioritaet(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	partieID := eineDreierpartie(t, umgebung)

	vorher := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	if vorher.Zug.Prioritaet != "anna" {
		t.Fatalf("Priorität hat %q", vorher.Zug.Prioritaet)
	}

	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "bert",
		"b1", "zug.aufgeben", `{}`)

	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
		"b2", "zug.passen", `{}`)

	danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	if danach.Zug.Prioritaet != "carla" {
		t.Errorf("nach anna hat %q Priorität — bert ist gegangen",
			danach.Zug.Prioritaet)
	}
}

// eineDreierpartie legt einen Tisch mit drei Sitzen an und bringt seinen
// Beginn zu Ende — für alle drei (#1851).
func eineDreierpartie(t *testing.T, umgebung []string) string {
	t.Helper()

	menge := strconv.Itoa(deckmenge)

	decks := make([]string, 0, 3)

	for _, name := range []string{"Eins", "Zwei", "Drei"} {
		kennung := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "neu", name, "magic"))

		mussKommandozeile(t, umgebung,
			"deck", "legen", kennung, druckkennung, menge, "hauptdeck")

		decks = append(decks, kennung)
	}

	return beginnAbschliessen(t, umgebung, bildAusCLI(t, mussKommandozeile(t,
		umgebung, "partie-anlegen", "magic",
		"anna:Anna:"+decks[0],
		"bert:Bert:"+decks[1],
		"carla:Carla:"+decks[2], "--json")).ID).ID
}
