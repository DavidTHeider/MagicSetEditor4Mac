package service

import (
	"errors"
	"regexp"
	"strings"
	"sync"
)

// **Ein Symbolcode darf ein regulärer Ausdruck sein — und MSEs Ausdrücke
// kennen Vorausschau.**
//
// `magic-mana-large` trägt 32 solcher Codes, im ganzen Pack sind es 73
// verschiedene Muster. Die Bauform ist immer dieselbe: eine Alternative aus
// Zweigen, und ein Zweig darf auf eine **negative Vorausschau** enden —
// `10(?!/)` heißt „die Zehn, aber nicht die in `10/W`".
//
// Gos `regexp` kennt keine Vorausschau, nach Absicht: RE2 garantiert lineare
// Laufzeit und lässt dafür Rückverweise und Vorausschau weg. Eine Bibliothek
// dafür wäre eine Abhängigkeit für 73 Muster, von denen **alle** dieselbe
// Sonderform haben.
//
// **Deshalb wird die Vorausschau zerlegt statt übersetzt:** der Zweig wird an
// seiner abschließenden `(?!…)` getrennt, der Kopf mit `regexp` geprüft und
// die Vorausschau anschließend als eigene Frage an den Rest gestellt. Was
// diese Form nicht trifft — eine Vorausschau mitten im Zweig, eine positive —,
// wird **benannt** und nicht heimlich falsch gelesen.
//
// Gelesen: `SymbolFont::split`, Zeilen 196–216, Fassung f9d9356 vom
// 2021-07-06 (#293).

// symbolmuster ist ein Symbolcode, der als Ausdruck gelesen wird.
type symbolmuster struct{ zweige []musterzweig }

// musterzweig ist ein Zweig der Alternative samt seiner Vorausschau.
type musterzweig struct {
	// kopf ist der Teil, der wirklich Text deckt — verankert am Anfang.
	kopf *regexp.Regexp

	// verbot darf **hinter** dem Treffer nicht greifen. Nil heißt „keine
	// Vorausschau".
	verbot *regexp.Regexp
}

// errVorausschau meldet eine Vorausschau, die diese Zerlegung nicht trifft.
var errVorausschau = errors.New("eine Vorausschau, die nicht am Zweigende steht")

// gemerkteMuster hält die schon übersetzten Ausdrücke.
//
// **Ein Muster ist eine reine Funktion seines Textes.** Eine Symbolschrift
// wird je Karte gelesen und trägt bis zu 32 Ausdrücke; sie bei jeder Karte neu
// zu übersetzen wäre Arbeit für dasselbe Ergebnis. Ein gescheitertes Muster
// wird als `nil` gemerkt — der zweite Versuch scheitert sonst genauso teuer
// wie der erste.
var gemerkteMuster sync.Map

// muster liefert das gelesene Muster zu einem Code.
func muster(code string) (*symbolmuster, error) {
	if fertig, da := gemerkteMuster.Load(code); da {
		gelesen, brauchbar := fertig.(*symbolmuster)
		if !brauchbar || gelesen == nil {
			return nil, errVorausschau
		}

		return gelesen, nil
	}

	gelesen, err := gelesenesMuster(code)
	if err != nil {
		gemerkteMuster.Store(code, (*symbolmuster)(nil))

		return nil, err
	}

	gemerkteMuster.Store(code, gelesen)

	return gelesen, nil
}

// gelesenesMuster übersetzt einen Code in seine Zweige.
func gelesenesMuster(code string) (*symbolmuster, error) {
	zweige := make([]musterzweig, 0, 2)

	for _, roh := range zweigeVon(code) {
		kopf, verbot, err := getrenntNachVorausschau(roh)
		if err != nil {
			return nil, err
		}

		gelesen, err := regexp.Compile("^(?:" + kopf + ")")
		if err != nil {
			return nil, err
		}

		zweig := musterzweig{kopf: gelesen}

		if verbot != "" {
			gelesen, err := regexp.Compile("^(?:" + verbot + ")")
			if err != nil {
				return nil, err
			}

			zweig.verbot = gelesen
		}

		zweige = append(zweige, zweig)
	}

	return &symbolmuster{zweige: zweige}, nil
}

// treffer sagt, wie viele Bytes am Anfang von `rest` dieses Muster deckt.
//
// **Der längste Zweig gewinnt.** MSE prüft mit `wxRegEx`, und das ist ein
// POSIX-Ausdruck: dort gilt „leftmost-longest", nicht „der erste Zweig, der
// passt". Bei `[0-9]+(?!/…)|.` ist das der Unterschied zwischen `12` und `1`.
// **Und die Gruppen des siegreichen Zweigs.** `draw text:` nennt eine
// Gruppennummer, und ihr Treffer wird auf das Bild geschrieben — bei
// `([XYZ0-9])/W` ist das die Ziffer im Kreis
// (#336).
func (m *symbolmuster) treffer(rest string) (laenge int, gruppen []string, passt bool) {
	for _, zweig := range m.zweige {
		stellen := zweig.kopf.FindStringSubmatchIndex(rest)
		if len(stellen) < 2 || stellen[1] == 0 {
			continue
		}

		if zweig.verbot != nil && zweig.verbot.MatchString(rest[stellen[1]:]) {
			continue
		}

		if stellen[1] <= laenge {
			continue
		}

		laenge = stellen[1]
		gruppen = gruppenAus(rest, stellen)
	}

	return laenge, gruppen, laenge > 0
}

// gruppenAus liest die Trefferzeichenketten — Gruppe 0 ist der ganze Treffer.
//
// **Eine Gruppe, die nicht getroffen hat, ist leer** und kein Fehler: MSE liest
// sie über `results.str(i)`, und das gibt dort eine leere Zeichenkette.
func gruppenAus(rest string, stellen []int) []string {
	gruppen := make([]string, 0, len(stellen)/2)

	for paar := 0; paar+1 < len(stellen); paar += 2 {
		if stellen[paar] < 0 {
			gruppen = append(gruppen, "")

			continue
		}

		gruppen = append(gruppen, rest[stellen[paar]:stellen[paar+1]])
	}

	return gruppen
}

// zweigeVon teilt eine Alternative an ihren **äußeren** Strichen.
//
// Ein `|` in einer Klammer oder in einer Zeichenklasse trennt keinen Zweig —
// `L/(\+[XYZ]|[0-9])` ist ein Zweig, nicht zwei.
func zweigeVon(code string) []string {
	var (
		zweige  []string
		laufend strings.Builder
		tiefe   int
		klasse  bool
	)

	runen := []rune(code)

	for stelle := 0; stelle < len(runen); stelle++ {
		zeichen := runen[stelle]

		switch {
		case zeichen == '\\' && stelle+1 < len(runen):
			laufend.WriteRune(zeichen)
			stelle++
			laufend.WriteRune(runen[stelle])

			continue

		case klasse:
			if zeichen == ']' {
				klasse = false
			}

		case zeichen == '[':
			klasse = true

		case zeichen == '(':
			tiefe++

		case zeichen == ')':
			tiefe--

		case zeichen == '|' && tiefe == 0:
			zweige = append(zweige, laufend.String())
			laufend.Reset()

			continue
		}

		laufend.WriteRune(zeichen)
	}

	return append(zweige, laufend.String())
}

// getrenntNachVorausschau schneidet eine abschließende `(?!…)` ab.
//
// Der zweite Rückgabewert ist der Ausdruck der Vorausschau, leer wenn keine da
// ist. Steht eine Vorausschau **nicht** am Ende, ist das ein Fehler — und kein
// stillschweigendes Weglassen: ein Muster, das die Hälfte seiner Bedingung
// verliert, trifft zu oft, und das sähe aus wie ein Treffer.
func getrenntNachVorausschau(zweig string) (kopf, verbot string, err error) {
	auf := strings.Index(zweig, "(?!")
	if auf < 0 {
		return zweig, "", nil
	}

	if !strings.HasSuffix(zweig, ")") {
		return "", "", errVorausschau
	}

	// **Die Klammer muss bis zum Zweigende reichen.** Bei `a(?!b)c` steht die
	// Vorausschau mittendrin; sie abzuschneiden ergäbe `a` mit dem Verbot `b`
	// und verlöre das `c`.
	tiefe := 0

	for stelle, zeichen := range zweig[auf:] {
		switch zeichen {
		case '(':
			tiefe++
		case ')':
			tiefe--
		}

		if tiefe == 0 {
			if auf+stelle != len(zweig)-1 {
				return "", "", errVorausschau
			}

			break
		}
	}

	if tiefe != 0 {
		return "", "", errVorausschau
	}

	return zweig[:auf], zweig[auf+len("(?!") : len(zweig)-1], nil
}
