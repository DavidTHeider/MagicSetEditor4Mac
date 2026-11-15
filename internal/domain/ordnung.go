package domain

import "unicode"

// **Wie MSE zwei Werte ordnet** — Werkstück 3 von Phase 3.7
// (#1168).
//
// # ported: `smart_compare`
//
// `src/util/string.cpp`, Zeilen 208–318, Fassung f9d9356 vom 2021-07-06 —
// `remove_accents`, `decompose_char2` und `smart_compare`. Drei Eigenschaften
// stehen darin, und alle drei gelten hier:
//
//  1. **Zahlen im Text zählen als Zahlen.** »Karte 10« steht hinter
//     »Karte 2« — der Grund, warum es diese Ordnung überhaupt gibt.
//  2. **Groß und klein sind gleich, Umlaute stehen beim Grundbuchstaben.**
//     `Ärger` steht zwischen `apfel` und `Banane`.
//  3. **Ein zusammengesetzter Buchstabe zerfällt in zwei.** `Æther` steht
//     zwischen `Aegis` und `Afar` — für Magic ist das der Fall, den es gibt.
//
// # adapted: zwei Tabellen weniger
//
// MSE führt vier Tabellen: *Latin 1* (`0xC0`–`0xFF`), *Latin Extended A*
// (`0x100`–`0x17E`), *Latin Extended B* und *Latin Additional*. **Hier sind
// die ersten beiden portiert**, und alles darüber fällt in MSEs eigenen
// `else`-Zweig: es ordnet nach seiner Kleinschreibung.
//
// **Was das kostet, gehört ausgesprochen** (§15, Regel 7): ein vietnamesischer
// Kartenname mit `ẹ` (`0x1EB9`) ordnet nach seinem Codepunkt statt unter `e`
// — er stünde hinter `z`. Die beiden portierten Tabellen decken Deutsch,
// Französisch, Spanisch, Italienisch, Polnisch, Tschechisch und die übrigen
// Sprachen, in denen die `.mse-locale` dieses Packs geschrieben sind.
//
// # Und MSEs letzte zwei Zweige haben kein Verhalten, das sich portieren ließe
//
// **Nachgerechnet am 2026-08-27** (Zweitblick über Phase 3.7, Befund 8). Der
// dritte Zweig ist **nicht** unerreichbar, und was wirklich dort steht, ist
// schlimmer:
//
//	} else if (c <= 0x180 && c <= 0x240) {   // Latin extended B
//	  dec = latin_B[c - 0x180];
//	} else if (c <= 0x1E00 && c <= 0x1EFF) { // Latin additional
//	  dec = latin_E[c - 0x1E00];
//	}
//
// `remove_accents`, `src/util/string.cpp`, Zeilen 208–222, Fassung `f9d9356`
// vom 2021-07-06. Beide Bedingungen prüfen zweimal dieselbe Richtung, und
// beide Male ist die **erste** die engere:
//
//   - *Latin Extended B* greift für genau zwei Zeichen, `0x17F` und `0x180`
//     — und für `ſ` (`0x17F`) ist der Index **`-1`**: gelesen wird vor der
//     Tabelle. Für `ƀ` (`0x180`) ist er **null**, und damit ist dieser eine
//     Punkt bestimmt: `latin_B[0]` ist `'b'`.
//   - *Latin Additional* greift für **alles** von `0x181` bis `0x1E00` und
//     rechnet dabei negativ; `ǆ` (`0x1C6`) landet bei `latin_E[-7226]`. Am
//     oberen Ende ist der Index wieder null: `Ḁ` (`0x1E00`) gibt `'a'`. Für
//     den Bereich, für den die Tabelle geschrieben ist, greift sie **nie**:
//     `ẹ` (`0x1EB9`) fällt an `c <= 0x1E00` vorbei in die Kleinschreibung.
//
// **Beide Nullpunkte sind portiert** (siehe [ohneAkzent]) — Befund 6 des
// Nachlaufs über Phase 3.7: an genau diesen zwei Stellen hat MSE ein
// bestimmtes Ergebnis.
//
// **Für alles dazwischen gibt es nichts Bestimmtes zu portieren.** Was MSE
// für `ǆ` zurückgibt, hängt davon ab, welche Bytes vor seinen Tabellen im
// Speicher liegen. Dass diese Zeichen hier kleingeschrieben ordnen, ist
// deshalb **eine Festlegung dieses Projekts** — nicht eine Portierung und
// auch keine Nachbildung eines Fehlers.

// latin1 ist MSEs `latin_1` — der Grundbuchstabe von `0xC0` bis `0xFF`.
//
// **Ein Leerzeichen heißt „keiner"** und fällt in die Kleinschreibung
// zurück; das sind `×` (`0xD7`), `Ø` (`0xD8`), `Þ` (`0xDE`), `ß` (`0xDF`)
// und `þ` (`0xFE`).
const latin1 = "aaaaaaaceeeeiiii" +
	"dnooooo ouuuuy  " +
	"aaaaaaaceeeeiiii" +
	"dnooooo ouuuuy y"

// latinA ist MSEs `latin_A` — von `0x100` bis `0x17F`.
const latinA = "aaaaaaccccccccdd" +
	"ddeeeeeeeeeegggg" +
	"gggghhhhiiiiiiii" +
	"iiiijjkkklllllll" +
	"lllnnnnnnnnnoooo" +
	"oooorrrrrrssssss" +
	"ssttttttuuuuuuuu" +
	"uuuuwwyyyzzzzzzs"

// ohneAkzent ist MSEs `remove_accents`.
//
//	return dec == ' ' ? toLower(c) : dec;
func ohneAkzent(zeichen rune) rune {
	switch {
	case zeichen >= 0xC0 && zeichen <= 0xFF:
		if grund := rune(latin1[zeichen-0xC0]); grund != ' ' {
			return grund
		}
	case zeichen >= 0x100 && zeichen <= 0x17E:
		if grund := rune(latinA[zeichen-0x100]); grund != ' ' {
			return grund
		}

	// **Zwei Zeichen aus MSEs kaputten Zweigen sind trotzdem bestimmt** —
	// Befund 6 des Nachlaufs über Phase 3.7. Der Absatz oben rechnet nach,
	// warum die beiden letzten Tabellen kein Verhalten haben; an **einem**
	// Punkt je Zweig ist der Index gerade null, und dort liest MSE das erste
	// Zeichen seiner Tabelle:
	//
	//	ƀ (0x180) → latin_B[0] == 'b'
	//	Ḁ (0x1E00) → latin_E[0] == 'a'
	//
	// **Zwei Zeichen brauchen keine zwei Tabellen.** Sie stehen hier, weil
	// die Portierung an jedem **bestimmten** Punkt gelten soll; der Rest der
	// beiden Bereiche hat keinen.
	case zeichen == 0x180:
		return 'b'
	case zeichen == 0x1E00:
		return 'a'
	}

	return unicode.ToLower(zeichen)
}

// zweiterTeil ist MSEs `decompose_char2` — der **zweite** Buchstabe eines
// zusammengesetzten; 0, wenn es keiner ist.
func zweiterTeil(zeichen rune) rune {
	switch {
	case zeichen < 0xC6:
		return 0
	case zeichen == 0xC6 || zeichen == 0xE6 || zeichen == 0x152 ||
		zeichen == 0x153 || zeichen == 0x1E2 || zeichen == 0x1E3 ||
		zeichen == 0x1FC || zeichen == 0x1FD:
		return 'e' // »ae« oder »oe«
	case zeichen == 0x132 || zeichen == 0x133 ||
		(zeichen >= 0x1C7 && zeichen <= 0x1CC):
		return 'j' // »ij«, »lj«, »nj«
	case (zeichen >= 0x1C4 && zeichen <= 0x1C6) ||
		(zeichen >= 0x1F1 && zeichen <= 0x1F3):
		return 'z' // »dz«
	default:
		return 0
	}
}

// Vergleiche ordnet zwei Werte wie MSEs `smart_compare`.
//
// Gibt −1, 0 oder 1 — die Gestalt, die `slices.SortFunc` erwartet.
func Vergleiche(links, rechts string) int {
	stand := ordnungsstand{a: []rune(links), b: []rune(rechts), gleich: true}

	for stand.pa < len(stand.a) && stand.pb < len(stand.b) {
		if ergebnis, fertig := stand.schritt(); fertig {
			return ergebnis
		}
	}

	return stand.schluss()
}

// ordnungsstand ist der Lauf durch beide Zeichenketten.
//
// **Ein Verbund und keine sechs Stellen**: MSEs Schleife trägt `in_num`, `lt`,
// `eq` und zwei Zeiger, und sie springt mit `goto next` in ihre eigene Mitte
// zurück. Ein Verbund macht daraus einen Schritt, den man einzeln lesen kann.
type ordnungsstand struct {
	a, b   []rune
	pa, pb int

	// inZahl sagt, ob beide zuletzt eine Ziffer hatten.
	inZahl bool

	// gleich und kleiner merken sich die **erste** Abweichung innerhalb einer
	// Zahl: `12` gegen `13` entscheidet sich an der zweiten Ziffer, gilt aber
	// erst, wenn beide Zahlen gleich lang enden. **`gleich` beginnt wahr**,
	// wie MSEs `bool eq = true`.
	gleich  bool
	kleiner bool
}

// schritt vergleicht ein Zeichenpaar. `fertig` heißt: die Ordnung steht fest.
//
// **Die innere Schleife ist MSEs `goto next`**: ein zusammengesetzter
// Buchstabe wird zerlegt, und dieselbe Stelle läuft noch einmal — mit den
// zweiten Buchstaben und **ohne** beide Zeiger weiterzurücken.
func (s *ordnungsstand) schritt() (int, bool) {
	zeichenA, zeichenB := s.a[s.pa], s.b[s.pb]

	for {
		ziffernA, ziffernB := unicode.IsDigit(zeichenA), unicode.IsDigit(zeichenB)

		switch {
		case ziffernA && ziffernB:
			s.inZahl = true

			if s.gleich && zeichenA != zeichenB {
				s.gleich, s.kleiner = false, zeichenA < zeichenB
			}

		case s.inZahl && ziffernA:
			// **Zwei Zahlen, eine ist länger** — die längere ist die größere.
			return 1, true

		case s.inZahl && ziffernB:
			return -1, true

		case s.inZahl && !s.gleich:
			// **Gleich lang, aber verschieden** — die gemerkte Abweichung gilt.
			if s.kleiner {
				return -1, true
			}

			return 1, true

		case zeichenA != zeichenB:
			s.gleich, s.kleiner = true, false

			ergebnis, fertig := s.zeichenpaar(zeichenA, zeichenB)
			if fertig {
				return ergebnis, true
			}

			if zerlegtA, zerlegtB, weiter := s.zerlegung(zeichenA, zeichenB); weiter {
				zeichenA, zeichenB = zerlegtA, zerlegtB

				continue
			} else if zerlegtA == abbruchLinks || zerlegtA == abbruchRechts {
				if zerlegtA == abbruchLinks {
					return 1, true
				}

				return -1, true
			}
		}

		// **Hier laufen alle Wege zusammen**, so wie bei MSE: `in_num = da &&
		// db;` steht am Ende der Schleife, hinter jedem `else if`, das nicht
		// zurückgekehrt ist. Wer den Zweig „gleicher Grundbuchstabe" daran
		// vorbeiführte, ließe `in_num` von der vorigen Runde stehen.
		s.inZahl = ziffernA && ziffernB
		s.pa++
		s.pb++

		return 0, false
	}
}

// Die beiden Marken für „eine Zeichenkette ist mitten in einer Zerlegung zu
// Ende" — sie können keine Rune sein, weil jede Rune ein Zeichen wäre.
const (
	abbruchLinks  rune = -1
	abbruchRechts rune = -2
)

// zeichenpaar entscheidet über zwei **verschiedene** Zeichen, soweit es geht.
//
// `fertig` heißt: entschieden. Sonst haben beide denselben Grundbuchstaben,
// und der Aufrufer sieht nach, ob einer von ihnen zerfällt.
func (s *ordnungsstand) zeichenpaar(zeichenA, zeichenB rune) (int, bool) {
	// **Steuerzeichen ordnen nach ihrem Codepunkt.** `\x01` kommt wirklich
	// vor: es ist MSEs geschütztes Kleinerzeichen in einem Feldwert.
	if zeichenA < 0x20 || zeichenB < 0x20 {
		if zeichenA < zeichenB {
			return -1, true
		}

		return 1, true
	}

	switch grundA, grundB := ohneAkzent(zeichenA), ohneAkzent(zeichenB); {
	case grundA < grundB:
		return -1, true
	case grundA > grundB:
		return 1, true
	default:
		return 0, false
	}
}

// zerlegung gibt die nächsten Zeichen, wenn eines von beiden zusammengesetzt
// war — `weiter` sagt, dass der Schritt sie noch einmal vergleichen soll.
//
// Ist eine Zeichenkette dabei zu Ende, kommt eine der beiden Abbruchmarken
// zurück: MSE gibt dort unmittelbar 1 beziehungsweise −1.
func (s *ordnungsstand) zerlegung(zeichenA, zeichenB rune) (
	weiterA, weiterB rune, nochmal bool,
) {
	zweiA, zweiB := zweiterTeil(zeichenA), zweiterTeil(zeichenB)
	if zweiA == 0 && zweiB == 0 {
		return 0, 0, false
	}

	if zweiA == 0 {
		s.pa++
		if s.pa >= len(s.a) {
			//	if (++pa >= na) return 1;
			return abbruchLinks, 0, false
		}

		zweiA = s.a[s.pa]
	}

	if zweiB == 0 {
		s.pb++
		if s.pb >= len(s.b) {
			return abbruchRechts, 0, false
		}

		zweiB = s.b[s.pb]
	}

	return zweiA, zweiB, true
}

// schluss ordnet, wenn eine der beiden Zeichenketten zu Ende ist.
//
//	// When we are at the end; shorter strings come first
//
// `src/util/string.cpp`, Zeile 298, Fassung f9d9356 vom 2021-07-06.
func (s *ordnungsstand) schluss() int {
	restA, restB := len(s.a)-s.pa, len(s.b)-s.pb

	if !s.inZahl {
		switch {
		case restA == restB:
			return 0
		case restA < restB:
			return -1
		default:
			return 1
		}
	}

	// **Beide endeten in einer Zahl.** Läuft die eine weiter, ist sie größer.
	if restA < restB && (unicode.IsDigit(s.b[s.pb]) || s.gleich) {
		return -1
	}

	if restA > restB && (unicode.IsDigit(s.a[s.pa]) || s.gleich) {
		return 1
	}

	switch {
	case s.gleich:
		return 0
	case s.kleiner:
		return -1
	default:
		return 1
	}
}
