package msescript

import (
	"context"
	"strconv"
	"strings"
)

// **`replace:` darf eine Funktion sein — dann wird sie je Treffer gerufen.**
//
//	if (replacement_function) {
//	  // set match results in context
//	  for (UInt sub = 0 ; sub < results.size() ; ++sub) {
//	    String name  = sub == 0 ? _("input") : String(_("_")) << sub;
//	    ctx.setVariable(name, to_script(results.str(sub)));
//	  }
//	  inside = replacement_function->eval(ctx)->toString();
//	}
//
// `RegexReplacer::apply`, `src/script/functions/regex.cpp`, Zeilen 78–99,
// Fassung f9d9356 vom 2021-07-06.
//
// **Der ganze Treffer heißt `input`, die Gruppen `_1` bis `_9`.**
//
// # Was das gekostet hat
//
// Ohne diesen Weg wurde die Funktion nach der einzigen Regel behandelt, die es
// gab — in Text verwandelt. Auf *Firebrand Incinerator* stand damit
// `Creature — Goblinfunktion<class-en>…` statt `Creature — Goblin Mage`, und
// im Regeltext `deal Funktion damage` statt `deal 2 damage`
// (#345).

// ErsetzeMitFunktion ruft für jeden Treffer die Funktion.
//
// **Rekursiv, wenn `recursive:` es sagt**, und dann höchstens zwanzig Ebenen
// tief — `if (recursive && level < 20)`. MSE bricht dort ab und nimmt das
// Ergebnis, wie es ist: keine Meldung, keine Schleife.
func (m *Muster) ErsetzeMitFunktion(
	ctx context.Context, text string, umgebung *Umgebung,
	je func(context.Context, *Umgebung) (Wert, error), rekursiv bool,
) (string, error) {
	return m.ersetzeMitFunktion(ctx, text, umgebung, je, rekursiv, 0)
}

// Rekursionsgrenze ist MSEs `level < 20`.
const Rekursionsgrenze = 20

func (m *Muster) ersetzeMitFunktion(
	ctx context.Context, text string, umgebung *Umgebung,
	je func(context.Context, *Umgebung) (Wert, error), rekursiv bool, ebene int,
) (string, error) {
	var (
		gebaut  string
		gelesen int
		zeichen = []rune(text)
		fehler  error
	)

	setzen := func(index, laenge int, gruppen []string) {
		if fehler != nil {
			return
		}

		gebaut += string(zeichen[gelesen:index])
		gelesen = index + laenge

		innen, err := m.einTreffer(ctx, umgebung, je, gruppen, rekursiv, ebene)
		if err != nil {
			fehler = err

			return
		}

		gebaut += innen
	}

	if err := m.jederTrefferMitNeustart(text, setzen); err != nil {
		return "", err
	}

	if fehler != nil {
		return "", fehler
	}

	return gebaut + string(zeichen[gelesen:]), nil
}

// ErsetzeRekursiv ist `replace` mit `recursive: true` und einem Text: die
// Vorlage wird **aus den Gruppen des Treffers** gefüllt, und erst ihr Ergebnis
// geht in die Tiefe — `results.format(replacement_string)`, dann
// `apply(ctx, inside, level + 1)` (`src/script/functions/regex.cpp`, Zeilen
// 96–104, Fassung f9d9356 vom 2021-07-06).
//
// **Bis zum Nachlauf über Phase 1 von 0.9.0 (Befund 7) wurde der Treffer für
// sich noch einmal durchsucht**, und ein Rückblick verlor dabei, was davor
// stand: aus `ab` mit `(?<=a)b` wurde nichts statt `a#`.
func (m *Muster) ErsetzeRekursiv(text, vorlage string) (string, error) {
	return m.ersetzeRekursiv(text, vorlage, 0)
}

func (m *Muster) ersetzeRekursiv(text, vorlage string, ebene int) (string, error) {
	var (
		gebaut  strings.Builder
		gelesen int
		zeichen = []rune(text)
		fehler  error
	)

	setzen := func(index, laenge int, gruppen []string) {
		if fehler != nil {
			return
		}

		gebaut.WriteString(string(zeichen[gelesen:index]))
		gelesen = index + laenge

		innen := fuelle(vorlage, func(nummer int) string {
			if nummer >= len(gruppen) {
				return ""
			}

			return gruppen[nummer]
		})

		if ebene < Rekursionsgrenze {
			innen, fehler = m.ersetzeRekursiv(innen, vorlage, ebene+1)
		}

		gebaut.WriteString(innen)
	}

	if err := m.jederTrefferMitNeustart(text, setzen); err != nil {
		return "", err
	}

	if fehler != nil {
		return "", fehler
	}

	gebaut.WriteString(string(zeichen[gelesen:]))

	return gebaut.String(), nil
}

// einTreffer bindet die Gruppen und ruft die Funktion.
func (m *Muster) einTreffer(
	ctx context.Context, umgebung *Umgebung,
	je func(context.Context, *Umgebung) (Wert, error),
	gruppen []string, rekursiv bool, ebene int,
) (string, error) {
	// **Eine eigene Ebene je Treffer.** MSE setzt die Variablen in den
	// laufenden Kontext; hier bleibt die Umgebung des Aufrufers unberührt, und
	// das ist der Unterschied zwischen „`input` gilt für diesen Ruf" und
	// „`input` gilt ab jetzt".
	darueber := umgebung.Darueber()

	for stelle, gruppe := range gruppen {
		name := "input"
		if stelle > 0 {
			name = "_" + strconv.Itoa(stelle)
		}

		darueber.Setze(name, Text(gruppe))
	}

	wert, err := je(ctx, darueber)
	if err != nil {
		return "", err
	}

	innen := wert.AlsText()

	if rekursiv && ebene < Rekursionsgrenze {
		return m.ersetzeMitFunktion(ctx, innen, umgebung, je, rekursiv, ebene+1)
	}

	return innen, nil
}
