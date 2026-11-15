package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Wahl ersetzt den Inhalt und lässt die Marken stehen** —
// #1074.
//
// Das ist die Zusage, an der alles hängt: ein Wert ohne seine
// `<word-list-…>`-Marken hätte beim nächsten Öffnen keine Auswahl mehr.
func TestEineWahlLaesstDieMarkenStehen(t *testing.T) {
	planer := service.NeuerPlanbauer()

	vorlage := vorlageMitListen(
		msepaket.Wortliste{Name: "type-en", Woerter: typenliste()},
		msepaket.Wortliste{Name: "class-en", Woerter: []msepaket.Wortlisteneintrag{
			{Name: "Bear"}, {Name: "Elf"},
		}},
	)

	roh := "<word-list-type-en>Legendary Creature</word-list-type-en>" +
		"<sep-soft> — </sep-soft><word-list-class-en>Bear</word-list-class-en>"

	faelle := []struct {
		name     string
		bereich  int
		weg      []int
		erwartet string
	}{
		{
			name: "ein anderer Typ", bereich: 0, weg: []int{3},
			erwartet: "<word-list-type-en>Legendary Instant</word-list-type-en>" +
				"<sep-soft> — </sep-soft><word-list-class-en>Bear</word-list-class-en>",
		},
		{
			name: "ein Präfix weg", bereich: 0, weg: []int{1},
			erwartet: "<word-list-type-en>Creature</word-list-type-en>" +
				"<sep-soft> — </sep-soft><word-list-class-en>Bear</word-list-class-en>",
		},
		{
			name: "der zweite Bereich", bereich: 1, weg: []int{1},
			erwartet: "<word-list-type-en>Legendary Creature</word-list-type-en>" +
				"<sep-soft> — </sep-soft><word-list-class-en>Elf</word-list-class-en>",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			neu, befunde, err := planer.Wortwahl(context.Background(),
				karteMitTyp(roh), "type", vorlage, fall.bereich, fall.weg)
			if err != nil {
				t.Fatalf("Wortwahl: %v", err)
			}

			if len(befunde) != 0 {
				t.Errorf("Befunde: %+v", befunde)
			}

			if neu != fall.erwartet {
				t.Errorf("daraus wird\n  %q\nerwartet\n  %q", neu, fall.erwartet)
			}
		})
	}
}

// **Ein Bereich oder ein Weg, den es nicht gibt, ändert nichts.**
func TestEineWahlInsLeereWirdAbgewiesen(t *testing.T) {
	planer := service.NeuerPlanbauer()
	vorlage := vorlageMitListen(msepaket.Wortliste{Name: "type-en", Woerter: typenliste()})
	karte := karteMitTyp("<word-list-type-en>Creature</word-list-type-en>")

	faelle := []struct {
		name    string
		bereich int
		weg     []int
	}{
		{name: "der Bereich ist zu groß", bereich: 1, weg: []int{0}},
		{name: "der Bereich ist negativ", bereich: -1, weg: []int{0}},
		{name: "der Weg ist leer", bereich: 0, weg: nil},
		{name: "der Weg zeigt vorbei", bereich: 0, weg: []int{99}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			neu, _, err := planer.Wortwahl(context.Background(),
				karte, "type", vorlage, fall.bereich, fall.weg)

			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("der Fehler ist %v", err)
			}

			if neu != "" {
				t.Errorf("zurück kam %q", neu)
			}
		})
	}
}

// **Der Bereich zählt die zugeordneten Stellen, nicht alle.**
//
// Eine Marke ohne Liste fällt aus [service.Planbauer.Wortlistenfeld] heraus.
// Wer beim Schreiben alle Stellen zählte, träfe die falsche — ein Klick auf
// den einen sichtbaren Bereich änderte den unsichtbaren daneben.
func TestDerBereichZaehltNurDieZugeordneten(t *testing.T) {
	planer := service.NeuerPlanbauer()

	roh := "<word-list-gibtsnicht>X</word-list-gibtsnicht>" +
		"<word-list-type-en>Creature</word-list-type-en>"

	neu, befunde, err := planer.Wortwahl(context.Background(), karteMitTyp(roh), "type",
		vorlageMitListen(msepaket.Wortliste{Name: "type-en", Woerter: typenliste()}),
		0, []int{3})
	if err != nil {
		t.Fatalf("Wortwahl: %v", err)
	}

	if len(befunde) != 1 {
		t.Errorf("Befunde: %+v", befunde)
	}

	erwartet := "<word-list-gibtsnicht>X</word-list-gibtsnicht>" +
		"<word-list-type-en>Instant</word-list-type-en>"
	if neu != erwartet {
		t.Errorf("daraus wird\n  %q\nerwartet\n  %q", neu, erwartet)
	}
}

// **Ein Kleinerzeichen im gewählten Wort wird geschützt.**
//
//	tve.replaceSelection(escape(new_value), …)
//
// `src/gui/value/text.cpp`, Zeile 300, Fassung f9d9356 vom 2021-07-06. Ein
// Skript kann alles zurückgeben; ohne den Schutz begänne beim nächsten Lesen
// eine Marke, und der Wert bräche ab.
func TestEinKleinerzeichenWirdGeschuetzt(t *testing.T) {
	planer := service.NeuerPlanbauer()

	vorlage := vorlageMitListen(msepaket.Wortliste{
		Name:    "land",
		Woerter: []msepaket.Wortlisteneintrag{{Skript: `"a<b"`}},
	})

	neu, _, err := planer.Wortwahl(context.Background(),
		karteMitTyp("<word-list-land></word-list-land>"), "type", vorlage, 0, []int{0})
	if err != nil {
		t.Fatalf("Wortwahl: %v", err)
	}

	erwartet := "<word-list-land>a\u0001b</word-list-land>"
	if neu != erwartet {
		t.Errorf("daraus wird %q, erwartet %q", neu, erwartet)
	}
}

// **Ein abgebrochener Kontext hält beides an.**
//
// Eine Wortliste kann Skripte auswerten, und die laufen über das ganze Set —
// `all_sub_types` geht bei `Multicolor.mse-set` über 1.400 Karten. Wer den
// Bereich wechselt, während das läuft, bekommt sonst ein Menü zur Karte von
// vorhin (§7).
func TestEinAbgebrochenerKontextHaeltAn(t *testing.T) {
	planer := service.NeuerPlanbauer()
	vorlage := vorlageMitListen(msepaket.Wortliste{Name: "type-en", Woerter: typenliste()})
	karte := karteMitTyp("<word-list-type-en>Creature</word-list-type-en>")

	ctx, abbrechen := context.WithCancel(context.Background())
	abbrechen()

	if _, _, err := planer.Wortlistenfeld(ctx, karte, "type", vorlage); !errors.Is(
		err, context.Canceled) {
		t.Errorf("Wortlistenfeld gab %v", err)
	}

	if _, _, err := planer.Wortwahl(ctx, karte, "type", vorlage, 0, []int{0}); !errors.Is(
		err, context.Canceled) {
		t.Errorf("Wortwahl gab %v", err)
	}
}
