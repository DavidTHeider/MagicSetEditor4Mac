// Schlüsselwörter im Kartentext
// (#1061).
//
// # Warum das hier steht und nicht in `msepaket` oder `msescript`
//
// `msepaket` liest, **was in der Datei steht** (#1060) — ein Paketleser, der
// Ausdrücke übersetzte und Skripte auswertete, wäre die Vermischung, die §2
// verhindert. `msescript` ist ein Sprachwerkzeug und kennt keine
// Kartenbegriffe. Was ein Schlüsselwort mit einem Regeltext **tut**, ist eine
// fachliche Regel über mehrere Einheiten — und die gehört nach `service`.

package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Schluesselworttreffer ist ein gefundenes Schlüsselwort samt dem, was der
// Aufrufer zum Zusammensetzen braucht.
type Schluesselworttreffer struct {
	// Text ist der getroffene Text, so wie er im Regeltext steht.
	Text string

	// Wort, Modus und Erinnerung kommen aus dem Schlüsselwort.
	Wort       string
	Modus      string
	Erinnerung string

	// Parameter sind die Werte der Parameter, in ihrer Reihenfolge —
	// `param1`, `param2` und so fort.
	Parameter []string

	// Fest sagt, ob es aus dem Spiel kommt.
	Fest bool

	// SchreibungStimmt sagt, ob der getroffene Text die **Schreibung** des
	// Musters hat — `correct_case` (`src/data/keyword.cpp`, Zeile 697,
	// Fassung f9d9356 vom 2021-07-06).
	//
	// **Die M15-Vorlage entscheidet daran**, ob ein Erinnerungstext überhaupt
	// erscheint: `correct_case or (mode != "pseudo" and not
	// used_placeholders)`. Ohne die Angabe stand auf einer echten Karte
	// zweimal derselbe Erinnerungstext — einmal hinter „Flying" und einmal
	// mitten im Satz hinter „gains flying".
	SchreibungStimmt bool
}

// vorbereitet ist ein Schlüsselwort mit seinem übersetzten Ausdruck.
type vorbereitet struct {
	wort     msepaket.Schluesselwort
	ausdruck *regexp.Regexp

	// genau ist derselbe Ausdruck **ohne** `(?i)` und auf den ganzen Treffer
	// verankert. Er beantwortet `correct_case`: passt der getroffene Text auch
	// dann, wenn Groß- und Kleinschreibung zählen?
	//
	// **Eine Näherung, und sie ist benannt.** MSE vergleicht die festen Teile
	// des Musters Zeichen für Zeichen (`keyword_match_detail`, Zeilen
	// 656–672); hier prüft ein zweiter Ausdruck denselben Text. Der
	// Unterschied läge bei einer Parameterart, deren eigener Ausdruck
	// Großschreibung verlangt — dann sagte diese Prüfung „falsch", wo MSE nur
	// die festen Teile ansieht.
	genau *regexp.Regexp

	// stelle ist die Stelle in der Eingabereihe. Sie entscheidet nur, wenn
	// alles andere gleich ist — MSE sortiert dort nach Namen, und das tut
	// dieselbe Arbeit ohne den Namen zu brauchen.
	stelle int
}

// Wortliste ist die vorbereitete Wortliste eines Spiels samt seines Sets.
type Wortliste struct{ woerter []vorbereitet }

// Anzahl sagt, wie viele Schlüsselwörter benutzbar sind.
func (l Wortliste) Anzahl() int { return len(l.woerter) }

// NeueWortliste bereitet die Schlüsselwörter vor — `Keyword::prepare`.
//
// # Was übersetzt wird
//
// `src/data/keyword.cpp`, Zeilen 148–208, Fassung f9d9356 vom 2021-07-06:
//
//   - `(?i)` vorn — Groß- und Kleinschreibung spielt keine Rolle
//     (`USE_CASE_INSENSITIVE_KEYWORDS 1`, Zeile 20).
//   - Der Text zwischen den Parametern wird escaped.
//   - `<atom-param>art</atom-param>` wird zur Gruppe der Parameterart; ist sie
//     wahlfrei, bekommt sie ein `?`.
//   - Vorn eine Wortgrenze, hinten `$` oder ein Zeichen, das keinen Wortlaut
//     hat.
//
// # Wo abgewichen wird — und warum
//
// **Go kennt keinen Vorausblick.** MSE hängt `(?=$|[^a-zA-Z0-9\(])` an; ein
// Vorausblick prüft, ohne zu verbrauchen. Gos `regexp` (RE2) hat ihn nicht.
//
// Hier steht deshalb dieselbe Bedingung als **letzte Gruppe**, die verbraucht
// wird — und der Treffer endet vor ihr. Das ist `adapted` und kein `ported`:
// es verhält sich gleich, solange zwei Treffer sich nicht **ein einziges
// Zeichen** teilen. Bei „Eile Eile" ist das Leerzeichen der Nachlauf des
// ersten und steht vor dem zweiten; der zweite beginnt erst danach, und beide
// werden gesetzt. Wo es auseinanderginge, wäre ein Treffer, der unmittelbar
// am Ende des vorigen begänne — und den schlösse MSEs Wortgrenze ebenso aus.
//
// **Ein Befund je abgeschaltetem Schlüsselwort**, nicht ein Abbruch: „throwing
// an error can mean a set will not be loaded! instead, simply disable the
// keyword" (Zeilen 174–177).
func NeueWortliste(
	woerter []msepaket.Schluesselwort, arten []msepaket.Schluesselwortparameterart,
) (Wortliste, []domain.Befund) {
	nachArt := map[string]msepaket.Schluesselwortparameterart{}
	for _, art := range arten {
		nachArt[art.Name] = art
	}

	var (
		liste   Wortliste
		befunde []domain.Befund
	)

	for stelle, wort := range woerter {
		ausdruck, befund := ausdruckFuer(wort, nachArt)
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		liste.woerter = append(liste.woerter, vorbereitet{
			wort: wort, ausdruck: ausdruck,
			genau:  genauerAusdruck(ausdruck),
			stelle: stelle,
		})
	}

	return liste, befunde
}

// trennerGefressen kürzt den festen Text um einen Trenner am Ende.
//
//	if (separator_before_eat.matches(result, text)) {
//	  // keep only stuff before the separator
//	  text.resize(result.position());
//	}
//
// `src/data/keyword.cpp`, Zeilen 111–119, Fassung f9d9356 vom 2021-07-06.
// Der Ausdruck dort ist `separator_before_is + "$"` (ebd. Zeile 100) — er
// greift also nur ganz hinten.
//
// **Ein unbrauchbarer Trennerausdruck lässt den Text stehen.** Er kommt aus
// einem fremden Paket (§8); ihn zum Fehler zu machen nähme das ganze
// Schlüsselwort mit, obwohl der Rest lesbar ist.
func trennerGefressen(text string, art msepaket.Schluesselwortparameterart) string {
	if !art.TrennerFressen || art.TrennerDavor == "" {
		return text
	}

	trenner, err := regexp.Compile("(?:" + art.TrennerDavor + ")$")
	if err != nil {
		return text
	}

	if stelle := trenner.FindStringIndex(text); stelle != nil {
		return text[:stelle[0]]
	}

	return text
}

// ohneTrennerDanach überspringt einen Trenner am Anfang des Restes.
//
//	separator_after_eat.assign(_("^") + separator_after_is);
//	…
//	i += result.length();
//
// `src/data/keyword.cpp`, Zeilen 107 und 126, Fassung f9d9356 vom 2021-07-06.
func ohneTrennerDanach(rest string, art msepaket.Schluesselwortparameterart) string {
	if !art.TrennerFressen || art.TrennerDanach == "" {
		return rest
	}

	trenner, err := regexp.Compile("^(?:" + art.TrennerDanach + ")")
	if err != nil {
		return rest
	}

	if stelle := trenner.FindStringIndex(rest); stelle != nil {
		return rest[stelle[1]:]
	}

	return rest
}

// ausdruckFuer übersetzt ein `match` in einen Ausdruck.
func ausdruckFuer(
	wort msepaket.Schluesselwort, arten map[string]msepaket.Schluesselwortparameterart,
) (*regexp.Regexp, *domain.Befund) {
	var gebaut strings.Builder

	rest := wort.Muster

	for {
		anfang := strings.Index(rest, "<atom-param>")
		if anfang < 0 {
			break
		}

		ende := strings.Index(rest[anfang:], "</atom-param>")
		if ende < 0 {
			break
		}

		name := rest[anfang+len("<atom-param>") : anfang+ende]

		art, bekannt := arten[name]
		if !bekannt {
			return nil, &domain.Befund{
				Ort: wort.Wort,
				Meldung: fmt.Sprintf(
					"die Parameterart %q gibt es nicht — das Schlüsselwort %q bleibt aus",
					name, wort.Wort),
			}
		}

		// **Der Trenner davor wird aus dem festen Text genommen.**
		//
		//	// remove the separator from the text to prevent duplicates
		//	param->eat_separator_before(text);
		//
		// `src/data/keyword.cpp`, Zeile 186, Fassung f9d9356 vom
		// 2021-07-06. Ohne das stünde er zweimal da: `Cumulative upkeep `
		// endet auf ein Leerzeichen, und das echte `cost`-Muster beginnt
		// mit `[ ]` — der Ausdruck verlangte **zwei**, und kein Kartentext
		// trägt beide. Der Erinnerungstext blieb aus.
		gebaut.WriteString(regexp.QuoteMeta(trennerGefressen(rest[:anfang], art)))
		gebaut.WriteString("(" + art.Muster + ")")

		if art.Wahlfrei {
			gebaut.WriteString("?")
		}

		rest = rest[anfang+ende+len("</atom-param>"):]

		// **Und der Trenner danach wird übersprungen** —
		// `param->eat_separator_after(match, i)`, ebd. Zeile 193.
		rest = ohneTrennerDanach(rest, art)
	}

	gebaut.WriteString(regexp.QuoteMeta(rest))

	kern := gebaut.String()

	// **Was auf den leeren Text passt, wird abgeschaltet** — „not valid if it
	// matches \"\", that would make MSE hang" (Zeilen 206–208).
	//
	// **Geprüft wird der Kern, nicht der umschlossene Ausdruck.** MSE prüft
	// den ganzen, und dort steht `\y` davor; in Go ist `\b` an einer leeren
	// Zeichenkette **nie** erfüllt, und die Prüfung wäre eine Zeile, die kein
	// Aufruf betreten kann (§3). Gemeint ist derselbe Fall: ein Muster, das
	// aus lauter wahlfreien Teilen besteht und deshalb überall trifft.
	if leer, fehler := regexp.Compile("^(?:" + kern + ")$"); fehler == nil && leer.MatchString("") {
		return nil, &domain.Befund{
			Ort: wort.Wort,
			Meldung: fmt.Sprintf(
				"das Muster von %q passt auf den leeren Text — das Schlüsselwort bleibt aus",
				wort.Wort),
		}
	}

	ausdruck, fehler := regexp.Compile(`(?i)\b` + kern + `($|[^a-zA-Z0-9(])`)
	if fehler != nil {
		return nil, &domain.Befund{
			Ort:     wort.Wort,
			Meldung: fmt.Sprintf("das Muster von %q ließ sich nicht übersetzen: %v", wort.Wort, fehler),
		}
	}

	return ausdruck, nil
}

// genauerAusdruck baut denselben Ausdruck ohne `(?i)` und verankert.
//
// **Kann nicht scheitern**: der Ausdruck ist schon übersetzt worden; ihm das
// `(?i)` zu nehmen und ihn zu verankern macht ihn nicht ungültig.
func genauerAusdruck(ausdruck *regexp.Regexp) *regexp.Regexp {
	quelle := strings.TrimPrefix(ausdruck.String(), "(?i)")

	// **`MustCompile` und kein Fehlerzweig.** Die Quelle ist ein bereits
	// übersetzter Ausdruck; ihm das `(?i)` zu nehmen und ihn zu verankern
	// macht ihn nicht ungültig. Ein Zweig, den kein Aufruf betreten kann, wäre
	// eine Behauptung darüber, was im Ernstfall geschieht (§3).
	return regexp.MustCompile("^" + quelle)
}

// atomklammer findet öffnende und schließende `<atom…>`-Auszeichnungen.
var atomklammer = regexp.MustCompile(`</?atom[^>]*>`)

// freieBereiche gibt die Bereiche des Textes, die **außerhalb** von
// `<atom…>` liegen.
//
// „don't expand keywords that are inside <atom> tags" — `src/data/keyword.cpp`,
// Zeile 526, Fassung f9d9356 vom 2021-07-06.
func freieBereiche(text string) [][2]int {
	var (
		frei  [][2]int
		tiefe int
		offen = 0
	)

	for _, klammer := range atomklammer.FindAllStringIndex(text, -1) {
		schliessend := strings.HasPrefix(text[klammer[0]:klammer[1]], "</")

		if !schliessend {
			if tiefe == 0 {
				frei = append(frei, [2]int{offen, klammer[0]})
			}

			tiefe++

			continue
		}

		if tiefe > 0 {
			tiefe--

			if tiefe == 0 {
				offen = klammer[1]
			}
		}
	}

	if tiefe == 0 {
		frei = append(frei, [2]int{offen, len(text)})
	}

	return frei
}

// imFreien sagt, ob ein Treffer ganz in einem freien Bereich liegt.
func imFreien(frei [][2]int, von, bis int) bool {
	for _, bereich := range frei {
		if von >= bereich[0] && bis <= bereich[1] {
			return true
		}
	}

	return false
}
