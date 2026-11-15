// **Was die vierzehn neuen Werkzeuge antworten — und wann sie abweisen**
// (#1107).
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// alleNeuenFaelle sind die Argumente, mit denen jedes neue Werkzeug gelingt.
var alleNeuenFaelle = map[string]string{
	// **Der Einstellungsaustausch** — Werkstück O3 von Phase 5.2 (#1427).
	"lesen.einstellungen.ausfuhr":        `{"abschnitte":["nutzer","profil"]}`,
	"lesen.einstellungen.vorschau":       `{"inhalt":"{\"fassung\":1}"}`,
	"schreiben.einstellungen.einfuhr":    `{"inhalt":"{\"fassung\":1}","revision":"abc123","nur":["profil:sprache"]}`,
	"lesen.stile":                        `{"setID": 1}`,
	"lesen.stilfolgen":                   `{"setID": 1, "paket": "probe-einzeiler"}`,
	"lesen.stiloptionen":                 `{"setID": 1}`,
	"schreiben.stil":                     `{"setID": 1, "paket": "probe-einzeiler"}`,
	"schreiben.stiloption":               `{"setID": 1, "feld": "rand", "wert": "breit"}`,
	"schreiben.stiloption.zuruecksetzen": `{"setID": 1, "feld": "rand"}`,
	"lesen.pakete":                       `{}`,
	"lesen.statistik":                    `{"setID": 1}`,
	"lesen.verteilung":                   `{"setID": 1, "feld": "name"}`,
	"lesen.einstellungen":                `{}`,
	"schreiben.paket.beziehen": `{"adresse": "https://example.invalid/x", ` +
		`"stand": "abc1234", "unterverzeichnis": "data", "name": "probe"}`,
	"schreiben.zuruecknehmen":    `{"setID": 1}`,
	"schreiben.wiederherstellen": `{"setID": 1}`,
	"schreiben.einstellung":      `{"schluessel": "sprache", "wert": "en"}`,
}

// **Jedes reicht den Fehler seiner Naht durch.**
//
// **Je Werkzeug und nicht einmal**: jedes hat sein eigenes `if fehler != nil`,
// und ein vergessenes gäbe eine leere Antwort statt einer Meldung.
func TestDieNeuenWerkzeugeReichenIhrenFehlerDurch(t *testing.T) {
	for name, argumente := range alleNeuenFaelle {
		t.Run(name, func(t *testing.T) {
			werkzeug := werkzeugMitBoesenNaehten(t, name)

			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente)); err == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Jedes gelingt über willigen Nähten.**
//
// **Ohne diesen Fall sagte der Test darüber nur, dass etwas fehlschlägt** —
// und das täte auch ein Werkzeug, das gar nichts kann.
func TestDieNeuenWerkzeugeAntwortenUeberWilligenNaehten(t *testing.T) {
	for name, argumente := range alleNeuenFaelle {
		t.Run(name, func(t *testing.T) {
			werkzeug := werkzeugMitNeuenNaehten(t, neueStilprobe(),
				neueBestandsprobeWeit(), name)

			aus, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
			if err != nil {
				t.Fatalf("Ruf: %v", err)
			}

			if aus == "" {
				t.Error("die Antwort ist leer")
			}
		})
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen, bevor die Naht sie sieht.**
func TestDieNeuenWerkzeugeWeisenNullkennungenAb(t *testing.T) {
	faelle := map[string]string{
		"lesen.stile":                        `{"setID": 0}`,
		"lesen.stilfolgen":                   `{"setID": 0, "paket": "p"}`,
		"lesen.stiloptionen":                 `{"setID": 0}`,
		"schreiben.stil":                     `{"setID": 0, "paket": "p"}`,
		"schreiben.stiloption":               `{"setID": 0, "feld": "rand", "wert": "a"}`,
		"schreiben.stiloption.zuruecksetzen": `{"setID": 0, "feld": "rand"}`,
		"lesen.statistik":                    `{"setID": 0}`,
		"lesen.verteilung":                   `{"setID": 0, "feld": "name"}`,
		"schreiben.zuruecknehmen":            `{"setID": 0}`,
		"schreiben.wiederherstellen":         `{"setID": 0}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			werkzeug := werkzeugMitNeuenNaehten(t, neueStilprobe(),
				neueBestandsprobeWeit(), name)

			_, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
			if err == nil {
				t.Fatal("die Null lief durch")
			}

			if !strings.Contains(err.Error(), "setID") {
				t.Errorf("die Meldung nennt das Feld nicht: %v", err)
			}
		})
	}
}

// **Und unlesbare Argumente ebenso.**
func TestDieNeuenWerkzeugeWeisenUnlesbareArgumenteAb(t *testing.T) {
	for _, name := range []string{
		"lesen.stilfolgen", "schreiben.stil", "schreiben.stiloption",
		"schreiben.stiloption.zuruecksetzen", "lesen.verteilung",
		"schreiben.paket.beziehen", "schreiben.einstellung",

		// **Die beiden Bewegungen seit #1393.** Sie lesen ihre Argumente
		// jetzt selbst, weil ein Vorgang danebensteht — vorher ging das über
		// [kennung], und dessen Fehlerweg war anderswo geprüft.
		"schreiben.zuruecknehmen", "schreiben.wiederherstellen",
	} {
		t.Run(name, func(t *testing.T) {
			werkzeug := werkzeugMitNeuenNaehten(t, neueStilprobe(),
				neueBestandsprobeWeit(), name)

			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{`)); err == nil {
				t.Error("ein unlesbarer Rumpf lief durch")
			}
		})
	}
}

// werkzeugMitNeuenNaehten sucht ein Werkzeug über Stil- und Bestandsprobe.
func werkzeugMitNeuenNaehten(
	t *testing.T, stil *stilprobe, bestand *bestandsprobeWeit, name string,
) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(), Ausfuhr: ruhendeAusfuhr{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: stil, Bestand: bestand,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("es gibt kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// werkzeugMitBoesenNaehten sucht ein Werkzeug über Nähten, die an allem scheitern.
func werkzeugMitBoesenNaehten(t *testing.T, name string) mcp.Werkzeug {
	t.Helper()

	boese := errors.New("die Naht ist zu")

	stil := neueStilprobe()
	stil.fehler = boese

	bestand := neueBestandsprobeWeit()
	bestand.fehler = boese

	return werkzeugMitNeuenNaehten(t, stil, bestand, name)
}
