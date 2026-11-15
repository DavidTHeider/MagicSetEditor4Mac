package flaechen_test

// **Der Kampfstand steht an jeder Fläche** — CR 508.1, CR 509.1
// (#1623).
//
// # Der Befund, für den es diesen Lauf gibt
//
// **`M6` (#1608)
// hat den Kampf gebaut, und niemand konnte ihn sehen.** Der Zugstand geht
// ungekürzt in die Sicht (`domain.Partiesicht`), aber kein DTO trug ihn
// weiter: wer angreift und wer blockt, stand an **keiner** der vier Flächen.
//
// **Das ist dieselbe Gestalt wie bei Leben und Vergiftungsmarken** (M3,
// 2026-09-17) und wie beim Zugstand selbst (M6): eine Regel, die greift,
// ohne dass jemand sieht, worauf — die erste Dimension aus §10.
//
// **Und es ist mehr als eine Anzeige.** Blocker zu erklären setzt voraus, zu
// wissen, **wer angreift**; ohne den Kampfstand ist CR 509.1 im Fenster
// nicht zu bedienen.

import (
	"net/http"
	"testing"
)

// **Ein erklärter Angriff steht an jeder Fläche.**
func TestDerKampfstandStehtAnJederFlaeche(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	partieID := spielbereitePartie(t, umgebung, ersterDeck, zweiterDeck).ID

	// **Ein Spielstein als Angreifer** (M12): er entsteht mit dem laufenden
	// Zug, und ein Zug weiter darf er angreifen (CR 302.6).
	mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna", "b1",
		"spieler.spielstein",
		`{"kennung":"anna/soldat/0","felder":{"name":"Soldat",`+
			`"super type":"Creature","power":"1","toughness":"1"}}`)

	// **Bis zum Kampf**: der Zug beginnt im Versorgungsschritt, und die
	// Abkürzung spart das Passen (CR 732).
	for range 8 {
		mussKommandozeile(t, umgebung, "partie-befehl", partieID, "anna",
			"a"+string(rune('a'+0)), "zug.abkuerzen", `{}`)
	}

	fuer := map[string]func() partiebild{
		"Kommandozeile": func() partiebild {
			return bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-zeigen", partieID, "anna", "--json"))
		},

		"MCP": func() partiebild {
			return bildAusFlach(t, mussMCP(t, umgebung, "lesen.partie",
				map[string]any{"partieID": partieID, "fuer": "anna"}))
		},

		"REST": func() partiebild {
			return bildAusFlach(t, mussREST(t, an, http.MethodGet,
				"/partien/"+partieID+"?fuer=anna", ""))
		},
	}

	for name, holen := range fuer {
		t.Run(name, func(t *testing.T) {
			// **Die leere Liste ist die Zusage** — sie steht da, auch wenn
			// niemand angreift. Ein `null` ließe jede Schleife darüber
			// umfallen (§8, und derselbe Satz wie bei den Mischpflichten).
			bild := holen()

			if bild.Zug.Angriffe == nil {
				t.Errorf("%s gibt die Angriffe als null statt als Liste", name)
			}

			if bild.Zug.Bloecke == nil {
				t.Errorf("%s gibt die Blöcke als null statt als Liste", name)
			}
		})
	}
}
