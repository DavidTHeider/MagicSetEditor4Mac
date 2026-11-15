package app

// **Die Zeile des Drucks an der Naht** — Werkstück A2 von Phase 5.3
// (#1441):
// `druckID` als Zeichenkette wie jede Kennung, `vorlaeufig` als Zustand.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDasExemplarDTOTraegtDieDruckzeileAlsText(t *testing.T) {
	t.Parallel()

	mit := zuExemplarDTO(&domain.Exemplar{ID: 1, Printing: "abc", DruckID: 42, Menge: 1})
	if mit.DruckID != "42" || mit.Vorlaeufig {
		t.Errorf("mit Alias: %+v", mit)
	}

	ohne := zuExemplarDTO(&domain.Exemplar{ID: 2, DruckID: 7, Menge: 1})
	if ohne.DruckID != "7" || !ohne.Vorlaeufig {
		t.Errorf("ohne Alias: %+v", ohne)
	}

	keiner := zuExemplarDTO(&domain.Exemplar{ID: 3, Printing: "abc", Menge: 1})
	if keiner.DruckID != "" || keiner.Vorlaeufig {
		t.Errorf("ohne Zeile: %+v", keiner)
	}
}

func TestDasExemplarDTOBringtDieDruckzeileZurueck(t *testing.T) {
	t.Parallel()

	exemplar, err := ausExemplarDTO(ExemplarDTO{DruckID: "7", Menge: 1})
	if err != nil || exemplar.DruckID != 7 {
		t.Errorf("Exemplar %+v, Fehler %v", exemplar, err)
	}

	leer, err := ausExemplarDTO(ExemplarDTO{Printing: "abc", Menge: 1})
	if err != nil || leer.DruckID != 0 {
		t.Errorf("ohne Zeile: %+v, Fehler %v", leer, err)
	}

	var feldfehler *domain.FieldError

	if _, err := ausExemplarDTO(ExemplarDTO{DruckID: "sieben", Menge: 1}); !errors.As(err, &feldfehler) ||
		feldfehler.Field != "druckID" {
		t.Errorf("eine unlesbare Kennung: %v", err)
	}
}

func TestDieAnderenNaehteTragenDieDruckzeile(t *testing.T) {
	t.Parallel()

	verweis := alsVerweisDTO(5, domain.Kartenverweis{DruckID: 9, Name: "Bolt"})
	if verweis.DruckID != "9" || !verweis.Vorlaeufig {
		t.Errorf("Verweis: %+v", verweis)
	}

	deck := alsDeck(domain.Deck{Plaetze: []domain.Deckplatz{
		{ID: 1, DruckID: 3, Menge: 1, Bereich: domain.BereichHauptdeck},
		{ID: 2, KarteID: 4, Menge: 1, Bereich: domain.BereichHauptdeck},
	}})
	if deck.Plaetze[0].DruckID != "3" || !deck.Plaetze[0].Vorlaeufig ||
		deck.Plaetze[1].DruckID != "" || deck.Plaetze[1].Vorlaeufig {
		t.Errorf("Deck: %+v", deck.Plaetze)
	}

	platz := alsDeckplatz(domain.Deckplatz{ID: 1, Printing: "abc", DruckID: 3, Menge: 1})
	if platz.DruckID != "3" || platz.Vorlaeufig {
		t.Errorf("Platz: %+v", platz)
	}

	preise := zuPreisstandDTOs([]domain.Preisstand{{Printing: "abc", DruckID: 3}}, 0)
	if preise[0].DruckID != "3" {
		t.Errorf("Preis: %+v", preise[0])
	}

	teilnehmer := alsTeilnehmer(domain.Teilnehmer{Deckabzug: []domain.Abzugsplatz{{Printing: "abc", DruckID: 3}}})
	if teilnehmer.Deckabzug[0].DruckID != "3" {
		t.Errorf("Abzug: %+v", teilnehmer.Deckabzug[0])
	}
}
