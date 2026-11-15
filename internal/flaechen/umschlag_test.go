// **Der portable Umschlag über jede Fläche ohne Fenster** — Werkstück N4 von
// Phase 5.4
// (#1457).
//
// # Was hier gemessen wird
//
// **Dass dieselbe Handlung überall dasselbe tut** (§16) — und der Umschlag
// ist der Fall, bei dem das zählt: was an einer Fläche hinausgeht, kommt an
// einer anderen wieder herein, und eine Fläche, die anders schnürte, wäre
// eine Fläche, deren Umschläge niemand öffnen kann.
//
// # Und warum ein eigener Schlüsselbund
//
// **Weil ein Umschlag unterschrieben wird.** Ohne `MSE_SCHLUESSELBUND` legte
// dieser Lauf eine Identität im Anmeldebund dieses Rechners an — genau der
// Fehler, den `internal/schluesselbund` in seinem Kopf benennt.
package flaechen_test

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Jede Fläche schnürt, und jede liest, was die andere geschnürt hat.**
func TestDenUmschlagUeberJedeFlaeche(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	setID := eingefahren(t, umgebung)

	// **Erst die Identität** — ohne sie wird nichts unterschrieben, und das
	// ist Absicht: ein Umschlag ohne Absender wäre ein Umschlag ohne Zweck.
	mussKommandozeile(t, umgebung, "peer-anlegen")

	an := starteDienst(t, umgebung)

	ordner := t.TempDir()

	for name, schnuere := range map[string]func(string){
		"Kommandozeile": func(pfad string) {
			mussKommandozeile(t, umgebung, "umschlag-schnueren", setID, pfad)
		},
		"MCP": func(pfad string) {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.umschlag.schnueren",
				map[string]any{"setID": setID, "ziel": pfad}); istFehler {
				t.Fatalf("schreiben.umschlag.schnueren: %s", text)
			}
		},
		"REST": func(pfad string) {
			status, rumpf := an.ruf(t, http.MethodPost, "/umschlag",
				`{"setID":"`+setID+`","ziel":"`+pfad+`"}`)
			if status != http.StatusCreated {
				t.Fatalf("POST umschlag: %d (%s)", status, rumpf)
			}
		},
	} {
		pfad := filepath.Join(ordner, strings.ToLower(name)+".msepackage")

		schnuere(pfad)

		// **Angesehen wird über alle drei** — derselbe Umschlag, und jede
		// muss dasselbe darüber sagen. **Der Verbund wird verglichen** und
		// nicht nach einem Wort durchsucht: eine Fläche, die den Urheber
		// wegließe, käme sonst durch.
		ausCLI := mussKommandozeile(t, umgebung, "--json", "umschlag-schauen", pfad)

		ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.umschlag",
			map[string]any{"pfad": pfad})
		if istFehler {
			t.Fatalf("lesen.umschlag: %s", ausMCP)
		}

		status, ausREST := an.ruf(t, http.MethodGet,
			"/umschlag?pfad="+url.QueryEscape(pfad), "")
		if status != http.StatusOK {
			t.Fatalf("GET umschlag: %d (%s)", status, ausREST)
		}

		gleicheAuskunft(t, "der Umschlag von "+name+" — Zeile gegen Modell", ausCLI, ausMCP)
		gleicheAuskunft(t, "der Umschlag von "+name+" — Modell gegen Dienst", ausMCP, ausREST)

		// **Und er sagt das Richtige.** Zwei gleiche Antworten können beide
		// falsch sein; ein Set steht darin, und ein Urheber auch.
		if !strings.Contains(ausCLI, `"art": "set"`) {
			t.Errorf("der Umschlag nennt kein Set:\n%s", ausCLI)
		}

		if !strings.Contains(ausCLI, `"urheber"`) {
			t.Errorf("der Umschlag nennt keinen Urheber:\n%s", ausCLI)
		}
	}
}

// **Übernommen wird an jeder Fläche — und jedes Mal als neues Set.**
//
// **Das ist die Abnahme des Werkstücks**, wörtlich: „Dieselbe freigegebene
// Karte/Set/… ergibt nach bewusster Übernahme fachlich denselben Inhalt und
// neue lokale IDs."
func TestEinUmschlagKommtAnJederFlaecheAlsNeuesSetHerein(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	setID := eingefahren(t, umgebung)

	mussKommandozeile(t, umgebung, "peer-anlegen")

	pfad := filepath.Join(t.TempDir(), "gemeinsam.msepackage")
	mussKommandozeile(t, umgebung, "umschlag-schnueren", setID, pfad)

	an := starteDienst(t, umgebung)

	gesehen := map[string]bool{setID: true}

	for name, uebernimm := range map[string]func() string{
		"Kommandozeile": func() string {
			return setkennungAus(t, mussKommandozeile(t, umgebung,
				"--json", "umschlag-uebernehmen", pfad))
		},
		"MCP": func() string {
			text, istFehler := mcpRuf(t, umgebung, "schreiben.umschlag.uebernehmen",
				map[string]any{"pfad": pfad})
			if istFehler {
				t.Fatalf("schreiben.umschlag.uebernehmen: %s", text)
			}

			return setkennungAus(t, text)
		},
		"REST": func() string {
			status, rumpf := an.ruf(t, http.MethodPost, "/umschlag/uebernahme",
				`{"pfad":"`+pfad+`"}`)
			if status != http.StatusCreated {
				t.Fatalf("POST umschlag/uebernahme: %d (%s)", status, rumpf)
			}

			return setkennungAus(t, rumpf)
		},
	} {
		neu := uebernimm()

		if neu == "" {
			t.Fatalf("%s: die Übernahme nennt keine Kennung", name)
		}

		// **Jede Übernahme legt ein eigenes Set an** — und keine trifft eine
		// Kennung, die es schon gab. Ohne diese Prüfung wäre „neue lokale
		// IDs" eine Behauptung über einen einzigen Ruf.
		if gesehen[neu] {
			t.Errorf("%s: die Kennung %q gab es schon", name, neu)
		}

		gesehen[neu] = true
	}
}

// **Was kein Umschlag ist, wird überall gleich abgewiesen.**
//
// Der Fehlerfall neben dem Erfolg (§16, §10): „wer nur den Erfolgsweg prüft,
// hat die Hälfte ausgelassen."
func TestEineFremdeDateiWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))

	fremd := filepath.Join(t.TempDir(), "fremd.msepackage")
	if fehler := os.WriteFile(fremd, []byte("kein ZIP"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "eine fremde Datei als Umschlag", abweisung{
		kern:   "kein Umschlag",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "umschlag-schauen", fremd)

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "lesen.umschlag", map[string]any{"pfad": fremd})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodGet,
				"/umschlag?pfad="+url.QueryEscape(fremd), "")
		},
		// **Der Ort kommt im Fenster aus dem Dialog** — dieselbe Datei.
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung, fremd).UmschlagWaehlen(t.Context()).Fehler
		},
	})

	vorher := mussKommandozeile(t, umgebung, "sets")

	// **Und der schreibende Weg** — hier stand nur das Ansehen, eine lesende
	// Bindung; die Übernahme, die ein Set anlegt, war an keiner Fläche
	// abgewiesen worden (Zweitblick über Phase C, Befund 3).
	abweisungGleich(t, "eine fremde Datei übernehmen", abweisung{
		kern:   "kein Umschlag",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "umschlag-uebernehmen", fremd)

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.umschlag.uebernehmen", map[string]any{"pfad": fremd})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPost, "/umschlag/uebernahme", `{"pfad":`+alsJSONText(t, fremd)+`}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).UmschlagUebernehmen(t.Context(), fremd).Fehler
		},
	})

	// **Und danach ist kein Set dazugekommen** — an keiner der vier Flächen.
	if nachher := mussKommandozeile(t, umgebung, "sets"); nachher != vorher {
		t.Errorf("eine abgewiesene Übernahme legte ein Set an:\nvorher\n%s\nnachher\n%s", vorher, nachher)
	}
}
