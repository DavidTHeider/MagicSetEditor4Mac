package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`position` über Text sucht `in` in `of`, nicht umgekehrt** —
// #2006.
//
//	if (of_t == SCRIPT_STRING || in_t == SCRIPT_STRING) {
//	  // string finding
//	  return (int)of->toString().find(in->toString()); // (int)npos == -1
//
// `src/script/functions/basic.cpp`, Zeilen 456–459, Fassung f9d9356 vom
// 2021-07-06. **Gefragt am Orakel** (`magicseteditor --cli … --raw`,
// 2026-10-03), jede Zeile unten ist seine Antwort. Hier stand bis dahin die
// umgekehrte Suche, erschlossen aus dem Namen der Argumente — und eine Liste
// neben einem Text wurde durchsucht, wo MSE abbricht. Keine Vorlage der Packs
// ruft das eingebaute `position` mit Text; `ar_position` ist eine eigene
// Skriptfunktion.
func TestPositionUeberTextWieMSE(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ quelle, erwartet string }{
		{`position(of:"c", in:"abc")`, "-1"},
		{`position(of:"abc", in:"c")`, "2"},
		{`position(of:"", in:"abc")`, "-1"},
		{`position(of:"1", in:"x1y")`, "-1"},
		{`position(of:"x1y", in:"1")`, "1"},
		{`position(of:12, in:"2")`, "1"},
		{`position(of:2, in:[1,2,3])`, "1"},
		{`position(of:[2], in:[[1],[2]])`, "1"},
	} {
		t.Run(fall.quelle, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%s ergab %s, MSE2 gibt %s", fall.quelle, wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Text neben einer Liste bricht ab, wie bei MSE** — „Can't convert from
// collection to string": die Textsuche greift, sobald **eine** Seite Text ist,
// und eine Liste hat keine Textform.
func TestPositionMitTextNebenEinerListeIstEinFehler(t *testing.T) {
	t.Parallel()

	for _, quelle := range []string{
		`position(of:"b", in: split_text(match:",", "a,b,c"))`,
		`position(of:"2", in:[1,2,3])`,
		`position(of:[1], in:"1")`,
	} {
		t.Run(quelle, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
			if err == nil || !strings.Contains(err.Error(), "Liste") {
				t.Errorf("%s: erwartet ein Fehler über die Liste, bekam %v", quelle, err)
			}
		})
	}
}
