package flaechen_test

// **Dieselbe Abwahl an jeder Fläche** — P1a, Befund C-01 aus dem Audit zur
// Produktionsreife 0.9.0
// (#2335).
//
// # Wogegen hier geprüft wird
//
// **Gegen eine Wahl, die nur eine Fläche treffen kann.** Bis zum 2026-10-13
// schickte das Fenster die Abwahl gar nicht, und Kommandozeile, MCP und
// Dienst kannten keine — die Regelfreigabe nahm sie an jeder Fläche an, der
// Anfang an keiner. §16 verlangt dieselbe Fähigkeit mit derselben Antwort.
//
// # Was hier geprüft wird
//
//   - **Angelegt**: jede der vier Flächen friert `zugstruktur` ein, und die
//     Antwort nennt es.
//   - **Gelesen**: die Partie nennt es danach an jeder Fläche, auch an der,
//     die sie nicht angelegt hat — die Wahl steht im Bestand und nicht nur
//     in der Antwort.
//   - **Abgewiesen**: eine unbekannte Abwahl ist an jeder Fläche derselbe
//     fachliche Fehler, und danach steht keine Partie da.
//
// # Was er nicht hält
//
// **Die wirksame Matrix.** Keine Fläche zeigt sie für eine Partie; dass der
// abgewählte Bereich dort `manuell` steht, prüft
// `internal/app/partieabwahl_intern_test.go` am Bestand.

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// abwahlAus liest die Abwahl aus einer Antwort — flach (MCP, Dienst) oder
// unter `partie.regelwerk` (Kommandozeile, Fenster).
func abwahlAus(t *testing.T, roh string) (kennung string, abgewaehlt []string) {
	t.Helper()

	var gelesen struct {
		ID         string   `json:"id"`
		Abgewaehlt []string `json:"abgewaehlt"`
		Partie     struct {
			ID        string `json:"id"`
			Regelwerk struct {
				Abgewaehlt []string `json:"abgewaehlt"`
			} `json:"regelwerk"`
		} `json:"partie"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if gelesen.Partie.ID != "" {
		return gelesen.Partie.ID, gelesen.Partie.Regelwerk.Abgewaehlt
	}

	return gelesen.ID, gelesen.Abgewaehlt
}

func TestDieAbwahlEntstehtAnJederFlaecheGleich(t *testing.T) {
	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	gewaehlt := []string{"zugstruktur"}

	fenster := fensterAus(t, umgebung)

	angelegt := map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "partie-anlegen",
			"magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck,
			"--ohne=zugstruktur", "--json"),
		"MCP": mussMCP(t, umgebung, "schreiben.partie.anlegen",
			map[string]any{
				"spiel":      "magic",
				"sitze":      sitzwuensche(ersterDeck, zweiterDeck),
				"abgewaehlt": gewaehlt,
			}),
		"Dienst": mussREST(t, an, http.MethodPost, "/partien",
			`{"spiel":"magic","abgewaehlt":["zugstruktur"],"sitze":`+
				string(mussJSON(t, sitzwuensche(ersterDeck, zweiterDeck)))+`}`),
		"Fenster": string(mussJSON(t, fenster.PartieAnlegen(t.Context(),
			app.PartieentwurfDTO{
				Spiel:     "magic",
				Regelwerk: app.RegelwerkDTO{Abgewaehlt: gewaehlt},
				Sitze: []app.SitzwunschDTO{
					{Kennung: "anna", Name: "Anna", DeckID: ersterDeck},
					{Kennung: "bert", Name: "Bert", DeckID: zweiterDeck},
				},
			}))),
	}

	for flaeche, roh := range angelegt {
		kennung, abgewaehlt := abwahlAus(t, roh)

		if !slices.Equal(abgewaehlt, gewaehlt) {
			t.Errorf("%s: angelegt mit %v statt %v:\n%s",
				flaeche, abgewaehlt, gewaehlt, roh)
		}

		// **Und gelesen über eine andere Fläche** — die Wahl steht im
		// Bestand und nicht nur in der Antwort, die sie wiederholt.
		_, gelesen := abwahlAus(t, mussREST(t, an, http.MethodGet,
			"/partien/"+kennung+"?fuer=anna", ""))

		if !slices.Equal(gelesen, gewaehlt) {
			t.Errorf("%s: gelesen am Dienst mit %v statt %v",
				flaeche, gelesen, gewaehlt)
		}
	}
}

// **Eine unbekannte Abwahl ist überall derselbe Fehler — und legt nichts an.**
func TestEineUnbekannteAbwahlWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ersterDeck, zweiterDeck := zweiDecks(t, umgebung)

	abweisungGleich(t, "unbekannte Abwahl", abweisung{
		klasse: klasseValidierung,
		kern:   "kein Regelbereich",
		cli: func() (string, int) {
			aus, fehlerstrom, code := kommandozeile(t, umgebung,
				"partie-anlegen", "magic", "anna:Anna:"+ersterDeck,
				"bert:Bert:"+zweiterDeck, "--ohne=quatsch")

			return aus + fehlerstrom, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.partie.anlegen",
				map[string]any{
					"spiel":      "magic",
					"sitze":      sitzwuensche(ersterDeck, zweiterDeck),
					"abgewaehlt": []string{"quatsch"},
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/partien",
				`{"spiel":"magic","abgewaehlt":["quatsch"],"sitze":`+
					string(mussJSON(t, sitzwuensche(ersterDeck, zweiterDeck)))+`}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).PartieAnlegen(t.Context(),
				app.PartieentwurfDTO{
					Spiel:     "magic",
					Regelwerk: app.RegelwerkDTO{Abgewaehlt: []string{"quatsch"}},
					Sitze: []app.SitzwunschDTO{
						{Kennung: "anna", Name: "Anna", DeckID: ersterDeck},
						{Kennung: "bert", Name: "Bert", DeckID: zweiterDeck},
					},
				}).Fehler
		},
	})

	// **Kein halber Stand**: nach vier Absagen steht keine Partie da.
	var liste []json.RawMessage

	if err := json.Unmarshal([]byte(mussREST(t, an, http.MethodGet,
		"/partien", "")), &liste); err != nil {
		t.Fatalf("die Partienliste ist kein JSON: %v", err)
	}

	if len(liste) != 0 {
		t.Errorf("nach den Absagen stehen %d Partien da", len(liste))
	}
}
