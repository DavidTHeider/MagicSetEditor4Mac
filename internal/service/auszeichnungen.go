package service

import (
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// ohneAuszeichnungen nimmt die spitzen Klammern aus einem Kartentext.
//
// # Der Befund
//
// Ein MSE-Kartentext trägt Auszeichnungen in spitzen Klammern:
//
//	Alliance Subversive — <sym>T</sym>: Target creature is lawbound…
//	<i-flavor>"What do you mean why? Burn! Burn!"</i-flavor>
//
// Unbehandelt stünden sie **wörtlich** im Textkasten. Gezählt in
// `magic.mse-game`: 2.165 `<atom-param>`, 108 `<sym>`, 16 `<i-auto>`, dazu
// `<atom-sep>` und `<atom-autorem>`.
//
// # Was hier geschieht, und was nicht
//
// **Die Klammern verschwinden, ihre Bedeutung wird noch nicht gezeichnet.**
// `render.Satz` führt `Fett` und `Kursiv`, und **niemand liest sie**: der
// Rasterer kennt keine Schnittvariante, und `schriften.Suche` nimmt nur einen
// Namen entgegen. Ein Symbol mitten im Text bräuchte zudem eine eigene Ebene
// an der richtigen Stelle.
//
// Deshalb bekommt jede weggelassene Auszeichnung einen **Befund**. Sichtbar
// falsch ist besser als unsichtbar fehlend, und wer die Befundzahl als Maß
// nimmt, soll nicht dadurch besser dastehen, dass etwas halb getan wurde (§8).
//
// # Ein `<` ohne schließende Klammer bleibt stehen
//
// `1 < 2` ist keine Auszeichnung. Wer alles ab `<` verwürfe, fräße den halben
// Satz — und ein Kartentext darf ein Kleinerzeichen enthalten.
// entschuetzt macht aus den beiden Steuerzeichen das, was sie meinen.
//
//	/// Escaped '<'
//	const Char ESCAPED_LANGLE = _('\1');
//	#define CONNECTION_SPACE _('\xEB00') // in private use area, untags to ' '
//
//	wxUniChar untag_char(wxUniChar c) {
//	  if (c == ESCAPED_LANGLE) return _('<');
//	  else if (c == CONNECTION_SPACE) return _(' ');
//	  else return c;
//	}
//
// `src/util/tagged_string.hpp` Zeile 20, `src/util/string.hpp` Zeile 85,
// `src/util/tagged_string.cpp` Zeilen 14–18, Fassung f9d9356 vom 2021-07-06.
//
// **Ein getipptes `<` steht so in der Datei** — sonst begänne beim nächsten
// Lesen eine Marke. Wer es zeichnete, wie es dasteht, zeichnete ein
// unsichtbares Zeichen an der Stelle eines gewöhnlichen
// (#687).
func entschuetzt(text string) string {
	return domain.Entschuetzt(text)
}

func ohneAuszeichnungen(feld, text string) ([]rohabschnitt, []domain.Befund) {
	if !strings.Contains(text, "<") {
		return []rohabschnitt{{Text: entschuetzt(text)}}, nil
	}

	var (
		abschnitte []rohabschnitt
		gebaut     strings.Builder
		genannt    = map[string]int{}
		rest       = text
		kursiv     int
		fett       int
		symbol     int
		linie      int
		weich      int

		// **Ein eigener Zähler.** `<soft>` und `<soft-line>` sind bei MSE
		// zwei Zähler und zwei Wirkungen — Ausrichtung gegen Umbruch.
		weicheZeile int

		// **Ein Stapel und kein Zähler.** Geschachtelte `<margin>` addieren
		// sich, und das schließende legt genau eine Lage wieder ab.
		raender []render.Margen

		imRand int
	)

	// schliesse legt das Angesammelte als Abschnitt ab.
	//
	// **Eine leere Stelle ergibt keinen Abschnitt.** `<i></i>` mitten im Text
	// zerschnitte den Zug sonst in drei, von denen einer nichts enthält — und
	// der Umbruch zählte ihn mit.
	schliesse := func() {
		if gebaut.Len() == 0 {
			return
		}

		// **Zwei gleiche Schnitte nebeneinander sind einer.** `<i>a<i>b</i>c`
		// schließt zweimal ab und ergäbe sonst drei kursive Abschnitte statt
		// eines — der Umbruch zählte drei Stücke, wo einer steht, und die
		// Laufweite fiele an zwei erfundenen Nähten an.
		if letzter := len(abschnitte) - 1; letzter >= 0 &&
			abschnitte[letzter].Kursiv == (kursiv > 0) &&
			abschnitte[letzter].Fett == (fett > 0) &&
			abschnitte[letzter].Weich == (weich > 0) &&
			abschnitte[letzter].Blockende == (linie > 0) &&
			abschnitte[letzter].WeicheZeile == (weicheZeile > 0) &&
			abschnitte[letzter].Margen == oberster(raender) &&
			abschnitte[letzter].ImRand == (imRand > 0) &&
			abschnitte[letzter].Symbolisch == (symbol > 0) {
			abschnitte[letzter].Text += gebaut.String()
			gebaut.Reset()

			return
		}

		abschnitte = append(abschnitte, rohabschnitt{
			Text: gebaut.String(), Kursiv: kursiv > 0, Fett: fett > 0,
			Symbolisch: symbol > 0, Blockende: linie > 0, Weich: weich > 0,
			WeicheZeile: weicheZeile > 0,
			Margen:      oberster(raender),
			ImRand:      imRand > 0,
		})
		gebaut.Reset()
	}

	for {
		auf := strings.IndexByte(rest, '<')
		if auf < 0 {
			gebaut.WriteString(rest)

			break
		}

		zu := strings.IndexByte(rest[auf:], '>')
		if zu < 0 {
			// **Kein Ende, keine Auszeichnung.** Der Rest ist Text.
			gebaut.WriteString(rest)

			break
		}

		inhalt := rest[auf+1 : auf+zu]

		name := grundform(inhalt)
		if !istAuszeichnungsname(name) {
			// **`1 < 2 und 3 > 2` ist Text**, keine Auszeichnung. Das
			// Kleinerzeichen geht mit, und gesucht wird hinter ihm weiter.
			gebaut.WriteString(rest[:auf+1])
			rest = rest[auf+1:]

			continue
		}

		gebaut.WriteString(rest[:auf])

		switch {
		case istUebergangen(name):
			// **Was MSE nicht kennt, übergeht es.** `TextElements::fromString`
			// führt eine Liste gedeuteter Namen; jeder andere wird
			// übersprungen und sein Inhalt bleibt Text
			// (`src/render/text/element.cpp`, Zeilen 78–130, Fassung f9d9356
			// vom 2021-07-06).
			//
			// **`sep` ist so einer**, und `combined_editor` klammert damit den
			// Trenner zwischen Regel- und Flavortext. Dieses Programm tat
			// schon dasselbe — und **meldete es**, 367-mal über die 190 Stile
			// von `M15-Magic-Pack`. Ein Befund für richtiges Verhalten
			// verdeckt die echten (§8).

		case istAufzaehlung(name):
			// **Gedeutet und verworfen** — siehe [istAufzaehlung].

		case istRand(name):
			// **Ein Rand gilt für den ganzen Absatz, nicht für eine Stelle
			// darin.** Er wird deshalb nicht gezählt wie ein Schnitt: MSE
			// schreibt ihn an `paragraphs.back()` und legt ihn beim
			// schließenden `</margin>` wieder ab
			// (`src/render/text/element.cpp`, Zeilen 177–196).
			//
			// Fassung f9d9356 vom 2021-07-06.
			schliesse()

			if strings.HasPrefix(inhalt, "/") {
				if len(raender) > 0 {
					raender = raender[:len(raender)-1]
				}
			} else if neu := margenAus(inhalt); neu != nil {
				// **Geschachtelte Ränder addieren sich** (ebd. Zeilen
				// 186–190).
				if len(raender) > 0 {
					vorher := raender[len(raender)-1]
					neu.Links += vorher.Links
					neu.Rechts += vorher.Rechts
					neu.Oben += vorher.Oben
				}

				raender = append(raender, *neu)
			}

		case istKursiv(name), istFett(name), istSymbolisch(name), istLinie(name),
			istWeich(name), istWeicheZeile(name), istImRand(name):
			schliesse()

			// **Gezählt und nicht geschaltet.** MSE hält einen Zähler:
			// `italic += 1` beim Öffnen, `-= 1` beim Schließen
			// (`src/render/text/element.cpp`, Zeilen 78–81, Fassung f9d9356 vom
			// 2021-07-06). Wer beim ersten `</i>` aufhörte, ließe den Rest
			// einer verschachtelten Stelle aufrecht stehen.
			schritt := 1
			if strings.HasPrefix(inhalt, "/") {
				schritt = -1
			}

			switch {
			case istKursiv(name):
				kursiv = max(kursiv+schritt, 0)
			case istFett(name):
				fett = max(fett+schritt, 0)
			// **`soft-line` vor `line` und vor `soft`.** Beide träfen es sonst
			// nicht — `istLinie` prüft `line`, und `<soft-line>` beginnt
			// anders —, wohl aber `istWeich`, wenn dort nicht ausdrücklich
			// ausgenommen. Die Reihenfolge ist MSEs eigene.
			case istWeicheZeile(name):
				weicheZeile = max(weicheZeile+schritt, 0)
			case istLinie(name):
				linie = max(linie+schritt, 0)
			case istWeich(name):
				weich = max(weich+schritt, 0)
			case istImRand(name):
				imRand = max(imRand+schritt, 0)
			default:
				symbol = max(symbol+schritt, 0)
			}

		case istSchlichtesAtom(name):
			// **Gedeutet und verworfen** — siehe [istSchlichtesAtom]: MSE malt
			// dafür nur im Editor einen Hintergrund und zeichnet den Inhalt.

		default:
			// **Gemeldet wird nur, was nicht gedeutet ist.** `<i>`, `<b>` und
			// `<sym>` werden gezeichnet; sie hier mitzuzählen ergäbe einen
			// Befund auf jeder Karte mit Flavortext — und ein Befund, der auf
			// jeder Karte steht, wird nicht mehr gelesen.
			genannt[name]++
		}

		rest = rest[auf+zu+1:]
	}

	schliesse()

	if len(abschnitte) == 0 {
		abschnitte = []rohabschnitt{{}}
	}

	// **Zuletzt und nicht unterwegs.** `ESCAPED_LANGLE` **ist** ein `<`; wer
	// es vor dem Lesen ersetzte, ließe an seiner Stelle eine Marke beginnen —
	// genau das, wovor MSE es schützt.
	for stelle := range abschnitte {
		abschnitte[stelle].Text = entschuetzt(abschnitte[stelle].Text)
	}

	return abschnitte, gemeldeteAuszeichnungen(feld, genannt)
}

// rohabschnitt ist ein Abschnitt, bevor seine Symbole aufgelöst sind.
//
// **Eigen und nicht [render.Abschnitt].** Dort stehen **Bildpfade**; welcher
// Code welches Bild meint, weiß erst [mitSymbolen] — und `internal/render`
// kennt keine Symbolschrift (§2).
type rohabschnitt struct {
	Text       string
	Kursiv     bool
	Fett       bool
	Symbolisch bool

	// Blockende ist `<line>` — siehe [render.Abschnitt.Blockende].
	Blockende bool

	// Weich ist `<soft>` — siehe [render.Abschnitt.Weich].
	Weich bool

	// WeicheZeile ist `<soft-line>` — siehe [render.Abschnitt.WeicheZeile].
	WeicheZeile bool

	// Margen ist `<margin:…>` — siehe [render.Abschnitt.Margen].
	Margen *render.Margen

	// ImRand ist `<bullet>` — siehe [render.Abschnitt.ImRand].
	ImRand bool
}

// oberster gibt die oberste Lage des Randstapels, oder nil.
func oberster(raender []render.Margen) *render.Margen {
	if len(raender) == 0 {
		return nil
	}

	return &raender[len(raender)-1]
}

// grundform macht aus `/i-auto` und `i-auto` denselben Namen.
//
// **Öffnend und schließend sind dieselbe Auszeichnung.** Sie getrennt zu zählen
// verdoppelte jede Meldung und sagte nichts dazu.
func grundform(inhalt string) string {
	// **Ohne `TrimSpace`.** Eine MSE-Auszeichnung trägt keine Leerzeichen; wer
	// sie wegschnitte, machte aus `a < b > c` die Auszeichnung `b` und fräße
	// den halben Satz.
	return strings.TrimPrefix(inhalt, "/")
}

// istAuszeichnungsname sagt, ob zwischen den Klammern eine Auszeichnung steht.
//
// **Buchstaben, Ziffern, Bindestrich, Unterstrich — und nicht leer.** Alles
// andere ist Text: ein Kartentext darf `1 < 2 und 3 > 2` enthalten, und wer das
// für eine Auszeichnung hielte, fräße den halben Satz.
func istAuszeichnungsname(name string) bool {
	if name == "" {
		return false
	}

	// **Sie beginnt mit einem Buchstaben.** `1<2>0` ist ein Vergleich und keine
	// Auszeichnung; der erste Anlauf machte daraus `10`, weil `2` als Name
	// durchging. Angestrichen vom Zweitblick über die Phasen B bis D — der
	// geprüfte Vergleich trug Leerzeichen und betrat diesen Weg nie.
	//
	// Jede Auszeichnung von MSE beginnt mit einem Buchstaben: `i`, `b`, `sym`,
	// `i-flavor`, `atom-param`, `kw-…`.
	erstes := rune(name[0])
	if (erstes < 'a' || erstes > 'z') && (erstes < 'A' || erstes > 'Z') {
		return false
	}

	// **Ein Doppelpunkt beendet den Namen; dahinter steht ein Parameter.**
	//
	// MSE2 deutet fünf Auszeichnungen mit Parameter — `<color:black>`,
	// `<font:Beleren Bold>`, `<size:…>`, `<margin:0:0:5>`, `<align:…>`
	// (`element.cpp:98–130`). Sie stehen auch im Pack. Der erste Anlauf ließ
	// weder Doppelpunkt noch Leerzeichen zu und **ließ sie roh auf der Karte
	// stehen** — angestrichen vom Nachlauf des Zweitblicks: der Fall `1<2>0`
	// war geschlossen, die Klasse „MSE-Tag gegen Nutztext" nicht.
	//
	// **Der Parameter wird nicht geprüft.** Was hinter dem Doppelpunkt steht,
	// darf alles sein — ein Dateiname mit Leerzeichen, drei Zahlen, ein
	// zusammengesetzter Ausdruck. Geprüft wird der **Name**, und das ist die
	// Stelle, an der sich eine Auszeichnung von einem Vergleich unterscheidet.
	// **Kein zweiter Leerprüfer dahinter.** Das erste Zeichen ist schon als
	// Buchstabe erkannt; `Cut` lässt es stehen, und der Name kann danach nicht
	// leer sein. Ein Zweig, den kein Aufruf betreten kann, ist einer zu viel
	// (§3).
	name, _, _ = strings.Cut(name, ":")

	for _, zeichen := range name {
		passt := zeichen == '-' || zeichen == '_' ||
			(zeichen >= 'a' && zeichen <= 'z') ||
			(zeichen >= 'A' && zeichen <= 'Z') ||
			(zeichen >= '0' && zeichen <= '9')
		if !passt {
			return false
		}
	}

	return true
}

// gemeldeteAuszeichnungen macht aus der Zählung **einen** Befund je Feld.
//
// **Einen und nicht einen je Vorkommen.** Ein Regeltext mit acht `<sym>` ist
// ein Mangel und nicht acht; acht Zeilen im Protokoll verdeckten die anderen
// Felder.
func gemeldeteAuszeichnungen(feld string, genannt map[string]int) []domain.Befund {
	if len(genannt) == 0 {
		return nil
	}

	namen := make([]string, 0, len(genannt))
	for name := range genannt {
		namen = append(namen, name)
	}

	sort.Strings(namen)

	return []domain.Befund{{
		Ort: feld,
		Meldung: "die Auszeichnungen " + strings.Join(namen, ", ") +
			" sind entfernt; gezeichnet wird der bloße Text",
	}}
}
