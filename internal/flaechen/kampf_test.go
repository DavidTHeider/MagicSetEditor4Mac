package flaechen_test

// **Angreifer werden an jeder Fläche erklärt** — Werkstück M6 der
// Regelmaschine
// (#1608).
//
// # Was dieser Lauf hält
//
// **Die Nutzlast.** `kampf.angreifer` trägt eine **Liste von Paaren** — die
// einzige Gestalt im ganzen Sandkasten, die das ist, und damit die Stelle,
// an der eine Fläche am ehesten anders aussieht als die andere (§16).

import (
	"net/http"
	"testing"
)

// **Dieselbe leere Erklärung, an jeder Fläche** — CR 508.1a.
//
// **Leer und nicht mit Angreifern**: eine Kreatur auf dem Schlachtfeld
// bräuchte in dieser Umgebung eine Karte mit Typzeile, und was dieser Lauf
// hält, ist der **Weg** und nicht die Regel. Dass eine Erklärung ohne
// Angreifer eine Erklärung ist, steht in CR 508.1a — und genau das prüft er.
func TestEineErklaerungGehtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	fuer := map[string]func(partieID string) string{
		"Kommandozeile": func(partieID string) string {
			return mussKommandozeile(t, umgebung, "partie-befehl", partieID,
				"anna", "b1", "kampf.angreifer", `{"angriffe":[]}`)
		},

		"MCP": func(partieID string) string {
			return mussMCP(t, umgebung, "schreiben.partie.befehl",
				map[string]any{
					"partieID": partieID, "absender": "anna",
					"kennung": "b1", "art": "kampf.angreifer",
					"nutzlast": map[string]any{"angriffe": []any{}},
				})
		},

		"REST": func(partieID string) string {
			return mussREST(t, an, http.MethodPost,
				"/partien/"+partieID+"/befehle",
				`{"absender":"anna","kennung":"b1","art":"kampf.angreifer",`+
					`"nutzlast":{"angriffe":[]}}`)
		},
	}

	for name, stelle := range fuer {
		t.Run(name, func(t *testing.T) {
			partieID := spielbereitePartie(t, umgebung, ersterDeck,
				zweiterDeck).ID

			bisZumAngreiferschritt(t, umgebung, partieID)

			stelle(partieID)

			// **Der Schritt geht weiter** — wer erklärt hat, passt, und
			// danach steht der Blockerschritt (CR 508.2).
			mussKommandozeile(t, umgebung, "partie-befehl", partieID,
				"anna", "p1", "zug.passen", `{}`)
			mussKommandozeile(t, umgebung, "partie-befehl", partieID,
				"bert", "p2", "zug.passen", `{}`)

			danach := bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, "anna", "--json"))

			if danach.Zug.Schritt != "blocker" {
				t.Fatalf("der Zug steht bei %q — diese Fläche hat die "+
					"Erklärung nicht angenommen", danach.Zug.Schritt)
			}
		})
	}
}

// bisZumAngreiferschritt passt, bis der Angreiferschritt dasteht.
func bisZumAngreiferschritt(t *testing.T, umgebung []string, partieID string) {
	t.Helper()

	for lauf := range 24 {
		bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
			"partie-zeigen", partieID, "anna", "--json"))

		if bild.Zug.Schritt == "angreifer" {
			return
		}

		mussKommandozeile(t, umgebung, "partie-befehl", partieID,
			bild.Zug.Prioritaet, "v"+string(rune('a'+lauf)),
			"zug.passen", `{}`)
	}

	t.Fatal("der Angreiferschritt kam nicht")
}
