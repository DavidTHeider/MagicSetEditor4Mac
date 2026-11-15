package domain_test

// **Die portable Referenz und ihre Auflösung** — Werkstück A2 von Phase 5.3
// (#1441). Was einen Bestand verlässt, trägt System, Semantik und Wert —
// nie eine lokale Zeilennummer; was hereinkommt, wird über Kennungen
// aufgelöst, und ein Wert, der auf zwei Drucke zeigt, ist ein Konflikt.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var (
	bolt = domain.Druck{
		ID: 7, Name: "Lightning Bolt", Setcode: "lea", Nummer: "161", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "e3285e6b-3e79-4d7c-bf96-d920f973b122"},
			{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"},
			{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"},
		},
	}
	boltBeta = domain.Druck{
		ID: 8, Name: "Lightning Bolt", Setcode: "leb", Nummer: "161", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "a3285e6b-3e79-4d7c-bf96-d920f973b122"},
			{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"},
		},
	}
	eigene = domain.Druck{ID: 9, Name: "Eigene Karte", Setcode: "xyz", Nummer: "1", Sprache: "de"}
)

func TestEinDruckWirdZurPortablenReferenz(t *testing.T) {
	t.Parallel()

	verweis := bolt.AlsVerweis()
	if verweis.Fassung != domain.Verweisfassung || verweis.System != domain.SystemScryfall ||
		verweis.Semantik != domain.SemantikDruck || verweis.Wert != "e3285e6b-3e79-4d7c-bf96-d920f973b122" {
		t.Errorf("die Referenz ist %+v", verweis)
	}

	// **Name, Setcode, Nummer und Sprache reisen mit** — der Rückfall, wenn
	// das Ziel das System nicht kennt.
	if verweis.Name != "Lightning Bolt" || verweis.Setcode != "lea" || verweis.Nummer != "161" || verweis.Sprache != "en" {
		t.Errorf("der Rückfall fehlt: %+v", verweis)
	}

	// **Nie die lokale Zeile.**
	if verweis.Wert == "7" {
		t.Error("die Referenz trägt die lokale Kennung")
	}

	// **Ohne eindeutige Kennung nur der Rückfall.**
	ohne := eigene.AlsVerweis()
	if ohne.System != "" || ohne.Wert != "" || ohne.Name != "Eigene Karte" {
		t.Errorf("die Referenz eines vorläufigen Drucks ist %+v", ohne)
	}

	if befunde := ohne.Validate(); len(befunde) != 0 {
		t.Errorf("ein Rückfall mit Name ist gültig — Befunde %v", befunde)
	}

	if befunde := (domain.Druckverweis{Fassung: domain.Verweisfassung}).Validate(); len(befunde) != 1 {
		t.Errorf("eine Referenz ohne alles gibt %v — erwartet genau einen Befund", befunde)
	}

	if befunde := (domain.Druckverweis{Fassung: 99, Name: "x"}).Validate(); len(befunde) != 1 || befunde[0].Field != "fassung" {
		t.Errorf("eine Referenz aus der Zukunft gibt %v", befunde)
	}
}

func TestDieAufloesungFindetGenauEinenDruckUeberDieKennung(t *testing.T) {
	t.Parallel()

	kandidaten := []domain.Druck{bolt, boltBeta, eigene}

	verweis := domain.Druckverweis{Fassung: 1, System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "A3285E6B-3E79-4D7C-BF96-D920F973B122"}
	ergebnis := domain.LoeseAuf(kandidaten, verweis)

	if ergebnis.Art != domain.AufloesungEindeutig || !slices.Equal(ergebnis.DruckIDs, []int64{8}) {
		t.Errorf("Ergebnis %+v — erwartet eindeutig 8 (Groß-/Kleinschreibung zählt nicht)", ergebnis)
	}
}

// **Ein Konzept nennt keinen Druck**: dieselbe Oracle-ID zeigt auf zwei
// Drucke, und das ist eine Mehrdeutigkeit, kein Treffer.
func TestEinKonzeptLoestNichtAuf(t *testing.T) {
	t.Parallel()

	verweis := domain.Druckverweis{Fassung: 1, System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"}
	ergebnis := domain.LoeseAuf([]domain.Druck{bolt, boltBeta}, verweis)

	if ergebnis.Art != domain.AufloesungMehrdeutig || !slices.Equal(ergebnis.DruckIDs, []int64{7, 8}) {
		t.Errorf("Ergebnis %+v — erwartet mehrdeutig 7, 8", ergebnis)
	}
}

// **Derselbe Textwert in zwei Systemen ist nicht dieselbe Kennung.**
func TestDerselbeWertInEinemAnderenSystemTrifftNicht(t *testing.T) {
	t.Parallel()

	tcg := domain.Druckverweis{Fassung: 1, System: domain.SystemTCGplayer, Semantik: domain.SemantikProdukt, Wert: "1"}
	if ergebnis := domain.LoeseAuf([]domain.Druck{bolt}, tcg); ergebnis.Art != domain.AufloesungUnbekannt {
		t.Errorf("Ergebnis %+v — »1« bei TCGplayer ist nicht »1« bei Cardmarket", ergebnis)
	}

	cardmarket := domain.Druckverweis{Fassung: 1, System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"}
	if ergebnis := domain.LoeseAuf([]domain.Druck{bolt}, cardmarket); ergebnis.Art != domain.AufloesungEindeutig {
		t.Errorf("Ergebnis %+v — erwartet eindeutig", ergebnis)
	}
}

// **Eine Produktkennung, die auf zwei Drucke zeigt, ist ein Konflikt** — kein
// Automerge, und schon gar keiner über den Namen.
func TestEineKollidierendeKennungIstEinKonflikt(t *testing.T) {
	t.Parallel()

	zweiter := boltBeta
	zweiter.Kennungen = append(slices.Clone(zweiter.Kennungen),
		domain.Druckkennung{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"})

	verweis := domain.Druckverweis{Fassung: 1, System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"}
	ergebnis := domain.LoeseAuf([]domain.Druck{bolt, zweiter}, verweis)

	if ergebnis.Art != domain.AufloesungMehrdeutig || len(ergebnis.DruckIDs) != 2 {
		t.Errorf("Ergebnis %+v — erwartet einen Konflikt über zwei Drucke", ergebnis)
	}
}

// **Ohne Kennung entscheidet der Rückfall** — Name, Setcode, Nummer und
// Sprache **zusammen**; der Name allein trifft nie.
func TestDerRueckfallTrifftNurVollstaendig(t *testing.T) {
	t.Parallel()

	kandidaten := []domain.Druck{bolt, boltBeta, eigene}

	voll := domain.Druckverweis{Fassung: 1, Name: "lightning bolt", Setcode: "LEB", Nummer: "161", Sprache: "en"}
	if ergebnis := domain.LoeseAuf(kandidaten, voll); ergebnis.Art != domain.AufloesungEindeutig || !slices.Equal(ergebnis.DruckIDs, []int64{8}) {
		t.Errorf("Ergebnis %+v — erwartet eindeutig 8", ergebnis)
	}

	nurName := domain.Druckverweis{Fassung: 1, Name: "Lightning Bolt"}
	if ergebnis := domain.LoeseAuf(kandidaten, nurName); ergebnis.Art != domain.AufloesungUnbekannt {
		t.Errorf("Ergebnis %+v — der Name allein trifft nicht", ergebnis)
	}

	// **Eine unbekannte Kennung fällt auf den Rückfall zurück**, statt zu scheitern.
	unbekannt := domain.Druckverweis{
		Fassung: 1, System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: "gibtsnicht",
		Name: "Eigene Karte", Setcode: "xyz", Nummer: "1", Sprache: "de",
	}
	if ergebnis := domain.LoeseAuf(kandidaten, unbekannt); ergebnis.Art != domain.AufloesungEindeutig || !slices.Equal(ergebnis.DruckIDs, []int64{9}) {
		t.Errorf("Ergebnis %+v — erwartet eindeutig 9 über den Rückfall", ergebnis)
	}

	if ergebnis := domain.LoeseAuf(nil, voll); ergebnis.Art != domain.AufloesungUnbekannt || len(ergebnis.DruckIDs) != 0 {
		t.Errorf("ohne Kandidaten ist das Ergebnis %+v", ergebnis)
	}
}
