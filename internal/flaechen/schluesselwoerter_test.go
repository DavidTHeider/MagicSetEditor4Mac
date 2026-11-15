// Die Schlüsselwörter über jede Fläche
// (#1062).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Anlegen, füllen, sehen, löschen — jede Fläche einen Schritt.**
//
// Der Reihe nach: was die eine hinterlässt, sieht die nächste. Das ist der
// Satz aus §16 — dieselbe Fähigkeit, überall dieselbe Antwort — und er ist
// erst dann geprüft, wenn eine Fläche die **Wirkung** einer anderen sieht.
func TestDieSchluesselwoerterUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)

	// **Angelegt über die Kommandozeile** — sie gibt die Kennung.
	wortID := strings.TrimSpace(
		mussKommandozeile(t, umgebung, "schluesselwort", setID, "anlegen"))

	if wortID == "" {
		t.Fatal("das Anlegen gab keine Kennung")
	}

	t.Run("gefüllt über MCP", func(t *testing.T) {
		for feld, wert := range map[string]string{
			"wort": "Flügelschlag", "muster": "Flügelschlag",
			"erinnerung": "Nur von Fliegern zu blocken.",
		} {
			text, istFehler := mcpRuf(t, umgebung, "schreiben.schluesselwort.setzen",
				map[string]any{
					"setID": zahl(t, setID), "wortID": zahl(t, wortID),
					"feld": feld, "wert": wert,
				})
			if istFehler {
				t.Fatalf("%s setzen: %s", feld, text)
			}
		}
	})

	t.Run("gesehen über REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet,
			"/sets/"+setID+"/schluesselwoerter", "")
		if status != http.StatusOK {
			t.Fatalf("GET schluesselwoerter: %d (%s)", status, rumpf)
		}

		for _, erwartet := range []string{
			`"wort":"Flügelschlag"`, `"muster":"Flügelschlag"`, `"fest":false`,
			`"gueltig":true`,
		} {
			if !strings.Contains(rumpf, erwartet) {
				t.Errorf("%s fehlt in\n%s", erwartet, rumpf)
			}
		}
	})

	t.Run("gesehen über die Kommandozeile", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "schluesselwoerter", setID)

		if !strings.Contains(aus, "Flügelschlag") {
			t.Errorf("das Schlüsselwort fehlt in\n%s", aus)
		}
	})

	t.Run("gelöscht über REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodDelete,
			"/sets/"+setID+"/schluesselwoerter/"+wortID, "")
		if status != http.StatusOK {
			t.Fatalf("DELETE: %d (%s)", status, rumpf)
		}
	})

	t.Run("und danach ist es überall weg", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "schluesselwoerter", setID)

		if strings.Contains(aus, "Flügelschlag") {
			t.Errorf("das Schlüsselwort steht noch in\n%s", aus)
		}

		text, istFehler := mcpRuf(t, umgebung, "lesen.schluesselwoerter",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen: %s", text)
		}

		if strings.Contains(text, "Flügelschlag") {
			t.Errorf("das Schlüsselwort steht noch in\n%s", text)
		}
	})
}

// **Und die des Spiels sind überall fest.**
//
// Die Saat bringt keine mit; geprüft wird deshalb, dass die Auskunft **fest**
// überhaupt durch jede Fläche kommt — und dass ein Schreiben auf eines ohne
// Kennung abgewiesen wird.
func TestEinFremdesSchluesselwortWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)

	t.Run("Kommandozeile", func(t *testing.T) {
		aus, fehler, code := kommandozeile(t, umgebung,
			"schluesselwort", setID, "setzen", "9999", "wort", "X")
		if code == 0 {
			t.Fatalf("das fremde Schlüsselwort kam durch: %s", aus)
		}

		if !strings.Contains(aus+fehler, "Schlüsselwort") {
			t.Errorf("die Meldung nennt es nicht: %s%s", aus, fehler)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.schluesselwort.setzen",
			map[string]any{
				"setID": zahl(t, setID), "wortID": 9999,
				"feld": "wort", "wert": "X",
			})
		if !istFehler {
			t.Errorf("das fremde Schlüsselwort kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, _ := an.ruf(t, http.MethodPut,
			"/sets/"+setID+"/schluesselwoerter/9999/felder/wort", `{"wert":"X"}`)
		if status == http.StatusOK {
			t.Error("das fremde Schlüsselwort kam durch")
		}
	})
}

// **Anlegen und Löschen kann jede Fläche für sich**
// (#1169).
//
// # Warum neben dem Lauf oben und nicht statt seiner
//
// Der oben fährt den **Reihum**: jede Fläche macht einen Schritt, und die
// nächste sieht seine Wirkung. Das ist die Zusage „alle sehen denselben
// Bestand" — und es lässt offen, ob **jede** Fläche den ganzen Weg kann.
//
// Gemessen: `schreiben.schluesselwort.anlegen` und `.loeschen` kamen in
// keinem Flächenlauf vor. Angelegt wurde stets über die Kommandozeile,
// gelöscht stets über den Dienst; wären die MCP-Werkzeuge kaputt, hätte es
// niemand gesehen.
func TestEinSchluesselwortAnlegenUndLoeschenUeberJedeFlaeche(t *testing.T) {
	for name, fall := range map[string]struct {
		anlegen  func(*testing.T, []string, string) string
		loeschen func(*testing.T, []string, string, string)
	}{
		"Kommandozeile": {
			anlegen: func(t *testing.T, umgebung []string, setID string) string {
				t.Helper()

				return strings.TrimSpace(
					mussKommandozeile(t, umgebung, "schluesselwort", setID, "anlegen"))
			},
			loeschen: func(t *testing.T, umgebung []string, setID, wortID string) {
				t.Helper()

				mussKommandozeile(t, umgebung, "schluesselwort", setID, "loeschen", wortID)
			},
		},
		"MCP": {
			anlegen: func(t *testing.T, umgebung []string, setID string) string {
				t.Helper()

				text, istFehler := mcpRuf(t, umgebung, "schreiben.schluesselwort.anlegen",
					map[string]any{"setID": zahl(t, setID)})
				if istFehler {
					t.Fatalf("schreiben.schluesselwort.anlegen: %s", text)
				}

				return wortkennungAus(t, text)
			},
			loeschen: func(t *testing.T, umgebung []string, setID, wortID string) {
				t.Helper()

				if text, istFehler := mcpRuf(t, umgebung, "schreiben.schluesselwort.loeschen",
					map[string]any{"setID": zahl(t, setID), "wortID": zahl(t, wortID)}); istFehler {
					t.Fatalf("schreiben.schluesselwort.loeschen: %s", text)
				}
			},
		},
		"REST": {
			anlegen: func(t *testing.T, umgebung []string, setID string) string {
				t.Helper()

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodPost,
					"/sets/"+setID+"/schluesselwoerter", "")
				if status != http.StatusOK && status != http.StatusCreated {
					t.Fatalf("POST schluesselwoerter: %d (%s)", status, rumpf)
				}

				return wortkennungAus(t, rumpf)
			},
			loeschen: func(t *testing.T, umgebung []string, setID, wortID string) {
				t.Helper()

				an := starteDienst(t, umgebung)

				if status, rumpf := an.ruf(t, http.MethodDelete,
					"/sets/"+setID+"/schluesselwoerter/"+wortID, ""); status != http.StatusOK {
					t.Fatalf("DELETE schluesselwort: %d (%s)", status, rumpf)
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := mitEigenerSaat(t)
			setID := setImProbebestand(t, umgebung)

			vorher := zeilenzahl(mussKommandozeile(t, umgebung, "schluesselwoerter", setID))

			wortID := fall.anlegen(t, umgebung, setID)
			if wortID == "" {
				t.Fatal("das Anlegen gab keine Kennung")
			}

			nachher := zeilenzahl(mussKommandozeile(t, umgebung, "schluesselwoerter", setID))
			if nachher != vorher+1 {
				t.Errorf("nach dem Anlegen %d Zeilen, vorher %d", nachher, vorher)
			}

			fall.loeschen(t, umgebung, setID, wortID)

			if wieder := zeilenzahl(
				mussKommandozeile(t, umgebung, "schluesselwoerter", setID)); wieder != vorher {
				t.Errorf("nach dem Löschen %d Zeilen, vor dem Anlegen %d", wieder, vorher)
			}
		})
	}
}

// wortkennungAus zieht die Kennung eines angelegten Schlüsselworts.
//
// **Aus zwei Gestalten**: MCP gibt `{"wortID":7}`, der Dienst `{"id":"7"}`.
// Das dürfen sie — gleich ist die Fachlichkeit, nicht die Ausgabe.
func wortkennungAus(t *testing.T, roh string) string {
	t.Helper()

	var antwort struct {
		WortID kennung `json:"wortID"`
		ID     kennung `json:"id"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if antwort.WortID != "" {
		return string(antwort.WortID)
	}

	return string(antwort.ID)
}

// zeilenzahl zählt die nicht leeren Zeilen einer Ausgabe.
func zeilenzahl(aus string) int {
	var gezaehlt int

	for _, zeile := range strings.Split(aus, "\n") {
		if strings.TrimSpace(zeile) != "" {
			gezaehlt++
		}
	}

	return gezaehlt
}
