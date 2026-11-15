package app_test

// **Die Zuordnung an der Naht** — Werkstück U9 von Phase 5.1
// (#1420):
// Kennungen als Text, jede Zahl an ihrer Stelle, und kein Fehler geht
// verloren.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zuordnungsprobe ist ein Doppel für `app.Deckzuordnungsnaht`.
type zuordnungsprobe struct {
	stand   service.Zuordnungsstand
	gesetzt domain.Deckzuordnung
	liste   string
	decks   []service.Deckbezug

	// gerufen hält fest, womit zuletzt gerufen wurde.
	deckID, platzID, exemplarID, zuordnungID int64
	menge                                    int

	fehler error
}

func (z *zuordnungsprobe) Stand(_ context.Context, deckID int64) (service.Zuordnungsstand, error) {
	z.deckID = deckID

	return z.stand, z.fehler
}

func (z *zuordnungsprobe) Zuordnen(
	_ context.Context, deckID, platzID, exemplarID int64, menge int,
) (domain.Deckzuordnung, error) {
	z.deckID, z.platzID, z.exemplarID, z.menge = deckID, platzID, exemplarID, menge

	return z.gesetzt, z.fehler
}

func (z *zuordnungsprobe) Loesen(_ context.Context, deckID, id int64) error {
	z.deckID, z.zuordnungID = deckID, id

	return z.fehler
}

func (z *zuordnungsprobe) Fehlkarten(_ context.Context, deckID int64) (string, error) {
	z.deckID = deckID

	return z.liste, z.fehler
}

func (z *zuordnungsprobe) DecksMitExemplar(_ context.Context, exemplarID int64) (
	[]service.Deckbezug, error,
) {
	z.exemplarID = exemplarID

	return z.decks, z.fehler
}

// zuordnungsanwendung baut eine Anwendung mit der Naht.
func zuordnungsanwendung(t *testing.T) (*app.App, *zuordnungsprobe) {
	t.Helper()

	probe := &zuordnungsprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitDeckzuordnung(probe)

	return anwendung, probe
}

// **Der Stand kommt mit Kennungen als Text und jeder Zahl an ihrer Stelle.**
func TestDerZuordnungsstandKommtAlsText(t *testing.T) {
	anwendung, probe := zuordnungsanwendung(t)

	probe.stand = service.Zuordnungsstand{
		Haben: service.Habenstand{Fehlend: 2, Unbekannt: 1},
		Zeilen: []service.Zuordnungszeile{{
			Habenzeile: service.Habenzeile{
				Printing: "abc", DruckID: 4, Bezeichnung: "Blitz",
				Gebraucht: 4, Vorhanden: 2, Fehlend: 2,
			},
			Reserviert: 1, Zugeordnet: 1, Frei: 0,
			Zuordnungen: []service.Zuordnungsangabe{{
				Deckzuordnung: domain.Deckzuordnung{ID: 9, PlatzID: 3, ExemplarID: 5, Menge: 1},
				Bezeichnung:   "Blitz · M11", Abweichend: true,
			}},
		}},
	}

	antwort := anwendung.DeckZuordnungen(t.Context(), "7")
	if antwort.Fehler != nil || probe.deckID != 7 {
		t.Fatalf("Antwort %+v, Deck %d", antwort.Fehler, probe.deckID)
	}

	if antwort.Fehlend != 2 || antwort.Unbekannt != 1 || len(antwort.Zeilen) != 1 {
		t.Fatalf("%+v", antwort)
	}

	zeile := antwort.Zeilen[0]
	if zeile.DruckID != "4" || zeile.Reserviert != 1 || zeile.Zugeordnet != 1 || zeile.Frei != 0 {
		t.Errorf("Zeile %+v", zeile)
	}

	if len(zeile.Zuordnungen) != 1 {
		t.Fatalf("%d Zuordnungen", len(zeile.Zuordnungen))
	}

	if zuordnung := zeile.Zuordnungen[0]; zuordnung.ID != "9" || zuordnung.PlatzID != "3" ||
		zuordnung.ExemplarID != "5" || !zuordnung.Abweichend || zuordnung.Bezeichnung != "Blitz · M11" {
		t.Errorf("Zuordnung %+v", zuordnung)
	}
}

// **Jede Kennung geht als Zahl weiter — und was keine ist, ist ein Feldbefund.**
func TestDieZuordnungLiestIhreDreiKennungen(t *testing.T) {
	anwendung, probe := zuordnungsanwendung(t)
	probe.gesetzt = domain.Deckzuordnung{ID: 1, PlatzID: 3, ExemplarID: 5, Menge: 2}

	antwort := anwendung.DeckZuordnen(t.Context(), "7", "3", "5", 2)
	if antwort.Fehler != nil || antwort.Zuordnung == nil {
		t.Fatalf("%+v", antwort)
	}

	if probe.deckID != 7 || probe.platzID != 3 || probe.exemplarID != 5 || probe.menge != 2 {
		t.Errorf("gerufen mit %+v", probe)
	}

	if antwort.Zuordnung.ID != "1" || antwort.Zuordnung.Menge != 2 {
		t.Errorf("Zuordnung %+v", antwort.Zuordnung)
	}

	for name, roh := range map[string][3]string{
		"Deck":     {"sieben", "3", "5"},
		"Platz":    {"7", "drei", "5"},
		"Exemplar": {"7", "3", "fünf"},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := anwendung.DeckZuordnen(t.Context(), roh[0], roh[1], roh[2], 1).Fehler; fehler == nil {
				t.Error("die unlesbare Kennung kam durch")
			}
		})
	}
}

// **Lösen, Fehlkarten und der Rückweg reichen ihre Kennungen durch.**
func TestDieUebrigenZuordnungswegeReichenDurch(t *testing.T) {
	anwendung, probe := zuordnungsanwendung(t)
	probe.liste = "2 Blitz (M11) 149\n"
	probe.decks = []service.Deckbezug{{DeckID: 1, Name: "Aggro", Menge: 2}}

	if antwort := anwendung.DeckZuordnungLoesen(t.Context(), "7", "9"); antwort.Fehler != nil ||
		probe.deckID != 7 || probe.zuordnungID != 9 {
		t.Errorf("Lösen: %+v", antwort.Fehler)
	}

	if antwort := anwendung.DeckFehlkarten(t.Context(), "7"); antwort.Fehler != nil ||
		antwort.Liste != probe.liste {
		t.Errorf("Fehlkarten: %+v", antwort)
	}

	antwort := anwendung.ExemplarDecks(t.Context(), "5")
	if antwort.Fehler != nil || len(antwort.Decks) != 1 || antwort.Decks[0].DeckID != "1" ||
		antwort.Decks[0].Menge != 2 || probe.exemplarID != 5 {
		t.Errorf("Rückweg: %+v", antwort)
	}
}

// **Kein Fehler geht verloren — und keine unlesbare Kennung durch.**
func TestJederZuordnungswegMeldetSeinenFehler(t *testing.T) {
	anwendung, probe := zuordnungsanwendung(t)
	probe.fehler = errors.New("die Naht schweigt")

	for name, ruf := range map[string]func() *app.Fehler{
		"Stand":      func() *app.Fehler { return anwendung.DeckZuordnungen(t.Context(), "7").Fehler },
		"Zuordnen":   func() *app.Fehler { return anwendung.DeckZuordnen(t.Context(), "7", "3", "5", 1).Fehler },
		"Loesen":     func() *app.Fehler { return anwendung.DeckZuordnungLoesen(t.Context(), "7", "9").Fehler },
		"Fehlkarten": func() *app.Fehler { return anwendung.DeckFehlkarten(t.Context(), "7").Fehler },
		"Rueckweg":   func() *app.Fehler { return anwendung.ExemplarDecks(t.Context(), "5").Fehler },
		"StandKennung": func() *app.Fehler {
			return anwendung.DeckZuordnungen(t.Context(), "sieben").Fehler
		},
		"LoesenKennung": func() *app.Fehler {
			return anwendung.DeckZuordnungLoesen(t.Context(), "7", "neun").Fehler
		},
		"LoesenDeck": func() *app.Fehler {
			return anwendung.DeckZuordnungLoesen(t.Context(), "sieben", "9").Fehler
		},
		"FehlkartenKennung": func() *app.Fehler {
			return anwendung.DeckFehlkarten(t.Context(), "sieben").Fehler
		},
		"RueckwegKennung": func() *app.Fehler {
			return anwendung.ExemplarDecks(t.Context(), "fünf").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			if ruf() == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
