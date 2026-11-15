package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Suchen und Ersetzen über einen Feldwert** — Werkstück 2 von Phase 3.7
// (#1168).

// frage baut eine Ersetzungsfrage mit Vorgaben.
func frage(text, ersatz string, abweichung ...func(*domain.Ersetzungsfrage)) domain.Ersetzungsfrage {
	gestellt := domain.Ersetzungsfrage{Text: text, Ersatz: ersatz}
	for _, ändere := range abweichung {
		ändere(&gestellt)
	}

	return gestellt
}

// **Gesucht wird im Text, nicht in den Auszeichnungen.**
//
//	String v = untag(value().value());
//
// `src/gui/value/text.cpp`, Zeile 1229, Fassung f9d9356 vom 2021-07-06. Ein
// `<i>` mitten im Wort darf einen Treffer nicht verhindern — und ein Muster,
// das auf `i` passt, darf die Auszeichnung nicht treffen.
func TestErsetzungSuchtImTextUndNichtInDenAuszeichnungen(t *testing.T) {
	t.Parallel()

	fälle := []struct {
		name, roh, was, ersatz, erwartet string
		treffer                          int
	}{
		{
			name: "schlichter Text",
			roh:  "Fliegend", was: "Flieg", ersatz: "Renn",
			erwartet: "Rennend", treffer: 1,
		},
		{
			name: "eine Auszeichnung außen herum bleibt stehen",
			roh:  "<i>Fliegend</i>", was: "Flieg", ersatz: "Renn",
			erwartet: "<i>Rennend</i>", treffer: 1,
		},
		{
			name: "der Name der Auszeichnung ist kein Text",
			roh:  "<i>ab</i>", was: "i", ersatz: "X",
			erwartet: "<i>ab</i>", treffer: 0,
		},
		{
			name: "zwei Treffer, beide ersetzt",
			roh:  "ab ab", was: "ab", ersatz: "cd",
			erwartet: "cd cd", treffer: 2,
		},
		{
			name: "ein geschütztes Kleinerzeichen ist ein Zeichen",
			roh:  "a\x01b", was: "a<b", ersatz: "X",
			erwartet: "X", treffer: 1,
		},
		{
			name: "der Ersatz wird geschützt",
			roh:  "ab", was: "ab", ersatz: "<c",
			erwartet: "\x01c", treffer: 1,
		},
		{
			// **Ein Zeichen aus mehreren Bytes verschiebt die Zuordnung.**
			//
			// `ß` steht in UTF-8 auf zwei Bytes. Die Stellenliste wird über
			// **Bytes** durchlaufen; eine Zuordnung, die je **Zeichen** einen
			// Eintrag führte, liefe dahinter aus dem Takt — und beim letzten
			// Treffer über das Ende hinaus.
			name: "ein Zeichen aus zwei Bytes",
			roh:  "<i>Große Kreatur</i>", was: "Kreatur", ersatz: "Artefakt",
			erwartet: "<i>Große Artefakt</i>", treffer: 1,
		},
		{
			name: "und derselbe Treffer ganz am Ende",
			roh:  "Größe", was: "e", ersatz: "en",
			erwartet: "Größen", treffer: 1,
		},
		{
			name: "nichts gefunden lässt den Wert unberührt",
			roh:  "<i>Fliegend</i>", was: "Schwimm", ersatz: "Renn",
			erwartet: "<i>Fliegend</i>", treffer: 0,
		},
	}

	for _, fall := range fälle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gestellt := frage(fall.was, fall.ersatz)

			if zahl := len(gestellt.Stellen(fall.roh)); zahl != fall.treffer {
				t.Errorf("%d Stellen, erwartet %d", zahl, fall.treffer)
			}

			if neu := gestellt.Ersetze(fall.roh); neu != fall.erwartet {
				t.Errorf("Ersetze(%q) = %q, erwartet %q", fall.roh, neu, fall.erwartet)
			}
		})
	}
}

// **Groß und klein ist erst gleich, wenn niemand es verlangt.**
//
//	if (!find.caseSensitive()) v.LowerCase();
//	… is_substr(s, pos, find.findString().Lower())
//
// dieselbe Datei, Zeilen 1230 und 1207.
func TestErsetzungAchtetAufGrossUndKleinNurAufVerlangen(t *testing.T) {
	t.Parallel()

	beachtend := func(f *domain.Ersetzungsfrage) { f.GrossKleinBeachten = true }

	if neu := frage("flieg", "Renn").Ersetze("Fliegend"); neu != "Rennend" {
		t.Errorf("ohne Beachtung: %q", neu)
	}

	if neu := frage("flieg", "Renn", beachtend).Ersetze("Fliegend"); neu != "Fliegend" {
		t.Errorf("mit Beachtung: %q", neu)
	}

	if neu := frage("Flieg", "Renn", beachtend).Ersetze("Fliegend"); neu != "Rennend" {
		t.Errorf("mit Beachtung, passend: %q", neu)
	}
}

// **Ein ganzes Wort steht zwischen Rand, Leerzeichen und Satzzeichen.**
//
//	bool is_word_end(const String& s, size_t pos) {
//	  if (pos == 0 || pos >= s.size()) return true;
//	  Char c = s.GetChar(pos);
//	  return isSpace(c) || isPunct(c);
//	}
//
// dieselbe Datei, Zeilen 1192–1197. **Der Rand zählt als Wortgrenze** — und
// `pos - 1` bei `pos == 0` läuft dort über, wird zu einer riesigen Zahl und
// fällt damit in denselben Zweig.
func TestErsetzungKannAufGanzeWoerterBeschraenken(t *testing.T) {
	t.Parallel()

	ganz := func(f *domain.Ersetzungsfrage) { f.GanzesWort = true }

	fälle := []struct {
		name, roh, erwartet string
	}{
		{"allein", "ab", "X"},
		{"am Anfang", "ab cd", "X cd"},
		{"am Ende", "cd ab", "cd X"},
		{"zwischen Leerzeichen", "cd ab ef", "cd X ef"},
		{"vor einem Satzzeichen", "ab.", "X."},
		{"nach einem Satzzeichen", ".ab", ".X"},
		{"mitten in einem Wort nicht", "abc", "abc"},
		{"am Wortanfang nicht", "abc cd", "abc cd"},

		// **Die berichtigte Grenze** (§15, Regel 7): MSE2 gäbe hier „xX".
		// Sein `is_word_end(s, pos - 1)` fragt bei einem Fund an Stelle 1
		// nach Stelle 0 — und der Zweig `pos == 0` gibt `true` zurück, ohne
		// das Zeichen anzusehen. Ein Fund, den ein Mensch überspringt, wiegt
		// anders als einer, der dreihundertmal geschrieben wird.
		{"am Wortende nicht", "xab", "xab"},

		// **Und ein Umlaut ist ein Buchstabe und kein Wortende.** In UTF-8
		// steht er auf zwei Bytes; wer davon das erste liest, bekommt ein
		// Zeichen, das es nicht gibt — und `unicode.IsSymbol` sagt dazu ja.
		{"nach einem Umlaut nicht", "äab", "äab"},
		{"vor einem Umlaut nicht", "abä", "abä"},
	}

	for _, fall := range fälle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if neu := frage("ab", "X", ganz).Ersetze(fall.roh); neu != fall.erwartet {
				t.Errorf("Ersetze(%q) = %q, erwartet %q", fall.roh, neu, fall.erwartet)
			}
		})
	}
}

// **Eine Frage ohne Suchtext ist keine.**
//
// **Ein leerer Ersatz dagegen ist einer**: „das Wort überall streichen" ist
// eine gewöhnliche Absicht, und wer sie verbietet, zwingt zu einem Umweg über
// die Massenänderung.
func TestErsetzungsfrageBrauchtEinenSuchtext(t *testing.T) {
	t.Parallel()

	if befunde := frage("", "X").Validate(); len(befunde) != 1 ||
		befunde[0].Field != "text" {
		t.Errorf("Befunde: %+v", befunde)
	}

	if befunde := frage("   ", "X").Validate(); len(befunde) != 1 {
		t.Errorf("nur Leerraum: %+v", befunde)
	}

	if befunde := frage("ab", "").Validate(); len(befunde) != 0 {
		t.Errorf("leerer Ersatz: %+v", befunde)
	}
}

// **Ein Suchtext, der sich selbst enthält, läuft nicht im Kreis.**
//
// Wer `a` durch `aa` ersetzt, bekommt `aa` und nicht endlos viele. Die Stellen
// stehen im **Ausgangstext** fest, bevor eine davon geschrieben wird — bei
// MSE2 steht daneben ein `// TODO: string might have changed when doing
// replace all` (Zeile 1223), und genau das ist hier vermieden.
func TestErsetzungLaeuftNichtImKreis(t *testing.T) {
	t.Parallel()

	if neu := frage("a", "aa").Ersetze("a"); neu != "aa" {
		t.Errorf("%q", neu)
	}

	if neu := frage("ab", "abab").Ersetze("abab"); neu != "abababab" {
		t.Errorf("%q", neu)
	}
}

// **Überlappende Treffer werden einmal gezählt.**
//
// `aa` in `aaa` steht an Stelle 0 und an Stelle 1; MSE geht nach einem Treffer
// weiter, ohne ihn zu überspringen — für eine Ersetzung über ein ganzes Set
// wäre das eine Ersetzung in einer schon ersetzten. Gezählt wird deshalb ohne
// Überlappung, wie es jeder Editor tut.
func TestUeberlappendeTrefferWerdenEinmalGezaehlt(t *testing.T) {
	t.Parallel()

	gestellt := frage("aa", "b")

	if zahl := len(gestellt.Stellen("aaa")); zahl != 1 {
		t.Errorf("%d Stellen in »aaa«, erwartet 1", zahl)
	}

	if neu := gestellt.Ersetze("aaaa"); neu != "bb" {
		t.Errorf("%q", neu)
	}
}

// **Eine Stelle nennt Anfang und Ende im rohen Wert.**
//
// Damit die Vorschau sie zeigen kann, ohne noch einmal zu suchen.
func TestEineStelleNenntIhrenPlatzImRohenWert(t *testing.T) {
	t.Parallel()

	stellen := frage("Flieg", "X").Stellen("<i>Fliegend</i>")
	if len(stellen) != 1 {
		t.Fatalf("%d Stellen", len(stellen))
	}

	if stellen[0].Anfang != 3 || stellen[0].Ende != 8 {
		t.Errorf("Stelle %+v, erwartet 3..8", stellen[0])
	}
}

// **Eine leere Frage findet nichts, statt überall zu treffen.**
//
// Ohne diese Wache stünde eine leere Suchzeichenkette an **jeder** Stelle —
// und `Ersetze` schöbe den Ersatz zwischen alle Zeichen.
func TestEineLeereFrageFindetNichts(t *testing.T) {
	t.Parallel()

	for _, was := range []string{"", "   "} {
		if stellen := frage(was, "X").Stellen("abc"); stellen != nil {
			t.Errorf("Stellen(%q) = %+v", was, stellen)
		}
	}
}

// **Und ein Befund ist ein Validierungsbefund.**
func TestEinErsetzungsbefundIstEineValidierung(t *testing.T) {
	t.Parallel()

	befunde := frage("", "").Validate()
	if len(befunde) == 0 || !errors.Is(befunde[0], domain.ErrValidation) {
		t.Errorf("Befunde: %+v", befunde)
	}
}

// **Ein Wert ohne seine Auszeichnungen — für eine Anzeige.**
//
// **Nicht zum Schreiben**: was hier herauskommt, ginge als Feldwert zurück und
// nähme jeder Karte ihre Kursivstellung.
func TestOhneAuszeichnungGibtDenReinenText(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		"<i>Fliegend</i>":          "Fliegend",
		"a\x01b":                   "a<b",
		"kein <sym>T</sym> Symbol": "kein T Symbol",
		"":                         "",
		"<unvollendet":             "",
	} {
		if text := domain.OhneAuszeichnung(roh); text != erwartet {
			t.Errorf("OhneAuszeichnung(%q) = %q, erwartet %q", roh, text, erwartet)
		}
	}
}
