package domain_test

// **Der Geldvertrag** — LOC-03.AK-02 und AK-03
// (#1689).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Stellen kommen aus der Währung und nicht aus einer Annahme.**
func TestDieStellenKommenAusDerWaehrung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		stellen int
		bekannt bool
	}{
		"EUR": {2, true},
		"USD": {2, true},
		"JPY": {0, true},
		"KRW": {0, true},
		"BHD": {3, true},
		"TND": {3, true},

		// **Kleinschreibung zählt auch** — ein CSV trägt, was jemand tippt.
		"eur": {2, true},

		// **Was keine drei Buchstaben hat, ist keine Währung** — und dann
		// wird zwei **geraten** und gesagt, dass geraten wurde.
		"":     {2, false},
		"Euro": {2, false},
		"EURO": {2, false},
		"€":    {2, false},
		"XYZ":  {2, true},

		// **Drei Zeichen, aber keine Buchstaben** — »1€2« ist keine
		// Währung, und `len` allein sähe drei Bytes (gefunden beim ersten
		// roten Lauf).
		"1-2": {2, false},

		// **Mit Leerzeichen drumherum** — ein CSV trägt, was jemand tippt.
		// **Der Rand steht im Fall und nicht im Schlüssel**: ein Name aus
		// Leerzeichen wäre in der Ausgabe nicht zu erkennen, und gocritic
		// strich ihn ohnehin an.
		"mit Rand": {2, true},
	}

	for code, erwartet := range faelle {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			// **Der Rand gehört zum Fall und nicht zum Namen** — ein
			// Unterlauf, der » EUR « hieße, wäre in der Ausgabe nicht zu
			// erkennen (gocritic strich den Schlüssel ohnehin an).
			gefragt := code
			if code == "mit Rand" {
				gefragt = "  EUR "
			}

			stellen, bekannt := domain.Waehrungsstellen(gefragt)

			if stellen != erwartet.stellen || bekannt != erwartet.bekannt {
				t.Errorf("%q gab %d/%v, erwartet %d/%v",
					gefragt, stellen, bekannt, erwartet.stellen, erwartet.bekannt)
			}
		})
	}
}

// **Was hineingeht, kommt heraus** — für jede der vier Währungen und in
// beiden Richtungen (AK-01: Roundtrip).
func TestDerGeldrundlaufHaeltInBeidenRichtungen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		text     string
		waehrung string
		betrag   int64
	}{
		{"19.99", "EUR", 1999},
		{"0.01", "EUR", 1},
		{"19.99", "USD", 1999},

		// **Yen hat keine Nachkommastellen** — 150 Yen sind 150 und nicht
		// 15.000.
		{"150", "JPY", 150},

		// **Dinar hat drei** — ein Fils ist ausdrückbar.
		{"1.234", "BHD", 1234},
		{"0.001", "BHD", 1},

		// **Ein Minus geht** — eine Gutschrift ist ein Betrag.
		{"-5.00", "EUR", -500},
	}

	for _, fall := range faelle {
		t.Run(fall.text+" "+fall.waehrung, func(t *testing.T) {
			t.Parallel()

			betrag, fehler := domain.GeldAus(fall.text, fall.waehrung)
			if fehler != nil {
				t.Fatalf("lesen: %v", fehler)
			}

			if betrag != fall.betrag {
				t.Errorf("gelesen wurde %d, erwartet %d", betrag, fall.betrag)
			}

			if zurueck := domain.GeldText(betrag, fall.waehrung); zurueck != fall.text {
				t.Errorf("zurück kam %q statt %q", zurueck, fall.text)
			}
		})
	}
}

// **Weniger Stellen als die Währung hat, werden aufgefüllt** — und nicht
// gerundet: »1.2« in Euro sind 120 Cent.
func TestWenigerStellenWerdenAufgefuellt(t *testing.T) {
	t.Parallel()

	betrag, fehler := domain.GeldAus("1.2", "EUR")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if betrag != 120 {
		t.Errorf("gelesen wurde %d statt 120", betrag)
	}
}

// **Überzählige Nullen gehen verlustfrei auf** — angestrichen vom Zweitblick
// über Phase 6.3 (M12): 0.5.x schrieb 150 ¥ in der CSV als »150.00«, und
// beim Einlesen fiel jeder Yen-Preis weg, weil der Yen keine Nachkommastelle
// kennt. Eine Null hinter der letzten Stelle ändert nichts am Betrag; eine
// andere Ziffer schon, und die bleibt abgewiesen.
func TestUeberzaehligeNullenGehenAuf(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		text     string
		waehrung string
		betrag   int64
	}{
		"Yen aus 0.5.x":    {"150.00", "JPY", 150},
		"Euro mit dreien":  {"19.990", "EUR", 1999},
		"Dinar mit vieren": {"1.2340", "BHD", 1234},
		"negativ":          {"-150.0", "JPY", -150},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			betrag, fehler := domain.GeldAus(fall.text, fall.waehrung)
			if fehler != nil || betrag != fall.betrag {
				t.Errorf("gelesen %d (%v) statt %d", betrag, fehler, fall.betrag)
			}
		})
	}
}

// **Was nicht kanonisch ist, wird erklärt** — und nicht still zu null
// gemacht (AK-03).
//
// **Das ist der Befund, für den es diesen Vertrag gibt**: vorher gab
// `strconv.ParseFloat` für »1,23« einen Fehler, und der Aufrufer machte
// daraus eine Null. Wer eine deutsch formatierte Liste einlas, verlor jeden
// Preis, ohne ein Wort zu lesen (gemessen am 2026-09-22).
func TestEinNichtKanonischerBetragWirdErklaert(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		text     string
		waehrung string
		teil     string
	}{
		"Komma statt Punkt":     {"1,23", "EUR", "Komma"},
		"Währungszeichen":       {"€1.23", "EUR", "Währungszeichen"},
		"Tausendertrennzeichen": {"1 234.56", "EUR", "kein Betrag"},
		"Buchstaben":            {"teuer", "EUR", "kein Betrag"},
		"nur ein Punkt":         {".", "EUR", "kein Betrag"},
		"Punkt am Ende":         {"12.", "EUR", "Punkt"},

		// **Zu viele Stellen für die Währung** — beim Yen ist schon eine zu
		// viel, und beim Euro die dritte.
		"drei Stellen im Euro": {"1.234", "EUR", "Nachkommastellen"},
		"eine Stelle im Yen":   {"150.5", "JPY", "Nachkommastellen"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			betrag, fehler := domain.GeldAus(fall.text, fall.waehrung)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}

			if betrag != 0 {
				t.Errorf("neben dem Fehler kam %d", betrag)
			}

			if !strings.Contains(fehler.Error(), fall.teil) {
				t.Errorf("die Meldung %q nennt %q nicht", fehler, fall.teil)
			}

			// **Und der Text steht im Satz** — sonst suchte jemand in
			// dreihundert Zeilen nach der einen (§8).
			if !strings.Contains(fehler.Error(), fall.text) {
				t.Errorf("die Meldung nennt den Wert nicht: %v", fehler)
			}
		})
	}
}

// **Kein Preis ist eine Null und kein Fehler** — die häufigste Zeile einer
// Sammlung hat keinen.
func TestEinLeererBetragIstEineNull(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"", "   "} {
		betrag, fehler := domain.GeldAus(text, "EUR")

		if fehler != nil || betrag != 0 {
			t.Errorf("%q gab %d (%v)", text, betrag, fehler)
		}
	}
}

// **Eine Zahl, die nicht in eine ganze passt, wird erklärt.**
func TestEinZuGrosserBetragWirdErklaert(t *testing.T) {
	t.Parallel()

	_, fehler := domain.GeldAus("99999999999999999999.99", "EUR")

	if !errors.Is(fehler, domain.ErrValidation) ||
		!strings.Contains(fehler.Error(), "zu groß") {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und ohne Währung gilt die Vorgabe — samt einem Satz, der sie nennt.**
func TestOhneWaehrungGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	betrag, fehler := domain.GeldAus("1.23", "")
	if fehler != nil || betrag != 123 {
		t.Fatalf("es kam %d (%v)", betrag, fehler)
	}

	_, fehler = domain.GeldAus("1.234", "")
	if fehler == nil ||
		!strings.Contains(fehler.Error(), "ohne Angabe") {
		t.Errorf("die Meldung lautet %v", fehler)
	}
}

// **Der Text schreibt so viele Stellen, wie die Währung hat.**
func TestDerGeldtextSchreibtDieStellenDerWaehrung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		betrag   int64
		waehrung string
		text     string
	}{
		"Euro":            {1999, "EUR", "19.99"},
		"Euro unter eins": {5, "EUR", "0.05"},
		"Yen":             {150, "JPY", "150"},
		"Dinar":           {1234, "BHD", "1.234"},
		"Dinar klein":     {1, "BHD", "0.001"},
		"negativ":         {-500, "EUR", "-5.00"},
		"null":            {0, "EUR", "0.00"},
		"null im Yen":     {0, "JPY", "0"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if text := domain.GeldText(
				fall.betrag, fall.waehrung); text != fall.text {
				t.Errorf("geschrieben wurde %q statt %q", text, fall.text)
			}
		})
	}
}
