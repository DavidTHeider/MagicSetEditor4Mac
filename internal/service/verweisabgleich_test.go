package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Abgleich folgt dem Verweis** — Werkstück 2 von Phase 3.6
// (#1167).
//
// # Warum das die eigentliche Zusage ist
//
// Ein Verweis, dem niemand folgt, ist eine Beschriftung. Erst hier wird er
// wirksam: der Abgleich holt **den** Druck, den jemand gewählt hat, statt
// eines, den `/cards/named` aussucht.

// **Mit Verweis wird der Druck geholt, nicht der Name.**
func TestDerAbgleichFolgtDemVerweis(t *testing.T) {
	t.Parallel()

	drucke := &druckdoppel{karte: probedruck()}
	ablage := neuesVerweisdoppel()
	ablage.verweise[7] = domain.Kartenverweis{
		Printing: "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1", Name: "Lightning Bolt",
	}

	abgleich := service.NeuerScryfallabgleich(fremdkarte(), drucke, ablage, nil)

	vorschlag, fehler := abgleich.VorschlagFuerKarte(t.Context(), "magic",
		map[string]string{}, 7, "ein ganz anderer Name")
	if fehler != nil {
		t.Fatalf("VorschlagFuerKarte: %v", fehler)
	}

	if drucke.gefragt != "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1" {
		t.Errorf("gefragt wurde nach %q", drucke.gefragt)
	}

	// **Und der Vorschlag sagt, aus welchem Druck er stammt.** Ohne diese
	// Angabe wäre nicht zu sehen, dass der Name übergangen wurde — und
	// „still" ist genau das, was §16 Regel 6 verbietet.
	if vorschlag.Druck != "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1" {
		t.Errorf("der Vorschlag nennt den Druck %q", vorschlag.Druck)
	}

	if vorschlag.Herkunft != probedruck().Adresse {
		t.Errorf("Herkunft %q", vorschlag.Herkunft)
	}
}

// **Ohne Verweis gilt der Name.** Der gewohnte Weg bleibt der gewohnte —
// verträglich im Sinne von 0033.
func TestOhneVerweisGiltDerName(t *testing.T) {
	t.Parallel()

	drucke := &druckdoppel{karte: probedruck()}

	abgleich := service.NeuerScryfallabgleich(
		fremdkarte(), drucke, neuesVerweisdoppel(), nil)

	vorschlag, fehler := abgleich.VorschlagFuerKarte(t.Context(), "magic",
		map[string]string{}, 7, "Prüfstein")
	if fehler != nil {
		t.Fatalf("VorschlagFuerKarte: %v", fehler)
	}

	if drucke.gefragt != "" {
		t.Errorf("es wurde nach dem Druck %q gefragt", drucke.gefragt)
	}

	if vorschlag.Herkunft != "https://beispiel.invalid/karte/1" {
		t.Errorf("Herkunft %q", vorschlag.Herkunft)
	}
}

// **Ein unlesbarer Verweisspeicher bricht den Abgleich ab.**
//
// # Warum nicht auf den Namen zurückfallen
//
// Weil ein Rückfall hier **still** die falsche Karte holte: der Verweis ist
// vielleicht da, nur nicht lesbar. Wer dann `/cards/named` fragte, bekäme
// irgendeinen Druck und schriebe ihn unter der Herkunft „gewählt" fort.
func TestEinUnlesbarerVerweisBrichtDenAbgleichAb(t *testing.T) {
	t.Parallel()

	ablage := neuesVerweisdoppel()
	ablage.fehler = domain.ErrBelegt

	abgleich := service.NeuerScryfallabgleich(
		fremdkarte(), &druckdoppel{karte: probedruck()}, ablage, nil)

	_, fehler := abgleich.VorschlagFuerKarte(t.Context(), "magic",
		map[string]string{}, 7, "Prüfstein")
	if fehler == nil {
		t.Fatal("der unlesbare Speicher kam als Erfolg zurück")
	}
}
