// **Was Scryfall auf eine Frage sagt, über jede Fläche ohne Fenster** —
// Suche, Regelhinweise, Manakosten
// (#1169).
//
// # Warum die drei zusammen stehen
//
// Sie sind die **reinen Auskünfte** an der Gegenstelle: sie ändern nichts, sie
// brauchen kein Set, und sie teilen dieselbe Vorrichtung — eine umgelenkte
// Adresse und ein eigenes Zuhause (§3). Drei Läufe daneben bauten dreimal
// dasselbe auf.
//
// **Kein Aufruf geht hinaus.** Die Gegenstelle steht in
// `scryfallgegenstelle_test.go`; warum das zweimal 25 MB gekostet hat, bevor
// es sie gab, steht dort.
package flaechen_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Eine Suche liefert an jeder Fläche dieselben Karten.**
//
// # Die Zusage
//
// Nicht „dieselbe Gestalt" — die Kommandozeile schreibt eine Tabelle, MCP und
// der Dienst JSON. Die Zusage ist, dass **dieselben Namen** darin stehen: die
// Gegenstelle gibt zwei Karten, und drei Flächen müssen beide nennen.
func TestEineScryfallsucheUeberJedeFlaeche(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	an := starteDienst(t, umgebung)

	for name, hole := range map[string]func() string{
		"Kommandozeile": func() string {
			return mussKommandozeile(t, umgebung, "scryfall-suche", "probe")
		},
		"MCP": func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.scryfall.suche",
				map[string]any{"frage": "probe"})
			if istFehler {
				t.Fatalf("lesen.scryfall.suche: %s", aus)
			}

			return aus
		},
		"REST": func() string {
			status, rumpf := an.ruf(t, http.MethodGet,
				"/scryfall/suche?frage="+url.QueryEscape("probe"), "")
			if status != http.StatusOK {
				t.Fatalf("GET suche: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		antwort := hole()

		for _, karte := range []string{"Erste Probe", "Zweite Probe"} {
			if !strings.Contains(antwort, karte) {
				t.Errorf("%s nennt %q nicht:\n%s", name, karte, antwort)
			}
		}
	}
}

// **Eine Suche ohne Treffer scheitert überall gleich.**
//
// Der Fehlerfall neben dem Erfolg (§16). Die Gegenstelle antwortet auf
// `gibtsnicht` mit 404 — und keine Fläche darf daraus eine leere Trefferliste
// machen, denn das hieße „nichts gefunden" statt „die Frage ging schief".
func TestEineErgebnisloseScryfallsucheScheitertUeberallGleich(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Suche ohne Treffer", abweisung{
		kern: "Scryfall kennt eine Karte zu dieser Suche nicht",
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).ScryfallSuche(t.Context(), "gibtsnicht").Fehler
		},
		klasse: klasseNichtGefunden,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "scryfall-suche", "gibtsnicht")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "lesen.scryfall.suche",
				map[string]any{"frage": "gibtsnicht"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodGet, "/scryfall/suche?frage=gibtsnicht", "")
		},
	})
}

// **Die Regelhinweise zu einer Karte stehen an jeder Fläche.**
//
// Sie sind der Grund, warum eine Karte manchmal anders funktioniert, als ihr
// Text vermuten lässt — und eine Fläche ohne Fenster, die sie nicht kennt,
// beantwortet die Frage „warum" gar nicht.
func TestDieRegelhinweiseUeberJedeFlaeche(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	an := starteDienst(t, umgebung)

	const hinweis = "Ein Hinweis aus der Probe."

	fuerAlleDrei(t, "die Regelhinweise", hinweis,
		func() string {
			return mussKommandozeile(t, umgebung, "scryfall-regeln", "Erste Probe")
		},
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.scryfall.regeln",
				map[string]any{"name": "Erste Probe"})
			if istFehler {
				t.Fatalf("lesen.scryfall.regeln: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet,
				"/scryfall/regeln?name="+url.QueryEscape("Erste Probe"), "")
			if status != http.StatusOK {
				t.Fatalf("GET regeln: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Ein Manakosten wird an jeder Fläche gegen dieselbe Symbolliste geprüft.**
//
// # Die Zusage, und warum sie zwei Fälle braucht
//
// Ein gültiger Kosten gibt **keinen** Befund, ein ungültiger einen. Nur den
// ersten zu fahren hieße prüfen, dass nichts gemeldet wird — und das täte
// auch eine Fläche, die gar nichts prüft. Der zweite Fall ist es, der die
// Zusage trägt.
func TestEinManakostenUeberJedeFlaeche(t *testing.T) {
	server := scryfallgegenstelle(t)
	umgebung := scryfallumgebung(t, server)

	an := starteDienst(t, umgebung)

	// **`{Q}` gibt es in der Symbolliste der Gegenstelle nicht.** Sie führt
	// `{R}` und `{2}`; ein drittes Symbol ist damit der Befund.
	for name, hole := range map[string]func(string) string{
		"Kommandozeile": func(kosten string) string {
			return mussKommandozeile(t, umgebung, "scryfall-mana", kosten)
		},
		"MCP": func(kosten string) string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.scryfall.manakosten",
				map[string]any{"kosten": kosten})
			if istFehler {
				t.Fatalf("lesen.scryfall.manakosten: %s", aus)
			}

			return aus
		},
		"REST": func(kosten string) string {
			status, rumpf := an.ruf(t, http.MethodGet,
				"/scryfall/manakosten?kosten="+url.QueryEscape(kosten), "")
			if status != http.StatusOK {
				t.Fatalf("GET manakosten: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		if antwort := hole("{Q}"); !strings.Contains(antwort, "{Q}") {
			t.Errorf("%s nennt das unbekannte Symbol nicht:\n%s", name, antwort)
		}

		if antwort := hole("{2}{R}"); strings.Contains(antwort, "{Q}") {
			t.Errorf("%s meldet einen Befund auf einen gültigen Kosten:\n%s", name, antwort)
		}
	}
}
