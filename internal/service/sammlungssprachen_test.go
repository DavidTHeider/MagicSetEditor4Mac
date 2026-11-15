package service_test

// **Ein CSV für jede Sprache** — LOC-03.AK-01
// (#1689):
// „de-DE, en-US sowie eine CJK-Locale besitzen Roundtriptests" — hier für
// das CSV-Trennzeichen und für Geld.
//
// # Was hier geprüft wird
//
// **Dass das Trennzeichen keine Sprache hat.** Ein Komma, immer
// (`docs/LOKALISIERUNG.md`) — und deshalb muss alles, was eine der drei
// Sprachen in eine Notiz schreibt, heil zurückkommen: ein deutsches
// Anführungszeichen, ein englisches in der Zeile, ein **japanisches
// Komma** (U+3001) und ein **breites** (U+FF0C). Die beiden letzten sehen
// aus wie ein Trennzeichen und sind keins.
//
// **Und dass jeder Betrag in den Stellen seiner Währung zurückkommt** — Yen
// ohne, Dinar mit drei (AK-02).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dreisprachig sind die Zeilen, die jede Sprache in ein CSV schriebe.
func dreisprachig() []*domain.Exemplar {
	return []*domain.Exemplar{
		{
			Printing: "33333333-3333-4333-8333-333333333333", Menge: 1,
			Zustand: domain.ZustandNearMint, Sprache: "de",
			Name: "Blitzschlag", Setcode: "m11", Nummer: "149",
			Erwerbspreis: 1999, Erwerbswaehrung: "EUR",
			Notiz: "getauscht, „fast neu“",
		},
		{
			Printing: "44444444-4444-4444-8444-444444444444", Menge: 2,
			Zustand: domain.ZustandNearMint, Sprache: "en",
			Name: "Counterspell", Setcode: "lea", Nummer: "54",
			Erwerbspreis: 1234, Erwerbswaehrung: "BHD",
			Notiz: `traded, "mint"`,
		},
		{
			Printing: "55555555-5555-4555-8555-555555555555", Menge: 3,
			Zustand: domain.ZustandNearMint, Sprache: "ja",
			Name: "稲妻", Setcode: "m11", Nummer: "149",
			Erwerbspreis: 150, Erwerbswaehrung: "JPY",
			Notiz: "交換済み、2026年，美品",
		},
	}
}

// **Was drei Sprachen hineinschreiben, kommt heraus.**
func TestDasCSVTraegtDreiSprachen(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	for _, exemplar := range dreisprachig() {
		if _, fehler := ablage.Anlegen(t.Context(), exemplar); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	text, fehler := service.NeueSammlungsausfuhr(ablage).
		Ausgeben(t.Context(), service.FormCSV)
	if fehler != nil {
		t.Fatalf("Ausgeben: %v", fehler)
	}

	leer := &sammlungsdoppel{}

	bericht, fehler := service.NeueSammlungsausfuhr(leer).
		Einfuegen(t.Context(), text, "")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(bericht.Angelegt) != 3 || len(bericht.Befunde) != 0 {
		t.Fatalf("%d angelegt, Befunde %+v — erwartet 3 und keine:\n%s",
			len(bericht.Angelegt), bericht.Befunde, text)
	}

	zurueck, lesefehler := leer.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	erwartet := map[string]*domain.Exemplar{}
	for _, exemplar := range dreisprachig() {
		erwartet[exemplar.Printing] = exemplar
	}

	for _, exemplar := range zurueck {
		soll := erwartet[exemplar.Printing]
		if soll == nil {
			t.Fatalf("unbekannter Druck %q zurückgelesen", exemplar.Printing)
		}

		if exemplar.Name != soll.Name || exemplar.Notiz != soll.Notiz {
			t.Errorf("%s kam als %q mit %q zurück", soll.Name,
				exemplar.Name, exemplar.Notiz)
		}

		if exemplar.Erwerbspreis != soll.Erwerbspreis ||
			exemplar.Erwerbswaehrung != soll.Erwerbswaehrung {
			t.Errorf("%s kostete %d %s und kam als %d %s zurück", soll.Name,
				soll.Erwerbspreis, soll.Erwerbswaehrung,
				exemplar.Erwerbspreis, exemplar.Erwerbswaehrung)
		}
	}
}
