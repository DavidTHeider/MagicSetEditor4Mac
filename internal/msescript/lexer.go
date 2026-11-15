package msescript

import (
	"unicode"
)

// Zerlege macht aus Quelltext eine Folge von Merkmalen.
//
// # Was hier bewusst nicht geschieht
//
// **Zeilen werden nicht verschluckt.** In einem `init script:` stehen mehrere
// Zuweisungen untereinander, getrennt allein durch den Umbruch — wer ihn
// wegwirft, kann sie danach nicht mehr auseinanderhalten. Der Zerteiler
// übergeht die überzähligen; das ist seine Aufgabe, nicht diese.
//
// **Ein Rückstrich wird durchgereicht, nicht gedeutet.** In `magic-m15` steht
// `contains(match:"\n", card.rule_text)`; ob `\n` dort ein Umbruch ist oder
// zwei Zeichen, entscheidet dieses eine Vorkommen nicht — `match:` ist anderswo
// in MSE ein Muster, und dort schriebe man den Umbruch genau so.
//
// Durchgereicht wird, weil das die **umkehrbare** Richtung ist: wer übersetzt,
// kann den Rückstrich nicht wiederherstellen. Steht in `docs/FORMATE.md` mit
// Testfall.
func Zerlege(quelle string) ([]Merkmal, error) {
	l := &lexer{zeichen: []rune(quelle), zeile: 1, spalte: 1}

	merkmale, err := l.lauf()
	if err != nil {
		return nil, err
	}

	return merkmale, nil
}

// Einfaerben gibt jedes Merkmal — **samt Kommentaren und so weit es geht**
// (#1164).
//
// # Wofür
//
// Für die Syntaxhervorhebung im Fenster. Sie braucht zwei Dinge, die dem
// Zerteiler zu Recht fehlen:
//
//   - **Kommentare.** Für ihn sind sie Zwischenraum.
//   - **Ein Ergebnis trotz Fehler.** Beim Tippen ist ein Ausdruck fast immer
//     unfertig; wer dann gar nichts einfärbt, lässt den Text bei jedem zweiten
//     Anschlag die Farbe verlieren.
//
// # Warum es derselbe Lexer ist
//
// Ein zweiter, der „ungefähr dieselben" Regeln kennt, liefe an dem Tag
// auseinander, an dem einer von beiden eine Regel lernt — und die Einfärbung
// zeigte dann eine Sprache, die es nicht gibt (§10). Der Unterschied ist ein
// Schalter, keine zweite Umsetzung.
func Einfaerben(quelle string) []Merkmal {
	l := &lexer{zeichen: []rune(quelle), zeile: 1, spalte: 1, kommentare: true}

	// **Der Fehler wird verworfen und nicht gemeldet.** Wo er steht, sagt
	// `Zerlege` — und der Aufrufer fragt ohnehin beides.
	merkmale, _ := l.lauf() //nolint:errcheck // siehe darüber

	return merkmale
}

type lexer struct {
	zeichen []rune
	stelle  int
	zeile   int
	spalte  int

	// kommentare heißt: `#` bis zum Zeilenende wird ein Merkmal statt
	// Zwischenraum. Gesetzt nur von [Einfaerben].
	kommentare bool
}

// operatoren stehen nach Länge sortiert: `:=` vor `:`, sonst zerfiele die
// Zuweisung in einen Doppelpunkt und ein Gleichheitszeichen.
var operatoren = []string{
	":=", "==", "!=", "<=", ">=",
	"+", "-", "*", "/", ".", "@", ",", ":", "<", ">",
	"(", ")", "{", "}", "[", "]",
}

func (l *lexer) lauf() ([]Merkmal, error) {
	var merkmale []Merkmal

	for l.stelle < len(l.zeichen) {
		zeichen := l.zeichen[l.stelle]

		switch {
		case zeichen == '\n':
			merkmale = append(merkmale, l.nimm(ArtZeilenende, "\n"))
			l.zeile++
			l.spalte = 1

		// **Das Semikolon trennt wie ein Umbruch.** Rund sechzig Stile eines
		// bezogenen Packs schreiben ihr `init script:` so — `{ a := 1; b := 2 }`
		// —, und ohne diese Zeile lief keines davon
		// (#186).
		//
		// Als **Zeilenende** und nicht als eigene Art: der Zerteiler trennt
		// seine Folge am Umbruch, und ein zweites Trennzeichen daneben hieße,
		// jede Stelle zu ändern, die eines überspringt.
		case zeichen == ';':
			merkmale = append(merkmale, l.nimm(ArtZeilenende, ";"))

		case zeichen == '\r' || zeichen == ' ' || zeichen == '\t':
			l.vor(1)

		case zeichen == '#':
			if merkmal, behalten := l.kommentar(); behalten {
				merkmale = append(merkmale, merkmal)
			}

		case zeichen == '"':
			merkmal, err := l.zeichenkette()
			if err != nil {
				// **Die schon gelesenen Merkmale reisen mit.** `Zerlege`
				// verwirft sie; [Einfaerben] färbt damit, so weit es geht.
				return merkmale, err
			}

			merkmale = append(merkmale, merkmal)

		case unicode.IsDigit(zeichen):
			merkmale = append(merkmale, l.zahl())

		case istNamensanfang(zeichen):
			merkmale = append(merkmale, l.name())

		default:
			merkmal, err := l.operator()
			if err != nil {
				return merkmale, err // siehe oben
			}

			merkmale = append(merkmale, merkmal)
		}
	}

	return merkmale, nil
}

// nimm baut ein Merkmal an der aktuellen Stelle und geht darüber hinweg.
func (l *lexer) nimm(art Art, text string) Merkmal {
	breite := len([]rune(text))
	merkmal := Merkmal{Art: art, Text: text, Zeile: l.zeile, Spalte: l.spalte, Breite: breite}
	l.stelle += breite
	l.spalte += breite

	return merkmal
}

func (l *lexer) vor(anzahl int) {
	l.stelle += anzahl
	l.spalte += anzahl
}

// kommentar überliest `#` bis zum Zeilenende — der Umbruch bleibt liegen.
// kommentar liest `#` bis zum Zeilenende.
//
// Er gibt ein Merkmal **und** ob es behalten wird: für den Zerteiler ist ein
// Kommentar Zwischenraum, für [Einfaerben] eine Farbe.
func (l *lexer) kommentar() (Merkmal, bool) {
	zeile, spalte, anfang := l.zeile, l.spalte, l.stelle

	for l.stelle < len(l.zeichen) && l.zeichen[l.stelle] != '\n' {
		l.vor(1)
	}

	if !l.kommentare {
		return Merkmal{}, false
	}

	text := string(l.zeichen[anfang:l.stelle])

	return Merkmal{
		Art: ArtKommentar, Text: text,
		Zeile: zeile, Spalte: spalte, Breite: l.stelle - anfang,
	}, true
}

// zahl liest Ziffern und höchstens einen Punkt **mit Ziffer dahinter**.
//
// Die Bedingung ist die ganze Feinheit: in `card.type_1` folgt auf den Punkt
// keine Ziffer, und `1.` gibt es in keinem gesichteten Paket. Ohne sie
// verschlänge eine Zahl den Punktzugriff, der ihr folgt.
func (l *lexer) zahl() Merkmal {
	ende := l.stelle
	for ende < len(l.zeichen) && unicode.IsDigit(l.zeichen[ende]) {
		ende++
	}

	if ende+1 < len(l.zeichen) && l.zeichen[ende] == '.' && unicode.IsDigit(l.zeichen[ende+1]) {
		ende++
		for ende < len(l.zeichen) && unicode.IsDigit(l.zeichen[ende]) {
			ende++
		}
	}

	return l.nimm(ArtZahl, string(l.zeichen[l.stelle:ende]))
}

func (l *lexer) name() Merkmal {
	ende := l.stelle
	for ende < len(l.zeichen) && istNamenszeichen(l.zeichen[ende]) {
		ende++
	}

	text := string(l.zeichen[l.stelle:ende])

	// Erst den **ganzen** Namen lesen, dann nachschlagen. Wer auf Präfix
	// vergleicht, zerlegt `notieren` in `not` und `ieren`.
	if Schluesselwoerter[text] {
		return l.nimm(ArtSchluesselwort, text)
	}

	return l.nimm(ArtBezeichner, text)
}

func (l *lexer) operator() (Merkmal, error) {
	// **Verglichen wird Rune für Rune, nicht über eine Zeichenkette.**
	//
	// Der erste Anlauf schrieb `string(l.zeichen[l.stelle:])` und prüfte mit
	// `strings.HasPrefix`. Das kostet je Merkmal die **ganze Restlänge** — über
	// eine Datei also ihr Quadrat. Bei den 582 KB, die `magic.mse-game` nach
	// dem Einbinden seiner Module mitbringt, waren das 19 Sekunden nur fürs
	// Zerteilen (#195). `TestDerZerteilerWaechstNichtQuadratisch` hält es fest.
	for _, zeichen := range operatoren {
		if l.beginntMit(zeichen) {
			return l.nimm(ArtOperator, zeichen), nil
		}
	}

	if l.zeichen[l.stelle] == '=' {
		return Merkmal{}, NeuerFehler(l.zeile, l.spalte, SchluesselEinzelnesGleich)
	}

	return Merkmal{}, NeuerFehler(l.zeile, l.spalte, SchluesselFremdesZeichen,
		string(l.zeichen[l.stelle]))
}

// istNamensanfang lässt auch Zeichen jenseits von ASCII zu.
//
// Die gesichteten Pakete tragen chinesische Kommentare; dass ein Bezeichner
// nicht ebenso aussehen darf, ist nirgends belegt. Abzuweisen, was man nicht
// gesehen hat, wäre eine Regel ohne Quelle.
func istNamensanfang(zeichen rune) bool {
	return zeichen == '_' || unicode.IsLetter(zeichen)
}

func istNamenszeichen(zeichen rune) bool {
	return istNamensanfang(zeichen) || unicode.IsDigit(zeichen)
}

// geflohen gibt, was hinter einem Rückstrich steht.
//
// **Unbekanntes bleibt, wie es dasteht — samt Rückstrich.** Eine Vorlage
// schreibt `"!promo\\b"` und meint das Regex-Wortende; wer den Rückstrich dort
// verschluckte, machte aus einem Muster ein anderes. Nur die drei Folgen, die
// in den Vorlagen des Packs vorkommen, werden gedeutet.
func geflohen(zeichen rune) string {
	switch zeichen {
	case 'n':
		return "\n"
	case 't':
		return "\t"
	case '"', '\\':
		return string(zeichen)
	}

	return "\\" + string(zeichen)
}

// beginntMit sagt, ob an der aktuellen Stelle dieser Text steht.
//
// **Über die Runen der Eingabe**, ohne eine Zeichenkette zu bauen — der Grund
// steht bei [lexer.operator].
func (l *lexer) beginntMit(text string) bool {
	stelle := l.stelle

	for _, zeichen := range text {
		if stelle >= len(l.zeichen) || l.zeichen[stelle] != zeichen {
			return false
		}

		stelle++
	}

	return true
}
