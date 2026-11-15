package service_test

// **Die Fehlerwege der Deckübernahme** — Werkstück 12 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `uebernahme_test.go` mit ihnen über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was ankommt**, hier,
// **was gemeldet wird, wenn der Bestand unterwegs abbricht**.
//
// # Warum jeder einzeln dasteht
//
// **Weil dieser Weg fünf Nähte hintereinander anfasst** — Zeiger, Übersicht,
// Entnahme, Speicher, Herkunft — und ein gemeinsamer Lauf nur zeigte, dass
// **einer** von ihnen durchreicht (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jeder Fehler auf dem Weg reicht durch und lässt nichts halb stehen.**
func TestJederFehlerwegDerUebernahmeReichtDurch(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("die Platte klemmt")

	faelle := map[string]func(*uebernahmedoppel){
		"beim Lesen des fremden Decks": func(d *uebernahmedoppel) {
			d.lesefehler = klemmt
		},
		"beim Anlegen und beim Platz": func(d *uebernahmedoppel) {
			d.schreibfehl = klemmt
		},
		"beim Aufschreiben der Herkunft": func(d *uebernahmedoppel) {
			d.herkunftfehl = klemmt
		},
	}

	for name, stellen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uebernahme, doppel := zweiProfileMitDeck(t, domain.RolleMitarbeiter)
			stellen(doppel)

			if _, _, fehler := uebernahme.DeckUebernehmen(
				t.Context(), 2, 7); !errors.Is(fehler, klemmt) {
				t.Errorf("Fehler = %v, erwartet den Bestandsfehler", fehler)
			}
		})
	}
}

// **Und ein Platz, der einzeln scheitert, reißt den ganzen Weg ab.**
//
// # Warum er nicht als Befund durchgeht
//
// **Weil ein Befund sagt „das konnte nicht mit" und ein Fehler „der Bestand
// antwortet nicht".** Wer den zweiten als ersten meldete, ließe jemanden
// glauben, sein Deck sei vollständig bis auf zwei Karten — während in
// Wahrheit die Platte klemmt (§8).
func TestEinScheiterndenPlatzMeldetDieUebernahme(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("die Platte klemmt")
	uebernahme, doppel := zweiProfileMitDeck(t, domain.RolleMitarbeiter)
	doppel.platzfehler = klemmt

	if _, _, fehler := uebernahme.DeckUebernehmen(
		t.Context(), 2, 7); !errors.Is(fehler, klemmt) {
		t.Errorf("Fehler = %v, erwartet den Bestandsfehler", fehler)
	}
}

// **Ohne gezeigtes Profil wird abgewiesen und nichts gelesen.**
func TestOhneGezeigtesProfilUebernimmtNiemand(t *testing.T) {
	t.Parallel()

	uebernahme, doppel := zweiProfileMitDeck(t, domain.RolleMitarbeiter)
	doppel.zeigerfehler(errors.New("kein Zeiger"))

	if _, _, fehler := uebernahme.DeckUebernehmen(
		t.Context(), 2, 7); fehler == nil {
		t.Error("ohne gezeigtes Profil kam die Übernahme durch")
	}

	if doppel.entnommen {
		t.Error("das fremde Deck wurde trotzdem gelesen")
	}
}

// **Und ein Profil, das es nicht gibt, ebenso.**
func TestEinFehlendesQuellprofilWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	uebernahme, _ := zweiProfileMitDeck(t, domain.RolleMitarbeiter)

	if _, _, fehler := uebernahme.DeckUebernehmen(
		t.Context(), 4711, 7); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler = %v, erwartet ErrNotFound", fehler)
	}
}

// **Und die Sammlungseinfuhr schreibt auf, woher sie kam.**
//
// # Warum das hier steht und nicht bei der Ausfuhr
//
// **Weil es dieselbe Frage ist wie oben**: was passiert, wenn eine Naht
// dazukommt oder klemmt. Die Ausfuhr selbst prüft
// `sammlungsausfuhr_test.go`.
func TestDieEinfuhrSchreibtIhreHerkunft(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	herkunft := &herkunftsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(ablage).
		MitHerkunft(herkunft).
		Einfuegen(t.Context(),
			"printing,menge,zustand,sprache\n"+probeprinting+",2,nearMint,en\n",
			"sammlung.csv")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Fatalf("es entstanden %d Zeilen", len(bericht.Angelegt))
	}

	if len(herkunft.gesetzt) != 1 ||
		herkunft.gesetzt[0].Art != domain.HerkunftEinfuhr ||
		herkunft.gesetzt[0].Quelle != "sammlung.csv" {
		t.Errorf("aufgeschrieben wurde %+v", herkunft.gesetzt)
	}
}

// **Ohne Quelle wird nichts behauptet.**
//
// **Die Zusage aus Vermerk 0071**: eine falsche Nennung ist schlimmer als
// keine. Wer aus der Zwischenablage einfügt, nennt nichts — und dann steht
// auch nichts da.
func TestOhneQuelleWirdNichtsAufgeschrieben(t *testing.T) {
	t.Parallel()

	herkunft := &herkunftsdoppel{}

	if _, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).
		MitHerkunft(herkunft).
		Einfuegen(t.Context(),
			"printing,menge,zustand,sprache\n"+probeprinting+",2,nearMint,en\n",
			""); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if len(herkunft.gesetzt) != 0 {
		t.Errorf("aufgeschrieben wurde %+v", herkunft.gesetzt)
	}
}

// **Und ein Fehlschlag beim Aufschreiben reicht durch — das Exemplar bleibt.**
//
// **§7, Regel 3: geschrieben ist geschrieben.** Die Zeilen sind da; was fehlt,
// ist die Notiz darüber. Der Bericht kommt deshalb **mit** dem Fehler zurück.
func TestEinFehlschlagBeimAufschreibenLaesstDieZeilenStehen(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("die Platte klemmt")

	bericht, fehler := service.NeueSammlungsausfuhr(&sammlungsdoppel{}).
		MitHerkunft(&herkunftsdoppel{kaputt: klemmt}).
		Einfuegen(t.Context(),
			"printing,menge,zustand,sprache\n"+probeprinting+",2,nearMint,en\n",
			"sammlung.csv")

	if !errors.Is(fehler, klemmt) {
		t.Errorf("Fehler = %v, erwartet den Bestandsfehler", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Errorf("der Bericht nennt %d Zeilen", len(bericht.Angelegt))
	}
}

// herkunftsdoppel hält fest, was aufgeschrieben wurde.
type herkunftsdoppel struct {
	gesetzt []domain.Herkunft
	kaputt  error
}

func (d *herkunftsdoppel) ExemplarherkunftSetzen(
	_ context.Context, _ int64, herkunft domain.Herkunft,
) error {
	if d.kaputt != nil {
		return d.kaputt
	}

	d.gesetzt = append(d.gesetzt, herkunft)

	return nil
}

// **Eine Übersicht, die klemmt, reißt ab.**
//
// **Beim zweiten Ruf und nicht beim ersten**: der erste holt das gezeigte
// Profil, der zweite sucht die Quelle. Ein Fehler schon beim ersten prüfte
// jenen Weg und ließe diesen unbetreten (§10).
func TestEineKlemmendeUebersichtReisstDieUebernahmeAb(t *testing.T) {
	t.Parallel()

	uebernahme, doppel := zweiProfileMitDeck(t, domain.RolleMitarbeiter)
	doppel.ablage.nachErlaubnis = errors.New("die Platte klemmt")

	if _, _, fehler := uebernahme.DeckUebernehmen(
		t.Context(), 2, 7); !errors.Is(fehler, doppel.ablage.nachErlaubnis) {
		t.Errorf("Fehler = %v, erwartet den Bestandsfehler", fehler)
	}
}
