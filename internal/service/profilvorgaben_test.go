package service_test

// **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Stand nennt alle fünf Arten — auch die leeren.**
//
// **Ohne die leeren könnte die Seite keine neue setzen**, ohne die Liste der
// Arten selbst zu kennen — und dann stünde sie an zwei Stellen.
func TestDerVorgabenstandNenntAlleFuenf(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueVorgabenverwaltung(
		neuesVorgabendoppel(), &angebotsdoppel{})

	staende, fehler := verwaltung.Stand(context.Background(), 1)
	if fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if len(staende) != 5 {
		t.Fatalf("es sind %d Stände", len(staende))
	}

	for _, stand := range staende {
		if stand.Werte == nil {
			t.Errorf("%s: Werte ist nil statt einer leeren Liste", stand.Art)
		}

		if stand.Verwaist == nil {
			t.Errorf("%s: Verwaist ist nil statt einer leeren Liste", stand.Art)
		}
	}
}

// **Was gesetzt wurde, kommt zurück — samt dem Angebot.**
func TestEineGesetzteVorgabeKommtZurueck(t *testing.T) {
	t.Parallel()

	ablage := neuesVorgabendoppel()
	angebot := &angebotsdoppel{wahlen: map[domain.Vorgabeart][]service.Vorgabewahl{
		domain.VorgabeDeck: {
			{Wert: "1", Beschriftung: "Erstes"},
			{Wert: "2", Beschriftung: "Zweites"},
		},
	}}

	verwaltung := service.NeueVorgabenverwaltung(ablage, angebot)

	if fehler := verwaltung.Setzen(context.Background(), 1, domain.Profilvorgabe{
		Art: domain.VorgabeDeck, Werte: []string{"2", "1"},
	}); fehler != nil {
		t.Fatalf("Setzen: %v", fehler)
	}

	staende, fehler := verwaltung.Stand(context.Background(), 1)
	if fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	for _, stand := range staende {
		if stand.Art != string(domain.VorgabeDeck) {
			continue
		}

		// **Die Reihenfolge ist die Aussage** — „Präferenz" heißt genau das.
		if !slices.Equal(stand.Werte, []string{"2", "1"}) {
			t.Errorf("Werte = %v", stand.Werte)
		}

		if len(stand.Moeglich) != 2 {
			t.Errorf("Moeglich = %v", stand.Moeglich)
		}

		if !stand.Mehrwertig {
			t.Error("Decks gelten als einwertig")
		}

		return
	}

	t.Fatal("die Art `deck` steht nicht im Stand")
}

// **Ein Wert, den es nicht gibt, wird beim Setzen abgewiesen.**
//
// **Und zwar gegen dieselbe Liste, aus der die Fläche ihre Auswahl baut** —
// zwei Listen liefen auseinander, sobald jemand ein Format löscht.
func TestEinUnbekannterWertWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	angebot := &angebotsdoppel{wahlen: map[domain.Vorgabeart][]service.Vorgabewahl{
		domain.VorgabeFormat: {{Wert: "5", Beschriftung: "Standard"}},
	}}

	verwaltung := service.NeueVorgabenverwaltung(neuesVorgabendoppel(), angebot)

	fehler := verwaltung.Setzen(context.Background(), 1, domain.Profilvorgabe{
		Art: domain.VorgabeFormat, Werte: []string{"4711"},
	})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
	}

	// **Der Beleg, dass die Prüfung nicht alles abweist** (§10).
	if fehler := verwaltung.Setzen(context.Background(), 1, domain.Profilvorgabe{
		Art: domain.VorgabeFormat, Werte: []string{"5"},
	}); fehler != nil {
		t.Errorf("der echte Wert wurde abgewiesen: %v", fehler)
	}
}

// **Eine Vorgabe überlebt das Löschen dessen, worauf sie zeigt — und wird
// gemeldet.**
func TestEineVerwaisteVorgabeWirdGemeldet(t *testing.T) {
	t.Parallel()

	ablage := neuesVorgabendoppel()
	angebot := &angebotsdoppel{wahlen: map[domain.Vorgabeart][]service.Vorgabewahl{
		domain.VorgabeDeck: {{Wert: "1", Beschriftung: "Erstes"}},
	}}

	verwaltung := service.NeueVorgabenverwaltung(ablage, angebot)

	if fehler := verwaltung.Setzen(context.Background(), 1, domain.Profilvorgabe{
		Art: domain.VorgabeDeck, Werte: []string{"1"},
	}); fehler != nil {
		t.Fatalf("Setzen: %v", fehler)
	}

	// **Das Deck verschwindet aus dem Angebot** — so, wie es ein Löschen tut.
	angebot.wahlen[domain.VorgabeDeck] = nil

	staende, fehler := verwaltung.Stand(context.Background(), 1)
	if fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	for _, stand := range staende {
		if stand.Art != string(domain.VorgabeDeck) {
			continue
		}

		// **Sie steht weiter da** — weggeräumt hätte ihm jemand seine Wahl
		// genommen, während er nicht hinsieht (§8).
		if !slices.Equal(stand.Werte, []string{"1"}) {
			t.Errorf("Werte = %v", stand.Werte)
		}

		if !slices.Equal(stand.Verwaist, []string{"1"}) {
			t.Errorf("Verwaist = %v", stand.Verwaist)
		}

		return
	}

	t.Fatal("die Art `deck` steht nicht im Stand")
}

// **Wirksam gibt den einen Wert einer einwertigen Art.**
func TestWirksamGibtDenEinenWert(t *testing.T) {
	t.Parallel()

	angebot := &angebotsdoppel{wahlen: map[domain.Vorgabeart][]service.Vorgabewahl{
		domain.VorgabeSpiel: {{Wert: "magic", Beschriftung: "Magic"}},
	}}

	verwaltung := service.NeueVorgabenverwaltung(neuesVorgabendoppel(), angebot)

	t.Run("ohne Vorgabe kommt nichts", func(t *testing.T) {
		// **Leer ist ein Zustand und kein Fehler** — ein frisches Profil
		// zeigt auf nichts.
		wert, fehler := verwaltung.Wirksam(
			context.Background(), 1, domain.VorgabeSpiel)
		if fehler != nil || wert != "" {
			t.Errorf("wert = %q, Fehler = %v", wert, fehler)
		}
	})

	if fehler := verwaltung.Setzen(context.Background(), 1, domain.Profilvorgabe{
		Art: domain.VorgabeSpiel, Werte: []string{"magic"},
	}); fehler != nil {
		t.Fatalf("Setzen: %v", fehler)
	}

	wert, fehler := verwaltung.Wirksam(context.Background(), 1, domain.VorgabeSpiel)
	if fehler != nil {
		t.Fatalf("Wirksam: %v", fehler)
	}

	if wert != "magic" {
		t.Errorf("wert = %q", wert)
	}

	t.Run("und eine mehrwertige Art wird abgewiesen", func(t *testing.T) {
		// **Weil die Frage falsch ist.** „Der eine Wert" der Präferenzdecks
		// gibt es nicht; wer sie will, fragt nach dem Stand.
		if _, fehler := verwaltung.Wirksam(
			context.Background(), 1, domain.VorgabeDeck); fehler == nil {
			t.Error("eine mehrwertige Art gab einen Wert")
		}
	})
}

// **Jeder Weg reicht sein Nein durch.**
func TestJederVorgabenwegReichtSeinNeinDurch(t *testing.T) {
	t.Parallel()

	stolpern := errors.New("die Platte ist voll")

	faelle := map[string]func() error{
		"Stand, der Bestand stolpert": func() error {
			ablage := neuesVorgabendoppel()
			ablage.kaputt = stolpern

			_, fehler := service.NeueVorgabenverwaltung(
				ablage, &angebotsdoppel{}).Stand(context.Background(), 1)

			return fehler
		},
		"Stand, das Angebot stolpert": func() error {
			_, fehler := service.NeueVorgabenverwaltung(
				neuesVorgabendoppel(), &angebotsdoppel{kaputt: stolpern},
			).Stand(context.Background(), 1)

			return fehler
		},
		"Setzen, das Angebot stolpert": func() error {
			return service.NeueVorgabenverwaltung(
				neuesVorgabendoppel(), &angebotsdoppel{kaputt: stolpern},
			).Setzen(context.Background(), 1, domain.Profilvorgabe{
				Art: domain.VorgabeSpiel, Werte: []string{"magic"},
			})
		},
		"Setzen, der Bestand stolpert": func() error {
			ablage := neuesVorgabendoppel()
			ablage.kaputt = stolpern

			return service.NeueVorgabenverwaltung(ablage, &angebotsdoppel{
				wahlen: map[domain.Vorgabeart][]service.Vorgabewahl{
					domain.VorgabeSpiel: {{Wert: "magic"}},
				},
			}).Setzen(context.Background(), 1, domain.Profilvorgabe{
				Art: domain.VorgabeSpiel, Werte: []string{"magic"},
			})
		},
		"Setzen, die Vorgabe ist krumm": func() error {
			return service.NeueVorgabenverwaltung(
				neuesVorgabendoppel(), &angebotsdoppel{},
			).Setzen(context.Background(), 1, domain.Profilvorgabe{
				Art: "kartenrücken", Werte: []string{"1"},
			})
		},
		"Wirksam, der Bestand stolpert": func() error {
			ablage := neuesVorgabendoppel()
			ablage.kaputt = stolpern

			_, fehler := service.NeueVorgabenverwaltung(
				ablage, &angebotsdoppel{}).Wirksam(
				context.Background(), 1, domain.VorgabeSpiel)

			return fehler
		},
	}

	for name, weg := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := weg(); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}
