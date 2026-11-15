package domain_test

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Wie MSE zwei Werte ordnet** — Werkstück 3 von Phase 3.7
// (#1168).

// **Zahlen in einem Text zählen als Zahlen.**
//
// Das ist der Grund, warum es diese Ordnung gibt: eine Kartenliste, in der
// »Karte 10« vor »Karte 2« steht, ist eine falsch sortierte Liste.
func TestZahlenImTextZaehlenAlsZahlen(t *testing.T) {
	t.Parallel()

	roh := []string{"Karte 10", "Karte 2", "Karte 1", "Karte 20", "Karte 3"}
	slices.SortFunc(roh, domain.Vergleiche)

	erwartet := []string{"Karte 1", "Karte 2", "Karte 3", "Karte 10", "Karte 20"}
	if !slices.Equal(roh, erwartet) {
		t.Errorf("%v", roh)
	}
}

// **Groß und klein sind gleich, Umlaute stehen bei ihrem Grundbuchstaben.**
//
//	return dec == ' ' ? toLower(c) : dec;
//
// `remove_accents`, `src/util/string.cpp`, Zeile 221, Fassung f9d9356 vom
// 2021-07-06.
func TestUmlauteStehenBeiIhremGrundbuchstaben(t *testing.T) {
	t.Parallel()

	roh := []string{"Zebra", "Ärger", "apfel", "Öl", "Banane", "Über", "ähnlich"}
	slices.SortFunc(roh, domain.Vergleiche)

	erwartet := []string{"ähnlich", "apfel", "Ärger", "Banane", "Öl", "Über", "Zebra"}
	if !slices.Equal(roh, erwartet) {
		t.Errorf("%v", roh)
	}
}

// **Ein zusammengesetzter Buchstabe zerfällt in zwei.**
//
//	} else if (c == 0xC6 || c == 0xE6 || …) {
//	  return _('e'); // "ae" or "oe"
//
// `decompose_char2`, dieselbe Datei, Zeilen 226–238. `Æther` steht damit
// zwischen `Aegis` und `Afar` — für Magic ist das der Fall, den es gibt.
func TestEinZusammengesetzterBuchstabeZerfaelltInZwei(t *testing.T) {
	t.Parallel()

	roh := []string{"Afar", "Æther", "Aegis", "Adler"}
	slices.SortFunc(roh, domain.Vergleiche)

	erwartet := []string{"Adler", "Aegis", "Æther", "Afar"}
	if !slices.Equal(roh, erwartet) {
		t.Errorf("%v", roh)
	}
}

// **Die kürzere Zeichenkette kommt zuerst.**
//
//	// When we are at the end; shorter strings come first
//
// dieselbe Datei, Zeile 298.
func TestDieKuerzereZeichenketteKommtZuerst(t *testing.T) {
	t.Parallel()

	roh := []string{"Blitzschlag", "Blitz", "Blitzen"}
	slices.SortFunc(roh, domain.Vergleiche)

	erwartet := []string{"Blitz", "Blitzen", "Blitzschlag"}
	if !slices.Equal(roh, erwartet) {
		t.Errorf("%v", roh)
	}
}

// **Zwei gleich lange Zahlen entscheiden über ihre Ziffern.**
func TestGleichLangeZahlenEntscheidenUeberIhreZiffern(t *testing.T) {
	t.Parallel()

	fuer := func(links, rechts string, erwartet int) {
		t.Helper()

		if got := domain.Vergleiche(links, rechts); got != erwartet {
			t.Errorf("Vergleiche(%q, %q) = %d, erwartet %d",
				links, rechts, got, erwartet)
		}
	}

	fuer("a12b", "a13b", -1)
	fuer("a13b", "a12b", 1)
	fuer("a12b", "a12b", 0)

	// **Eine längere Zahl ist die größere** — auch mit führenden Nullen davor.
	fuer("a9b", "a10b", -1)
	fuer("a010", "a9", 1)

	// **Und eine Zahl, die weiterläuft, gewinnt am Ende.**
	fuer("a1", "a12", -1)
	fuer("a12", "a1", 1)
}

// **Zwei leere Werte sind gleich, und ein leerer steht vorn.**
func TestEinLeererWertStehtVorn(t *testing.T) {
	t.Parallel()

	if domain.Vergleiche("", "") != 0 {
		t.Error("zwei leere sind nicht gleich")
	}

	if domain.Vergleiche("", "a") != -1 || domain.Vergleiche("a", "") != 1 {
		t.Error("der leere steht nicht vorn")
	}
}

// **Steuerzeichen ordnen nach ihrem Codepunkt.**
//
//	} else {
//	  // control characters
//	  if (a < b) return -1;
//	  else       return 1;
//	}
//
// dieselbe Datei, Zeilen 290–294. **Und `\x01` kommt hier wirklich vor**: es
// ist MSEs geschütztes Kleinerzeichen in einem Feldwert.
func TestSteuerzeichenOrdnenNachIhremCodepunkt(t *testing.T) {
	t.Parallel()

	if domain.Vergleiche("a\x01b", "a\x02b") != -1 {
		t.Error("die Steuerzeichen ordnen nicht")
	}
}

// **Ein Zeichen jenseits der Tabellen ordnet nach seiner Kleinschreibung.**
//
// Das ist MSEs `else`-Zweig in `remove_accents` — und zugleich die Stelle, an
// der diese Portierung weniger kann als das Original: MSE führt zusätzlich
// Tabellen für *Latin Extended B* und *Latin Additional*. Was das kostet,
// steht bei [domain.Vergleiche].
func TestEinFremdesZeichenOrdnetNachSeinerKleinschreibung(t *testing.T) {
	t.Parallel()

	// Griechisch: Σ und σ sind dasselbe Zeichen in zwei Schreibungen.
	if domain.Vergleiche("Σ", "σ") != 0 {
		t.Error("groß und klein unterscheiden sich")
	}
}

// **Die übrigen Zweige der Ordnung** — jeder einmal betreten.
//
// **Warum als Tabelle und nicht als Prosa**: es sind die Fälle, die MSEs
// Schleife mit `return` verlässt, und jeder von ihnen ist ein Satz über zwei
// Zeichenketten. Ein Zweig ohne Lauf wäre eine Behauptung (§3).
func TestJederZweigDerOrdnungWirdBetreten(t *testing.T) {
	t.Parallel()

	fälle := []struct {
		name, links, rechts string
		erwartet            int
	}{
		// **Latin Extended A** — die zweite portierte Tabelle. `ł` (0x142)
		// steht unter `l`.
		{"Latin Extended A", "łampe", "lampe", 0},

		// **Ein Buchstabe, für den die Tabelle ein Leerzeichen führt**:
		// `ß` (0xDF) hat keinen Grundbuchstaben und fällt in die
		// Kleinschreibung.
		{"kein Grundbuchstabe", "ß", "ß", 0},

		// **Zusammengesetzt gegen einfach, und danach ist Schluss.**
		//
		//	if (la2) a = la2;
		//	else { if (++pa >= na) return 1; a = sa.GetChar(pa); }
		//	if (lb2) b = lb2;
		//	else { if (++pb >= nb) return -1; b = sb.GetChar(pb); }
		//
		// `Æ` und `A` haben denselben Grundbuchstaben; `Æ` zerfällt und legt
		// ein `e` nach, `A` nicht — und `A` ist damit zu Ende. MSE gibt dann
		// **−1**: die kürzere Zeichenkette kommt zuerst, gerechnet von der
		// **linken** Seite aus.
		//
		// **Nachgerechnet, nicht erwartet.** Der erste Anlauf stand hier auf
		// 1, weil „Æ ist mehr als A" plausibel klingt.
		{"Æ am Ende gegen A", "Æ", "A", -1},
		{"A gegen Æ am Ende", "A", "Æ", 1},

		// **Zusammengesetzt auf beiden Seiten** — beide zerfallen, und die
		// zweiten Buchstaben sind gleich.
		{"Æ gegen æ", "Æ", "æ", 0},

		// **»ij« zerfällt in »i« und »j«.**
		{"ĳ gegen ij", "ĳ", "ij", 0},

		// **»dz« zerfällt gerade nicht — und das ist eine Festlegung dieses
		// Projekts, keine Portierung.**
		//
		// Hier stand, MSEs dritter Zweig sei unerreichbar und `ǆ` falle
		// deshalb in die Kleinschreibung. Der Zweitblick über Phase 3.7 hat
		// das angestrichen (Befund 8), und nachgerechnet stimmt es nicht:
		// `ǆ` (`0x1C6`) landet in MSEs viertem Zweig bei `latin_E[-7226]` —
		// **vor** der Tabelle. Was MSE dort zurückgibt, hängt vom Speicher
		// ab; ein Erwartungswert lässt sich daraus nicht ableiten.
		//
		// **Die Zeile bleibt trotzdem**, denn sie hält fest, was *hier*
		// gilt: alles jenseits der zwei portierten Tabellen ordnet nach
		// seiner Kleinschreibung, und damit steht `ǆ` hinter `dz`. Die
		// Begründung steht ausführlich bei [domain.Vergleiche].
		{"ǆ gegen dz", "ǆ", "dz", 1},

		// **Und die zwei Punkte, an denen MSE doch etwas Bestimmtes tut** —
		// Befund 6 des Nachlaufs über Phase 3.7. In beiden kaputten Zweigen
		// ist der Index an **einer** Stelle gerade null, und dort liest MSE
		// das erste Zeichen seiner Tabelle: `latin_B[0]` ist `'b'`,
		// `latin_E[0]` ist `'a'`.
		//
		// **Sie sind portiert**, und dieser Lauf hält sie: `ƀ` ordnet unter
		// `b`, `Ḁ` unter `a` — nicht nach ihrem Codepunkt.
		{"ƀ ordnet unter b", "ƀa", "bb", -1},
		{"Ḁ ordnet unter a", "Ḁz", "ab", 1},

		// **Zahl gegen Nichtzahl mitten im Wort.**
		{"Ziffer gegen Buchstabe", "a1", "ab", -1},

		// **Zwei Zahlen gleicher Länge, verschieden, danach geht es weiter.**
		{"Abweichung wirkt nach", "a12x", "a13a", -1},

		// **Und eine gleich lange Zahl, gefolgt von Verschiedenem.**
		{"gleiche Zahl, dann Text", "a12b", "a12c", -1},

		// **Zwei »dz«-Buchstaben zerfallen doch** — wenn beide Seiten einen
		// tragen. Erst dann sind die Grundbuchstaben gleich, und
		// `decompose_char2` kommt zum Zug.
		{"ǆ gegen ǅ", "ǆ", "ǅ", 0},

		// **Eine Zahl läuft weiter, mitten im Wort**: `12` gegen `1`, und
		// danach steht bei einem noch ein Buchstabe.
		//
		//	} else if (in_num && da) {
		//	  // comparing numbers, one is longer, therefore it is greater
		//	  return 1;
		{"längere Zahl mitten im Wort", "a12b", "a1b", 1},
		{"kürzere Zahl mitten im Wort", "a1b", "a12b", -1},

		// **Steuerzeichen in der anderen Richtung.**
		{"größeres Steuerzeichen", "a\x02", "a\x01", 1},

		// **Beide enden in einer gleich langen, verschiedenen Zahl.** Dann
		// entscheidet erst der Schluss — die gemerkte Abweichung.
		{"Zahl am Ende", "a12", "a13", -1},
		{"Zahl am Ende, umgekehrt", "a13", "a12", 1},

		// **Und die eine ist hinter der Zahl noch länger.**
		//
		//	} else if (na - pa > nb - pb) {
		//	  Char a = sa.GetChar(pa);
		//	  if (isDigit(a) || eq) return 1;  // a is longer
		{"nach gleicher Zahl noch ein Zeichen", "a12x", "a12", 1},

		// **Und zwei Werte, die in derselben Zahl enden, sind gleich.**
		// Ohne den Buchstaben dahinter — mit ihm entscheidet schon die
		// Schleife, und der Schluss käme gar nicht dran.
		{"dieselbe Zahl am Ende", "a12", "a12", 0},
	}

	for _, fall := range fälle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := domain.Vergleiche(fall.links, fall.rechts); got != fall.erwartet {
				t.Errorf("Vergleiche(%q, %q) = %d, erwartet %d",
					fall.links, fall.rechts, got, fall.erwartet)
			}
		})
	}
}

// **Die Ausrichtung einer Spalte kommt aus einem Wort in mehreren.**
func TestDieListenausrichtungLiestIhrWort(t *testing.T) {
	t.Parallel()

	for angabe, erwartet := range map[string]domain.Listenausrichtung{
		"":             domain.ListeLinks,
		"left":         domain.ListeLinks,
		"right":        domain.ListeRechts,
		"center":       domain.ListeMitte,
		"centre":       domain.ListeMitte,
		"middle right": domain.ListeRechts,
		"top center":   domain.ListeMitte,
		"MIDDLE LEFT":  domain.ListeLinks,
		"schräg":       domain.ListeLinks,
		"middle":       domain.ListeLinks,
	} {
		if got := domain.ListenausrichtungAus(angabe); got != erwartet {
			t.Errorf("ListenausrichtungAus(%q) = %q, erwartet %q", angabe, got, erwartet)
		}
	}
}
