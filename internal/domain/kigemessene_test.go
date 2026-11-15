package domain_test

// **Ein gemessenes Modell, das der Anbieter nur nennt** —
// #1541.
//
// # Was hier geprüft wird
//
// **Die Grenze, die [domain.MitGemessenen] zieht.** Aufgenommen wird ein
// Alias, den der Anbieter in **dieser** Antwort nennt und für den eine
// gültige Messung vorliegt — sonst nichts. „Kein Katalogeintrag ohne
// Anbieterantwort" (0097) bleibt damit stehen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// messungFuer baut eine gültige Messung dieses Anbieters.
//
// **Neben [eineProbe] und nicht darin**: jene nimmt Änderungen als
// Funktionen, hier sind es immer dieselben zwei Felder.
func messungFuer(anbieter, alias string) domain.KIModellprobe {
	return domain.KIModellprobe{
		Anbieter: anbieter, Alias: alias,
		Faehigkeit: domain.KITextgenerierung,
		Datum:      "2026-09-13",
		Beleg:      "gemessen an 2026-09-13: schemakonform zurück",
		Laufzeit:   "Responses-API /v1/responses",
		Modell:     alias,
	}
}

// **Ein genanntes Modell mit Messung kommt dazu.**
func TestEinGenanntesModellMitMessungKommtDazu(t *testing.T) {
	t.Parallel()

	ergaenzt := domain.MitGemessenen(nil,
		[]domain.KIModellprobe{messungFuer("openai", "gpt-probe")},
		[]string{"gpt-probe"}, "openai", domain.KIOrtCloud)

	if len(ergaenzt) != 1 {
		t.Fatalf("es kamen %d Modelle heraus: %+v", len(ergaenzt), ergaenzt)
	}

	modell := ergaenzt[0]
	if modell.Anbieter != "openai" || modell.Alias != "gpt-probe" ||
		modell.Ort != domain.KIOrtCloud ||
		modell.Laufzeit != "Responses-API /v1/responses" ||
		len(modell.Geprueft) != 1 {
		t.Fatalf("das ergänzte Modell ist %+v", modell)
	}
}

// **Ohne Messung kommt nichts dazu.**
//
// **Das ist 0097**: ein Katalogeintrag ohne Beleg wäre eine Behauptung.
func TestOhneMessungKommtNichtsDazu(t *testing.T) {
	t.Parallel()

	if ergaenzt := domain.MitGemessenen(nil, nil,
		[]string{"gpt-probe"}, "openai", domain.KIOrtCloud); len(ergaenzt) != 0 {
		t.Fatalf("ohne Messung kamen %+v heraus", ergaenzt)
	}
}

// **Und eine Messung ohne genanntes Modell ebenfalls nicht.**
//
// **Auch das ist 0097**, von der anderen Seite: wer ein Modell abschaltet,
// verschwindet aus dem Katalog, obwohl die Messung im Bestand liegt.
func TestEineMessungOhneGenanntesModellKommtNichtDazu(t *testing.T) {
	t.Parallel()

	ergaenzt := domain.MitGemessenen(nil,
		[]domain.KIModellprobe{messungFuer("openai", "gpt-weg")},
		[]string{"gpt-da"}, "openai", domain.KIOrtCloud)

	if len(ergaenzt) != 0 {
		t.Fatalf("das abgeschaltete Modell kam durch: %+v", ergaenzt)
	}
}

// **Ein Modell, das schon dasteht, kommt nicht zweimal.**
func TestEinVorhandenesModellKommtNichtZweimal(t *testing.T) {
	t.Parallel()

	vorhanden := []domain.KIModell{{
		Anbieter: "openai", Alias: "gpt-probe", Ort: domain.KIOrtCloud,
	}}

	ergaenzt := domain.MitGemessenen(vorhanden,
		[]domain.KIModellprobe{messungFuer("openai", "gpt-probe")},
		[]string{"gpt-probe", "gpt-probe"}, "openai", domain.KIOrtCloud)

	if len(ergaenzt) != 1 {
		t.Fatalf("es kamen %d Modelle heraus: %+v", len(ergaenzt), ergaenzt)
	}
}

// **Die Messung eines anderen Anbieters zählt nicht.**
//
// **Zwei Anbieter dürfen denselben Namen führen** — deshalb steht der
// Anbieter seit A8 überall daneben.
func TestDieMessungEinesAnderenAnbietersZaehltNicht(t *testing.T) {
	t.Parallel()

	ergaenzt := domain.MitGemessenen(nil,
		[]domain.KIModellprobe{messungFuer("xai", "gpt-probe")},
		[]string{"gpt-probe"}, "openai", domain.KIOrtCloud)

	if len(ergaenzt) != 0 {
		t.Fatalf("die fremde Messung kam durch: %+v", ergaenzt)
	}
}

// **Eine ungültige Messung trägt nichts.**
func TestEineUngueltigeMessungTraegtNichts(t *testing.T) {
	t.Parallel()

	kaputt := messungFuer("openai", "gpt-probe")
	kaputt.Datum = "vorgestern"

	if ergaenzt := domain.MitGemessenen(nil,
		[]domain.KIModellprobe{kaputt}, []string{"gpt-probe"},
		"openai", domain.KIOrtCloud); len(ergaenzt) != 0 {
		t.Fatalf("die ungültige Messung kam durch: %+v", ergaenzt)
	}
}

// **Und eine Messung ohne Laufzeitangabe ist keine gültige.**
//
// **Deshalb steht am ergänzten Modell nie eine leere Laufzeit**:
// [domain.KIModellprobe.Validate] verlangt sie, und was sie abweist, trägt
// nichts. Ein Zweig „keine Laufzeit gefunden" wäre unbetretbar — und eine
// Zusage, die niemand prüft (§10).
func TestEineMessungOhneLaufzeitTraegtNichts(t *testing.T) {
	t.Parallel()

	ohne := messungFuer("openai", "gpt-probe")
	ohne.Laufzeit = ""

	if ergaenzt := domain.MitGemessenen(nil, []domain.KIModellprobe{ohne},
		[]string{"gpt-probe"}, "openai", domain.KIOrtCloud); len(ergaenzt) != 0 {
		t.Fatalf("die Messung ohne Laufzeit kam durch: %+v", ergaenzt)
	}
}
