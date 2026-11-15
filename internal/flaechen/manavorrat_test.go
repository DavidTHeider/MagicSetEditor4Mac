package flaechen_test

// **Der Manavorrat steht an jeder Fläche** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Warum dieser Lauf einzeln steht
//
// **Weil ein Wächter zählt, dass eine Fähigkeit da ist, und nicht, dass sie
// dasselbe tut** (§16). Hier wird es getan: an jeder der drei Flächen ohne
// Fenster bekommt derselbe Sitz dasselbe Mana, und jede nennt denselben
// Vorrat.
//
// **Und er prüft die zweite Hälfte gleich mit**: der Vorrat leert sich am
// Ende des Abschnitts (CR 500.5) — an jeder Fläche.

import (
	"net/http"
	"testing"
)

// **Mana kommt an, und jede Fläche nennt es gleich.**
func TestDerManavorratStehtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	fuer := map[string]func(partieID string){
		"Kommandozeile": func(partieID string) {
			mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
				"b1", "spieler.mana", `{"art":"R","um":2}`)
		},

		"MCP": func(partieID string) {
			mussMCP(t, umgebung, "schreiben.partie.befehl", map[string]any{
				"partieID": partieID, "absender": "anna", "kennung": "b1",
				"art":      "spieler.mana",
				"nutzlast": map[string]any{"art": "R", "um": 2},
			})
		},

		"REST": func(partieID string) {
			mussREST(t, an, http.MethodPost, "/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"b1","art":"spieler.mana",`+
					`"nutzlast":{"art":"R","um":2}}`)
		},
	}

	for name, stelle := range fuer {
		t.Run(name, func(t *testing.T) {
			partieID := spielbereitePartie(t, umgebung, ersterDeck,
				zweiterDeck).ID

			stelle(partieID)

			danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, "anna", "--json"))

			if danach.Sitze[0].Manavorrat != "RR" {
				t.Fatalf("der Vorrat steht bei %q — diese Fläche legt kein "+
					"Mana in den Vorrat", danach.Sitze[0].Manavorrat)
			}
		})
	}
}

// **Und der Abschnitt leert ihn** — CR 500.5, an jeder Fläche derselbe Weg.
//
// **Die Gegenprobe zum Lauf darüber** (§10): ohne sie sagte er nur, dass
// Mana **ankommt** — ein Vorrat, der nie verginge, käme genauso durch, und
// dann wäre jedes Land unbegrenzt.
func TestDerManavorratVergehtMitDemAbschnitt(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
		"b1", "spieler.mana", `{"art":"G","um":1}`)

	// **Beide passen** — damit endet der Abschnitt (CR 500.2), und der
	// Vorrat geht mit ihm.
	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
		"b2", "zug.passen", `{}`)
	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "bert",
		"b3", "zug.passen", `{}`)

	danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", partieID, "anna", "--json"))

	if danach.Sitze[0].Manavorrat != "" {
		t.Errorf("der Vorrat steht nach dem Abschnitt bei %q",
			danach.Sitze[0].Manavorrat)
	}
}
