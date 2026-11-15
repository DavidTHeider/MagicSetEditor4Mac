package render_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// festeBreite ist eine Schrift, in der jedes Zeichen gleich breit ist.
//
// **Damit der Test den Umbruch prüft und nicht die Schrift.** Mit einer echten
// Schrift stünde in jedem Sollwert eine gemessene Zahl, und wer sie ändert,
// weiß nicht, ob er den Umbruch oder die Schriftfassung verschoben hat.
type festeBreite struct{}

func (festeBreite) Breite(_ rune) float64 { return 1 }
func (festeBreite) Zeilenhoehe() float64  { return 1.2 }

func TestDerUmbruchFaelltAnDenErwartetenStellen(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		breite   float64
		erwartet []string
	}{
		{
			name: "was hineinpasst, bleibt eine Zeile",
			text: "abc def", breite: 100,
			erwartet: []string{"abc def"},
		},
		{
			// Bei Größe 10 und fester Breite 1 ist ein Zeichen 10 Punkte breit.
			// "abc def" misst 70, "abc" misst 30.
			name: "umgebrochen wird am Leerzeichen",
			text: "abc def", breite: 50,
			erwartet: []string{"abc", "def"},
		},
		{
			name: "mehrere Umbrüche",
			text: "aa bb cc dd", breite: 50,
			erwartet: []string{"aa bb", "cc dd"},
		},
		{
			// **Ein zu langes Wort wird zwischen zwei Zeichen gebrochen**, wie
			// bei MSE (`viewer.cpp`, Zeilen 614–633, Fassung f9d9356 vom
			// 2021-07-06). Hier stand bis zum Nachlauf über E/F, es rage
			// hinaus, weil eine Trennung eine Zeile ergäbe, die niemand
			// geschrieben hat — erschlossen statt nachgelesen. Ein Symbolzug
			// wie `{T}` bleibt ganz: er ist ein Zeichen ([render.Setze]).
			name: "ein Wort, das allein nicht passt",
			text: "unteilbarkeit ab", breite: 50,
			erwartet: []string{"untei", "lbark", "eit", "ab"},
		},
		{
			name: "ein Umbruch im Text bleibt einer",
			text: "ab\ncd", breite: 100,
			erwartet: []string{"ab", "cd"},
		},
		{
			// Eine leere Zeile trennt zwei Absätze. Wer sie wegwirft, schiebt
			// den Regeltext an den Namen.
			name: "eine leere Zeile bleibt stehen",
			text: "ab\n\ncd", breite: 100,
			erwartet: []string{"ab", "", "cd"},
		},
		{
			// **Mehrere Leerzeichen bleiben mehrere.**
			//
			// Hier stand bis zum 2026-08-16 „zählen als eines", und das war
			// unsere Annahme, keine Beobachtung: der alte Umbruch benutzte
			// `strings.Fields`, das jede Folge von Leerraum zusammenzieht.
			//
			// MSE2 tut das nicht. Es bricht **an** einem Leerzeichen und
			// versteckt allein das eine, das den Umbruch verursacht hat —
			// `hide_breaker` in `src/render/text/viewer.cpp`, Zeile 589,
			// Fassung f9d9356 vom 2021-07-06. Jedes andere Leerzeichen ist ein
			// Zeichen wie jedes andere und wird gesetzt.
			name: "mehrere Leerzeichen bleiben mehrere",
			text: "ab    cd", breite: 100,
			erwartet: []string{"ab    cd"},
		},
		{
			name: "leerer Text ergibt eine leere Zeile",
			text: "", breite: 100,
			erwartet: []string{""},
		},
	}

	satz := render.Satz{Groesse: 10, Ausrichtung: render.LinksOben}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ergebnis := render.Setze(satz,
				render.Rechteck{Breite: fall.breite, Hoehe: 100}, fall.text, festeBreite{})

			gesetzt := make([]string, 0, len(ergebnis.Zeilen))
			for _, zeile := range ergebnis.Zeilen {
				gesetzt = append(gesetzt, zeile.Text)
			}

			if strings.Join(gesetzt, "|") != strings.Join(fall.erwartet, "|") {
				t.Errorf("%q, erwartet war %q", gesetzt, fall.erwartet)
			}
		})
	}
}

// Die Ausrichtung setzt die Zeile an die richtige Stelle.
func TestDieAusrichtungSetztDieZeile(t *testing.T) {
	// "ab" misst bei Größe 10 zwanzig Punkte, der Rahmen ist 100 breit und
	// 100 hoch, die Zeilenhöhe ist 12.
	faelle := []struct {
		name        string
		ausrichtung render.Ausrichtung
		links, oben float64
	}{
		{"links oben", render.LinksOben, 0, 0},
		{"mitte oben", render.MitteOben, 40, 0},
		{"rechts oben", render.RechtsOben, 80, 0},
		{"links mitte", render.LinksMitte, 0, 44},
		{"mitte mitte", render.MitteMitte, 40, 44},
		{"rechts unten", render.RechtsUnten, 80, 88},
		{"links unten", render.LinksUnten, 0, 88},
		{"mitte unten", render.MitteUnten, 40, 88},
		{"rechts mitte", render.RechtsMitte, 80, 44},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ergebnis := render.Setze(
				render.Satz{Groesse: 10, Ausrichtung: fall.ausrichtung},
				render.Rechteck{Breite: 100, Hoehe: 100}, "ab", festeBreite{})

			zeile := ergebnis.Zeilen[0]
			if zeile.Links != fall.links || zeile.Oben != fall.oben {
				t.Errorf("(%g,%g), erwartet war (%g,%g)",
					zeile.Links, zeile.Oben, fall.links, fall.oben)
			}
		})
	}
}

// Die Laufweite steht **zwischen** den Zeichen, nicht hinter dem letzten.
//
// Sonst säße eine rechtsbündige Zeile um eine Laufweite zu weit links — bei den
// 0,2 Punkten aus `magic-m15` fiele das erst im Vergleich zweier Karten auf.
func TestDieLaufweiteStehtZwischenDenZeichen(t *testing.T) {
	satz := render.Satz{Groesse: 10, Laufweite: 2, Ausrichtung: render.LinksOben}

	ergebnis := render.Setze(satz, render.Rechteck{Breite: 1000, Hoehe: 100}, "abc", festeBreite{})

	// Drei Zeichen à 10 Punkte, dazu **zwei** Zwischenräume à 2.
	if breite := ergebnis.Zeilen[0].Breite; breite != 34 {
		t.Errorf("Breite %g, erwartet waren 34", breite)
	}

	// Und ein einzelnes Zeichen hat gar keine Laufweite.
	einzeln := render.Setze(satz, render.Rechteck{Breite: 1000, Hoehe: 100}, "a", festeBreite{})
	if breite := einzeln.Zeilen[0].Breite; breite != 10 {
		t.Errorf("Breite %g, erwartet waren 10", breite)
	}
}

// Der Zeilenabstand schlägt die Zeilenhöhe der Schrift, wenn er gesetzt ist.
func TestDerZeilenabstandSchlaegtDieSchrift(t *testing.T) {
	rahmen := render.Rechteck{Breite: 25, Hoehe: 100}

	ohne := render.Setze(render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		rahmen, "ab cd", festeBreite{})
	if ohne.Zeilen[1].Oben != 12 {
		t.Errorf("ohne Angabe %g, erwartet waren 12 (Größe 10 × Zeilenhöhe 1,2)", ohne.Zeilen[1].Oben)
	}

	mit := render.Setze(
		render.Satz{Groesse: 10, Zeilenabstand: 2, Ausrichtung: render.LinksOben},
		rahmen, "ab cd", festeBreite{})
	if mit.Zeilen[1].Oben != 20 {
		t.Errorf("mit Angabe %g, erwartet waren 20", mit.Zeilen[1].Oben)
	}
}

// Mehr Text als Platz ergibt eine Höhe über dem Rahmen — und keine Kürzung.
//
// Der Aufrufer entscheidet, ob er kürzt oder verkleinert; hier wird gemessen,
// nicht entschieden.
func TestMehrTextAlsPlatzWirdGemeldetStattGekuerzt(t *testing.T) {
	ergebnis := render.Setze(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 25, Hoehe: 20}, "ab cd ef gh", festeBreite{})

	if len(ergebnis.Zeilen) != 4 {
		t.Errorf("%d Zeilen, erwartet waren 4", len(ergebnis.Zeilen))
	}

	if ergebnis.Hoehe <= 20 {
		t.Errorf("Höhe %g — sie müsste über der Rahmenhöhe von 20 liegen", ergebnis.Hoehe)
	}
}

// schriften ist eine Schriftquelle aus dem Arbeitsspeicher.
//
// **Sie unterscheidet den Schnitt so, wie `internal/schriften` es tut**: fett
// heißt `<Name> Bold`, und findet sich der nicht, gilt der aufrechte
// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
// Eine Vorrichtung, die das Gewicht wegwürfe, wäre großzügiger als der Betrieb
// (§3).
type schriften map[string]render.Schrift

func (s schriften) Suche(name string, fett bool) (render.Schrift, error) {
	if fett {
		if schrift, da := s[name+" Bold"]; da {
			return schrift, nil
		}
	}

	schrift, da := s[name]
	if !da {
		return nil, errors.New("nicht im System")
	}

	return schrift, nil
}

// Eine fehlende Schrift wird **benannt**, nicht stillschweigend ersetzt.
//
// **Eine Karte mit falscher Schrift sieht richtig aus und ist es nicht.** Wer
// sie mit Helvetica statt MPlantin gedruckt bekommt, merkt es an dem Tag, an
// dem die Karten aus der Druckerei kommen.
func TestEineFehlendeSchriftWirdBenannt(t *testing.T) {
	faelle := []struct {
		name    string
		quelle  render.Schriftquelle
		schrift string
		nennt   string
	}{
		{
			name:    "eine Schrift, die es im System nicht gibt",
			quelle:  schriften{"Beleren": festeBreite{}},
			schrift: "MPlantin",
			nennt:   "MPlantin",
		},
		{
			name:    "gar keine Schriftsuche eingerichtet",
			quelle:  nil,
			schrift: "MPlantin",
			nennt:   "keine Schriftsuche",
		},
		{
			name:    "eine Textebene ohne Schriftnamen",
			quelle:  schriften{},
			schrift: "  ",
			nennt:   "ohne Schriftnamen",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
				{
					Art: render.ArtText, Rahmen: render.Rechteck{Breite: 10, Hoehe: 10},
					Abschnitte: render.Aufrecht("Feuerball"), Satz: render.Satz{Schrift: fall.schrift, Groesse: 5},
					Deckung: 1,
				},
			}}

			rasterer := render.NeuerRasterer(bilder{})
			if fall.quelle != nil {
				rasterer = rasterer.MitSchriften(fall.quelle)
			}

			_, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
			}

			if !strings.Contains(befunde[0].Meldung, fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", befunde[0].Meldung, fall.nennt)
			}
		})
	}
}

// Eine vorhandene Schrift wird gesetzt — und dass nichts gezeichnet wurde,
// steht als Befund da.
//
// **Der erste Anlauf ließ eine Textebene als erfolgreich gelten.** Die Karte
// blieb an der Stelle leer, ohne dass jemand es erfuhr — ein Mangel, der wie
// ein Erfolg aussieht. Angestrichen vom Zweitblick über Phase 4.
func TestEineSchriftOhneGlyphenWirdGemeldet(t *testing.T) {
	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{
		{
			Art: render.ArtText, Rahmen: render.Rechteck{Breite: 40, Hoehe: 20},
			Abschnitte: render.Aufrecht("Feuerball wirken"), Satz: render.Satz{Schrift: "MPlantin", Groesse: 5},
			Deckung: 1,
		},
	}}

	// `festeBreite` misst und zeichnet nicht — genau die Schrift, mit der der
	// Umbruch geprüft ist.
	_, befunde, err := render.NeuerRasterer(bilder{}).
		MitSchriften(schriften{"MPlantin": festeBreite{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %v", len(befunde), befunde)
	}

	// **Gesetzt wurde trotzdem** — die Meldung sagt beides.
	if !strings.Contains(befunde[0].Meldung, "gesetzt") ||
		!strings.Contains(befunde[0].Meldung, "keine Glyphen") {
		t.Errorf("%q", befunde[0].Meldung)
	}
}

// **Ein Rahmen der Größe null ist eine Ansage, ein negativer ein Fehler.**
//
// `magic-m15-saga.mse-style` schreibt am Textfeld
// `width: {if is_saga() then 0 else 153}` — die Vorlage **verbirgt** das Feld
// so, und MSE zeichnet in einen Kasten ohne Fläche nichts und meldet nichts.
// Ein Befund darüber stünde auf jeder Saga-Karte in derselben Liste wie die
// echten (#270).
//
// **Eine negative Größe bleibt ein Befund.** Sie kann keine Vorlage meinen; sie
// entsteht, wenn `right` links von `left` liegt — und dann ist die Rechnung
// falsch, nicht die Absicht.
func TestEinTextrahmenOhneBreiteIstEinBefund(t *testing.T) {
	for _, rahmen := range []render.Rechteck{
		{Breite: -1, Hoehe: 10},
		{Breite: 10, Hoehe: -1},
	} {
		plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
			{
				Art: render.ArtText, Rahmen: rahmen, Abschnitte: render.Aufrecht("a"),
				Satz: render.Satz{Schrift: "MPlantin", Groesse: 5}, Deckung: 1,
			},
		}}

		_, befunde, err := render.NeuerRasterer(bilder{}).
			MitSchriften(schriften{"MPlantin": festeBreite{}}).
			Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("abgewiesen: %v", err)
		}

		if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Textrahmen") {
			t.Errorf("Rahmen %s: Befunde %v", rahmen, befunde)
		}
	}
}

// **Ein Rahmen der Größe null zeichnet nichts — schweigend.**
func TestEinTextrahmenDerGroesseNullSchweigt(t *testing.T) {
	for _, rahmen := range []render.Rechteck{
		{Breite: 0, Hoehe: 10},
		{Breite: 10, Hoehe: 0},
	} {
		plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
			{
				Art: render.ArtText, Rahmen: rahmen, Abschnitte: render.Aufrecht("a"),
				Satz: render.Satz{Schrift: "MPlantin", Groesse: 5}, Deckung: 1,
			},
		}}

		bild, befunde, err := render.NeuerRasterer(bilder{}).
			MitSchriften(schriften{"MPlantin": festeBreite{}}).
			Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("abgewiesen: %v", err)
		}

		if len(befunde) != 0 {
			t.Errorf("Rahmen %s: Befunde %v", rahmen, befunde)
		}

		if bild.NRGBAAt(5, 5).A != 0 {
			t.Errorf("Rahmen %s: es wurde gezeichnet", rahmen)
		}
	}
}
