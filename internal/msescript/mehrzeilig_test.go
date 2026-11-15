package msescript_test

// **`^` und `$` treffen an jeder Zeile**
// (#435).
//
// MSE benutzt `boost::regex` (`src/util/regex.cpp`, Zeile 20, Fassung f9d9356
// vom 2021-07-06) und übergibt **keine** Match-Flags — weder bei
// `regex_search` (`regex.hpp`, Zeilen 74 und 80) noch bei `regex_replace`
// (`regex.cpp`, Zeile 28, wo `boost::format_sed` eine **Format**-Flagge ist).
// Es gilt `match_default`, und darin ist `match_single_line` **nicht** gesetzt;
// genau dafür gibt es diese Flagge.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestMusterTreffenAnJederZeile(t *testing.T) {
	for _, fall := range []struct {
		name     string
		muster   string
		text     string
		erwartet []string
	}{
		{
			// **Der Fall aus dem Pack**: `magic.mse-game/script`, Zeile 1943,
			// klammert jede Zeile, die mit „• " beginnt.
			name:     "der Aufzählungspunkt aus dem Pack",
			muster:   `^(• )([^\n]+\n?)`,
			text:     "Choose one —\n• A\n• B",
			erwartet: []string{"• A\n", "• B"},
		},
		{
			name:     "das Ende einer Zeile",
			muster:   `x$`,
			text:     "ax\nbx",
			erwartet: []string{"x", "x"},
		},
		{
			// **Die Gegenprobe:** `\A` bleibt der Textanfang. Ohne sie wäre
			// eine Übersetzung, die jedes Anker-Zeichen mehrzeilig macht,
			// ebenso grün.
			name:     "der Textanfang bleibt der Textanfang",
			muster:   `\Aa`,
			text:     "b\na",
			erwartet: nil,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			muster, err := msescript.UebersetzeMuster(fall.muster)
			if err != nil {
				t.Fatalf("übersetzen: %v", err)
			}

			gefunden, err := muster.AlleTreffer(fall.text)
			if err != nil {
				t.Fatalf("suchen: %v", err)
			}

			if strings.Join(gefunden, "|") != strings.Join(fall.erwartet, "|") {
				t.Errorf("Treffer %q, erwartet %q", gefunden, fall.erwartet)
			}
		})
	}
}
