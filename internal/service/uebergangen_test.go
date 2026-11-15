package service_test

// **Was MSE2 übergeht, meldet dieses Programm nicht**
// (#429).
//
// `TextElements::fromString` (`src/render/text/element.cpp`, Zeilen 78–200,
// Fassung f9d9356 vom 2021-07-06) ist **eine** `else if`-Kette. Was sie nicht
// trifft, fällt durch — kommentarlos, und der Inhalt bleibt Text. Die Kette
// ist damit eine **abschließende** Liste, und sie steht dort.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Gegenprobe geht in beide Richtungen** (§10): eine Liste, die alles
// übergeht, wäre so falsch wie die alte, die nichts übergangen hat. Deshalb
// steht hier beides — was schweigen **muss** und was melden **muss**.
func TestNurWasMSE2DeutetWirdGemeldet(t *testing.T) {
	for _, fall := range []struct {
		name     string
		tag      string
		gemeldet bool
		warum    string
	}{
		// **Steht nicht in der Kette** — MSE2 überspringt sie ebenso.
		{name: "kw-a", tag: "kw-a", warum: "Schlüsselwortmarkierung des Editors"},
		{name: "word-list-type", tag: "word-list-type", warum: "Auswahlliste des Editors"},
		{name: "nospellcheck", tag: "nospellcheck", warum: "Rechtschreibprüfung"},
		{name: "param-cost", tag: "param-cost", warum: "weder atom-param noch ref-param"},
		{name: "sep", tag: "sep", warum: "der Trenner von combined_editor"},

		// **Steht in der Kette** — MSE2 wertet sie aus, wir noch nicht.
		// **`soft-line` seit #431
		// gedeutet** — es steht hier weiter, weil die Präfixregel es von
		// `soft` trennen muss.
		{name: "soft-line", tag: "soft-line", warum: "von diesem Programm gedeutet"},
		// **`margin` seit #433
		// gedeutet.**
		{name: "margin", tag: "margin:130:0:0", warum: "von diesem Programm gedeutet"},
		// **`li` und `bullet` seit
		// #431
		// gedeutet** — `li` warnt bei MSE nur, `bullet` legt seine Zeichen in
		// den Rand.
		{name: "li", tag: "li", warum: "von diesem Programm gedeutet"},
		{name: "bullet", tag: "bullet", warum: "von diesem Programm gedeutet"},
		{name: "align", tag: "align:right", gemeldet: true, warum: "Ausrichtung"},
		{name: "code", tag: "code", gemeldet: true, warum: "Quelltextsatz"},
		{name: "size", tag: "size:12", gemeldet: true, warum: "Schriftgröße"},

		// **Die Grenzfälle der Präfixregel.** `is_tag` verlangt hinter dem
		// Präfix ein `>`, `-`, `:` oder ein Leerzeichen (`is_tag_end_char`,
		// `src/util/tagged_string.cpp`, Zeilen 28–30 und 165–171).
		// **`line` deutet dieses Programm selbst** — es schneidet einen Block.
		// Es steht hier, weil die Präfixregel es von `li` trennen muss.
		//
		// Fassung f9d9356 vom 2021-07-06.
		{name: "line trifft li nicht", tag: "line", warum: "von diesem Programm gedeutet"},
		{name: "linear trifft weder li noch line", tag: "linear", warum: "n und a sind keine Endzeichen"},
		{name: "bold trifft b nicht", tag: "bold", warum: "o ist kein Endzeichen"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			befunde := befundeFuerText(t, "<"+fall.tag+">Text</"+fall.tag+">")

			gemeldet := false
			for _, befund := range befunde {
				if strings.Contains(befund.Meldung, "sind entfernt") {
					gemeldet = true
				}
			}

			if gemeldet != fall.gemeldet {
				t.Errorf("gemeldet %t, erwartet %t — %s (%+v)",
					gemeldet, fall.gemeldet, fall.warum, befunde)
			}
		})
	}
}

// befundeFuerText zeichnet eine Karte mit **einem** Textfeld und gibt, was
// dabei anfiel.
func befundeFuerText(t *testing.T, text string) []domain.Befund {
	t.Helper()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	_, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: text},
		}}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return befunde
}

// **`<soft-line>` erreicht den Satz**
// (#431).
//
// Bis hierher wurde die Auszeichnung gemeldet und weggeworfen; der Umbruch
// darin galt als hart und bekam `line height hard`.
func TestEineWeicheZeileErreichtDenSatz(t *testing.T) {
	abschnitte := textabschnitte(t, "a<soft-line>\n</soft-line>b", false)

	if len(abschnitte) != 3 {
		t.Fatalf("%d Abschnitte, erwartet drei: %+v", len(abschnitte), abschnitte)
	}

	// **Der mittlere trägt den Umbruch.** `<line>\n</line>` ist bei
	// `magic.mse-game` genauso gebaut, und die Zuordnung ist dieselbe.
	if !abschnitte[1].WeicheZeile {
		t.Errorf("der Abschnitt mit dem Umbruch ist nicht weich: %+v", abschnitte[1])
	}

	// **Und die anderen nicht** — die Gegenprobe: ohne sie bestünde der Test
	// auch, wenn jeder Abschnitt weich wäre.
	if abschnitte[0].WeicheZeile || abschnitte[2].WeicheZeile {
		t.Errorf("die Abschnitte daneben sind weich: %+v", abschnitte)
	}
}

// **`<margin:l:r:t>` erreicht den Satz — samt MSEs Fehlgriff**
// (#433).
//
// MSE zieht bei der Länge **2** ab, wo für ein durch `:` beendetes Feld **1**
// richtig wäre (`src/render/text/element.cpp`, Zeilen 181–183, Fassung f9d9356
// vom 2021-07-06). Die ersten beiden Zahlen verlieren damit ihr letztes
// Zeichen. Nachgebaut, nicht behoben — der Grund steht in Vermerk 0027.
func TestEinAbsatzrandErreichtDenSatz(t *testing.T) {
	for _, fall := range []struct {
		name     string
		tag      string
		erwartet render.Margen
	}{
		{
			name: "die Angabe aus dem Pack", tag: "margin:130:0:0",
			// 130 → „13", 0 → „" (bleibt null), 0 → „0".
			erwartet: render.Margen{Links: 13},
		},
		{
			name: "zweistellig", tag: "margin:20:30:5",
			erwartet: render.Margen{Links: 2, Rechts: 3, Oben: 5},
		},
		{
			name:     "einstellig verliert alles",
			tag:      "margin:9:8:7",
			erwartet: render.Margen{Oben: 7},
		},
		{
			// **Leere Felder sind null**, wie `ToDouble` auf einer leeren
			// Zeichenkette: es meldet einen Fehlschlag und lässt den Wert
			// stehen.
			name: "leere Felder", tag: "margin:::",
			erwartet: render.Margen{},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			abschnitte := textabschnitte(t, "a<"+fall.tag+">b</margin>", false)

			var gefunden *render.Margen

			for i := range abschnitte {
				if abschnitte[i].Margen != nil {
					gefunden = abschnitte[i].Margen
				}
			}

			if gefunden == nil {
				t.Fatalf("kein Abschnitt trägt Ränder: %+v", abschnitte)
			}

			if *gefunden != fall.erwartet {
				t.Errorf("Ränder %+v, erwartet %+v", *gefunden, fall.erwartet)
			}
		})
	}
}

// **Eine unvollständige Angabe setzt keinen Rand.**
//
// MSE prüft `colon < pos - 1 && text.GetChar(colon) == ':'` und lässt den
// Block sonst ganz aus (`src/render/text/element.cpp`, Zeilen 178–180) — der
// Absatz behält dann die Ränder, die er hatte.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEineUnvollstaendigeRandangabeSetztNichts(t *testing.T) {
	for _, tag := range []string{"margin", "margin:1:2"} {
		t.Run(tag, func(t *testing.T) {
			for _, abschnitt := range textabschnitte(t, "a<"+tag+">b</margin>", false) {
				if abschnitt.Margen != nil {
					t.Errorf("%q setzt Ränder %+v", tag, *abschnitt.Margen)
				}
			}
		})
	}
}

// **Geschachtelte Ränder addieren sich.**
//
//	if (!margins.empty()) {
//	  m.left  += margins.back().left;
//	  m.right += margins.back().right;
//	  m.top   += margins.back().top;
//	}
//
// `src/render/text/element.cpp`, Zeilen 186–190, Fassung f9d9356 vom
// 2021-07-06. Und das schließende `</margin>` legt genau **eine** Lage ab.
func TestGeschachtelteRaenderAddierenSich(t *testing.T) {
	// `20`→2, `30`→3, `5`→5; innen `40`→4, `50`→5, `1`→1.
	abschnitte := textabschnitte(t,
		"a<margin:20:30:5>b<margin:40:50:1>c</margin>d</margin>e", false)

	nach := map[string]render.Margen{}

	for i := range abschnitte {
		if abschnitte[i].Margen != nil {
			nach[abschnitte[i].Text] = *abschnitte[i].Margen
		}
	}

	if innen := nach["c"]; innen != (render.Margen{Links: 6, Rechts: 8, Oben: 6}) {
		t.Errorf("innen %+v, erwartet {6 8 6}", innen)
	}

	// **Und nach dem schließenden ist der äußere wieder da** — die Gegenprobe:
	// ohne sie bestünde der Test auch, wenn der Stapel nie abgelegt würde.
	if aussen := nach["d"]; aussen != (render.Margen{Links: 2, Rechts: 3, Oben: 5}) {
		t.Errorf("nach dem Schließen %+v, erwartet {2 3 5}", aussen)
	}

	if _, da := nach["e"]; da {
		t.Errorf("nach dem zweiten Schließen steht noch ein Rand: %+v", nach["e"])
	}
}

// **`<bullet>` erreicht den Satz**
// (#431).
//
// `magic.mse-game` klammert jede Zeile, die mit „• " beginnt, in
// `<li><bullet>• </bullet>…</li>` (`script`, Zeilen 1943–1946) — jeder modale
// Zauber trägt das.
func TestEinAufzaehlungspunktErreichtDenSatz(t *testing.T) {
	abschnitte := textabschnitte(t, "<li><bullet>• </bullet>Text</li>", false)

	if len(abschnitte) != 2 {
		t.Fatalf("%d Abschnitte, erwartet zwei: %+v", len(abschnitte), abschnitte)
	}

	if !abschnitte[0].ImRand || abschnitte[0].Text != "• " {
		t.Errorf("der Punkt steht nicht im Rand: %+v", abschnitte[0])
	}

	// **Und der Text daneben nicht** — die Gegenprobe.
	if abschnitte[1].ImRand {
		t.Errorf("der Text steht im Rand: %+v", abschnitte[1])
	}
}
