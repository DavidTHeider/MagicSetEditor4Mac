package msescript

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// **Eine Zeichenkette darf Ausdrücke enthalten** — `"a{e}b"`.
//
// MSEs Zerteiler zerlegt sie in drei Marken und setzt sie mit `+` wieder
// zusammen:
//
//	// for smart strings: "x" {{ e }} "y"
//	"a{e}b"  -->  "a"  "{  e  }"  "b"
//
// `src/script/parser.cpp`, Zeilen 302–307 und 842–860, Fassung f9d9356 vom
// 2021-07-06. Aus `"a{e}b"` wird `"a" + e + "b"`.
//
// **`magic.mse-game` baut damit seine Auszeichnungen** —
// `"<{tag}>{input}</{tag}>"` und
// `"<nospellcheck><key>{keyword}</key></nospellcheck>"`. Ohne die Zerlegung
// stand die geschweifte Klammer wörtlich auf der Karte
// (#345).
//
// # Warum das hier steht und nicht im Zerteiler für Marken
//
// MSE löst es beim **Lesen** der Marke: der Leser bricht an der Klammer ab und
// gibt drei Marken heraus. Dieses Programm liest eine Zeichenkette als **eine**
// Marke; die Zerlegung geschieht deshalb hier, wo aus der Marke ein Knoten
// wird. Das Ergebnis ist dasselbe — eine Kette aus `+`.

// textbaustein macht aus einer Zeichenkettenmarke einen Knoten.
//
// Ohne eingebetteten Ausdruck ist das ein Blatt; sonst eine Kette aus `+`.
func (z *zerteiler) textbaustein(merkmal Merkmal) (*Knoten, error) {
	teile := textteile(merkmal.Text)
	if len(teile) == 1 && !teile[0].ausdruck {
		return blatt(FormZeichenkette, Merkmal{
			Text:  ohneKlammerschutz(teile[0].text),
			Zeile: merkmal.Zeile, Spalte: merkmal.Spalte,
		}), nil
	}

	var kette *Knoten

	for _, teil := range teile {
		glied, err := z.textglied(teil, merkmal)
		if err != nil {
			return nil, err
		}

		if kette == nil {
			kette = glied

			continue
		}

		kette = &Knoten{
			Form: FormBinaer, Text: "+", Kinder: []*Knoten{kette, glied},
			Zeile: merkmal.Zeile, Spalte: merkmal.Spalte,
		}
	}

	return kette, nil
}

// textglied macht aus einem Teil den Knoten, der ihn beisteuert.
func (z *zerteiler) textglied(teil textteil, merkmal Merkmal) (*Knoten, error) {
	if !teil.ausdruck {
		return blatt(FormZeichenkette, Merkmal{
			Text:  ohneKlammerschutz(teil.text),
			Zeile: merkmal.Zeile, Spalte: merkmal.Spalte,
		}), nil
	}

	// **Der eingebettete Ausdruck wird für sich zerteilt.** Er ist ein Skript
	// wie jedes andere; ihn hier von Hand zu lesen hieße, den Zerteiler ein
	// zweites Mal zu schreiben.
	//
	// **Mit der Tiefe, in der er steht** (BK-08): ein eigener Zerteiler mit
	// frischer Zählung ließe `"{"{"{…}"}"}"` an der Grenze vorbei, eine Ebene
	// je Zeichenkette. Zu tief ist kein Text, sondern ein Befund — an der
	// Stelle der Zeichenkette, in der es geschah.
	knoten, fehler := z.eingebettet(teil.text)
	if befund := (*Fehler)(nil); errors.As(fehler, &befund) && befund.Schluessel == SchluesselZuTiefGeschachtelt {
		return nil, NeuerFehler(merkmal.Zeile, merkmal.Spalte, SchluesselZuTiefGeschachtelt,
			strconv.Itoa(HoechsteSchachtelung))
	}

	if fehler != nil {
		// **Was sich nicht zerteilen lässt, ist Text.**
		//
		// Eine geschweifte Klammer steht auch dort, wo niemand einen Ausdruck
		// meint: `replace(…, replace: "\x{20AC}")` schreibt eine
		// Ersatzzeichenfolge im sed-Stil, und `{20AC}` ist darin eine Angabe
		// für den Mustervergleich. MSE unterscheidet die beiden gar nicht —
		// es liest jede Klammer als Ausdruck und scheitert an solchen Stellen.
		//
		// Hier fällt sie auf Text zurück, und das ist eine **Abweichung mit
		// Ansage**: sie hält eine Vorlage am Leben, die MSE abweisen würde,
		// und sie kann eine echte Verwechslung verdecken. Der Preis ist
		// geringer als der Nutzen — 212 Stellen im Pack brauchen die
		// Einbettung, und keine davon sieht wie ein Muster aus
		// (#345).
		// **Und dann auch ganz**: `\{` schützt eine Klammer und fällt dabei
		// weg (`docs/FORMATE.md`). Ohne diesen Schritt behielte gerade der
		// Auffangfall den Rückstrich, den jeder andere Weg entfernt
		// (#468).
		//
		//nolint:nilerr // ein unlesbarer Ausdruck ist Text, siehe darüber
		return blatt(FormZeichenkette, Merkmal{
			Text:  ohneKlammerschutz("{" + teil.text + "}"),
			Zeile: merkmal.Zeile, Spalte: merkmal.Spalte,
		}), nil
	}

	return knoten, nil
}

// eingebettet zerteilt einen eingebetteten Ausdruck mit einem eigenen
// Zerteiler, der die Tiefe dieses erbt.
//
// **Der Fehler ist hier meist die Antwort und keine Störung**: was sich nicht
// zerteilen lässt, war kein Ausdruck ([zerteiler.textglied] macht daraus
// Text) — außer, er war zu tief geschachtelt.
func (z *zerteiler) eingebettet(quelle string) (*Knoten, error) {
	merkmale, err := Zerlege(quelle)
	if err != nil {
		return nil, err
	}

	unter := &zerteiler{merkmale: merkmale, ende: endstelle(merkmale), tiefe: z.tiefe}

	return unter.folge(nil)
}

// textteil ist ein Stück einer Zeichenkette — Text oder Ausdruck.
type textteil struct {
	text     string
	ausdruck bool
}

// textteile zerlegt eine Zeichenkette an ihren **ungeschützten** Klammern.
//
// **Eine Klammer ohne Ende ist Text.** Ein Kartentext darf `{` enthalten —
// `{T}` steht in jedem Regeltext, und eine halbe Klammer soll ihn nicht
// verschlucken.
//
// **Anführungszeichen im Ausdruck zählen mit.** `"{if x then "a" else "b"}"`
// ist eine Zeichenkette, deren Ausdruck zwei weitere enthält; wer die
// schließende Klammer ohne diese Buchhaltung suchte, fände die falsche.
func textteile(text string) []textteil {
	var (
		teile   []textteil
		laufend strings.Builder
	)

	runen := []rune(text)

	for stelle := 0; stelle < len(runen); stelle++ {
		if runen[stelle] == '\\' && stelle+1 < len(runen) {
			laufend.WriteRune(runen[stelle])
			stelle++
			laufend.WriteRune(runen[stelle])

			continue
		}

		if runen[stelle] != '{' {
			laufend.WriteRune(runen[stelle])

			continue
		}

		ende, gefunden := klammerende(runen, stelle)
		if !gefunden || nachMusterflucht(laufend.String()) {
			laufend.WriteRune(runen[stelle])

			continue
		}

		teile = append(teile, textteil{text: laufend.String()})
		laufend.Reset()

		teile = append(teile, textteil{text: string(runen[stelle+1 : ende]), ausdruck: true})
		stelle = ende
	}

	return append(teile, textteil{text: laufend.String()})
}

// nachMusterflucht sagt, ob vor der Klammer eine Fluchtfolge steht.
//
// **`\x{20AC}` ist keine Einbettung, sondern ein Muster.** Der Ersatztext von
// `replace` folgt sed: `\x{DDDD}` nennt ein Zeichen über seine Nummer, und die
// Klammern gehören dazu. Ohne diese Frage würde `{20AC}` als Ausdruck gelesen —
// und `{110000}` sogar erfolgreich, als Zahl.
//
// **Eng gefasst**: nur ein Rückstrich mit **einem** Buchstaben dahinter. Das
// trifft `\x` und `\c` und nichts, was wie ein Ausdruck aussieht
// (#345).
func nachMusterflucht(bisher string) bool {
	runen := []rune(bisher)
	if len(runen) < 2 {
		return false
	}

	return runen[len(runen)-2] == '\\' &&
		((runen[len(runen)-1] >= 'a' && runen[len(runen)-1] <= 'z') ||
			(runen[len(runen)-1] >= 'A' && runen[len(runen)-1] <= 'Z'))
}

// klammerende sucht die schließende Klammer zu der an `auf`.
//
// **Zeichenketten werden übersprungen**
// (#468). Ein
// Anführungszeichen steht hier sehr wohl: seit der Leser die geschweiften
// Klammern mitzählt, kommt eine **verschachtelte** Zeichenkette roh in der
// Marke an — `"a{f("}")}b"`. Wer nur die Klammern zählte, fände die `}` in
// `"}"` und schnitte den Ausdruck mittendurch.
func klammerende(runen []rune, auf int) (int, bool) {
	tiefe := 0

	for stelle := auf; stelle < len(runen); stelle++ {
		switch runen[stelle] {
		case '\\':
			// **Eine Fluchtfolge nimmt ihr Zeichen mit** — `\"` beginnt keine
			// Zeichenkette und `\{` keine Einbettung.
			stelle++

		case '"':
			schluss, geschlossen := zeichenkettenende(runen, stelle)
			if !geschlossen {
				return 0, false
			}

			stelle = schluss

		case '{':
			tiefe++

		case '}':
			tiefe--
			if tiefe == 0 {
				return stelle, true
			}
		}
	}

	return 0, false
}

// zeichenkettenende sucht das Anführungszeichen, das die Zeichenkette an
// `auf` schließt.
func zeichenkettenende(runen []rune, auf int) (int, bool) {
	for stelle := auf + 1; stelle < len(runen); stelle++ {
		if runen[stelle] == '\\' && stelle+1 < len(runen) {
			stelle++

			continue
		}

		if runen[stelle] == '"' {
			return stelle, true
		}
	}

	return 0, false
}

// ohneKlammerschutz macht aus `\{` ein `{` und aus `\}` ein `}`.
//
// **Beide Klammern** — MSE nimmt beide:
//
//	else if (c == '\\' || c == '"' || c == '\'' || c == '{' || c == '}') {
//	  str += c; // don't warn about these, since they look escape-worthy
//
// `src/script/parser.cpp`, Zeile 287, Fassung f9d9356 vom 2021-07-06.
//
// `magic.mse-game` schreibt seine Parameterverweise als `\{cmc({input})\}`
// (#1342).
// **Gemessen am 2026-08-31** über die Datendateien des Packs: 337-mal `\{`,
// **253-mal `\}`**, in 17 Dateien.
//
// **Nur die geschweiften Klammern.** Die übrigen Schutzzeichen liest der Leser
// der Marke; hier geht es allein um die, die dieses Zerlegen erst nötig machen.
func ohneKlammerschutz(text string) string {
	return strings.NewReplacer(`\{`, "{", `\}`, "}").Replace(text)
}

// AuswertenTextskript wertet einen **Textskript** aus — eine Zeichenkette mit
// eingebetteten Ausdrücken, ohne umschließende Anführungszeichen.
//
// # Wofür
//
// MSE nennt das `StringScript`. Ein `reminder:` eines Schlüsselworts ist einer
// (`src/data/keyword.hpp`, Zeile 95, Fassung f9d9356 vom 2021-07-06), ebenso
// der `default:` eines Kartenfelds. Der Wert steht **roh** in der Paketdatei:
//
//	reminder: {this_or_that("creature", upper:true)} can’t be blocked …
//
// # Warum nicht einfach in Anführungszeichen setzen und zerteilen
//
// **Weil ein Anführungszeichen darin vorkommt.** Der Erinnerungstext oben
// trägt zwei, und in Anführungszeichen gesetzt endete die Zeichenkette an der
// falschen Stelle. MSE hat dasselbe Problem und löst es genauso: es zerteilt
// im **Zeichenkettenmodus**, in dem ein `"` ein Zeichen ist wie jedes andere.
//
// **Ein leerer Textskript ist eine leere Zeichenkette**, kein Fehler — MSE
// gibt für einen nicht gesetzten `StringScript` ebenfalls `""`.
func AuswertenTextskript(
	ctx context.Context, quelle string, umgebung *Umgebung,
) (Wert, error) {
	if quelle == "" {
		return Text(""), nil
	}

	var (
		leser   zerteiler
		merkmal = Merkmal{Text: quelle, Zeile: 1, Spalte: 1}
	)

	// **Scheitert nur an der Schachtelungsgrenze** (BK-08, Vermerk 0177):
	// was sich als Ausdruck nicht lesen lässt, wird Text (siehe
	// [zerteiler.textglied]) — eine zu tiefe Schachtelung nicht. Hier stand
	// „kann nicht scheitern", und mit der Grenze wurde aus dem verworfenen
	// Fehler ein leerer Knoten und ein `panic` (Zweitblick über E, Befund 1).
	knoten, fehler := leser.textbaustein(merkmal)
	if fehler != nil {
		return Nichts, fehler
	}

	return Auswerten(ctx, knoten, umgebung)
}
