package msescript

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// `sort_text` ordnet die Zeichen eines Textes nach einer Ordnungszeichenkette.
//
// # Was die Zeichenkette sagt — und was nicht
//
// Sie ist eine kleine Sprache für sich, und
// #98 hielt
// fest, dass **ein** Vorkommen ihre Regeln nicht entscheidet. Die bezogenen
// Pakete tragen sechsundzwanzig verschiedene; zwei der drei Fragen entscheiden
// sie. Die Belege stehen bei den Prüfstücken, nicht hier — dort gehören sie
// hin, weil sie prüfbar sein müssen.
//
//	Zeichen     ein eigener Rang, in der genannten Reihenfolge
//	[…]         **ein gemeinsamer** Rang: die Eingabereihenfolge bleibt
//	<…>         filtern und entdoppeln: nur diese Zeichen, jedes einmal
//	\c          `c` als gewöhnliches Zeichen
//	(…)         **nicht erschlossen** — wird gemeldet
//
// **Was nicht in der Ordnung steht, kommt hinten an** und behält seinen Platz
// unter seinesgleichen. Das ist eine **Vermutung**; sie steht mit ihrem
// Prüfstück in `docs/FORMATE.md`. Die andere Lesart — es fällt weg — machte aus
// einer Manakosten mit einem unbekannten Symbol stillschweigend eine ohne, und
// ein stiller Verlust ist der teurere Irrtum.
func init() {
	eintragen("sort_text", ordneText)
}

// ordnung ist eine zerlegte Ordnungszeichenkette.
type ordnung struct {
	// rang gibt je Zeichen seinen Rang. Zeichen einer Klammergruppe teilen ihn.
	rang map[rune]int

	// menge sagt, dass die ganze Ordnung `<…>` war: filtern und entdoppeln.
	menge bool

	// zyklen sind die `cycle(…)`-Gruppen. Ihre Ränge entscheidet erst die
	// Eingabe — siehe [zyklus].
	zyklen []zyklus

	// anOrt ist die Unterordnung eines `in_place(…)`. Sie ordnet die Zeichen,
	// die sie nennt, **untereinander** und lässt sie an ihren Stellen; alles
	// andere bleibt unberührt stehen ([ordneAnOrt]).
	anOrt *ordnung
}

func ordneText(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("order")
	if !benannt {
		return Nichts, ruf.fehlerAn("sort_text braucht ein Argument `order:`")
	}

	if err := ruf.eineEingabe("sort_text"); err != nil {
		return Nichts, err
	}

	regel, fehler := zerlegeOrdnung(roh.AlsText())
	if fehler != nil {
		return Nichts, ruf.fehlerAn(fehler.Error())
	}

	return Text(regel.wende(ruf.Stellung[0].AlsText())), nil
}

// wende ordnet einen Text nach der Regel.
func (o ordnung) wende(eingabe string) string {
	if o.anOrt != nil {
		return o.anOrt.ordneAnOrt(eingabe)
	}

	if o.menge {
		return o.filtere(eingabe)
	}

	// **Was die Ordnung nicht nennt, fällt weg.**
	//
	//	} else { // single char
	//	  size_t count = count_and_remove(it.value, input);
	//	  ret.append(count, it.value);
	//	}
	//	…
	//	return ret;
	//
	// `spec_sort`, `src/util/spec_sort.cpp`, Zeilen 266–352 und 355, Fassung
	// f9d9356 vom 2021-07-06: die Schleife hängt an `ret` nur an, was die
	// Spezifikation nennt — was in `input` übrig bleibt, wird verworfen.
	//
	// **Was das gekostet hat:** ein kanonisches MSE2-Set speichert Manakosten
	// als `[3][W]`, und `mana_filter` schickt sie durch `sort_text`. Die
	// Klammern stehen in der Ordnung nicht als Zeichen, sondern als
	// Gruppenklammer — sie fallen bei MSE weg. Hier blieben sie stehen und
	// wurden als Symbole gezeichnet: vier leere Kästen neben den Manakosten,
	// auf **jeder** gewöhnlichen Karte
	// (#411).
	zeichen := behalteGenannte(o, []rune(eingabe))

	// **Ein Zyklus wird erst hier gedreht**, denn seine Reihenfolge hängt an
	// der Eingabe (siehe [zyklus]). Die übrigen Ränge stehen schon.
	raenge := o.rang

	if len(o.zyklen) > 0 {
		vorkommen := map[rune]int{}
		for _, einzelnes := range zeichen {
			vorkommen[einzelnes]++
		}

		raenge = make(map[rune]int, len(o.rang))
		for wo, rang := range o.rang {
			raenge[wo] = rang
		}

		for _, gruppe := range o.zyklen {
			for wo, rang := range gruppe.drehe(vorkommen) {
				raenge[wo] = rang
			}
		}
	}

	// **Stabil**, und darauf beruht die Klammergruppe: Zeichen gleichen Ranges
	// behalten ihre Eingabereihenfolge. Ohne das würde aus `10` ein `01`, und
	// eine Karte kostete plötzlich eins statt zehn.
	sort.SliceStable(zeichen, func(a, b int) bool {
		return o.rangIn(raenge, zeichen[a]) < o.rangIn(raenge, zeichen[b])
	})

	return string(zeichen)
}

// behalteGenannte wirft weg, was die Ordnung nicht kennt.
//
// **Ein Zyklus zählt mit.** Seine Zeichen stehen in `rang`, bevor er gedreht
// wird; wer nur die feste Tafel fragte, würfe `WUBRG` weg.
func behalteGenannte(o ordnung, zeichen []rune) []rune {
	behalten := zeichen[:0]

	for _, einzelnes := range zeichen {
		if _, bekannt := o.rang[einzelnes]; bekannt {
			behalten = append(behalten, einzelnes)
		}
	}

	return behalten
}

// rangIn gibt den Rang eines Zeichens in einer gedrehten Rangtafel.
//
// **Ohne Ausweichrang**, seit die Ordnung verwirft, was sie nicht nennt
// ([behalteGenannte]): jedes Zeichen, das hier ankommt, steht in der Tafel. Ein
// Zweig, den kein Aufruf betreten kann, wäre eine Behauptung darüber, was im
// Ernstfall geschieht (§3).
func (o ordnung) rangIn(raenge map[rune]int, zeichen rune) int {
	return raenge[zeichen]
}

// filtere gibt die genannten Zeichen, jedes einmal, in der Reihe der Ordnung.
//
// **Entdoppelt**, und das ist keine Feinheit: `mana_is_wedge` zählt das
// Ergebnis und vergleicht mit drei. Wer `WWU` für dreifarbig hielte, gäbe der
// Karte den falschen Rahmen.
func (o ordnung) filtere(eingabe string) string {
	da := map[rune]bool{}

	for _, zeichen := range eingabe {
		if _, bekannt := o.rang[zeichen]; bekannt {
			da[zeichen] = true
		}
	}

	gefunden := make([]rune, 0, len(da))
	for zeichen := range da {
		gefunden = append(gefunden, zeichen)
	}

	sort.Slice(gefunden, func(a, b int) bool {
		return o.rang[gefunden[a]] < o.rang[gefunden[b]]
	})

	return string(gefunden)
}

// errOrdnungOffen meldet einen Teil der Ordnungssprache, der nicht erschlossen
// ist.
//
// **Gemeldet und nicht geraten** (#98). Eine Ordnung zu erfinden, die plausibel
// aussieht, hieße Manakosten in einer Reihenfolge zu setzen, die MSE anders
// setzt — und das fiele niemandem auf, außer beim Vergleich zweier Karten
// nebeneinander. Weiter oben wird daraus ein Vorbehalt an der Ebene (§8).
func errOrdnungOffen(stelle string) error {
	return fmt.Errorf(
		"die Ordnung %q ist nicht erschlossen: eine runde Klammer trennt eine "+
			"Gruppe ab, deren Regel aus keinem gesichteten Paket hervorgeht "+
			"(#98). Sie wird gemeldet statt geraten", stelle)
}

// zerlegeOrdnung liest eine Ordnungszeichenkette.
func zerlegeOrdnung(quelle string) (ordnung, error) {
	if mengenform, ist := alsMenge(quelle); ist {
		return mengenform, nil
	}

	if strings.HasPrefix(quelle, "<") {
		return ordnung{}, fmt.Errorf(
			"die Ordnung %q beginnt mit `<` und schließt nicht", quelle)
	}

	// **`in_place(…)` ist die ganze Ordnung, nicht ein Glied darin.**
	//
	// So steht es in den Paketen (`magic.mse-game/script:521`:
	// `sort_text(order:"in_place((WUBRG))")`), und so ist es hier gebaut. MSEs
	// Schleife erlaubt es an jeder Stelle — käme es dort vor, meldete der
	// Wächter unten es weiterhin als ungebaut, statt es zu raten (§10).
	if innen, ist := strings.CutPrefix(quelle, "in_place("); ist {
		rumpf, zu := strings.CutSuffix(innen, ")")
		if !zu {
			return ordnung{}, fmt.Errorf(
				"die Ordnung %q öffnet `in_place(` und schließt es nicht", quelle)
		}

		unter, err := zerlegeOrdnung(rumpf)
		if err != nil {
			return ordnung{}, err
		}

		return ordnung{rang: map[rune]int{}, anOrt: &unter}, nil
	}

	regel := ordnung{rang: map[rune]int{}}
	zeichen := []rune(quelle)
	rang := 0

	for i := 0; i < len(zeichen); i++ {
		// **`cycle(` ist die ausgeschriebene Form von `(`** — `spec_sort.cpp`
		// behandelt beide im selben Zweig (Zeile 292 und 294). Beide Formen
		// stehen in `M15-Magic-Pack`.
		rest := string(zeichen[i:])

		if strings.HasPrefix(rest, "cycle(") {
			i += len("cycle") - 1

			continue
		}

		// **Die übrigen Schlüsselwörter von MSE werden gemeldet, nicht
		// stillschweigend als Zeichenfolge gelesen.**
		//
		// `spec_sort.cpp` kennt neben `cycle(` noch `ordered(`, `mixed(`,
		// `once(`, `compound(`, `pattern(`, `in_place(` und `any()`. Keines
		// davon steht in `M15-Magic-Pack`; sie hier zu bauen wäre Vorrat ohne
		// Fundstelle. Ein `in_place((WUBRG))` als Buchstabenfolge zu lesen wäre
		// aber schlimmer als ein Fehler — es ergäbe eine Ordnung, die
		// plausibel aussieht und falsch ist.
		if wort := vorSchluesselwort(rest); wort != "" {
			return ordnung{}, fmt.Errorf(
				"die Ordnung %q benutzt %s…; das kennt MSE, dieses Programm hat es "+
					"nicht gebaut — kein Paket von M15-Magic-Pack führt es. "+
					"Gemeldet statt geraten", quelle, wort)
		}

		switch zeichen[i] {
		case '\\':
			// Ein Gegenschrägstrich nimmt dem Nächsten seine Bedeutung. Steht
			// er am Ende, ist er selbst gemeint — ein Fehler wäre hier eine
			// Zusage über eine Schreibweise, die kein Paket führt.
			if i+1 < len(zeichen) {
				i++
			}

			regel.setze(zeichen[i], rang)
			rang++

		case '(':
			zu := suche(zeichen, i, ')')
			if zu < 0 {
				return ordnung{}, errOrdnungOffen(quelle)
			}

			gruppe := zyklus{zeichen: zeichen[i+1 : zu], rang: rang}
			regel.zyklen = append(regel.zyklen, gruppe)

			for _, inhalt := range gruppe.zeichen {
				regel.setze(inhalt, rang)
			}

			// **Jedes Zeichen der Gruppe bekommt einen eigenen Rang**, anders
			// als bei `[…]`: der Zyklus ordnet innerhalb, er mischt nicht.
			rang += len(gruppe.zeichen)
			i = zu

		case '[':
			zu := suche(zeichen, i, ']')
			if zu < 0 {
				return ordnung{}, fmt.Errorf(
					"die Ordnung %q öffnet eine eckige Klammer und schließt sie nicht", quelle)
			}

			// **Ein gemeinsamer Rang für die ganze Gruppe.** `mana_unsort`
			// setzt alles in eine Klammer und heißt so, weil dann nichts mehr
			// sortiert wird.
			for _, inhalt := range zeichen[i+1 : zu] {
				regel.setze(inhalt, rang)
			}

			rang++
			i = zu

		default:
			regel.setze(zeichen[i], rang)
			rang++
		}
	}

	return regel, nil
}

// schluesselwoerter sind die Ordnungsfunktionen aus `spec_sort.cpp`, die
// dieses Programm **nicht** baut. `cycle(` steht nicht dabei — es ist gebaut.
var schluesselwoerter = []string{
	"ordered(", "mixed(", "once(", "compound(", "pattern(", "in_place(", "any()",
}

// vorSchluesselwort gibt das erkannte Schlüsselwort samt Klammer, sonst "".
func vorSchluesselwort(rest string) string {
	for _, wort := range schluesselwoerter {
		if strings.HasPrefix(rest, wort) {
			return wort
		}
	}

	return ""
}

// alsMenge liest die Form `<…>`, die die ganze Ordnung ausmacht.
func alsMenge(quelle string) (ordnung, bool) {
	if !strings.HasPrefix(quelle, "<") || !strings.HasSuffix(quelle, ">") || len(quelle) < 2 {
		return ordnung{}, false
	}

	regel := ordnung{rang: map[rune]int{}, menge: true}

	for rang, zeichen := range []rune(quelle[1 : len(quelle)-1]) {
		regel.setze(zeichen, rang)
	}

	return regel, true
}

// setze trägt einen Rang ein — **das erste Vorkommen gilt**.
//
// Mehrere Ordnungen führen dasselbe Zeichen zweimal, etwa die `0` in
// `[…01234567890…]`. Wer das zweite gelten ließe, hinge ein Zeichen ans Ende,
// das vorn gemeint war.
func (o ordnung) setze(zeichen rune, rang int) {
	if _, schon := o.rang[zeichen]; !schon {
		o.rang[zeichen] = rang
	}
}

// suche gibt die Stelle des schließenden Zeichens — oder -1.
func suche(zeichen []rune, ab int, zu rune) int {
	for i := ab + 1; i < len(zeichen); i++ {
		if zeichen[i] == zu {
			return i
		}
	}

	return -1
}
