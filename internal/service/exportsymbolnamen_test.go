package service_test

import (
	"errors"
	"strings"
	"testing"
)

// **Der Dateiname und der `alt`-Text eines Symbolbilds** —
// #658.
//
// `exportsymbole_test.go` daneben prüft den Weg: Zerlegen, Rastern, Merkliste,
// Ablegen. Hier stehen die beiden Zeichenketten, die am Ende in der Ausgabe
// landen — und sie folgen **zwei verschiedenen** Regeln aus MSE, `clean_filename`
// und `html_escape`.

// **MSEs `html_escape` — fünf Zweige und die Zahlenmarke.**
//
//	if (c == ESCAPED_LANGLE || c == _('<')) ret += _("&lt;");
//	else if (c == _('>')) ret += _("&gt;");
//	else if (c == _('&')) ret += _("&amp;");
//	else if (c == _('\'')) ret += _("&#39;");
//	else if (c == _('\"')) ret += _("&quot;");
//	else if (c >= 0x80) ret += String(_("&#")) << (int)c << _(';');
//
// `src/script/functions/export.cpp`, Zeilen 158–178, Fassung f9d9356 vom
// 2021-07-06.
//
// **Strenger als der Weg in `to_html`**: dort bleiben `>`, `"` und `'` roh,
// hier nicht — der Text steht in einem **Attribut**, und ein Hochkomma darin
// beendete es.
func TestDerAltTextWirdMaskiert(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, zug, alt string }{
		{"kleiner", `<`, "&lt;"},
		{"geschütztes kleiner", "\u0001", "&lt;"},
		{"größer", `>`, "&gt;"},
		{"kaufmännisches Und", `&`, "&amp;"},
		{"Hochkomma", `'`, "&#39;"},
		{"Anführungszeichen", `\"`, "&quot;"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ausgabe, _, err := mitSymbolen(&beigabenmitschrift{}, &symbolrastererprobe{}).
				Fuehre(t.Context(),
					vorlageMit(`symbols_to_html("`+fall.zug+`", symbol_font: "mana")`),
					eineKarte(), nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v", err)
			}

			if !strings.Contains(ausgabe, "alt='"+fall.alt+"'") {
				t.Errorf("%s — erwartet alt='%s'", ausgabe, fall.alt)
			}
		})
	}
}

// **Der Dateiname bleibt einer, auch wenn nichts Brauchbares übrig ist.**
//
//	if (clean.empty()) { clean = _("no-name") + clean; }
//
// `clean_filename`, `src/util/file_utils.cpp`, Zeilen 68–71, Fassung f9d9356
// vom 2021-07-06. Ohne diesen Rückfall hieße die Datei `mana-.png`, und zwei
// verschiedene Symbole ohne brauchbare Zeichen teilten sie sich.
func TestEinDateinameOhneBrauchbareZeichen(t *testing.T) {
	t.Parallel()

	faelle := []struct{ name, zug, datei string }{
		{"nur ein Schrägstrich", `/`, "mana-no-name.png"},
		{"ein führendes Leerzeichen fällt weg", ` X`, "mana-X.png"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ausgabe, _, err := mitSymbolen(&beigabenmitschrift{}, &symbolrastererprobe{}).
				Fuehre(t.Context(),
					vorlageMit(`symbols_to_html("`+fall.zug+`", symbol_font: "mana")`),
					eineKarte(), nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v", err)
			}

			if !strings.Contains(ausgabe, "src='"+fall.datei+"'") {
				t.Errorf("%s — erwartet src='%s'", ausgabe, fall.datei)
			}
		})
	}
}

// **Ohne Eingabe wird der Name genannt**, und ein Schreibfehler kommt durch.
func TestSymbolsToHTMLOhneEingabeUndMitSchreibfehler(t *testing.T) {
	t.Parallel()

	t.Run("ohne Eingabe", func(t *testing.T) {
		t.Parallel()

		_, _, err := mitSymbolen(&beigabenmitschrift{}, &symbolrastererprobe{}).
			Fuehre(t.Context(), vorlageMit(`symbols_to_html(symbol_font: "mana")`),
				eineKarte(), nil, nil)
		if err == nil || !strings.Contains(err.Error(), "symbols_to_html") {
			t.Fatalf("Fehler = %v", err)
		}
	})

	// **Eine halbe Ausfuhr ist eine, die jemand für vollständig hält** (§8).
	t.Run("die Platte ist voll", func(t *testing.T) {
		t.Parallel()

		_, _, err := mitSymbolen(
			&beigabenmitschrift{schreibfehte: errors.New("kein Platz")},
			&symbolrastererprobe{},
		).Fuehre(t.Context(), vorlageMit(`symbols_to_html("W", symbol_font: "mana")`),
			eineKarte(), nil, nil)
		if err == nil || !strings.Contains(err.Error(), "kein Platz") {
			t.Fatalf("Fehler = %v", err)
		}
	})
}
