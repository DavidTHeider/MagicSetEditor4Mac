package domain_test

// **Die quellneutrale Druckidentität** — Werkstück A2 von Phase 5.3
// (#1441),
// entschieden in 0086: ein Druck ist eine lokale Zeile; wie andere ihn nennen,
// sind Kennungen mit System **und** Semantik — und nur die Semantik »druck«
// nennt genau einen.

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineKennungKenntIhrSystemUndIhreSemantik(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		kennung   domain.Druckkennung
		befunde   []string
		eindeutig bool
	}{
		{"Scryfall-Druck", domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "e3285e6b-3e79-4d7c-bf96-d920f973b122"}, nil, true},
		{"Oracle-Konzept", domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"}, nil, false},
		{"Cardmarket-Produkt", domain.Druckkennung{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1234"}, nil, false},
		{"Kunstwerk", domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikKunstwerk, Wert: "7ed2be7b-1c4a-4d8c-9f6f-3c2f6b0d8e11"}, nil, false},
		{"ohne Wert", domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikDruck}, []string{"wert"}, true},
		{"erfundenes System", domain.Druckkennung{System: "kartenkiosk", Semantik: domain.SemantikDruck, Wert: "1"}, []string{"system"}, true},
		{"erfundene Semantik", domain.Druckkennung{System: domain.SystemCardmarket, Semantik: "sonstwas", Wert: "1"}, []string{"semantik"}, false},
		// **Eine Scryfall-Druckkennung ist eine UUID** — dieselbe Prüfung wie am Deckplatz.
		{"Scryfall-Druck ohne UUID", domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "bolt"}, []string{"wert"}, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var felder []string
			for _, befund := range fall.kennung.Validate() {
				felder = append(felder, befund.Field)
			}

			if !slices.Equal(felder, fall.befunde) {
				t.Errorf("Befunde %v, erwartet %v", felder, fall.befunde)
			}

			if fall.kennung.Eindeutig() != fall.eindeutig {
				t.Errorf("Eindeutig ist %v, erwartet %v", fall.kennung.Eindeutig(), fall.eindeutig)
			}
		})
	}
}

// **Normalisieren macht aus derselben Kennung dieselbe Zeichenkette** — eine
// UUID in Großbuchstaben ist dieselbe wie in kleinen; eine Händlernummer bleibt.
func TestEineKennungWirdNormalisiert(t *testing.T) {
	t.Parallel()

	uuid := domain.Druckkennung{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "  E3285E6B-3E79-4D7C-BF96-D920F973B122 "}
	uuid.Normalisieren()

	if uuid.Wert != "e3285e6b-3e79-4d7c-bf96-d920f973b122" {
		t.Errorf("die UUID ist %q", uuid.Wert)
	}

	nummer := domain.Druckkennung{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: " 12AB "}
	nummer.Normalisieren()

	if nummer.Wert != "12AB" {
		t.Errorf("die Händlernummer ist %q — sie wird nur getrimmt", nummer.Wert)
	}
}

func TestEinDruckNenntSeineKennungen(t *testing.T) {
	t.Parallel()

	druck := domain.Druck{
		Name: " Lightning Bolt ", Setcode: "LEA", Nummer: "161", Sprache: "en",
		Kennungen: []domain.Druckkennung{
			{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "E3285E6B-3E79-4D7C-BF96-D920F973B122"},
			{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"},
			{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "1"},
			{System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt, Wert: "2"},
		},
	}
	druck.Normalisieren()

	if befunde := druck.Validate(); len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if druck.Name != "Lightning Bolt" || druck.Setcode != "lea" {
		t.Errorf("normalisiert ist %q / %q", druck.Name, druck.Setcode)
	}

	// **Der Scryfall-Alias** — das heutige `printing`, falls es eines gibt.
	if druck.Scryfall() != "e3285e6b-3e79-4d7c-bf96-d920f973b122" {
		t.Errorf("der Alias ist %q", druck.Scryfall())
	}

	// **Mehrere Produktkennungen je Druck sind erlaubt** — CardTrader nennt
	// ein Array von Cardmarket-IDs.
	if produkte := druck.Werte(domain.SystemCardmarket, domain.SemantikProdukt); !slices.Equal(produkte, []string{"1", "2"}) {
		t.Errorf("die Produkte sind %v", produkte)
	}

	if _, gibts := druck.Kennung(domain.SystemMTGJSON, domain.SemantikDruck); gibts {
		t.Error("ein Druck ohne MTGJSON-Kennung nennt eine")
	}
}

func TestEinDruckBrauchtEinenNamenOderEineDruckkennung(t *testing.T) {
	t.Parallel()

	leer := domain.Druck{}
	if befunde := leer.Validate(); len(befunde) != 1 || befunde[0].Field != "name" {
		t.Errorf("ein leerer Druck gibt %v — erwartet genau einen Befund am Namen", befunde)
	}

	nurKennung := domain.Druck{Kennungen: []domain.Druckkennung{
		{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "e3285e6b-3e79-4d7c-bf96-d920f973b122"},
	}}
	if befunde := nurKennung.Validate(); len(befunde) != 0 {
		t.Errorf("ein Druck mit Kennung und ohne Namen gibt %v", befunde)
	}

	// **Ein Befund an einer Kennung nennt ihre Stelle** — `kennungen[1].wert`.
	kaputt := domain.Druck{Name: "Bolt", Kennungen: []domain.Druckkennung{
		{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "e3285e6b-3e79-4d7c-bf96-d920f973b122"},
		{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: ""},
	}}
	befunde := kaputt.Validate()
	if len(befunde) != 1 || befunde[0].Field != "kennungen[1].wert" {
		t.Errorf("Befunde %v — erwartet einer an kennungen[1].wert", befunde)
	}

	// **Zwei eindeutige Kennungen desselben Systems an einem Druck** wären
	// zwei Wahrheiten — ein Befund, kein stilles Nebeneinander.
	doppelt := domain.Druck{Name: "Bolt", Kennungen: []domain.Druckkennung{
		{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "e3285e6b-3e79-4d7c-bf96-d920f973b122"},
		{System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "f3285e6b-3e79-4d7c-bf96-d920f973b122"},
	}}
	befunde = doppelt.Validate()
	if len(befunde) != 1 || !strings.Contains(befunde[0].Message, "zwei") {
		t.Errorf("Befunde %v — erwartet einer über zwei Druckkennungen desselben Systems", befunde)
	}
}

// **Vorläufig ist ein Zustand, kein Fehler**: ein Druck ohne eindeutige
// Kennung ist vorläufig — und sagt es.
func TestEinDruckOhneEindeutigeKennungIstVorlaeufig(t *testing.T) {
	t.Parallel()

	vorlaeufig := domain.Druck{Name: "Eigene Karte", Setcode: "xyz", Nummer: "1"}
	if !vorlaeufig.Vorlaeufig() {
		t.Error("ein Druck ohne Kennung gilt als aufgelöst")
	}

	nurKonzept := domain.Druck{Name: "Bolt", Kennungen: []domain.Druckkennung{
		{System: domain.SystemScryfall, Semantik: domain.SemantikKonzept, Wert: "4457ed35-7c10-48c8-9776-456485fdf070"},
	}}
	if !nurKonzept.Vorlaeufig() {
		t.Error("eine Konzeptkennung allein nennt keinen Druck")
	}

	mtgjson := domain.Druck{Name: "Bolt", Kennungen: []domain.Druckkennung{
		{System: domain.SystemMTGJSON, Semantik: domain.SemantikDruck, Wert: "5f8287b1-5bb6-5f4c-ad17-316a40d5bb0c"},
	}}
	if mtgjson.Vorlaeufig() {
		t.Error("eine MTGJSON-Druckkennung löst auf — auch ohne Scryfall")
	}
}
