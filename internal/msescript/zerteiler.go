package msescript

import (
	"errors"
	"strconv"
)

// Zerteile macht aus Quelltext einen Ausdrucksbaum.
//
// # Der Vorrang, und woher er kommt
//
// Von unten nach oben, schwächste Bindung zuerst:
//
//	Zuweisung   :=            (rechts)
//	Bedingung   if…then…else  (rechts)
//	Oder        or            (links)
//	Und         and           (links)
//	Vergleich   == != < > <=  (links)
//	Summe       + -           (links)
//	Produkt     * /           (links)
//	Vorzeichen  not -         (rechts)
//	Nachsatz    . ( ) @       (links)
//
// **Zwei Stufen sind an Paketen belegt, die übrigen sind übernommen.** Belegt:
// dass der Vergleich stärker bindet als `or` (`a=="x" or a=="y"` ergäbe sonst
// Unsinn) und dass `not` stärker bindet als `and` (yugioh Zeile 224). Für alles
// andere gilt die Reihenfolge, die fast jede Sprache hat — und das ist eine
// Übernahme, keine Beobachtung. `docs/FORMATE.md` sagt es so.
//
// Die eine Zeile, die mehr entschieden hätte, entscheidet nichts:
// `no_shadow() and is_spell_card() or is_trap_card()` liest sich unter
// „`and` vor `or`" genauso wie unter „gleichstark, nach links" — beide ergeben
// `(a and b) or c`. Nur „`or` vor `and`" wäre anders, und dafür spricht nichts.
//
// **Der Baum wird gemerkt** — der Grund und die Bedingung stehen in
// `zerteilerspeicher.go`.
func Zerteile(quelle string) (*Knoten, error) {
	if baum, da := gemerkterBaum(quelle); da {
		baumtreffer.Add(1)

		return baum, nil
	}

	baumfehlgriffe.Add(1)

	baum, err := zerteileFrisch(quelle)
	if err != nil {
		return nil, err
	}

	merkeBaum(quelle, baum)

	return baum, nil
}

// Befund holt den [Fehler] aus einem Fehler dieses Pakets — **nie nil**.
//
// # Warum es das gibt
//
// Wer Zeile und Spalte eines Zerteilfehlers will, muss ihn erst wieder aus
// dem `error` holen. Das ist ein `errors.As` mit einem Zweig für den Fall,
// dass es keiner war — und **diesen Fall gibt es beim Zerteilen nicht**: der
// Zerteiler und der Lexer geben nichts anderes.
//
// Ein Zweig ohne Fall ist eine Behauptung darüber, was im Ernstfall geschieht
// (§3). Bei jedem Aufrufer stünde er einzeln da und bliebe unbetreten;
// **hier** ist er zu betreten, denn hier darf jeder Fehler hereinkommen.
//
// Gefunden am 2026-09-02 in `internal/service/paketpruefung.go`
// (#1324).
//
// # Was der Rückfall sagt
//
// **Die Stelle 1:1 und den Satz, der ankam.** Ein Fehler ohne Stelle bekommt
// damit die erste Zeile — sichtbar falsch statt still verschwunden. Der
// Schlüssel ist [SchluesselDurchgereicht]: das Fenster zeigt dann den Satz,
// wie er ist.
func Befund(fehler error) *Fehler {
	var befund *Fehler
	if errors.As(fehler, &befund) {
		return befund
	}

	return NeuerFehler(1, 1, SchluesselDurchgereicht, fehler.Error())
}

// zerteileFrisch rechnet den Baum wirklich aus.
func zerteileFrisch(quelle string) (*Knoten, error) {
	merkmale, err := Zerlege(quelle)
	if err != nil {
		return nil, err
	}

	z := &zerteiler{merkmale: merkmale, ende: endstelle(merkmale)}

	// **Ohne Prüfung auf einen Rest.** `folge(nil)` läuft bis zum Ende; was
	// nicht in einen Ausdruck passt, meldet sie selbst. Eine zweite Prüfung
	// träfe nie zu — ein Zweig, den kein Test betreten kann, ist eine
	// Behauptung darüber, was im Ernstfall geschieht.
	return z.folge(nil)
}

// HoechsteSchachtelung ist, wie tief ein Ausdruck höchstens geschachtelt sein
// darf — gezählt über jede Klasse der Rekursion (BK-08, Vermerk 0177).
//
// **Eine neue Absicherung und keine Regel aus MSE2**: dessen Zerteiler
// (`src/script/parser.cpp`, Zeilen 444–454 und 718–719, Fassung f9d9356 vom
// 2021-07-06) steigt ebenso ungezählt hinab. Gemessen am 2026-10-07: die
// Pakete von `full-magic-pack` schachteln höchstens 13 Klammern tief
// (`magic.mse-game/script`); 256 lassen jedem echten Skript Raum und halten
// den Stapel klein.
const HoechsteSchachtelung = 256

type zerteiler struct {
	merkmale []Merkmal
	stelle   int
	ende     Merkmal

	// tiefe zählt die offenen Ebenen der Schachtelung ([hinab]) — über
	// eingebettete Textskripte hinweg, die einen eigenen Zerteiler bekommen.
	tiefe int
}

// hinab zählt eine Ebene der Schachtelung — und weist eine über
// [HoechsteSchachtelung] ab, an der Stelle, an der sie beginnt.
//
// **Gezählt wird, wo geschachtelt wird**: jeder Ausdruck und jedes weitere
// Vorzeichen einer Kette. **Die rechte Seite eines Operators nicht** — sie
// liest immer eine Stufe höher (`rechterRand`), und so steigt sie höchstens
// um die Zahl der Vorrangstufen; eine Folge gleichrangiger Operatoren läuft
// in einer Schleife. Gemessen mit `1 or else 1 or else …`, der einzigen
// Kette, deren rechte Seite tiefer beginnt: sie schachtelt nicht.
func (z *zerteiler) hinab() error {
	if z.tiefe >= HoechsteSchachtelung {
		return z.hier(SchluesselZuTiefGeschachtelt, strconv.Itoa(HoechsteSchachtelung))
	}

	z.tiefe++

	return nil
}

// hinauf schließt eine Ebene.
func (z *zerteiler) hinauf() { z.tiefe-- }

// endstelle gibt die Stelle **hinter** dem letzten Merkmal.
//
// Sie ist es, die ein unvollständiger Ausdruck meldet: „a :=" hört nach dem
// Zuweisungszeichen auf, und der Finger zeigt dorthin, wo der Wert fehlt — nicht
// auf das letzte Zeichen, das noch da war.
func endstelle(merkmale []Merkmal) Merkmal {
	if len(merkmale) == 0 {
		return Merkmal{Zeile: 1, Spalte: 1}
	}

	letztes := merkmale[len(merkmale)-1]
	if letztes.Art == ArtZeilenende {
		return Merkmal{Zeile: letztes.Zeile + 1, Spalte: 1}
	}

	// Über `Breite`, nicht über den Text: eine Zeichenkette ist im Quelltext
	// zwei Zeichen breiter als ihr Inhalt.
	return Merkmal{Zeile: letztes.Zeile, Spalte: letztes.Spalte + letztes.Breite}
}

// ------------------------------------------------------------------- Ablesen

func (z *zerteiler) amEnde() bool { return z.stelle >= len(z.merkmale) }

func (z *zerteiler) jetzt() Merkmal {
	if z.amEnde() {
		return z.ende
	}

	return z.merkmale[z.stelle]
}

func (z *zerteiler) hier(schluessel Meldungsschluessel, werte ...string) *Fehler {
	stelle := z.jetzt()

	return NeuerFehler(stelle.Zeile, stelle.Spalte, schluessel, werte...)
}

// istOperator sieht über Zeilenumbrüche hinweg **nicht** hinweg.
func (z *zerteiler) istOperator(zeichen ...string) bool {
	jetzt := z.jetzt()
	if z.amEnde() || jetzt.Art != ArtOperator {
		return false
	}

	for _, gesucht := range zeichen {
		if jetzt.Text == gesucht {
			return true
		}
	}

	return false
}

func (z *zerteiler) istWort(woerter ...string) bool {
	jetzt := z.jetzt()
	if z.amEnde() || jetzt.Art != ArtSchluesselwort {
		return false
	}

	for _, gesucht := range woerter {
		if jetzt.Text == gesucht {
			return true
		}
	}

	return false
}

// umbruecheUeberspringen geht über Zeilenenden hinweg.
//
// **Mitten in einem Ausdruck ist ein Umbruch nichts.** In `yugioh` steht die
// else-Kette über sechs Zeilen, in `magic-m15` die and-Kette über vier. Wer den
// Umbruch dort für ein Ende hält, liest die Hälfte jeder Vorlage falsch.
func (z *zerteiler) umbruecheUeberspringen() {
	for !z.amEnde() && z.merkmale[z.stelle].Art == ArtZeilenende {
		z.stelle++
	}
}

// klammerOffen meldet eine Klammer, die niemand schließt — **an ihrem
// Anfang**.
//
// # Warum an fünf Stellen dieselbe Funktion
//
// Es sind fünf: die Gruppierung `(…)`, der Funktionsrumpf `{…}`, die
// Abbildung `[…]`, der Index `x[…]` und der Aufruf `f(…)`. Jede hatte vorher
// ihre eigene Zeile, und jede zeigte auf eine andere Stelle — mal aufs
// Quellende, mal auf das Zeichen, an dem der Zerteiler stehen blieb. Genau so
// entsteht eine Klasse, die halb behoben ist.
//
// # Die Vorlage
//
//	if (opening) {
//	  size_t open_pos = opening->pos - begin;
//	  errors.push_back(ScriptParseError(open_pos, error_pos, …,
//	                                    opening->value, expected, peek(0).value));
//	}
//
// `src/script/parser.cpp`, Zeilen 334–337 in `TokenIterator::expected`,
// Fassung f9d9356 vom 2021-07-06. MSE übergibt das öffnende Merkmal überall
// dort, wo es eine Klammer schließen will, und legt den Fehler auf dessen
// Stelle.
//
// # Zwei Abweichungen, beide gewollt
//
// **MSE nennt zusätzlich, was es stattdessen gefunden hat** („instead of 'end
// of input'"). Das kann es sich leisten: sein Fehler trägt eine Spanne von
// der öffnenden Klammer bis dorthin, und der Editor hebt sie ganz hervor.
// Dieses Programm trägt **eine** Stelle; ein Fundstück ohne seine Stelle
// dazuzuschreiben schickte den Lesenden auf die Suche.
//
// **Und `min(1, 2` nennt hier die Klammer, dort das Wort.** MSE gibt
// `(0): Expected closing ')' for this 'min'` — weil `min`, `max` und `assert`
// dort feste Formen des Zerteilers sind und die Variable `token` noch auf dem
// Wort steht. In diesem Programm sind es Funktionen der Standardbibliothek
// (`bibliothek_werte.go`); ihr Aufruf geht denselben Weg wie jeder andere,
// und die Stelle ist die Klammer. Das ist eine Vereinheitlichung, keine
// Verschlechterung — die Stelle liegt ein Zeichen weiter rechts.
func klammerOffen(beginn Merkmal, schluss string) *Fehler {
	return NeuerFehler(
		beginn.Zeile, beginn.Spalte, SchluesselKlammerOffen,
		"`"+beginn.Text+"`", "`"+schluss+"`",
	)
}
