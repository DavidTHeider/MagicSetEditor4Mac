// **Die Rolle über weitere Schreibklassen** — C04
// (#2229).
//
// Die Rollenmatrix daneben fährt je Wirkung **einen** Weg: ein Deck, eine
// Datei, eine Suche. Für die Klassen, an denen fremde Arbeit am meisten
// verliert — ein Kartenfeld, ein Exemplar, ein Turnier —, stand an keiner
// Fläche ohne Fenster ein Lauf mit der Rolle. Hier gehen sie in denselben
// vier Lagen über Kommandozeile, MCP und Dienst, **an der Wirkung** geprüft:
// danach zählt das Profil der Lage genau, was entstehen durfte.
//
// **Was er nicht fährt**: das Fenster — es fährt die Lagen in `internal/app`
// (die Matrix daneben sagt, wo); und jede andere Schreibklasse — die
// Zuordnung jeder Fähigkeit zu ihrer Klasse hält `AUTORISIERUNG.md`.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestWeitereSchreibklassenFolgenDerRolle(t *testing.T) {
	t.Parallel()

	for _, lage := range []rollenlage{
		{"eigenes Profil", "", false, true},
		{"fremd als Mitarbeiter", "mitarbeiter", true, true},
		{"fremd als Leser", "leser", true, false},
		{"fremd ohne Rolle", "", true, false},
	} {
		t.Run(lage.name, func(t *testing.T) {
			t.Parallel()

			umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
			setID := richteLageEin(t, umgebung, lage)
			karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
			mcp := starteMCPProzess(t, umgebung)
			an := starteDienst(t, umgebung)

			kommandozeileGelang := func(argumente ...string) bool {
				_, _, code := kommandozeile(t, umgebung, argumente...)

				return code == 0
			}
			mcpGelang := func(werkzeug string, argumente map[string]any) bool {
				_, istFehler := mcp.ruf(werkzeug, argumente)

				return !istFehler
			}
			dienstGelang := func(methode, weg, rumpf string) bool {
				status, _ := an.ruf(t, methode, weg, rumpf)

				return status >= http.StatusOK && status < http.StatusMultipleChoices
			}

			exemplar := func(druck string) map[string]any {
				return map[string]any{"printing": druck, "menge": 1, "zustand": "played", "sprache": "de", "finish": "nonfoil"}
			}

			// **Der Wert davor** (C04-Rest, #2284): ein Leser muss ihn stehen
			// lassen — nicht nur das „moos" des Dienstes nicht schreiben.
			farbweltDavor := einstellungswert(t, mussKommandozeile(t, umgebung, "einstellungen", "--json"), "farbwelt")

			for wirkung, gelang := range map[string]map[string]bool{
				"ein Kartenfeld": {
					"Kommandozeile": kommandozeileGelang("feld", setID, karteID, "name", "Aus der Zeile"),
					"MCP": mcpGelang("schreiben.feld", map[string]any{
						"setID": json.Number(setID), "karteID": json.Number(karteID), "feld": "name", "wert": "Aus dem Modell",
					}),
					"Dienst": dienstGelang(http.MethodPut, "/sets/"+setID+"/karten/"+karteID+"/felder/name", `{"wert":"Aus dem Dienst"}`),
				},
				"ein Exemplar": {
					"Kommandozeile": kommandozeileGelang("exemplar", "00000000-0000-4000-8000-00000000c401", "1", "played", "de", "nonfoil"),
					"MCP":           mcpGelang("schreiben.exemplar", exemplar("00000000-0000-4000-8000-00000000c402")),
					"Dienst":        dienstGelang(http.MethodPost, "/sammlung", string(mussJSON(t, exemplar("00000000-0000-4000-8000-00000000c403")))),
				},
				"ein Turnier": {
					"Kommandozeile": kommandozeileGelang("turnier", "neu", "Aus der Zeile", "magic", ""),
					"MCP":           mcpGelang("schreiben.turnier.anlegen", map[string]any{"name": "Aus dem Modell", "spiel": "magic"}),
					"Dienst":        dienstGelang(http.MethodPost, "/turniere", `{"name":"Aus dem Dienst","spiel":"magic"}`),
				},
				"ein Format": {
					"Kommandozeile": kommandozeileGelang("format", "neu", "commander", "Aus der Zeile", "magic", "commander"),
					"MCP": mcpGelang("schreiben.format.ausbauart", map[string]any{
						"bauart": "commander", "name": "Aus dem Modell", "spiel": "magic", "scryfallname": "commander",
					}),
					"Dienst": dienstGelang(http.MethodPost, "/formate",
						`{"bauart":"commander","name":"Aus dem Dienst","spiel":"magic","scryfallname":"commander"}`),
				},
				"ein Set": {
					"Kommandozeile": kommandozeileGelang("set-neu", "Aus der Zeile", "probe", "probe-schlicht"),
					"MCP": mcpGelang("schreiben.set.neu", map[string]any{
						"name": "Aus dem Modell", "spiel": "probe", "stil": "probe-schlicht",
					}),
					"Dienst": dienstGelang(http.MethodPost, "/sets/neu", `{"name":"Aus dem Dienst","spiel":"probe","stil":"probe-schlicht"}`),
				},
				"ein Schlüsselwort": {
					"Kommandozeile": kommandozeileGelang("schluesselwort", setID, "anlegen"),
					"MCP":           mcpGelang("schreiben.schluesselwort.anlegen", map[string]any{"setID": json.Number(setID)}),
					"Dienst":        dienstGelang(http.MethodPost, "/sets/"+setID+"/schluesselwoerter", ""),
				},
				"eine Autoersetzung": {
					"Kommandozeile": kommandozeileGelang("autoersetzung", setID, "setzen", "(Z)", "zeile"),
					"MCP": mcpGelang("schreiben.autoersetzung", map[string]any{
						"setID": json.Number(setID), "muster": "(M)", "ersatz": "modell", "ganzesWort": true, "aktiv": true,
					}),
					"Dienst": dienstGelang(http.MethodPut, "/sets/"+setID+"/autoersetzungen/(D)",
						`{"ersatz":"dienst","ganzesWort":true,"aktiv":true}`),
				},
				// **Eine Einstellung des Profils** — bis C04 nur an der
				// Kommandozeile gefahren (`nurlesen_test.go`; Zweitblick über
				// Phase C, Befund 3). Jede Fläche setzt einen anderen Wert,
				// damit die Wirkung zeigt, wer geschrieben hat.
				"eine Einstellung": {
					"Kommandozeile": kommandozeileGelang("einstellungen", "setzen", "farbwelt", "klar"),
					"MCP":           mcpGelang("schreiben.einstellung", map[string]any{"schluessel": "farbwelt", "wert": "schiefer"}),
					"Dienst":        dienstGelang(http.MethodPut, "/einstellungen/farbwelt", `{"wert":"moos"}`),
				},
				// **Eine Rücknahme schreibt**, auch wenn nichts zurückzunehmen ist:
				// sie geht durch dieselbe Wache wie jede Änderung am Set.
				"eine Rücknahme": {
					"Kommandozeile": kommandozeileGelang("zuruecknehmen", setID),
					"MCP":           mcpGelang("schreiben.zuruecknehmen", map[string]any{"setID": json.Number(setID)}),
					"Dienst":        dienstGelang(http.MethodPost, "/sets/"+setID+"/zuruecknehmen", ""),
				},
			} {
				for flaeche, ja := range gelang {
					if ja != lage.darf {
						t.Errorf("%s über %s ging %v, erwartet %v", wirkung, flaeche, ja, lage.darf)
					}
				}
			}

			// **An der Wirkung**: im Profil der Lage steht genau, was durfte.
			karten := mussKommandozeile(t, umgebung, "karten", setID, "--json")
			if geschrieben := strings.Contains(karten, "Aus de"); geschrieben != lage.darf {
				t.Errorf("ein Kartenfeld wurde geschrieben: %v, erwartet %v", geschrieben, lage.darf)
			}

			if zahl := len(alleAusHuelle(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))); zahl != map[bool]int{true: 3, false: 0}[lage.darf] {
				t.Errorf("%d Exemplare im Profil der Lage", zahl)
			}

			// **Die Einstellung**: die Elemente der Tabelle werden der Reihe
			// nach gerufen, der Dienst zuletzt — wer schreiben darf, hinterlässt
			// seinen Wert.
			// **Und wer nicht darf, hinterlässt gar nichts** — auch kein
			// „klar" oder „schiefer" neben einem zurückgegebenen Fehler
			// (C04-Rest, #2284).
			erwartet := map[bool]string{true: "moos", false: farbweltDavor}[lage.darf]
			if farbwelt := einstellungswert(t, mussKommandozeile(t, umgebung, "einstellungen", "--json"), "farbwelt"); farbwelt != erwartet {
				t.Errorf("die Farbwelt der Lage ist %q, erwartet %q", farbwelt, erwartet)
			}

			for _, liste := range []struct {
				aufruf []string
				marke  string
			}{
				{[]string{"turniere"}, "Aus de"},
				{[]string{"formate", "--json"}, "Aus de"},
				{[]string{"sets"}, "Aus de"},
				{[]string{"autoersetzung", setID, "--json"}, "zeile"},
			} {
				if inhalt := mussKommandozeile(t, umgebung, liste.aufruf...); strings.Contains(inhalt, liste.marke) != lage.darf {
					t.Errorf("%v der Lage:\n%s", liste.aufruf, inhalt)
				}
			}
		})
	}
}
