package domain_test

// **Weiterarbeiten** — Werkstück U11 von Phase 5.1
// (#1422):
// was zuletzt geöffnet, eingeführt oder ausgegeben wurde, und was jemand
// angeheftet hat — je Profil, geordnet, gekappt.

import (
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func um(stunde int) time.Time {
	return time.Date(2026, time.September, 9, stunde, 0, 0, 0, time.UTC)
}

func schritt(art domain.Arbeitsart, kennung string, vorgang domain.Arbeitsvorgang, stunde int) domain.Arbeitsschritt {
	return domain.Arbeitsschritt{
		Objekt:    domain.Arbeitsobjekt{Art: art, Kennung: kennung, Name: "Objekt " + kennung},
		Vorgang:   vorgang,
		Zeitpunkt: um(stunde),
	}
}

func TestEinArbeitsobjektNenntArtUndKennung(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		objekt  domain.Arbeitsobjekt
		befunde []string
	}{
		{"ein Set", domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: "7", Name: "Probe"}, nil},
		{"ein Deck", domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: "3"}, nil},
		{"ein Paket", domain.Arbeitsobjekt{Art: domain.ArbeitsartPaket, Kennung: "magic-m15"}, nil},
		{"ohne Kennung", domain.Arbeitsobjekt{Art: domain.ArbeitsartSet}, []string{"kennung"}},
		{"erfundene Art", domain.Arbeitsobjekt{Art: "karte", Kennung: "1"}, []string{"art"}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var felder []string
			for _, befund := range fall.objekt.Validate() {
				felder = append(felder, befund.Field)
			}

			if !slices.Equal(felder, fall.befunde) {
				t.Errorf("Befunde %v, erwartet %v", felder, fall.befunde)
			}
		})
	}
}

func TestEinSchrittNenntSeinenVorgang(t *testing.T) {
	t.Parallel()

	gut := schritt(domain.ArbeitsartSet, "1", domain.VorgangGeoeffnet, 9)
	if befunde := gut.Validate(); len(befunde) != 0 {
		t.Errorf("ein guter Schritt gibt %v", befunde)
	}

	schlecht := gut
	schlecht.Vorgang = "angesehen"

	if befunde := schlecht.Validate(); len(befunde) != 1 || befunde[0].Field != "vorgang" {
		t.Errorf("ein erfundener Vorgang gibt %v", befunde)
	}

	ohneZeit := gut
	ohneZeit.Zeitpunkt = time.Time{}

	if befunde := ohneZeit.Validate(); len(befunde) != 1 || befunde[0].Field != "zeitpunkt" {
		t.Errorf("ein Schritt ohne Zeitpunkt gibt %v", befunde)
	}

	// **Ein Befund am Objekt trägt seine Stelle.**
	ohneKennung := gut
	ohneKennung.Objekt.Kennung = ""

	if befunde := ohneKennung.Validate(); len(befunde) != 1 || befunde[0].Field != "objekt.kennung" {
		t.Errorf("ein Schritt ohne Kennung gibt %v", befunde)
	}
}

// **Geordnet wird nach Zeit, gruppiert nach Vorgang, gekappt je Gruppe** —
// und »Weiterarbeiten« zeigt auf den jüngsten Schritt überhaupt.
func TestDieUebersichtOrdnetUndKappt(t *testing.T) {
	t.Parallel()

	schritte := []domain.Arbeitsschritt{
		schritt(domain.ArbeitsartSet, "1", domain.VorgangGeoeffnet, 8),
		schritt(domain.ArbeitsartSet, "2", domain.VorgangGeoeffnet, 11),
		schritt(domain.ArbeitsartDeck, "3", domain.VorgangGeoeffnet, 10),
		schritt(domain.ArbeitsartSet, "4", domain.VorgangEingefuehrt, 9),
		schritt(domain.ArbeitsartSet, "2", domain.VorgangAusgegeben, 12),
		schritt(domain.ArbeitsartSet, "5", domain.VorgangGeoeffnet, 7),
	}
	angeheftet := []domain.Arbeitsobjekt{{Art: domain.ArbeitsartDeck, Kennung: "3", Name: "Deck"}}

	uebersicht := domain.Weiterarbeit(schritte, angeheftet, 2)

	if uebersicht.Zuletzt == nil || uebersicht.Zuletzt.Objekt.Kennung != "2" ||
		uebersicht.Zuletzt.Vorgang != domain.VorgangAusgegeben {
		t.Errorf("Zuletzt ist %+v — erwartet Set 2, ausgegeben um zwölf", uebersicht.Zuletzt)
	}

	kennungen := func(schritte []domain.Arbeitsschritt) []string {
		alle := make([]string, 0, len(schritte))
		for _, eines := range schritte {
			alle = append(alle, eines.Objekt.Kennung)
		}

		return alle
	}

	// Geöffnet: 2 (11), 3 (10), 1 (8), 5 (7) → gekappt auf zwei.
	if !slices.Equal(kennungen(uebersicht.Geoeffnet), []string{"2", "3"}) {
		t.Errorf("geöffnet %v", kennungen(uebersicht.Geoeffnet))
	}

	if !slices.Equal(kennungen(uebersicht.Eingefuehrt), []string{"4"}) ||
		!slices.Equal(kennungen(uebersicht.Ausgegeben), []string{"2"}) {
		t.Errorf("eingeführt %v, ausgegeben %v", kennungen(uebersicht.Eingefuehrt), kennungen(uebersicht.Ausgegeben))
	}

	if len(uebersicht.Angeheftet) != 1 || uebersicht.Angeheftet[0].Kennung != "3" {
		t.Errorf("angeheftet %v", uebersicht.Angeheftet)
	}

	// **Leer ist leer** — kein Zuletzt, leere Listen statt nil, damit das
	// Fenster nicht über `null` stolpert.
	leer := domain.Weiterarbeit(nil, nil, 5)
	if leer.Zuletzt != nil || leer.Geoeffnet == nil || leer.Angeheftet == nil {
		t.Errorf("die leere Übersicht ist %+v", leer)
	}

	// **Angeheftet weiß, ob es angeheftet ist.**
	if !uebersicht.IstAngeheftet(domain.ArbeitsartDeck, "3") || uebersicht.IstAngeheftet(domain.ArbeitsartSet, "3") {
		t.Error("IstAngeheftet irrt")
	}
}
