// **Ein gewechselter Stil steht in der gesicherten Datei — in der Form, die
// MSE2 liest** (F05,
// #2247;
// Zweitblick über Phase G, zweiter Teil, Befund 2).
//
// # Was hier schiefging
//
// Der Schreiber übernahm die Kopfzeilen aus der Quelldatei. Wer den Stil
// wechselte und sicherte, bekam den **alten** — der Wechsel stand im Bestand
// und nirgends sonst. Und ein Name in voller Paketform (`probe-kunst`) ist für
// MSE2 kein Stil: es sucht `<spiel>-<wert>.mse-style`
// (`src/data/set.cpp`, `src/data/stylesheet.cpp`, Fassung f9d9356 vom
// 2021-07-06).
//
// # Warum über alle vier Flächen
//
// Jede baut ihren Schreiber selbst (`verdrahtung.go`, `mcpnaht.go`); wer die
// Auflösung nur an einer einsetzt, schreibt an den anderen die volle Form.
package flaechen_test

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinGewechselterStilStehtInDerGesichertenDatei(t *testing.T) {
	sichern := map[string]func(t *testing.T, umgebung []string, setID, ziel string){
		"Kommandozeile": func(t *testing.T, umgebung []string, setID, ziel string) {
			mussKommandozeile(t, umgebung, "sichern", setID, ziel)
		},
		"MCP": func(t *testing.T, umgebung []string, setID, ziel string) {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.set.sichern",
				map[string]any{"setID": zahl(t, setID), "ziel": ziel}); istFehler {
				t.Fatalf("sichern: %s", text)
			}
		},
		"Dienst": func(t *testing.T, umgebung []string, setID, ziel string) {
			an := starteDienst(t, umgebung)
			if status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/sicherung",
				`{"ziel":"`+ziel+`"}`); status != http.StatusCreated {
				t.Fatalf("POST sicherung: %d (%s)", status, rumpf)
			}
		},
		"Fenster": func(t *testing.T, umgebung []string, setID, ziel string) {
			// **Der Dialog antwortet mit dem Ziel** — wie die Hülle es tut.
			if antwort := fensterAus(t, umgebung, ziel).SetSpeichernUnter(t.Context(), setID); antwort.Fehler != nil {
				t.Fatalf("SetSpeichernUnter: %+v", antwort.Fehler)
			}
		},
	}

	for flaeche, sichere := range sichern {
		t.Run(flaeche, func(t *testing.T) {
			umgebung := mitEigenerSaat(t)
			setID := eingefahren(t, umgebung)

			mussKommandozeile(t, umgebung, "stile", setID, "wechseln", "probe-kunst")

			ziel := filepath.Join(t.TempDir(), "gewechselt.mse-set")
			sichere(t, umgebung, setID, ziel)

			kopf := setdateiAus(t, ziel)
			if !strings.Contains(kopf, "\nstylesheet: kunst\n") {
				t.Errorf("der gewechselte Stil steht nicht in Kurzform in der Datei:\n%s", kopf)
			}
		})
	}
}

// setdateiAus liest die Datei `set` aus einem gesicherten Set. **Ein Ordner**,
// denn die Quelle ist einer, und so schreibt der Schreiber sie wieder.
func setdateiAus(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := fs.ReadFile(os.DirFS(pfad), "set")
	if err != nil {
		t.Fatalf("die Datei set fehlt: %v", err)
	}

	return string(inhalt)
}

// **Eine Stiloption wirkt nach Sichern und Wiederöffnen** — auch wenn die
// Quelle noch einen Block unter dem falschen Namen trägt, wie ihn der
// Schreiber vor F05 anlegte (`probe-probe-schlicht`; Nachlauf über Phase G,
// zweiter Teil, Befund 1). Gelesen wird die Option unter dem Namen des
// Pakets; schrieb der Schreiber in den alten Block, stand nach dem
// Wiederöffnen wieder die Vorgabe da.
func TestEineStiloptionWirktNachSichernUndOeffnen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	quelle := filepath.Join(t.TempDir(), "alt.mse-set")
	if fehler := os.MkdirAll(quelle, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\nset info:\n\ttitle: Alter Block\n" +
		"card:\n\tname: Feuerball\nstyling:\n\tprobe-probe-schlicht:\n\t\thintergrund: hell\n"
	if fehler := os.WriteFile(filepath.Join(quelle, "set"), []byte(inhalt), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	setID := setkennungAus(t, mussKommandozeile(t, umgebung, "oeffnen", quelle, "--json"))
	mussKommandozeile(t, umgebung, "stiloption", setID, "setzen", "hintergrund", "dunkel")

	ziel := filepath.Join(t.TempDir(), "wieder.mse-set")
	mussKommandozeile(t, umgebung, "sichern", setID, ziel)

	wieder := mitEigenerSaat(t)
	neu := setkennungAus(t, mussKommandozeile(t, wieder, "oeffnen", ziel, "--json"))

	if aus := mussKommandozeile(t, wieder, "stiloption", neu); !strings.Contains(aus, "dunkel") {
		t.Errorf("die Option wirkt nach dem Wiederöffnen nicht:\n%s\n\nDatei:\n%s", aus, setdateiAus(t, ziel))
	}
}
