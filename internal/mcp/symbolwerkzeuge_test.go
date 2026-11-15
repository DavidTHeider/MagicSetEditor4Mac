// **Die Rufe der Symbolwerkzeuge und der Massenänderung** — Werkstücke
// #1390 und
// #1391.
//
// Was hier geprüft wird: dass ein Werkzeug seine Argumente liest, die Naht
// ruft und ihre Antwort weiterreicht — samt den Wegen, auf denen es abweist.
// Dass die Rechnung stimmt, prüft `internal/app`; dass alle Flächen dasselbe
// sagen, `internal/flaechen`.
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// errBoese ist der Fehler, mit dem die Proben scheitern.
var errBoese = errors.New("die Naht sagt nein")

// werkzeugMitNamenUndSymbolfehler baut die Werkzeuge mit einer Symbolprobe,
// die scheitert.
func werkzeugMitNamenUndSymbolfehler(
	t *testing.T, bestand *bestandsprobe, name string,
) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(
		naehteMitSymbolprobe(bestand, &symbolprobe{fehler: errBoese}),
	) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// rufeText ruft ein Werkzeug und gibt seine rohe Antwort.
//
// **Als Text und nicht als Abbildung**: `rufeWerkzeug` daneben zerlegt die
// Antwort, und hier zählt, ob ein Name überhaupt vorkommt — in einer
// verschachtelten Gestalt, die je Werkzeug anders aussieht.
func rufeText(t *testing.T, name, argumente string) string {
	t.Helper()

	werkzeug := werkzeugMitNamen(t, probenbestand(), name)

	antwort, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	return antwort
}

// **Die Symbolwerkzeuge reichen durch.**
func TestDieSymbolwerkzeugeReichenDurch(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		werkzeug  string
		argumente string
		erwartet  string
	}{
		"lesen": {
			"lesen.symbol", `{"setID": 1, "name": "probe.mse-symbol"}`,
			"probe.mse-symbol",
		},
		"palette": {
			"lesen.symbolpalette", `{"karteID": 1, "feld": "casting cost"}`,
			"casting cost",
		},
		"rechnen": {
			"lesen.symbol.grundform", `{"art": "Rechteck"}`,
			"lesen.symbol.grundform",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := rufeText(t, fall.werkzeug, fall.argumente)

			if !strings.Contains(antwort, fall.erwartet) {
				t.Errorf("die Antwort nennt %q nicht: %s", fall.erwartet, antwort)
			}
		})
	}
}

// **Das Ablegen reicht die Teile weiter und gibt den Namen.**
func TestSchreibenSymbolReichtDieTeileWeiter(t *testing.T) {
	t.Parallel()

	antwort := rufeText(t, "schreiben.symbol",
		`{"setID": 1, "teile": [{"art":"Form"}], "unter": "eigen.mse-symbol"}`)

	if !strings.Contains(antwort, "eigen.mse-symbol") {
		t.Errorf("die Antwort lautet %s", antwort)
	}
}

// **Und ohne Namen kommt der Vorgabename.**
func TestSchreibenSymbolOhneNamen(t *testing.T) {
	t.Parallel()

	antwort := rufeText(t, "schreiben.symbol", `{"setID": 1, "teile": []}`)

	if !strings.Contains(antwort, "symbol.mse-symbol") {
		t.Errorf("die Antwort lautet %s", antwort)
	}
}

// **Die Fehlerwege der Symbolwerkzeuge und der Massenänderung.**
//
// Jeder einzeln, und jeder betreten: ein Zweig, der nur gebaut und nie
// begangen wird, ist eine Behauptung darüber, was im Ernstfall geschieht.
func TestDieFehlerwegeDerSymbolwerkzeuge(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		werkzeug  string
		argumente string
	}{
		"lesen.symbol: kein JSON":        {"lesen.symbol", `{`},
		"lesen.symbol: Kennung null":     {"lesen.symbol", `{"setID": 0}`},
		"schreiben.symbol: kein JSON":    {"schreiben.symbol", `{`},
		"schreiben.symbol: Kennung null": {"schreiben.symbol", `{"setID": 0, "teile": []}`},
		"palette: kein JSON":             {"lesen.symbolpalette", `{`},
		"palette: Kennung null":          {"lesen.symbolpalette", `{"karteID": 0, "feld": "x"}`},
		"rechnen: kein JSON":             {"lesen.symbol.grundform", `{`},
		"felder: kein JSON":              {"schreiben.felder", `{`},
		"felder: Setkennung null":        {"schreiben.felder", `{"setID": 0, "karteIDs": [1], "feld": "a", "wert": "b"}`},
		"felder: Kartenkennung null":     {"schreiben.felder", `{"setID": 1, "karteIDs": [0], "feld": "a", "wert": "b"}`},
		"vorschau: kein JSON":            {"lesen.felder.vorschau", `{`},
		"vorschau: Setkennung null":      {"lesen.felder.vorschau", `{"setID": 0, "karteIDs": [1], "feld": "a", "wert": "b"}`},
		"vorschau: Kartenkennung null":   {"lesen.felder.vorschau", `{"setID": 1, "karteIDs": [0], "feld": "a", "wert": "b"}`},
		"verlauf: kein JSON":             {"lesen.verlauf", `{`},
		"verlauf: Setkennung null":       {"lesen.verlauf", `{"setID": 0}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werkzeug := werkzeugMitNamen(t, probenbestand(), fall.werkzeug)

			if _, fehler := werkzeug.Ruf(
				t.Context(), json.RawMessage(fall.argumente),
			); fehler == nil {
				t.Error("der Ruf ging durch, erwartet war eine Abweisung")
			}
		})
	}
}

// **Und die Massenänderung sagt, wie viele Karten sie geändert hat.**
func TestSchreibenFelderNenntDieZahl(t *testing.T) {
	t.Parallel()

	probe := probenbestand()

	werkzeug := werkzeugMitNamen(t, probe, "schreiben.felder")

	antwort, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(
		`{"setID": 1, "karteIDs": [7, 8], "feld": "artist", "wert": "X"}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, "2") {
		t.Errorf("die Antwort lautet %s", antwort)
	}

	// **Und beide Karten kamen an der Naht an.** Ohne diese Zusicherung wäre
	// der Lauf auch grün, wenn das Werkzeug die Liste fallen ließe und die
	// Zahl aus seinen Argumenten zählte.
	if len(probe.massengesetzt) != 2 {
		t.Errorf("weitergereicht: %v", probe.massengesetzt)
	}
}

// **Und ein Fehler der Naht kommt durch — an jedem Weg.**
//
// Die vier Zweige sind je zwei Zeilen und sehen aus wie eine Formalie. Ohne
// diesen Lauf bliebe unbelegt, dass ein Modell den Fehler überhaupt zu sehen
// bekommt statt einer leeren Antwort.
func TestEinFehlerDerSymbolnahtKommtDurch(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"lesen.symbol":           `{"setID": 1}`,
		"schreiben.symbol":       `{"setID": 1, "teile": []}`,
		"lesen.symbolpalette":    `{"karteID": 1, "feld": "x"}`,
		"lesen.symbol.grundform": `{"art": "Rechteck"}`,
		"schreiben.felder":       `{"setID": 1, "karteIDs": [1], "feld": "a", "wert": "b"}`,
		"lesen.felder.vorschau":  `{"setID": 1, "karteIDs": [1], "feld": "a", "wert": "b"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bestand := probenbestand()
			bestand.schreibfehler = errBoese

			werkzeug := werkzeugMitNamenUndSymbolfehler(t, bestand, name)

			if _, fehler := werkzeug.Ruf(
				t.Context(), json.RawMessage(argumente),
			); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}
