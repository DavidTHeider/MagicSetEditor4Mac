package msescript

import "unicode"

// **Wie MSE zwei Ordnungsschlüssel vergleicht** —
// #2006.
//
// # ported: `smart_compare`
//
// `src/util/string.cpp`, Zeilen 208–318, Fassung f9d9356 vom 2021-07-06 —
// `remove_accents`, `decompose_char2` und `smart_compare`. `sort_list` ordnet
// damit (`smart_less_first`, `src/script/functions/basic.cpp`, Zeilen
// 478–520): Zahlen im Text zählen als Zahlen, groß und klein sind gleich,
// Umlaute stehen beim Grundbuchstaben, `Æ` zerfällt in `a` und `e`.
//
// # Zweimal da, und das mit Absicht
//
// **Dieselbe Portierung steht als `domain.Vergleiche`** — dort ordnet sie
// Kartenlisten und Statistiken. Dieses Paket importiert nichts aus dem
// Projekt, auch nicht `domain` (§2), und `domain` nichts außer der
// Standardbibliothek; eine gemeinsame Stelle gäbe es nur, wenn eine der beiden
// Regeln fiele. **Was zwei Stellen kosten, hält ein Lauf fest**:
// `internal/service/ordnungsgleich_test.go` fährt beide über dieselben Paare
// und über gefuzzte, und eine Abweichung ist rot.
//
// Die Einzelheiten — die zwei portierten Tabellen, die zwei bestimmten Punkte
// in MSEs kaputten Zweigen, warum alles andere kleingeschrieben ordnet — stehen
// ausführlich an `domain.Vergleiche` (`internal/domain/ordnung.go`).

// grundLatin1 ist MSEs `latin_1` — der Grundbuchstabe von `0xC0` bis `0xFF`;
// ein Leerzeichen heißt „keiner".
const grundLatin1 = "aaaaaaaceeeeiiii" +
	"dnooooo ouuuuy  " +
	"aaaaaaaceeeeiiii" +
	"dnooooo ouuuuy y"

// grundLatinA ist MSEs `latin_A` — von `0x100` bis `0x17F`; MSE liest sie
// nur bis `0x17E`.
const grundLatinA = "aaaaaaccccccccdd" +
	"ddeeeeeeeeeegggg" +
	"gggghhhhiiiiiiii" +
	"iiiijjkkklllllll" +
	"lllnnnnnnnnnoooo" +
	"oooorrrrrrssssss" +
	"ssttttttuuuuuuuu" +
	"uuuuwwyyyzzzzzzs"

// grundbuchstabe ist MSEs `remove_accents`:
//
//	return dec == ' ' ? toLower(c) : dec;
func grundbuchstabe(zeichen rune) rune {
	switch {
	case zeichen >= 0xC0 && zeichen <= 0xFF:
		if grund := rune(grundLatin1[zeichen-0xC0]); grund != ' ' {
			return grund
		}
	case zeichen >= 0x100 && zeichen <= 0x17E:
		return rune(grundLatinA[zeichen-0x100])
	case zeichen == 0x180:
		return 'b' // latin_B[0]
	case zeichen == 0x1E00:
		return 'a' // latin_E[0]
	}

	return unicode.ToLower(zeichen)
}

// zweiterBuchstabe ist MSEs `decompose_char2` — der zweite Buchstabe eines
// zusammengesetzten, 0 sonst.
func zweiterBuchstabe(zeichen rune) rune {
	switch zeichen {
	case 0xC6, 0xE6, 0x152, 0x153, 0x1E2, 0x1E3, 0x1FC, 0x1FD:
		return 'e' // »ae« oder »oe«
	case 0x132, 0x133, 0x1C7, 0x1C8, 0x1C9, 0x1CA, 0x1CB, 0x1CC:
		return 'j' // »ij«, »lj«, »nj«
	case 0x1C4, 0x1C5, 0x1C6, 0x1F1, 0x1F2, 0x1F3:
		return 'z' // »dz«
	}

	return 0
}

// KlugVergleiche ordnet zwei Schlüssel wie MSEs `smart_compare`: −1, 0 oder 1.
func KlugVergleiche(links, rechts string) int {
	lauf := vergleichslauf{a: []rune(links), b: []rune(rechts), gleich: true}

	for lauf.pa < len(lauf.a) && lauf.pb < len(lauf.b) {
		if ergebnis, fertig := lauf.schritt(); fertig {
			return ergebnis
		}
	}

	return lauf.schluss()
}

// vergleichslauf trägt MSEs `in_num`, `eq`, `lt` und die zwei Zeiger.
type vergleichslauf struct {
	a, b   []rune
	pa, pb int

	inZahl bool

	// gleich und kleiner merken die erste Abweichung innerhalb einer Zahl;
	// `gleich` beginnt wahr wie MSEs `bool eq = true`.
	gleich  bool
	kleiner bool
}

// schritt vergleicht ein Zeichenpaar; die innere Schleife ist MSEs
// `goto next` nach einer Zerlegung.
func (l *vergleichslauf) schritt() (int, bool) {
	zeichenA, zeichenB := l.a[l.pa], l.b[l.pb]

	for {
		ziffernA, ziffernB := unicode.IsDigit(zeichenA), unicode.IsDigit(zeichenB)

		switch {
		case ziffernA && ziffernB:
			if l.gleich && zeichenA != zeichenB {
				l.gleich, l.kleiner = false, zeichenA < zeichenB
			}
		case l.inZahl && ziffernA:
			return 1, true // die längere Zahl ist die größere
		case l.inZahl && ziffernB:
			return -1, true
		case l.inZahl && !l.gleich:
			return richtung(l.kleiner), true
		case zeichenA != zeichenB:
			l.gleich, l.kleiner = true, false

			if ergebnis, fertig := buchstabenpaar(zeichenA, zeichenB); fertig {
				return ergebnis, true
			}

			naechstesA, naechstesB, ergebnis := l.zerlege(zeichenA, zeichenB)
			if ergebnis != 0 {
				return ergebnis, true
			}

			if naechstesA != 0 {
				zeichenA, zeichenB = naechstesA, naechstesB

				continue
			}
		}

		// Alle Wege laufen hier zusammen, wie MSEs `in_num = da && db;`.
		l.inZahl = ziffernA && ziffernB
		l.pa++
		l.pb++

		return 0, false
	}
}

// richtung macht aus „kleiner" −1 und sonst 1.
func richtung(kleiner bool) int {
	if kleiner {
		return -1
	}

	return 1
}

// buchstabenpaar entscheidet über zwei verschiedene Zeichen, soweit es geht.
// Steuerzeichen ordnen nach dem Codepunkt — `\x01` ist MSEs geschütztes
// Kleinerzeichen in einem Feldwert.
func buchstabenpaar(zeichenA, zeichenB rune) (int, bool) {
	if zeichenA < 0x20 || zeichenB < 0x20 {
		return richtung(zeichenA < zeichenB), true
	}

	grundA, grundB := grundbuchstabe(zeichenA), grundbuchstabe(zeichenB)
	if grundA == grundB {
		return 0, false
	}

	return richtung(grundA < grundB), true
}

// zerlege gibt die nächsten zwei Zeichen, wenn eines zusammengesetzt war,
// oder ein Ergebnis, wenn eine Kette dabei zu Ende ist:
//
//	if (++pa >= na) return 1;
//
// `src/util/string.cpp`, Zeile 280, Fassung f9d9356 vom 2021-07-06.
func (l *vergleichslauf) zerlege(zeichenA, zeichenB rune) (naechstesA, naechstesB rune, ergebnis int) {
	zweitA, zweitB := zweiterBuchstabe(zeichenA), zweiterBuchstabe(zeichenB)
	if zweitA == 0 && zweitB == 0 {
		return 0, 0, 0
	}

	if zweitA == 0 {
		l.pa++
		if l.pa >= len(l.a) {
			return 0, 0, 1
		}

		zweitA = l.a[l.pa]
	}

	if zweitB == 0 {
		l.pb++
		if l.pb >= len(l.b) {
			return 0, 0, -1
		}

		zweitB = l.b[l.pb]
	}

	return zweitA, zweitB, 0
}

// schluss ordnet, wenn eine Kette zu Ende ist — kürzere zuerst, und eine Zahl,
// die weiterläuft, ist größer.
func (l *vergleichslauf) schluss() int {
	restA, restB := len(l.a)-l.pa, len(l.b)-l.pb

	if !l.inZahl {
		switch {
		case restA == restB:
			return 0
		case restA < restB:
			return -1
		default:
			return 1
		}
	}

	if restA < restB && (unicode.IsDigit(l.b[l.pb]) || l.gleich) {
		return -1
	}

	if restA > restB && (unicode.IsDigit(l.a[l.pa]) || l.gleich) {
		return 1
	}

	if l.gleich {
		return 0
	}

	return richtung(l.kleiner)
}
