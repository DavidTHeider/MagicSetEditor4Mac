package app

// **Der Bestand hinter der Übersicht** — Werkstück U11 von Phase 5.1 (#1422):
// ob es ein Set, ein Deck oder ein Paket noch gibt, sagt die Anwendung über
// die Nähte, die sie ohnehin hat.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type innererBestand struct{ fehler error }

func (b innererBestand) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Name: "Probeset"}}, b.fehler
}

func (innererBestand) Karten(context.Context, int64) ([]*domain.CardRow, error) { return nil, nil }

func (innererBestand) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) { return nil, nil }

type innereDecks struct{ fehler error }

func (d innereDecks) Anlegen(context.Context, domain.Deck) (domain.Deck, error) {
	return domain.Deck{}, nil
}

func (d innereDecks) Umbenennen(context.Context, domain.Deck) (domain.Deck, error) {
	return domain.Deck{}, nil
}

func (d innereDecks) Deck(_ context.Context, kennung int64) (domain.Deck, error) {
	if d.fehler != nil {
		return domain.Deck{}, d.fehler
	}

	if kennung != 3 {
		return domain.Deck{}, domain.ErrNotFound
	}

	return domain.Deck{ID: 3, Name: "Mono Rot"}, nil
}

func (d innereDecks) Decks(context.Context, string) ([]domain.Deck, error) { return nil, nil }

func (d innereDecks) Kopieren(context.Context, int64, string) (domain.Deck, error) {
	return domain.Deck{}, nil
}

func (d innereDecks) PlatzLegen(context.Context, int64, domain.Deckplatz) (domain.Deckplatz, error) {
	return domain.Deckplatz{}, nil
}

func (d innereDecks) PlatzNehmen(context.Context, int64, int64) error { return nil }

func (d innereDecks) PlatzVerschieben(
	context.Context, int64, int64, domain.Deckbereich,
) (domain.Deckplatz, error) {
	return domain.Deckplatz{}, nil
}

func (d innereDecks) FormatWaehlen(context.Context, int64, int64) (domain.Deck, error) {
	return domain.Deck{}, nil
}

func (d innereDecks) Loeschen(context.Context, int64) error { return nil }

type innerePakete struct {
	planlos

	fehler error
}

// planlos gibt Plan, Ersetzen und Zurück ohne Wirkung — für Doppel, deren
// Läufe sie nicht brauchen (DOM-01.AK-02, #1696).
type planlos struct{}

// Herkunftsart ohne Wirkung: mitgeliefert, wie ein Paket ohne Ort.
func (planlos) Herkunftsart(domain.Paket) domain.Paketherkunftsart { return domain.PaketMitgeliefert }

func (planlos) PlanEntfernen(context.Context, string, string) (service.Paketplan, error) {
	return service.Paketplan{}, nil
}

// Kopieren gibt eine leere Kopie — FMT-02 (#1691).
func (planlos) Kopieren(context.Context, string, string, string) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (planlos) Weitergeben(context.Context, []string, string) (string, error) { return "", nil }

func (planlos) Einlesen(context.Context, string) (*service.Uebernahme, error) { return nil, nil }

func (planlos) PlanEinbinden(context.Context, string) (service.Paketplan, error) {
	return service.Paketplan{}, nil
}

func (planlos) PlanZurueck(context.Context, string, string) (service.Paketplan, error) {
	return service.Paketplan{}, nil
}

func (planlos) Ersetzen(context.Context, string, bool) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (planlos) Zurueck(context.Context, string, string) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (planlos) Setbedarf(context.Context, int64) ([]service.Abhaengigkeitsbefund, error) {
	return nil, nil
}

func (planlos) Kopfbedarf(context.Context, domain.SetHead) ([]service.Abhaengigkeitsbefund, error) {
	return nil, nil
}

func (p innerePakete) Auffrischen(context.Context) ([]domain.Befund, error) { return nil, nil }

func (p innerePakete) Vermerken(context.Context, ...service.Vermerk) ([]domain.Befund, error) {
	return nil, nil
}

func (p innerePakete) Uebernehmen(context.Context, string) (*service.Uebernahme, error) {
	return nil, nil
}

func (p innerePakete) Abhaengigkeiten(context.Context) ([]service.Abhaengigkeitsbefund, error) {
	return nil, nil
}

func (p innerePakete) Liste(context.Context) ([]domain.Paket, error) {
	return []domain.Paket{{Art: "mse-style", Name: "magic-m15.mse-style"}}, p.fehler
}

// Uebersicht kommt in diesem Lauf nicht vor — DOM-01.AK-01 (#1696).
func (p innerePakete) Uebersicht(context.Context) ([]service.Paketzeile, error) {
	return nil, nil
}

func (p innerePakete) Installieren(context.Context, string) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func (p innerePakete) Entfernen(context.Context, string, string, string) error { return nil }

func (p innerePakete) Anlegen(context.Context, service.Neuanlage) (domain.Paket, error) {
	return domain.Paket{}, nil
}

func TestDerArbeitsbestandFragtDieNaehte(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.bestand = innererBestand{}
	anwendung.decks = innereDecks{}
	anwendung.pakete = innerePakete{}

	bestand := arbeitsbestand{app: anwendung}

	faelle := []struct {
		name   string
		objekt domain.Arbeitsobjekt
		gibt   bool
	}{
		{"ein Set, das es gibt", domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "7"}, true},
		{"ein Set, das es nicht gibt", domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "9"}, false},
		{"eine Setkennung, die keine ist", domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "sieben"}, false},
		{"ein Deck, das es gibt", domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "3"}, true},
		{"ein Deck, das es nicht gibt", domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "4"}, false},
		{"eine Deckkennung, die keine ist", domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "drei"}, false},
		{"ein Paket, das es gibt", domain.Arbeitsobjekt{Art: domain.ArbeitsartPaket, Kennung: "magic-m15.mse-style"}, true},
		{"ein Paket, das es nicht gibt", domain.Arbeitsobjekt{Art: domain.ArbeitsartPaket, Kennung: "fremd"}, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gibt, err := bestand.Gibt(t.Context(), fall.objekt)
			if err != nil || gibt != fall.gibt {
				t.Errorf("Gibt = %v, %v — erwartet %v", gibt, err, fall.gibt)
			}
		})
	}
}

func TestDerArbeitsbestandReichtFehlerDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("kaputt")
	anwendung := NeueApp("")
	anwendung.bestand = innererBestand{fehler: kaputt}
	anwendung.decks = innereDecks{fehler: kaputt}
	anwendung.pakete = innerePakete{fehler: kaputt}
	bestand := arbeitsbestand{app: anwendung}

	for _, objekt := range []domain.Arbeitsobjekt{
		{Art: domain.ArbeitsartSet, Kennung: "7"},
		{Art: domain.ArbeitsartDeck, Kennung: "3"},
		{Art: domain.ArbeitsartPaket, Kennung: "magic-m15.mse-style"},
	} {
		if _, err := bestand.Gibt(t.Context(), objekt); !errors.Is(err, kaputt) {
			t.Errorf("%s: %v", objekt.Art, err)
		}
	}

	// **Ein Set ohne Bestand hat keinen Namen — die Kennung bleibt.**
	anwendung.bestand = nil
	if name := anwendung.arbeitsname(t.Context(), domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "7"}); name != "7" {
		t.Errorf("der Name ohne Bestand ist %q", name)
	}

	anwendung.bestand = innererBestand{fehler: kaputt}
	if name := anwendung.arbeitsname(t.Context(), domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "7"}); name != "7" {
		t.Errorf("der Name bei kaputtem Bestand ist %q", name)
	}

	// **Ein Deck ohne Naht, eine Deckkennung, die keine ist, ein Paket.**
	anwendung.decks = nil
	if name := anwendung.arbeitsname(t.Context(), domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "3"}); name != "3" {
		t.Errorf("der Deckname ohne Naht ist %q", name)
	}

	anwendung.decks = innereDecks{}
	if name := anwendung.arbeitsname(t.Context(), domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "drei"}); name != "drei" {
		t.Errorf("der Deckname bei unlesbarer Kennung ist %q", name)
	}

	if name := anwendung.arbeitsname(t.Context(), domain.Arbeitsobjekt{Art: domain.ArbeitsartPaket, Kennung: "magic-m15.mse-style"}); name != "magic-m15.mse-style" {
		t.Errorf("der Paketname ist %q", name)
	}
}
