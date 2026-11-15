// **Der Modellkatalog über jede Fläche ohne Fenster** — Werkstück A6 von
// Phase 5.3 (#1445).
//
// # Was hier gemessen wird
//
// **Dass alle vier dieselbe Auskunft geben — und dass sie »keines« lautet.**
// Das ist kein schwacher Lauf, sondern genau die Zusage des Issues: „Unbekannt
// bleibt unbekannt; Modelllisten garantieren keine Capability." Käme hier ein
// Modell heraus, wäre es eines, das niemand geprüft hat.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der Katalog ist an allen vier leer — und die Regel steht dabei.**
func TestDerModellkatalogUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "der Modellkatalog",
		[3]string{"nur lokal", `"modelle": []`, `"modelle":[]`},
		func() string {
			return mussKommandozeile(t, umgebung, "ki-modelle")
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.ki.modelle",
				map[string]any{})
			if istFehler {
				t.Fatalf("lesen.ki.modelle: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/ki/modelle", "")
			if status != http.StatusOK {
				t.Fatalf("GET ki/modelle: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Ohne eingeschalteten Anbieter geht kein Ruf hinaus.**
//
// # Hier stand „Der Modellkatalog fragt niemanden", und das war zu viel
//
// **Seit A7 fragt er jeden eingeschalteten Anbieter** — er muss es, denn
// eine belegte Fähigkeit entsteht nur aus einer Antwort
// ([0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)).
// Der Kopf sagte das Gegenteil, und der Rumpf prüfte es gar nicht: er sah
// nach, ob das Wort `nurLokal` in der Antwort steht. Gefunden bei A14
// (#1453),
// zusammen mit derselben Zusage in der MCP-Werkzeugbeschreibung.
//
// **Was bleibt, ist die engere Aussage** — und die ist die wichtige: das
// Lesen ist die Handlung, die jemand tut, **bevor** er sich entscheidet, und
// solange nichts eingeschaltet ist, löst sie nichts aus. Gemessen wird sie
// jetzt an einem Zähler und nicht an einem Wort: die Adresse zeigt auf eine
// Gegenstelle, der Schalter bleibt aus.
//
// **Die Gegenprobe ist gefahren**: mit `ki.laufzeit an` zählt dieselbe
// Gegenstelle drei Anfragen.
func TestDerModellkatalogOhneAnbieterFragtNiemanden(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())
	laufzeit := starteLaufzeit(t, `{}`)

	// **Die Adresse steht, der Schalter nicht.** Ohne Adresse ginge der Ruf
	// an die Vorgabe, und ein Zähler, den niemand erreichen kann, misst
	// nichts (§10).
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.KIAdresse, laufzeit.adresse)

	aus := mussKommandozeile(t, umgebung, "ki-modelle", "--json")

	if !strings.Contains(aus, `"nurLokal"`) {
		t.Errorf("die Antwort nennt die Regel nicht:\n%s", aus)
	}

	if angefasst := laufzeit.anfragen(); angefasst != 0 {
		t.Errorf("%d Anfragen an eine abgeschaltete Laufzeit", angefasst)
	}
}

// **Das Kontextfenster steht an allen drei Flächen** — AI-01.AK-02
// (#1742):
// dieselbe Zahl, die das Fenster zeigt, an Kommandozeile, MCP und Dienst.
func TestDasKontextfensterUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())
	laufzeit := starteLaufzeit(t, `{}`)

	kiEinschalten(t, umgebung, laufzeit.adresse)

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "das Kontextfenster",
		[3]string{"kontextfenster\t4096", `"kontextfenster": 4096`, `"kontextfenster":4096`},
		func() string {
			return mussKommandozeile(t, umgebung, "ki-modelle")
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.ki.modelle", map[string]any{})
			if istFehler {
				t.Fatalf("lesen.ki.modelle: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/ki/modelle", "")
			if status != http.StatusOK {
				t.Fatalf("GET ki/modelle: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Ein Vorschlag ohne eingeschaltete Laufzeit scheitert überall gleich.**
//
// # Was hier gemessen wird
//
// **Derselbe fachliche Fehler an allen drei Flächen** (§16) — und zwar der
// einzige, der sich ohne eine laufende Ollama-Installation messen lässt: die
// Vorgabe von `ki.laufzeit` ist `aus`, und „aus" heißt, dass kein einziger
// Aufruf hinausgeht.
//
// **Was er nicht deckt**, steht hier und nicht in einer Fußnote: ein Lauf
// gegen eine **echte** Laufzeit. Er braucht ein installiertes Ollama, steht
// deshalb hinter dem Tag `netzlauf` und läuft nicht in der Pipeline
// (`internal/ollama/netzlauf_test.go`, §3).
func TestEinKIVorschlagOhneLaufzeitUeberJedeFlaeche(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein KI-Vorschlag ohne Laufzeit", abweisung{
		klasse: klasseValidierung,
		kern:   "eingeschaltet",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).KIVorschlagen(t.Context(), "1", app.KIVorschlagsbezug{
				KarteID: 1, Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "Schreibe den Regeltext.",
			}).Fehler
		},
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"ki-vorschlag", "1", "ollama", "gemma4:8b",
				"Schreibe den Regeltext.")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.ki.vorschlag",
				map[string]any{
					"karteId": 1, "anbieter": "ollama", "alias": "gemma4:8b",
					"auftrag": "Schreibe den Regeltext.",
				})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/karten/1/ki-vorschlag",
				`{"anbieter":"ollama","alias":"gemma4:8b",`+
					`"auftrag":"Schreibe den Regeltext."}`)
		},
	})
}
