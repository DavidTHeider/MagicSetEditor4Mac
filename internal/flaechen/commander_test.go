package flaechen_test

// **Die Commandervariante an jeder Fläche** — CR 903, Werkstück M19
// (#1620).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Commanderpartie überhaupt entsteht.** Ein Profil, das niemand
// wählen kann, ist ein Paket ohne Aufrufer (§10) — und genau das war es bis
// zum 2026-09-17: `Commanderprofil` stand da, und `Partieaufbau` nahm immer
// das Grundprofil.
//
// **Und dass sie mit vierzig Leben anfängt** (CR 903.7). Das ist die
// sichtbare Zusage; die Steuer und der Commanderschaden sind im Kern
// gemessen.

import (
	"net/http"
	"testing"
)

// **Eine Commanderpartie fängt an jeder Fläche mit vierzig Leben an** —
// CR 903.7.
func TestEineCommanderpartieFaengtAnJederFlaecheMitVierzigAn(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	fuer := map[string]func() partiebild{
		"Kommandozeile": func() partiebild {
			return bildAusCLI(t, mussKommandozeile(t, umgebung,
				"partie-anlegen", "magic",
				"anna:Anna:"+ersterDeck+":held",
				"bert:Bert:"+zweiterDeck+":held", "--json"))
		},

		"MCP": func() partiebild {
			return bildAusFlach(t, mussMCP(t, umgebung,
				"schreiben.partie.anlegen", map[string]any{
					"spiel": "magic",
					"sitze": []map[string]string{
						{
							"kennung": "anna", "name": "Anna",
							"deckID": ersterDeck, "commander": "held",
						},
						{
							"kennung": "bert", "name": "Bert",
							"deckID": zweiterDeck, "commander": "held",
						},
					},
				}))
		},

		"REST": func() partiebild {
			return bildAusFlach(t, mussREST(t, an, http.MethodPost,
				"/partien",
				`{"spiel":"magic","sitze":[`+
					`{"kennung":"anna","name":"Anna","deckID":"`+
					ersterDeck+`","commander":"held"},`+
					`{"kennung":"bert","name":"Bert","deckID":"`+
					zweiterDeck+`","commander":"held"}]}`))
		},
	}

	for name, anlegen := range fuer {
		t.Run(name, func(t *testing.T) {
			bild := anlegen()

			for _, sitz := range bild.Sitze {
				if sitz.Leben != 40 {
					t.Errorf("%s: %q fängt mit %d Leben an",
						name, sitz.Kennung, sitz.Leben)
				}
			}
		})
	}
}

// **Und ohne Commander bleibt es bei zwanzig** — die Gegenprobe.
//
// **Ohne sie sagte der Lauf darüber nichts**: eine Fläche, die immer
// vierzig gäbe, bestünde ihn ebenso.
func TestOhneCommanderBleibtEsBeiZwanzig(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	bild := bildAusCLI(t, mussKommandozeile(t, umgebung,
		"partie-anlegen", "magic", "anna:Anna:"+ersterDeck,
		"bert:Bert:"+zweiterDeck, "--json"))

	for _, sitz := range bild.Sitze {
		if sitz.Leben != 20 {
			t.Errorf("%q fängt mit %d Leben an", sitz.Kennung, sitz.Leben)
		}
	}
}
