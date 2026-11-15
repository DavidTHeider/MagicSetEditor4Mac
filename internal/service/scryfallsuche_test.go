// **Die Kartensuche bei Scryfall** — Werkstück 1 von Phase 3.6
// (#1167).
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// suchquelle gibt feste Treffer und merkt sich die Frage.
type suchquelle struct {
	gefragt string
	treffer scryfall.Treffer
	alter   time.Duration
	befunde []domain.Befund
	fehler  error
}

func (s *suchquelle) Suche(_ context.Context, frage string) (
	scryfall.Treffer, time.Duration, []domain.Befund, error,
) {
	s.gefragt = frage

	return s.treffer, s.alter, s.befunde, s.fehler
}

// **Die Frage geht unverändert hinaus.**
//
// # Warum das eine Zusage ist und keine Nachlässigkeit
//
// Scryfalls Suchsyntax ist die des Suchfelds seiner Website — `t:creature c:r
// cmc<=3`. Sie hier nachzubauen wäre eine zweite Sprache, die niemand kennt
// und die bei jeder Erweiterung auseinanderliefe. Wer sie kennt, soll sie
// benutzen können; wer nicht, tippt einen Namen, und der ist eine gültige
// Frage.
func TestDieFrageGehtUnveraendertHinaus(t *testing.T) {
	t.Parallel()

	quelle := &suchquelle{}

	if _, err := service.NeueSuche(quelle).
		Suche(t.Context(), "t:creature c:r cmc<=3"); err != nil {
		t.Fatalf("Suche: %v", err)
	}

	if quelle.gefragt != "t:creature c:r cmc<=3" {
		t.Errorf("gefragt wurde %q", quelle.gefragt)
	}
}

// **Ein Treffer trägt, was zum Wählen reicht.**
func TestEinTrefferTraegtWasZumWaehlenReicht(t *testing.T) {
	t.Parallel()

	quelle := &suchquelle{
		alter: 3 * time.Hour,
		treffer: scryfall.Treffer{
			Karten: []*scryfall.Karte{{
				Kennung: "abc", Name: "Lightning Bolt",
				Set: "lea", Setname: "Limited Edition Alpha", Nummer: "161",
				Seltenheit: "common", Manakosten: "{R}", Typzeile: "Instant",
				Bilder: scryfall.Bilder{Klein: "https://bilder/klein.jpg"},
			}},
			Weitere: true,
		},
	}

	auskunft, err := service.NeueSuche(quelle).Suche(t.Context(), "bolt")
	if err != nil {
		t.Fatalf("Suche: %v", err)
	}

	erwartet := service.Suchtreffer{
		Kennung: "abc", Name: "Lightning Bolt",
		Setcode: "lea", Setname: "Limited Edition Alpha", Nummer: "161",
		Seltenheit: "common", Manakosten: "{R}", Typzeile: "Instant",
		Bild: "https://bilder/klein.jpg",
	}

	if len(auskunft.Treffer) != 1 || auskunft.Treffer[0] != erwartet {
		t.Errorf("der Treffer ist %+v", auskunft.Treffer)
	}

	// **Und das Alter reist mit** (§16): eine Auskunft ohne ihr Alter ist eine
	// Behauptung über den heutigen Stand.
	if auskunft.Alter != 3*time.Hour {
		t.Errorf("das Alter ist %v", auskunft.Alter)
	}

	if !auskunft.Weitere {
		t.Error("»es gibt mehr« ging verloren")
	}
}

// **Eine zweiseitige Karte bekommt das Bild ihrer Vorderseite.**
//
// `image_uris` steht bei einer DFC **nicht** auf der obersten Ebene, sondern
// an jeder Seite (`card_faces`). Wer nur oben nachsähe, ließe jede zweiseitige
// Karte ohne Bild — und das sind bei modernen Sätzen viele.
func TestEineZweiseitigeKarteBekommtIhrVorderbild(t *testing.T) {
	t.Parallel()

	quelle := &suchquelle{treffer: scryfall.Treffer{
		Karten: []*scryfall.Karte{{
			Kennung: "dfc", Name: "Delver of Secrets // Insectile Aberration",
			Seiten: []scryfall.Seite{
				{Name: "Delver of Secrets", Bilder: scryfall.Bilder{Klein: "vorn.jpg"}},
				{Name: "Insectile Aberration", Bilder: scryfall.Bilder{Klein: "hinten.jpg"}},
			},
		}},
	}}

	auskunft, err := service.NeueSuche(quelle).Suche(t.Context(), "delver")
	if err != nil {
		t.Fatalf("Suche: %v", err)
	}

	if auskunft.Treffer[0].Bild != "vorn.jpg" {
		t.Errorf("das Bild ist %q", auskunft.Treffer[0].Bild)
	}
}

// **Eine Karte ganz ohne Bild bekommt keines — und keinen Fehler.**
func TestEineKarteOhneBildBekommtKeines(t *testing.T) {
	t.Parallel()

	quelle := &suchquelle{treffer: scryfall.Treffer{
		Karten: []*scryfall.Karte{{Kennung: "leer", Name: "Ohne Bild"}},
	}}

	auskunft, err := service.NeueSuche(quelle).Suche(t.Context(), "ohne")
	if err != nil {
		t.Fatalf("Suche: %v", err)
	}

	if auskunft.Treffer[0].Bild != "" {
		t.Errorf("es kam ein Bild: %q", auskunft.Treffer[0].Bild)
	}
}

// **Eine leere Frage geht gar nicht erst hinaus.**
//
// Scryfall antwortete darauf mit einem Fehler; ihn erst hinausgehen zu lassen
// wäre eine Anfrage, deren Antwort feststeht — und sie zählte gegen die
// Grenze, die §16 einzuhalten verlangt.
func TestEineLeereFrageGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	quelle := &suchquelle{}

	_, err := service.NeueSuche(quelle).Suche(t.Context(), "   ")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v — will ErrValidation", err)
	}

	if quelle.gefragt != "" {
		t.Errorf("es ging etwas hinaus: %q", quelle.gefragt)
	}
}

// **Ein Fehler der Quelle kommt durch.**
func TestEinFehlerDerSuchquelleKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("Scryfall antwortet nicht")

	_, err := service.NeueSuche(&suchquelle{fehler: gestolpert}).
		Suche(t.Context(), "bolt")

	if !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Und die Befunde reisen mit** (§8).
func TestDieBefundeDerSucheReisenMit(t *testing.T) {
	t.Parallel()

	quelle := &suchquelle{
		treffer: scryfall.Treffer{Karten: []*scryfall.Karte{{Kennung: "a"}}},
		befunde: []domain.Befund{{Ort: "karte", Meldung: "ein Feld war nicht zu deuten"}},
	}

	auskunft, err := service.NeueSuche(quelle).Suche(t.Context(), "a")
	if err != nil {
		t.Fatalf("Suche: %v", err)
	}

	if len(auskunft.Befunde) != 1 {
		t.Errorf("die Befunde sind %+v", auskunft.Befunde)
	}
}
