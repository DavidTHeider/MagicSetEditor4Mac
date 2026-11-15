// **Der Symboleditor über jede Fläche**
// (#1391).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Die Zusage, um die es geht
//
// **Dieselbe Rechnung gibt überall dasselbe Symbol.** Neun Fähigkeiten hatte
// bis zum 2026-09-02 nur das Fenster, mit der Begründung „ohne Zeichenfläche
// gibt es keinen Punkt, auf den man zeigt" — und das beschrieb die Bedienung,
// nicht die Fähigkeit. Die Rechnungen sind reine Funktionen.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Eine Grundform, eine Symmetrie darum — über jede Fläche.**
//
// Der Ablauf ist die kürzeste Kette, die etwas zeigt: aus nichts wird ein
// Rechteck, und aus dem Rechteck eine Symmetrie mit zwei Kopien. Beide
// Schritte sind reine Rechnung; keiner fasst den Bestand an.
func TestEinSymbolRechnenUeberJedeFlaeche(t *testing.T) {
	for name, rechne := range map[string]func(*testing.T, []string, string, string) string{
		"Kommandozeile": func(t *testing.T, umgebung []string, was, argumente string) string {
			t.Helper()

			// **Die Teile kommen über die Standardeingabe**, die Schalter als
			// Argumente. `--roh` gibt es nicht: was hier hineingeht, ist der
			// JSON-Rumpf, den die anderen beiden auch schicken.
			aus, fehler, code := mitEingabeUndCode(t, umgebung, argumente,
				"symbol-rechnen", was, "--json")
			if code != 0 {
				t.Fatalf("symbol-rechnen %s: %d (%s)", was, code, fehler)
			}

			return aus
		},
		"MCP": func(t *testing.T, umgebung []string, was, argumente string) string {
			t.Helper()

			var wahl map[string]any
			if err := json.Unmarshal([]byte(argumente), &wahl); err != nil {
				t.Fatalf("die Argumente lesen: %v", err)
			}

			text, istFehler := mcpRuf(t, umgebung, was, wahl)
			if istFehler {
				t.Fatalf("%s: %s", was, text)
			}

			return text
		},
		"REST": func(t *testing.T, umgebung []string, was, argumente string) string {
			t.Helper()

			an := starteDienst(t, umgebung)

			status, rumpf := an.ruf(t, http.MethodPost,
				"/symbol/"+strings.TrimPrefix(was, "lesen.symbol."), argumente)
			if status != http.StatusOK {
				t.Fatalf("POST: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := eigenerBestand(t)

			// **Die Kommandozeile nennt die Rechnung anders als die anderen
			// beiden** — ihr Argument ist derselbe Name, dem REST sein
			// Präfix abnimmt. Das ist erlaubt: gleich sein muss die
			// Fachlichkeit, nicht die Gestalt.
			rechteck := rechne(t, umgebung, "lesen.symbol.grundform", `{"art":"Rechteck"}`)

			if !strings.Contains(rechteck, "Rechteck") {
				t.Fatalf("keine Grundform: %s", rechteck)
			}

			teile := teileAus(t, rechteck)

			// **Die Punktart, über jede Fläche** (#2436): ein Rechteck hat
			// freie Ecken, und »size« macht die erste symmetrisch.
			mitSperre := rechne(t, umgebung, "lesen.symbol.sperre",
				`{"teile":`+teile+`,"teil":0,"punkt":0,"sperre":"size"}`)
			if !strings.Contains(strings.ReplaceAll(mitSperre, " ", ""), `"sperre":"size"`) {
				t.Errorf("die Punktart kam nicht an: %s", mitSperre)
			}

			mitSymmetrie := rechne(t, umgebung, "lesen.symbol.symmetrie",
				`{"teile":`+teile+`,"stelle":0,"weise":"anlegen"}`)

			for _, erwartet := range []string{"Symmetrie", "rotation"} {
				if !strings.Contains(mitSymmetrie, erwartet) {
					t.Errorf("die Antwort nennt %q nicht: %s", erwartet, mitSymmetrie)
				}
			}
		})
	}
}

// **Und eine Rechnung, die es nicht gibt, wird überall abgewiesen.**
//
// Die Gegenprobe: ohne sie wäre der Lauf darüber auch dann grün, wenn eine
// Fläche jeden Namen annähme und eine leere Antwort gäbe.
func TestEineUnbekannteSymbolrechnungWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := eigenerBestand(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		// **Der Name wird zusammengesetzt und steht nicht als Literal da.**
		// Ein Wächter hält jede Werkzeugkennung in diesen Läufen gegen den
		// Katalog (`flaechenlauf_test.go`); eine erfundene sähe darin aus
		// wie eine Fläche ohne Katalogzeile.
		_, _, code := mitEingabeUndCode(t, umgebung, `{}`,
			"symbol-rechnen", "lesen.symbol."+"gibtsnicht")
		if code == 0 {
			t.Error("der Aufruf ging durch")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		_, istFehler := mcpRuf(t, umgebung, "lesen.symbol.grundform",
			map[string]any{"art": "GibtsNicht"})
		if !istFehler {
			t.Error("eine unbekannte Grundform ging durch")
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, _ := an.ruf(t, http.MethodPost, "/symbol/gibtsnicht", `{}`)
		if status == http.StatusOK {
			t.Error("die Rechnung ging durch")
		}
	})
}

// teileAus holt die Teilliste aus der Antwort einer beliebigen Fläche.
//
// **Drei Gestalten, eine Fachlichkeit.** Die Kommandozeile gibt die rohe
// Liste, MCP legt sie unter »Ergebnis«, der Dienst unter »ergebnis« — das
// dürfen sie (§16 sagt: dieselbe **Antwort**, nicht dieselbe Verpackung).
func teileAus(t *testing.T, roh string) string {
	t.Helper()

	var alsListe json.RawMessage
	if err := json.Unmarshal([]byte(roh), &alsListe); err == nil &&
		strings.HasPrefix(strings.TrimSpace(roh), "[") {
		return string(alsListe)
	}

	var umschlag struct {
		Gross json.RawMessage `json:"Ergebnis"`
		Klein json.RawMessage `json:"ergebnis"`
	}

	if err := json.Unmarshal([]byte(roh), &umschlag); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if len(umschlag.Gross) > 0 {
		return string(umschlag.Gross)
	}

	if len(umschlag.Klein) > 0 {
		return string(umschlag.Klein)
	}

	t.Fatalf("kein Ergebnis in %s", roh)

	return ""
}

// **Und die Palette über jede Fläche.**
//
// Sie hängt an einer Karte und ihrer Vorlage — deshalb braucht dieser Lauf
// ein Set. Geprüft wird die **Auskunft**, nicht ihre Verpackung: die drei
// Flächen benennen die Liste verschieden, und das dürfen sie.
func TestDieSymbolpaletteUeberJedeFlaeche(t *testing.T) {
	for name, hole := range map[string]func(*testing.T, []string, string) string{
		"Kommandozeile": func(t *testing.T, umgebung []string, karteID string) string {
			t.Helper()

			return mussKommandozeile(t, umgebung, "symbolpalette", karteID, "name", "--json")
		},
		"MCP": func(t *testing.T, umgebung []string, karteID string) string {
			t.Helper()

			text, istFehler := mcpRuf(t, umgebung, "lesen.symbolpalette",
				map[string]any{"karteID": zahl(t, karteID), "feld": "name"})
			if istFehler {
				t.Fatalf("lesen.symbolpalette: %s", text)
			}

			return text
		},
		"REST": func(t *testing.T, umgebung []string, karteID string) string {
			t.Helper()

			an := starteDienst(t, umgebung)

			status, rumpf := an.ruf(t, http.MethodGet,
				"/karten/"+karteID+"/symbolpalette/name", "")
			if status != http.StatusOK {
				t.Fatalf("GET: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung, _, karteIDs := setMitKarten(t)

			antwort := hole(t, umgebung, karteIDs[0])

			// **Eine Liste, und sei sie leer.** Das Probeset bringt keine
			// Vorlage mit Palette mit; was hier zählt, ist, dass der Weg
			// trägt und eine Liste zurückgibt — keine `null` (§3).
			if !strings.Contains(antwort, "[") {
				t.Errorf("keine Liste: %s", antwort)
			}
		})
	}
}
