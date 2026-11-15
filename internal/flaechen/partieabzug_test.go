// **Eine laufende Partie behält ihren Abzug** — DOM-03.AK-04
// (#1698),
// der 0.6.0-Nachtest zu G1
// (#1458).
//
// # Der Ablauf
//
// Zwei Decks → eine Partie → das erste Deck bekommt vier Karten mehr und
// einen neuen Namen → die Partie liest sich an jeder Fläche wie vorher.
//
// **Und die Gegenprobe, dass die Änderung wirklich geschah**: eine zweite
// Partie aus demselben Deck trägt die vier Karten mehr. Ohne sie bewiese ein
// gleiches Bild auch ein Deck, das sich gar nicht geändert hat (§10).
package flaechen_test

import (
	"net/http"
	"strconv"
	"testing"
)

func TestEineLaufendePartieBehaeltIhrenAbzug(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	vorher := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json"))

	// **Das Deck ändert sich, während die Partie läuft.**
	mussKommandozeile(t, umgebung, "deck", "legen", ersterDeck, druckkennung,
		strconv.Itoa(deckmenge+4), "hauptdeck")
	mussKommandozeile(t, umgebung, "deck", "umbenennen", ersterDeck, "1", "Aggro, umgebaut")

	nachCLI := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-zeigen", vorher.ID, "anna", "--json"))
	nachMCP := bildAusFlach(t, mussMCP(t, umgebung, "lesen.partie",
		map[string]any{"partieID": vorher.ID, "fuer": "anna"}))
	nachREST := bildAusFlach(t, mussREST(t, an, http.MethodGet,
		"/partien/"+vorher.ID+"?fuer=anna", ""))

	gleichesBild(t, "die Partie nach der Deckänderung, über die Kommandozeile", vorher, nachCLI)
	gleichesBild(t, "die Partie nach der Deckänderung, über MCP", vorher, nachMCP)
	gleichesBild(t, "die Partie nach der Deckänderung, über den Dienst", vorher, nachREST)

	neu := bildAusCLI(t, mussKommandozeile(t, umgebung, "partie-anlegen",
		"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck, "--json"))
	if len(neu.Objekte) != len(vorher.Objekte)+4 {
		t.Errorf("die neue Partie hat %d Objekte, erwartet %d — die Deckänderung kam nicht an",
			len(neu.Objekte), len(vorher.Objekte)+4)
	}
}
