// **Die Symbolbibliothek an jeder Fläche** — DOM-02
// (#1697).
//
// # Der Ablauf
//
// Eine Zeichnung einlesen → in ein Set einsetzen → aus dem Set wieder
// aufnehmen → duplizieren und umbenennen → **alle drei Einträge tragen
// dasselbe Set**, denn sie haben denselben Abdruck → Löschen ohne
// Bestätigung nennt das Set, mit Bestätigung geht es. **Dieselbe Antwort an jeder Fläche** (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// bibliothekswege sind die Schritte des Ablaufs, je Fläche einmal geschrieben.
type bibliothekswege struct {
	einlesen   func(inhalt, name string) string
	einsetzen  func(id, setID string)
	aufnehmen  func(setID, name string) string
	kopieren   func(id, name string) string
	umbenennen func(id, name string)
	liste      func() string
	text       func(id string) string

	// loeschen gibt, ob es ging, und was die Fläche dazu sagte.
	loeschen func(id string, bestaetigt bool) (gut bool, gesagt string)
}

// pruefeBibliotheksablauf fährt den Ablauf über eine Fläche.
func pruefeBibliotheksablauf(t *testing.T, flaeche, setID string, wege bibliothekswege) {
	t.Helper()

	inhalt := string(msepaket.SchreibeSymbolzeichnung(msepaket.Standardsymbol()))

	krone := wege.einlesen(inhalt, "Krone")
	wege.einsetzen(krone, setID)
	reif := wege.aufnehmen(setID, "Reif")
	wege.umbenennen(wege.kopieren(reif, "Zacke"), "Zinne")

	var symbole []struct {
		Name      string   `json:"name"`
		Verwendet []string `json:"verwendet"`
	}

	if err := json.Unmarshal([]byte(wege.liste()), &symbole); err != nil {
		t.Fatalf("%s: die Liste: %v", flaeche, err)
	}

	// **Alle drei tragen dasselbe Set** — sie haben denselben Abdruck, und
	// die Kopie heißt jetzt anders.
	if len(symbole) != 3 || symbole[2].Name != "Zinne" {
		t.Errorf("%s: die Liste: %+v", flaeche, symbole)
	}

	for _, symbol := range symbole {
		if !slices.Equal(symbol.Verwendet, []string{"Kronenset"}) {
			t.Errorf("%s: %s trägt %v", flaeche, symbol.Name, symbol.Verwendet)
		}
	}

	if text := wege.text(reif); !strings.Contains(text, "mse_version") {
		t.Errorf("%s: als Text: %q", flaeche, text)
	}

	if gut, gesagt := wege.loeschen(krone, false); gut || !strings.Contains(gesagt, "Kronenset") {
		t.Errorf("%s: ohne Bestätigung: %v %q", flaeche, gut, gesagt)
	}

	if gut, gesagt := wege.loeschen(krone, true); !gut || !strings.Contains(gesagt, "Kronenset") {
		t.Errorf("%s: mit Bestätigung: %v %q", flaeche, gut, gesagt)
	}
}

func TestDieSymbolbibliothekReistUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := neuesSet(t, umgebung, "Kronenset")

		pruefeBibliotheksablauf(t, "Kommandozeile", setID, bibliothekswege{
			einlesen: func(inhalt, name string) string {
				datei := filepath.Join(t.TempDir(), "zeichnung.mse-symbol")
				if err := os.WriteFile(datei, []byte(inhalt), 0o600); err != nil {
					t.Fatalf("schreiben: %v", err)
				}

				return strings.TrimSpace(mussKommandozeile(t, umgebung, "symbolbibliothek", "einlesen", datei, name))
			},
			einsetzen: func(id, set string) {
				mussKommandozeile(t, umgebung, "symbolbibliothek", "einsetzen", id, set)
			},
			aufnehmen: func(set, name string) string {
				return strings.TrimSpace(mussKommandozeile(t, umgebung, "symbolbibliothek", "aufnehmen", set, name))
			},
			kopieren: func(id, name string) string {
				return strings.TrimSpace(mussKommandozeile(t, umgebung, "symbolbibliothek", "duplizieren", id, name))
			},
			umbenennen: func(id, name string) {
				mussKommandozeile(t, umgebung, "symbolbibliothek", "umbenennen", id, name)
			},
			liste: func() string { return mussKommandozeile(t, umgebung, "symbolbibliothek", "--json") },
			text:  func(id string) string { return mussKommandozeile(t, umgebung, "symbolbibliothek", "ausgeben", id) },
			loeschen: func(id string, bestaetigt bool) (bool, string) {
				argumente := []string{"symbolbibliothek", "loeschen", id}
				if bestaetigt {
					argumente = append(argumente, "--bestaetigt")
				}

				aus, fehler, code := kommandozeile(t, umgebung, argumente...)

				return code == 0, aus + fehler
			},
		})
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := neuesSet(t, umgebung, "Kronenset")

		pruefeBibliotheksablauf(t, "MCP", setID, bibliothekswege{
			einlesen: func(inhalt, name string) string {
				return ausDemFeld(t, mussMCP(t, umgebung, "schreiben.symbolbibliothek.einlesen",
					map[string]any{"inhalt": inhalt, "name": name}), "id")
			},
			einsetzen: func(id, set string) {
				mussMCP(t, umgebung, "schreiben.symbolbibliothek.einsetzen",
					map[string]any{"id": zahlAus(t, id), "setID": zahlAus(t, set)})
			},
			aufnehmen: func(set, name string) string {
				return ausDemFeld(t, mussMCP(t, umgebung, "schreiben.symbolbibliothek.aufnehmen",
					map[string]any{"setID": zahlAus(t, set), "name": name}), "id")
			},
			kopieren: func(id, name string) string {
				return ausDemFeld(t, mussMCP(t, umgebung, "schreiben.symbolbibliothek.duplizieren",
					map[string]any{"id": zahlAus(t, id), "name": name}), "id")
			},
			umbenennen: func(id, name string) {
				mussMCP(t, umgebung, "schreiben.symbolbibliothek.umbenennen",
					map[string]any{"id": zahlAus(t, id), "name": name})
			},
			liste: func() string { return mussMCP(t, umgebung, "lesen.symbolbibliothek", map[string]any{}) },
			text: func(id string) string {
				return ausDemFeld(t, mussMCP(t, umgebung, "lesen.symbolbibliothek.symbol",
					map[string]any{"id": zahlAus(t, id)}), "inhalt")
			},
			loeschen: func(id string, bestaetigt bool) (bool, string) {
				text, istFehler := mcpRuf(t, umgebung, "schreiben.symbolbibliothek.loeschen",
					map[string]any{"id": zahlAus(t, id), "bestaetigt": bestaetigt})

				return !istFehler, text
			},
		})
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := neuesSet(t, umgebung, "Kronenset")
		an := starteDienst(t, umgebung)

		pruefeBibliotheksablauf(t, "REST", setID, bibliothekswege{
			einlesen: func(inhalt, name string) string {
				rumpf, _ := json.Marshal(map[string]string{"inhalt": inhalt, "name": name}) //nolint:errcheck // Text geht immer

				return ausDemFeld(t, mussREST(t, an, http.MethodPost, "/symbole/einlesen", string(rumpf)), "id")
			},
			einsetzen: func(id, set string) {
				mussREST(t, an, http.MethodPut, "/sets/"+set+"/bibliothekssymbol", `{"id":`+id+`}`)
			},
			aufnehmen: func(set, name string) string {
				return ausDemFeld(t, mussREST(t, an, http.MethodPost, "/symbole",
					`{"setID":`+set+`,"name":"`+name+`"}`), "id")
			},
			kopieren: func(id, name string) string {
				return ausDemFeld(t, mussREST(t, an, http.MethodPost, "/symbole/"+id+"/kopie",
					`{"name":"`+name+`"}`), "id")
			},
			umbenennen: func(id, name string) {
				mussREST(t, an, http.MethodPut, "/symbole/"+id+"/name", `{"name":"`+name+`"}`)
			},
			liste: func() string { return mussREST(t, an, http.MethodGet, "/symbole", "") },
			text: func(id string) string {
				return ausDemFeld(t, mussREST(t, an, http.MethodGet, "/symbole/"+id, ""), "inhalt")
			},
			loeschen: func(id string, bestaetigt bool) (bool, string) {
				status, antwort := an.ruf(t, http.MethodDelete,
					"/symbole/"+id+"?bestaetigt="+strconv.FormatBool(bestaetigt), "")

				return status == http.StatusOK, antwort
			},
		})
	})
}

// zahlAus liest eine Kennung als Zahl — MCP nimmt sie so.
func zahlAus(t *testing.T, kennung string) int64 {
	t.Helper()

	zahl, err := strconv.ParseInt(kennung, 10, 64)
	if err != nil {
		t.Fatalf("die Kennung %q: %v", kennung, err)
	}

	return zahl
}
