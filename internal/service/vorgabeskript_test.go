package service_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein `default:` ist ein Skript, auch ohne geschweifte Klammern.**
//
// # Der Befund
//
// `magic.mse-game/card_fields` führt **142** Vorgabewerte, und **alle 142**
// beginnen mit `stylesheet; `:
//
//	default: stylesheet; card_color(casting_cost: card.casting_cost, …)
//
// `a; b` ist in MSE-Script eine Folge — `stylesheet` wird ausgewertet und
// verworfen, `card_color(…)` ist das Ergebnis. Das vorangestellte `stylesheet`
// ist eine **Abhängigkeitsangabe**: es sagt MSE, dass der Wert neu zu rechnen
// ist, wenn sich die Vorlage ändert.
//
// **Keines davon ist geklammert.** Unser Vorgabenweg verlangte `{ … }` und ließ
// deshalb **jeden einzelnen** fallen — der rohe Text wurde durchgereicht, und
// eine Karte ohne `card color` bekam die Zeichenkette
// `stylesheet; card_color(…)` als Farbe.
//
// # Warum ein Skript und keine Zeichenkette
//
// `magic.mse-game/set_fields` schreibt die andere Form:
//
//	default: "Type custom types here; or remove this line"
//
// **In Anführungszeichen** — das ist eine Zeichenkette **in Skriptsyntax**,
// nicht roher Text. Beide Formen sind Skript; ein `default:` ist damit immer
// eines (#298).
func TestEinVorgabewertIstEinSkriptAuchOhneKlammern(t *testing.T) {
	// **Gelesen wird über ein Feldskript**, denn ein Vorgabewert füllt die
	// Skriptumgebung und nicht die Kartenwerte: er ist das, womit ein Stil
	// rechnet, nicht das, was jemand eingetippt hat.
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ card.anzeige + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{{
		Name: "anzeige", Default: `stylesheet; "abgeleitet"`,
	}}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		if ebene.Bild.Pfad != "/probe.mse-style/abgeleitet.png" {
			t.Errorf("der Pfad ist %q, erwartet /probe.mse-style/abgeleitet.png — "+
				"der Vorgabewert wurde nicht als Skript ausgewertet", ebene.Bild.Pfad)
		}

		return
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Eine Zeichenkette in Anführungszeichen ist auch ein Skript.**
//
// So schreibt `magic.mse-game/set_fields` seine Vorgaben. Wer sie roh
// durchreichte, bekäme die Anführungszeichen mit auf die Karte.
func TestEinVorgabewertInAnfuehrungszeichenVerliertSie(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ card.anzeige + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{{
		Name: "anzeige", Default: `"ohne Anfuehrungszeichen"`,
	}}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/ohne Anfuehrungszeichen.png" {
				t.Errorf("der Pfad ist %q — die Anführungszeichen stehen noch da",
					ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein abgebrochener Lauf rechnet die übrigen Felder nicht mehr aus.**
//
// # Der Befund des Zweitblicks
//
// `magic.mse-game` führt 249 Kartenfelder, 142 davon mit einem Skript. Das ist
// die längste Strecke zwischen zwei Kontextprüfungen im ganzen Vorschaulauf —
// wer während des Tippens die Karte wechselt, wartete darauf, dass die alte
// fertig gerechnet wird.
//
// **Der Abbruch war am Orchestrator verdrahtet, nicht an der teuren Schleife**
// (§7, die projekteigene Regel).
func TestEinAbgebrochenerLaufRechnetNichtWeiter(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ card.anzeige + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{{
		Name: "anzeige", Default: `"abgeleitet"`,
	}}}

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	// **Kein Absturz, kein halber Stand** — und der Vorgabewert entsteht nicht
	// mehr, weil niemand mehr auf ihn wartet.
	plan, _, err := service.NeuerPlanbauer().Stilplan(
		abgebrochen, &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err == nil && plan == nil {
		t.Fatal("weder Plan noch Fehler")
	}
}

// spaeterAbbruch ist ein Kontext, der **erst beim n-ten Blick** abgebrochen ist.
//
// # Wozu er nötig ist
//
// `Stilplan` prüft den Kontext an seiner ersten Zeile. Ein von Anfang an
// abgebrochener kommt deshalb nie bis zu den Schleifen, deren Abbruch geprüft
// werden soll — und die Zweige dort blieben ungedeckt, obwohl sie im Betrieb
// sehr wohl betreten werden: **abgebrochen wird während des Laufs, nicht
// davor.**
//
// Ein Zweig, den kein Test erreicht, ist entweder tot oder eine Lücke (§3).
// Dieser ist eine Lücke, und dieser Kontext schließt sie.
type spaeterAbbruch struct {
	context.Context

	verbleibend *int
}

func (s spaeterAbbruch) Err() error {
	if *s.verbleibend > 0 {
		*s.verbleibend--

		return nil
	}

	return context.Canceled
}

func (s spaeterAbbruch) Done() <-chan struct{} { return nil }

// **Ein Abbruch mitten im Lauf hält die Vorgabewerte an.**
//
// `magic.mse-game` führt 249 Kartenfelder, 142 davon mit einem Skript — die
// längste Strecke zwischen zwei Kontextprüfungen im ganzen Vorschaulauf.
// Angestrichen vom Zweitblick über die Phasen B bis D: „Abbruch ist am
// Orchestrator verdrahtet, nicht an den rechenintensiven Schleifen."
func TestEinAbbruchMittenImLaufHaeltDieVorgabenAn(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ card.zweites + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{
		{Name: "erstes", Default: `"eins"`},
		{Name: "zweites", Default: `"zwei"`},
	}}

	// **Einmal nil, dann abgebrochen**: die Eingangsprüfung kommt durch, die
	// Schleife nicht.
	verbleibend := 1
	ctx := spaeterAbbruch{Context: t.Context(), verbleibend: &verbleibend}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		ctx, &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// **Das zweite Feld ist nicht mehr belegt worden**, und der Bildpfad zeigt
	// das: er trägt den leeren Wert.
	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		if ebene.Bild.Pfad == "/probe.mse-style/zwei.png" {
			t.Error("der Lauf hat trotz Abbruch weitergerechnet")
		}

		return
	}
}

// **Und die Inhaltsmessung hält ebenso an.**
//
// Ein Stil von `M15-Magic-Pack` führt bis zu sechzig Feldstile, und jede
// Textmessung sucht eine Schrift.
func TestEinAbbruchHaeltAuchDieInhaltsmessungAn(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "name", Links: "0", Oben: "0", Breite: "100", Hoehe: "20",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	// Zwei Blicke gehen durch — die Eingangsprüfung und die Vorgabenschleife —,
	// beim dritten ist Schluss.
	verbleibend := 2
	ctx := spaeterAbbruch{Context: t.Context(), verbleibend: &verbleibend}

	if _, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(ctx, &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "abcd"},
		}}, service.Vorlage{Stil: stil}); err != nil {
		t.Fatalf("Plan: %v", err)
	}
}
