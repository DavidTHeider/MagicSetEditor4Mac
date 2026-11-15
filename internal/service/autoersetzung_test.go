package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Der Autoersetzer, portiert aus `TextValueEditor::tryAutoReplace`
// (#897).
//
// Die Fälle beschreiben, was MSE2 tut — nachgelesen in
// `src/gui/value/text.cpp`, Zeilen 1033–1052 und 1116–1165, Fassung f9d9356 vom
// 2021-07-06.

func regeln() []msepaket.AutoErsetzung {
	return []msepaket.AutoErsetzung{
		{Muster: "(C)", Ersatz: "©", GanzesWort: true},
		{Muster: "AE", Ersatz: "Æ"},
		{Muster: "TM", Ersatz: "™"},
		{Muster: "--", Ersatz: "—", GanzesWort: true},
	}
}

// **Ersetzt wird rückwärts von der Schreibmarke, nicht im ganzen Text.**
//
// Das ist der Kern der Portierung: MSE2 prüft je Regel, ob der Text **vor**
// der Marke auf das Muster endet.
func TestErsetztWirdAnDerSchreibmarke(t *testing.T) {
	faelle := []struct {
		name    string
		text    string
		stelle  int
		nachher string
		danach  int
	}{
		// **Die Stellen sind Bytes**, nicht Zeichen: `©` wiegt zwei Bytes,
		// `Æ` ebenfalls. Die Umrechnung auf das, was das Fenster zählt,
		// gehört an die Naht und nicht hierher.
		{"das Muster endet an der Marke", "(C)", 3, "©", 2},
		{"hinter der Marke bleibt alles stehen", "(C)x", 3, "©x", 2},
		{
			"vor der Marke wird nichts angefasst",
			"(C) und (C)", 11, "(C) und ©", 10,
		},
		{
			"das Muster endet nicht an der Marke",
			"(C)x", 4, "(C)x", 4,
		},
		{"die Marke steht am Anfang", "(C)", 0, "(C)", 0},
		{"leerer Text", "", 0, "", 0},
		{
			"ohne ganzes Wort greift es auch mitten im Wort",
			"xAE", 3, "xÆ", 3,
		},

		// **Und mit ganzem Wort auch** — das überrascht und ist portiert.
		// `isWordBoundary(1)` sieht auf `(`, und `(` steht in
		// `word_bound_chars` (`src/gui/value/text.cpp`, Zeile 1116, Fassung f9d9356
		// vom 2021-07-06). Die
		// Grenze gilt damit, obwohl links ein `x` steht.
		{
			"eine Klammer ist selbst eine Wortgrenze",
			"x(C)", 4, "x©", 3,
		},
	}

	ersetzer := service.NeuerAutoersetzer(regeln())

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			text, stelle := ersetzer.AnStelle(fall.text, fall.stelle)

			if text != fall.nachher || stelle != fall.danach {
				t.Errorf("%q@%d → %q@%d, erwartet %q@%d",
					fall.text, fall.stelle, text, stelle, fall.nachher, fall.danach)
			}
		})
	}
}

// **Die Regeln laufen der Reihe nach, und eine kann auf die nächste treffen.**
//
// MSE2 setzt nach einer Ersetzung `end = selection_start_i` und geht mit der
// **nächsten** Regel weiter — die Schleife läuft einmal durch alle, nicht bis
// zur Ruhe.
func TestDieRegelnLaufenDerReiheNachEinmalDurch(t *testing.T) {
	// `a` → `b`, danach `b` → `c`: die zweite Regel sieht das Ergebnis der
	// ersten, weil sie **danach** kommt.
	nachher, stelle := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "a", Ersatz: "b"},
		{Muster: "b", Ersatz: "c"},
	}).AnStelle("a", 1)

	if nachher != "c" || stelle != 1 {
		t.Errorf("%q@%d, erwartet \"c\"@1", nachher, stelle)
	}

	// Umgekehrt nicht: `b` → `c` steht **vor** `a` → `b` und ist durch, bevor
	// das `b` entsteht.
	nachher, _ = service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "b", Ersatz: "c"},
		{Muster: "a", Ersatz: "b"},
	}).AnStelle("a", 1)

	if nachher != "b" {
		t.Errorf("%q, erwartet \"b\" — die Schleife läuft nicht bis zur Ruhe", nachher)
	}
}

// **Die Wortgrenze ist eine Oder-Bedingung, und sie sieht auf das Muster
// selbst** — `TextValueEditor::isWordBoundary`, `src/gui/value/text.cpp`,
// Zeilen 1138–1165, Fassung f9d9356 vom 2021-07-06.
//
// Zwei Eigenschaften, die zusammen überraschen und beide portiert sind:
//
//  1. Sie gilt an einer Stelle, wenn **entweder** das Zeichen dahinter ein
//     Grenzzeichen ist (oder der Text endet) **oder** das davor (oder der Text
//     beginnt). Nicht beides.
//  2. `isWordBoundary(start)` sieht auf das Zeichen **an** `start` — und das
//     ist das **erste Zeichen des Musters**. Ein Muster, das mit einem
//     Satzzeichen beginnt, erfüllt seine linke Grenze damit immer.
//
// **Für `--` heißt das: `whole word: true` ändert nichts.** Beide Striche sind
// Satzzeichen, also ist links wie rechts eine Grenze, wo immer das Muster
// steht. Das sieht nach einem Fehler der Vorlage aus; portiert wird trotzdem
// das Verhalten (§15), und dieser Kasten hält es fest, damit niemand es später
// für einen Fehler **dieses** Programms hält.
func TestDieWortgrenzeSiehtAufDasMusterSelbst(t *testing.T) {
	ersetzer := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "--", Ersatz: "—", GanzesWort: true},
	})

	faelle := map[string]struct {
		text     string
		stelle   int
		erwartet string
	}{
		"am Textanfang":       {"--", 2, "—"},
		"nach einem Leerraum": {"a --", 4, "a —"},

		// **Auch hier** — der Strich ist selbst ein Grenzzeichen.
		"zwischen Buchstaben": {"a--b", 3, "a—b"},
		"vor einem Leerraum":  {"a-- b", 3, "a— b"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			text, _ := ersetzer.AnStelle(fall.text, fall.stelle)
			if text != fall.erwartet {
				t.Errorf("%q@%d → %q, erwartet %q",
					fall.text, fall.stelle, text, fall.erwartet)
			}
		})
	}
}

// **Ein Muster aus Buchstaben verhält sich dagegen wie erwartet.**
//
// Der Gegenbeleg zum Kasten darüber: dort greift `whole word` sehr wohl, und
// die Eigenschaft ist keine allgemeine Unwirksamkeit, sondern eine Folge des
// ersten Musterzeichens.
func TestBeiEinemBuchstabenmusterGreiftGanzesWort(t *testing.T) {
	ersetzer := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "AE", Ersatz: "Æ", GanzesWort: true},
	})

	if text, _ := ersetzer.AnStelle("xAEy", 3); text != "xAEy" {
		t.Errorf("%q, erwartet %q — mitten im Wort darf nicht ersetzt werden", text, "xAEy")
	}

	if text, _ := ersetzer.AnStelle("x AE", 4); text != "x Æ" {
		t.Errorf("%q, erwartet %q", text, "x Æ")
	}
}

// **Auszeichnungen werden übersprungen** — `isWordBoundary` überliest `<…>`
// in beide Richtungen (ebenda, Zeilen 1144 und 1154–1162).
//
// Ohne das stünde eine Auszeichnung zwischen Muster und Wortgrenze und machte
// jede `whole word`-Regel im ausgezeichneten Text unwirksam.
func TestAuszeichnungenZaehlenNichtAlsWortzeichen(t *testing.T) {
	ersetzer := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "(C)", Ersatz: "©", GanzesWort: true},
	})

	text, _ := ersetzer.AnStelle("<b>(C)</b>", 6)
	if text != "<b>©</b>" {
		t.Errorf("%q, erwartet %q", text, "<b>©</b>")
	}
}

// **Ohne Regeln bleibt alles, wie es ist.**
func TestOhneRegelnGeschiehtNichts(t *testing.T) {
	text, stelle := service.NeuerAutoersetzer(nil).AnStelle("(C)", 3)
	if text != "(C)" || stelle != 3 {
		t.Errorf("%q@%d, erwartet \"(C)\"@3", text, stelle)
	}
}

// **Eine Stelle außerhalb des Textes wird nicht ausgeführt.**
//
// Sie kommt aus dem Fenster, und was von dort kommt, ist fremde Eingabe (§8).
func TestEineUnmoeglicheStelleAendertNichts(t *testing.T) {
	for _, stelle := range []int{-1, 4} {
		text, danach := service.NeuerAutoersetzer(regeln()).AnStelle("(C)", stelle)
		if text != "(C)" || danach != stelle {
			t.Errorf("Stelle %d: %q@%d", stelle, text, danach)
		}
	}
}

// **Wortgrenzen über mehrbytige Zeichen hinweg.**
//
// Bei `äab` ist das letzte Byte von `ä` kein Buchstabe. Wer Bytes liest, lässt
// ein Whole-Word-`ab` mitten im Wort greifen. Gefunden vom Zweitblick über
// Phase 3, damals an der Maschine, die es nicht mehr gibt — die Fälle sind
// hierher mitgenommen worden, weil sie über die Maschine hinaus gelten.
func TestWortgrenzenUeberMehrbytigeZeichen(t *testing.T) {
	ersetzer := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "ab", Ersatz: "X", GanzesWort: true},
	})

	faelle := []struct{ name, text, erwartet string }{
		{"ein Buchstabe davor", "äab", "äab"},
		{"eine Ziffer davor", "1ab", "1ab"},
		{"ein Satzzeichen davor", "„ab", "„X"},
		{"allein stehend", "ab", "X"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			text, _ := ersetzer.AnStelle(fall.text, len(fall.text))
			if text != fall.erwartet {
				t.Errorf("%q → %q, erwartet %q", fall.text, text, fall.erwartet)
			}
		})
	}
}

// **Kombinierende Zeichen sind keine Wortgrenze.**
//
// `a` gefolgt von U+0301 ist ein `á` aus zwei Runen. Wäre die Marke eine
// Grenze, griffe ein Whole-Word-Muster mitten in einem Buchstaben. Gefunden
// vom Nachlauf des Zweitblicks über Phase 3.
func TestKombinierendeZeichenSindKeineWortgrenze(t *testing.T) {
	ersetzer := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "ab", Ersatz: "X", GanzesWort: true},
	})

	if text, _ := ersetzer.AnStelle("áab", len("áab")); text != "áab" {
		t.Errorf("%q, erwartet unverändert — ein Akzent ist keine Wortgrenze", text)
	}
}

// **Die Regeln kommen aus dem Spielpaket des Sets** —
// der Weg, den es bis zum 2026-08-23 nicht gab.
func TestDieRegelnKommenAusDemSpielpaketDesSets(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 1, Game: "probe", Style: "probe-schlicht"},
		}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	regeln, err := waehler.Autoersetzungen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Autoersetzungen: %v", err)
	}

	if len(regeln) == 0 {
		t.Fatal("die Saat führt auto-replace-Regeln, und keine kam an")
	}

	// **Und sie wirken.** Ein Test, der nur zählt, prüft die Zahl und nicht den
	// Weg (§10).
	if text, _ := service.NeuerAutoersetzer(regeln).AnStelle("->", 2); text != "→" {
		t.Errorf("%q, erwartet %q", text, "→")
	}
}

// **Ein Spiel ohne Paket gibt keine Regeln und keinen Fehler.**
//
// Es fehlt auf einem frischen Rechner, und dann wird eben nichts ersetzt. Wer
// daraus einen Fehler machte, hielte das Tippen an, weil eine Vorlage nicht
// daliegt (§8).
func TestOhneSpielpaketGibtEsKeineRegelnUndKeinenFehler(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 1, Game: "gibtsnicht", Style: "probe-schlicht"},
		}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	regeln, err := waehler.Autoersetzungen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Autoersetzungen: %v", err)
	}

	if len(regeln) != 0 {
		t.Errorf("%d Regeln aus einem Spiel, das es nicht gibt", len(regeln))
	}
}

// **Und ein Set, das es nicht gibt, wird gemeldet.**
func TestOhneSetWirdDerFehlerGemeldet(t *testing.T) {
	waehler := service.NeuerStilwaehler(
		&koepfequelle{}, saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	if _, err := waehler.Autoersetzungen(t.Context(), 99); err == nil {
		t.Error("ein Set, das es nicht gibt, kam durch")
	}
}

// **Ein `<` ohne sein `>` ist keine Auszeichnung** — und ein `>` ohne sein `<`
// auch nicht.
//
// MSEs `isWordBoundary` überliest Auszeichnungen und gibt beim unvollständigen
// Paar auf: nach vorn endet es beim Textende (das **ist** eine Grenze), nach
// hinten gibt es `false` (`src/gui/value/text.cpp`, Zeilen 1144 und 1154–1162,
// Fassung f9d9356 vom 2021-07-06).
//
// **Die beiden Richtungen antworten also verschieden**, und das ist portiert:
// ein `>` ohne sein `<` links vom Muster verhindert die Ersetzung, ein `<`
// ohne sein `>` rechts davon nicht. Die Fälle stehen hier, weil sie in fremden
// Daten vorkommen: ein Regeltext mit einem mathematischen `<` ist keine
// Seltenheit.
func TestUnvollstaendigeAuszeichnungen(t *testing.T) {
	ersetzer := service.NeuerAutoersetzer([]msepaket.AutoErsetzung{
		{Muster: "ab", Ersatz: "X", GanzesWort: true},
	})

	faelle := []struct {
		name     string
		text     string
		stelle   int
		erwartet string
	}{
		// Nach vorn: das offene `<` führt bis zum Textende, und das ist eine
		// Grenze.
		{"ein offenes Kleiner-als dahinter", "ab<c", 2, "X<c"},

		// Nach hinten: das `>` ohne sein `<` bricht die Suche mit „keine
		// Grenze" ab — und damit greift die Regel nicht.
		{"ein einzelnes Größer-als davor", "x>ab", 4, "x>ab"},

		// Zwei `>` hintereinander: dieselbe Antwort, anderer Weg dorthin.
		{"zwei Größer-als davor", "a>>ab", 5, "a>>ab"},

		// Und mit vollständiger Auszeichnung greift sie doch.
		{"eine vollständige Auszeichnung davor", "<b>ab", 5, "<b>X"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			text, _ := ersetzer.AnStelle(fall.text, fall.stelle)

			if text != fall.erwartet {
				t.Errorf("%q@%d → %q, erwartet %q",
					fall.text, fall.stelle, text, fall.erwartet)
			}
		})
	}
}
