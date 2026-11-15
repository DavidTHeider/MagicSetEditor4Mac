// **Die Stände der bezogenen Pakete — über jede Fläche** — G04
// (#2251).
//
// Dieselbe Auskunft und dieselben Abweisungen am Fenster, an der
// Kommandozeile, über MCP und REST (§16). **Kein Aufruf geht hinaus**: der
// Bezug kommt von der Gegenstelle in `paketbezug_test.go`, und deren Adresse
// ist keine von GitHub — nach einem neueren Stand lässt sich bei ihr nicht
// fragen, und genau diese Abweisung ist der zweite Lauf hier.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// bezogeneUmgebung bezieht das Probearchiv als Baum `bezogen`.
func bezogeneUmgebung(t *testing.T) (umgebung []string, adresse string) {
	t.Helper()

	server := bezugsgegenstelle(t)
	umgebung = mitEigenerSaat(t)
	mussKommandozeile(t, umgebung, "paket-beziehen", server.URL, bezugsstand, "pakete", "bezogen")

	return umgebung, server.URL
}

// **Jede Fläche nennt denselben Baum, ungefragt.**
func TestDieBezuegeStehenAnJederFlaeche(t *testing.T) {
	umgebung, adresse := bezogeneUmgebung(t)
	an := starteDienst(t, umgebung)

	var zeile struct {
		Bezuege []app.PaketbezugDTO `json:"bezuege"`
	}

	if err := json.Unmarshal([]byte(mussKommandozeile(t, umgebung, "paketbezuege", "--json")), &zeile); err != nil {
		t.Fatalf("paketbezuege: %v", err)
	}

	if len(zeile.Bezuege) != 1 || zeile.Bezuege[0].Name != "bezogen" || zeile.Bezuege[0].Stand != bezugsstand ||
		zeile.Bezuege[0].Adresse != adresse || zeile.Bezuege[0].Lage != "ungefragt" {
		t.Fatalf("Kommandozeile: %+v", zeile.Bezuege)
	}

	if text, istFehler := mcpRuf(t, umgebung, "lesen.paket.bezuege", map[string]any{}); istFehler ||
		!strings.Contains(text, "bezogen") || !strings.Contains(text, bezugsstand) {
		t.Errorf("MCP: %s", text)
	}

	if status, rumpf := an.ruf(t, http.MethodGet, "/paketbezuege", ""); status != http.StatusOK ||
		!strings.Contains(rumpf, `"name":"bezogen"`) || !strings.Contains(rumpf, `"lage":"ungefragt"`) {
		t.Errorf("REST: %d %s", status, rumpf)
	}

	if antwort := fensterAus(t, umgebung).PaketBezuege(t.Context(), false); antwort.Fehler != nil ||
		len(antwort.Bezuege) != 1 || antwort.Bezuege[0].Stand != bezugsstand {
		t.Errorf("Fenster: %+v", antwort)
	}
}

// **Nach einem neueren Stand fragen lässt sich nur bei GitHub** — und das
// sagt jede Fläche gleich, statt still „aktuell" zu melden.
func TestEineQuelleOhneAuskunftWirdUeberallGleichAbgewiesen(t *testing.T) {
	umgebung, _ := bezogeneUmgebung(t)
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine Quelle ohne Auskunft", abweisung{
		kern:   "lässt sich kein neuerer Stand erfragen",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "paketbezug-vorschau", "bezogen", bezugsstand)

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "lesen.paket.bezug.vorschau",
				map[string]any{"name": "bezogen", "stand": bezugsstand})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodGet, "/paketbezuege/bezogen/vorschau?stand="+bezugsstand, "")
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).PaketBezugVorschau(t.Context(), "bezogen", bezugsstand).Fehler
		},
	})
}

// **Ein Zweig wird nicht aktualisiert** — an keiner Fläche.
func TestEinZweigWirdNirgendsAktualisiert(t *testing.T) {
	umgebung, _ := bezogeneUmgebung(t)
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Zweig statt eines Commits", abweisung{
		kern:   "ist kein Commit",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "paketbezug-aktualisieren", "bezogen", "main", "--bestaetigt")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.paket.bezug.aktualisieren",
				map[string]any{"name": "bezogen", "stand": "main", "bestaetigt": true})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/paketbezuege/bezogen/aktualisierung", `{"stand":"main","bestaetigt":true}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).PaketBezugAktualisieren(t.Context(), "bezogen", "main", true).Fehler
		},
	})
}

// **Ohne aufgehobenen Stand gibt es kein Zurück** — an keiner Fläche.
func TestOhneAufgehobenenStandGibtEsKeinZurueck(t *testing.T) {
	umgebung, _ := bezogeneUmgebung(t)
	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Zurück ohne aufgehobenen Stand", abweisung{
		kern:   "kein aufgehobener Stand",
		klasse: klasseNichtGefunden,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "paketbezug-zurueck", "bezogen")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.paket.bezug.zurueck", map[string]any{"name": "bezogen"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/paketbezuege/bezogen/zurueck", "")
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).PaketBezugZurueck(t.Context(), "bezogen").Fehler
		},
	})
}
