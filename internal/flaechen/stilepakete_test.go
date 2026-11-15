// **Stile, Stiloptionen, Pakete und Paketdateien über jede Fläche ohne
// Fenster** (#1169).
//
// # Warum die vier zusammen stehen
//
// Sie betreffen das **Aussehen** und das, woraus es kommt: welcher Stil an
// einem Set hängt, welche Optionen dieser Stil hat, welche Pakete installiert
// sind und welche Dateien darin bearbeitbar sind. Alle vier brauchen dieselbe
// Vorrichtung — die Saat als Paketverzeichnis und ein eingefahrenes Set.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Stile eines Sets stehen an jeder Fläche — und der Wechsel wirkt.**
//
// # Die Zusage
//
// Erst nennen alle drei dieselben wählbaren Stile. Dann wechselt **eine**, und
// die anderen beiden sehen den neuen — die Wahl gehört dem Set, und es gibt
// nur eines davon.
func TestDieStileUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	fuerAlleDrei(t, "die Stile", "probe-kunst",
		func() string { return mussKommandozeile(t, umgebung, "stile", setID) },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.stile",
				map[string]any{"setID": zahl(t, setID)})
			if istFehler {
				t.Fatalf("lesen.stile: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/stile", "")
			if status != http.StatusOK {
				t.Fatalf("GET stile: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Gewechselt wird über den Dienst, gelesen über alle drei.**
	//
	// Der Stil gehört dem Set; wer die Wirkung nur an der Fläche prüfte, die
	// gewechselt hat, sähe nicht, dass die anderen denselben Bestand lesen.
	if status, rumpf := an.ruf(t, http.MethodPut, "/sets/"+setID+"/stil",
		`{"paket":"probe-kunst"}`); status != http.StatusOK {
		t.Fatalf("PUT stil: %d (%s)", status, rumpf)
	}

	// **Der Stern markiert den gewählten** — so schreibt es die
	// Kommandozeile, und das ist die Auskunft, um die es geht.
	if aus := mussKommandozeile(t, umgebung, "stile", setID); !strings.Contains(aus, "* probe-kunst") {
		t.Errorf("die Kommandozeile sieht den Wechsel nicht:\n%s", aus)
	}

	// **An MCP, Dienst und Fenster der gewählte Stil, nicht ein Name** —
	// „probe-kunst" stand schon vor dem Wechsel in der Liste (Zweitblick über
	// Phase C, Befund 3).
	aus, istFehler := mcpRuf(t, umgebung, "lesen.stile", map[string]any{"setID": zahl(t, setID)})
	if istFehler {
		t.Fatalf("lesen.stile: %s", aus)
	}

	_, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/stile", "")
	fenster := fensterAus(t, umgebung)

	for flaeche, roh := range map[string]string{
		"MCP": aus, "Dienst": rumpf, "Fenster": string(mussJSON(t, fenster.Stile(t.Context(), setID))),
	} {
		if gewaehlt := gewaehlterStil(t, roh); gewaehlt != "probe-kunst" {
			t.Errorf("%s nennt %q als gewählt, gewechselt war zu probe-kunst:\n%s", flaeche, gewaehlt, roh)
		}
	}

	// **Und zurück über das Fenster** — gesehen an der Kommandozeile.
	if antwort := fenster.StilWechseln(t.Context(), setID, "probe-schlicht"); antwort.Fehler != nil {
		t.Fatalf("StilWechseln: %+v", antwort.Fehler)
	}

	if aus := mussKommandozeile(t, umgebung, "stile", setID); !strings.Contains(aus, "* probe-schlicht") {
		t.Errorf("die Kommandozeile sieht den Wechsel im Fenster nicht:\n%s", aus)
	}
}

// gewaehlterStil liest den gewählten Stil — aus einer nackten Liste (MCP,
// Dienst) oder aus der Hülle des Fensters (`{"stile":[…]}`).
func gewaehlterStil(t *testing.T, roh string) string {
	t.Helper()

	type stil struct {
		Paket    string `json:"paket"`
		Gewaehlt bool   `json:"gewaehlt"`
	}

	var huelle struct {
		Stile []stil `json:"stile"`
	}

	liste := []stil{}
	if json.Unmarshal([]byte(roh), &huelle) == nil && len(huelle.Stile) > 0 {
		liste = huelle.Stile
	} else if fehler := json.Unmarshal([]byte(roh), &liste); fehler != nil {
		t.Fatalf("kein JSON: %v\n%s", fehler, roh)
	}

	for _, eins := range liste {
		if eins.Gewaehlt {
			return eins.Paket
		}
	}

	return ""
}

// **Ein Stil, den es nicht gibt, wird überall gleich abgewiesen.**
func TestEinFremderStilWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Stil, den es nicht gibt", abweisung{
		kern:   "liegt in keinem Paketverzeichnis",
		klasse: klasseNichtGefunden,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung,
				"stile", setID, "wechseln", "gibtsnicht")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.stil",
				map[string]any{"setID": zahl(t, setID), "paket": "gibtsnicht"})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPut, "/sets/"+setID+"/stil",
				`{"paket":"gibtsnicht"}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).StilWechseln(t.Context(), setID, "gibtsnicht").Fehler
		},
	})
}

// **Eine Stiloption, an einer Fläche gesetzt, gilt an allen — und lässt sich
// zurücksetzen.**
//
// # Warum das Zurücksetzen dazugehört
//
// Es ist die **dritte** Bindung derselben Fähigkeit, und sie ist die
// unauffällige: „auf die Vorgabe stellen" ist etwas anderes als „auf den
// Vorgabewert setzen" — danach steht dort wieder, was der Stil sagt, auch
// wenn der Stil sich ändert.
func TestDieStiloptionenUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	fuerAlleDrei(t, "die Stiloptionen", "hintergrund",
		func() string { return mussKommandozeile(t, umgebung, "stiloption", setID) },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.stiloptionen",
				map[string]any{"setID": zahl(t, setID)})
			if istFehler {
				t.Fatalf("lesen.stiloptionen: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/stiloptionen", "")
			if status != http.StatusOK {
				t.Fatalf("GET stiloptionen: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Gesetzt über MCP, gesehen auf der Zeile und im Dienst.**
	if text, istFehler := mcpRuf(t, umgebung, "schreiben.stiloption",
		map[string]any{"setID": zahl(t, setID), "feld": "hintergrund", "wert": "dunkel"}); istFehler {
		t.Fatalf("schreiben.stiloption: %s", text)
	}

	if aus := mussKommandozeile(t, umgebung, "stiloption", setID); !strings.Contains(aus, "dunkel") {
		t.Errorf("die Kommandozeile sieht die gesetzte Option nicht:\n%s", aus)
	}

	// **Und zurückgesetzt über den Dienst.** Danach steht wieder die Vorgabe
	// da — `hell`, wie der Stil sie nennt.
	if status, rumpf := an.ruf(t, http.MethodDelete,
		"/sets/"+setID+"/stiloptionen/hintergrund", ""); status != http.StatusOK {
		t.Fatalf("DELETE stiloption: %d (%s)", status, rumpf)
	}

	aus := mussKommandozeile(t, umgebung, "stiloption", setID)
	if strings.Contains(aus, "dunkel") {
		t.Errorf("die Option steht nach dem Zurücksetzen noch auf dem gesetzten Wert:\n%s", aus)
	}
}

// **Die installierten Pakete und ihre Dateien stehen an jeder Fläche.**
//
// # Warum beide in einem Lauf
//
// Die zweite Auskunft baut auf der ersten: welche Dateien ein Paket hat, kann
// nur fragen, wer weiß, dass es das Paket gibt. Zwei Läufe daneben bauten
// dieselbe Saat zweimal auf.
func TestDiePaketeUndIhreDateienUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	an := starteDienst(t, umgebung)

	const paket = "probe-schlicht.mse-style"

	fuerAlleDrei(t, "die Pakete", "probe-schlicht",
		func() string { return mussKommandozeile(t, umgebung, "pakete") },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.pakete", map[string]any{})
			if istFehler {
				t.Fatalf("lesen.pakete: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/pakete", "")
			if status != http.StatusOK {
				t.Fatalf("GET pakete: %d (%s)", status, rumpf)
			}

			return rumpf
		})

	// **Und die bearbeitbaren Dateien darin.** Ein `.mse-style` führt genau
	// eine — `style` —, und das ist die Auskunft, mit der die Paketwerkstatt
	// beginnt.
	fuerAlleDrei(t, "die Paketdateien", "style",
		func() string { return mussKommandozeile(t, umgebung, "paketdateien", paket) },
		func() string {
			aus, istFehler := mcpRuf(t, umgebung, "lesen.paketdateien",
				map[string]any{"paket": paket})
			if istFehler {
				t.Fatalf("lesen.paketdateien: %s", aus)
			}

			return aus
		},
		func() string {
			status, rumpf := an.ruf(t, http.MethodGet, "/pakete/"+paket+"/dateien", "")
			if status != http.StatusOK {
				t.Fatalf("GET paketdateien: %d (%s)", status, rumpf)
			}

			return rumpf
		})
}

// **Ein Paket, das es nicht gibt, wird überall gleich abgewiesen.**
func TestEinFremdesPaketWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	const fremd = "gibtsnicht.mse-style"

	an := starteDienst(t, umgebung)

	// **Der Kern ist der Satz und nicht mehr der Go-Fehler.** Bis zum
	// 2026-09-06 stand hier `"öffnen: stat"` — der Text von `os.Stat`, und
	// damit prüfte dieser Lauf, dass ein **Go-Fehler** an allen drei Flächen
	// ankommt. Genau das ist seit Werkstück 1 von Phase 4.13 nicht mehr so
	// (§8, #1352):
	// der Grund geht ins Protokoll, der Satz an die Fläche.
	//
	// **Der Ort bleibt im Satz**, und das ist eine Entscheidung: ein Paket
	// liegt dort, wo jemand es hingestellt hat, und „welches der
	// installierten fehlt?" beantwortet nur der ganze Pfad (Vermerk 0075).
	abweisungGleich(t, "ein Paket, das es nicht gibt", abweisung{
		kern:   "ist nicht zu finden",
		klasse: klasseNichtGefunden,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "paketdateien", fremd)

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "lesen.paketdateien",
				map[string]any{"paket": fremd})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodGet, "/pakete/"+fremd+"/dateien", "")
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).Paketdateien(t.Context(), fremd).Fehler
		},
	})
}
