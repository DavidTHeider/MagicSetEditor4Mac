package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`default:` schlägt `initial:`** — der Vorgabeeintrag einer Stiloption
// (#590).
//
// # Nachgelesen, nicht erschlossen
//
//	ColorValue::ColorValue(const ColorFieldP& field)
//	  : Value(field)
//	  , value( !field->initial.isDefault() ? field->initial()
//	         : !field->choices.empty()     ? field->choices[0]->color
//	         :                               Color()
//	         , true)
//	{}
//
//	bool ColorValue::update(Context& ctx) {
//	  bool change = field().default_script.invokeOnDefault(ctx, value)
//	              | field().        script.invokeOn(ctx, value);
//
// `src/data/field/color.cpp`, Zeilen 82–101, Fassung f9d9356 vom 2021-07-06.
//
// **Das zweite Argument von `Defaultable` ist `true`**: der Konstruktor setzt
// `initial` und markiert den Wert zugleich als *Vorgabe*. `invokeOnDefault`
// läuft genau dann (`src/script/scriptable.hpp`, Zeilen 71–78) — `default:`
// überschreibt `initial:` also, solange niemand die Option angefasst hat.
//
// # Und `default:` ist ein Skript, keine Zeichenkette
//
// `OptionalScript default_script` (`color.hpp`, Zeile 33). Am Orakel
// nachgefragt (`magicseteditor --cli … -q --raw`):
//
//	"black"      ->  "black"
//	rgb(0,0,0)   ->  rgb(0,0,0)
//	black        ->  ERROR: Variable not set: black
//
// **Die drei Formen sind nicht ausgedacht, sondern gezählt.** Über die 376
// Stilpakete von `full-magic-pack` tragen **zehn** Stiloptionen ein `default:`,
// alle vom Typ `color`: achtmal `"black"`, einmal `rgb(0,0,0)`, einmal das
// nackte `black` — und das letzte steht in `magic-m15-bigtext` neben einem
// `initial: black`, das es auffängt.
func TestDieVorgabeEinerStiloption(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		feld     domain.StyleField
		erwartet string
	}{{
		name:     "eine Zeichenkette als Vorgabe — achtmal im Pack",
		feld:     domain.StyleField{Kind: domain.KindColor, Default: `"black"`},
		erwartet: "black",
	}, {
		// **`rgba(…)` und nicht `rgb(…)`**: die Textform ist die des
		// Auswerters ([msescript.Wert.AlsText]), und `farbeAus` liest genau
		// sie. Eine zweite Schreibweise daneben liefe früher oder später
		// auseinander.
		name:     "ein Funktionsaufruf als Vorgabe — einmal im Pack",
		feld:     domain.StyleField{Kind: domain.KindColor, Default: "rgb(0,0,0)"},
		erwartet: "rgba(0,0,0,255)",
	}, {
		name: "eine Vorgabe, die scheitert, lässt initial stehen",
		feld: domain.StyleField{
			Kind: domain.KindColor, Default: "black", Initial: "black",
		},
		erwartet: "black",
	}, {
		name:     "die Vorgabe schlägt initial",
		feld:     domain.StyleField{Kind: domain.KindColor, Default: `"white"`, Initial: "rgb(9,9,9)"},
		erwartet: "white",
	}, {
		// **`nil` wertet fehlerfrei aus und ergibt trotzdem nichts** — eine
		// Vorgabe, die keine ist. Sie darf `initial:` nicht verdrängen.
		name:     "eine Vorgabe, die zu nichts auswertet",
		feld:     domain.StyleField{Kind: domain.KindColor, Default: "nil", Initial: "rgb(7,7,7)"},
		erwartet: "rgb(7,7,7)",
	}, {
		// **Ein Block ist eine Funktion**, kein Wert — im Pack steht diese
		// Form an Kartenfeldern (`default: { nil }`). Als Farbe taugt sie
		// nicht.
		name:     "eine Vorgabe, die eine Funktion ist",
		feld:     domain.StyleField{Kind: domain.KindColor, Default: "{ nil }", Initial: "rgb(8,8,8)"},
		erwartet: "rgb(8,8,8)",
	}, {
		name:     "ohne default gilt initial wie bisher",
		feld:     domain.StyleField{Kind: domain.KindChoice, Initial: "hell", Choices: []string{"dunkel", "hell"}},
		erwartet: "hell",
	}, {
		name:     "ohne beides gilt die erste Wahl wie bisher",
		feld:     domain.StyleField{Kind: domain.KindChoice, Choices: []string{"dunkel", "hell"}},
		erwartet: "dunkel",
	}}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if wert := stiloptionsvorgabe(t.Context(), fall.feld); wert != fall.erwartet {
				t.Errorf("Vorgabe %q, erwartet %q", wert, fall.erwartet)
			}
		})
	}
}

// **Und die Textform muss als Farbe ankommen** — sie ist ein Zwischenschritt,
// kein Selbstzweck (§10: die Zusage dort prüfen, wo sie eingelöst wird).
//
// # Warum hier Weiß steht, wo das Pack Schwarz schreibt
//
// **Weil eine Prüfung auf Schwarz nichts hielte.** Der erste Anlauf nahm die
// Werte des Packs — `"black"` und `rgb(0,0,0)` — und war **auch ohne die
// Auswertung grün**: ohne sie gibt [stiloptionsvorgabe] die leere Zeichenkette,
// und `farbeAus("")` fällt auf deckendes Schwarz zurück (`farbdeutung.go`).
// Die Erwartung war also schon vor der Behebung erfüllt.
//
// Gefunden hat es die Gegenprobe, nicht das Lesen. Mit Weiß trennt der Fall:
// ohne die Auswertung kommt Schwarz an, mit ihr Weiß.
//
// **Beide Schreibweisen des Packs sind dabei**: eine Zeichenkette über die
// Namensliste, ein `rgb(…)` über `rgba(…)`.
func TestDieVorgabeKommtAlsFarbeAn(t *testing.T) {
	t.Parallel()

	weiss := render.Farbe{R: 255, G: 255, B: 255, A: 255}

	for _, quelle := range []string{`"white"`, "rgb(255,255,255)"} {
		t.Run(quelle, func(t *testing.T) {
			t.Parallel()

			vorgabe := stiloptionsvorgabe(t.Context(),
				domain.StyleField{Kind: domain.KindColor, Default: quelle})

			if farbe := farbeAus(vorgabe); farbe != weiss {
				t.Errorf("%q ergibt %+v, erwartet %+v", vorgabe, farbe, weiss)
			}
		})
	}
}
