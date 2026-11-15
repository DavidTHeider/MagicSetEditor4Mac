package domain_test

// **Die Ordnung einer Sammlung folgt ihrer Sprache** — LOC-03.AK-01
// (#1689).
//
// # Was hier geprüft wird
//
// **Dass die Sprache wirklich wirkt** — und nicht nur mitgereicht wird.
// Für Deutsch und Englisch ist die Ordnung **dieselbe**; beide legen »ä«
// neben »a«. Der Unterschied fängt bei der dritten Sprache an: im
// Schwedischen steht »ä« hinter »z«.
//
// **Ohne diesen Lauf wäre das Feld Zierde** (§10): ein Wert, der
// durchgereicht wird und nichts bewirkt, sieht aus wie eine Zusage.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dreiBezeichnungen sind die Zeilen, an denen sich Sprachen unterscheiden.
func dreiBezeichnungen() []*domain.Exemplar {
	return []*domain.Exemplar{
		{Printing: "1", Menge: 1, Name: "Zebra", Setcode: "zen", Nummer: "1"},
		{Printing: "2", Menge: 1, Name: "Ärger", Setcode: "zen", Nummer: "2"},
		{Printing: "3", Menge: 1, Name: "Apfel", Setcode: "zen", Nummer: "3"},
	}
}

// **Deutsch und Englisch ordnen gleich — Schwedisch nicht.**
//
// **Der Nachbar daneben** (`sammlungssicht_test.go`) prüft, **dass** nach
// Sprache und nicht nach Bytes geordnet wird; hier steht, **welche**
// Sprache es ist.
func TestDieOrdnungFolgtDerGewaehltenSprache(t *testing.T) {
	t.Parallel()

	faelle := map[string][]string{
		// **»Ä« neben »A«** — so ordnen beide.
		"":   {"Apfel", "Ärger", "Zebra"},
		"de": {"Apfel", "Ärger", "Zebra"},
		"en": {"Apfel", "Ärger", "Zebra"},

		// **Und im Schwedischen hinter »Z«.**
		"sv": {"Apfel", "Zebra", "Ärger"},

		// **Ein Kürzel, das es nicht gibt, ordnet wie Deutsch** — geraten
		// wird nicht.
		"klingonisch": {"Apfel", "Ärger", "Zebra"},
	}

	for sprache, erwartet := range faelle {
		t.Run("Sprache "+sprache, func(t *testing.T) {
			t.Parallel()

			pruefeOrdnung(t, sprache, dreiBezeichnungen(), erwartet)
		})
	}
}

// **Und Kanji ordnet jede Sprache anders** — LOC-03.AK-01 verlangt eine
// CJK-Locale.
//
// **Gemessen am 2026-09-22 mit `x/text/collate`**: Deutsch und Englisch
// legen 光 vor 稲妻, Japanisch umgekehrt — dieselbe Reihenfolge, die `Intl`
// im Fenster gibt. Die Namen sind die japanischen von vier Karten, wie sie
// in einer Sammlung stehen.
func TestKanjiOrdnenSichNachDerSprache(t *testing.T) {
	t.Parallel()

	faelle := map[string][]string{
		"de-DE": {"光", "嵐", "稲妻", "電撃"},
		"en-US": {"光", "嵐", "稲妻", "電撃"},
		"ja-JP": {"稲妻", "光", "電撃", "嵐"},
	}

	for sprache, erwartet := range faelle {
		t.Run("Sprache "+sprache, func(t *testing.T) {
			t.Parallel()

			pruefeOrdnung(t, sprache, []*domain.Exemplar{
				{Printing: "1", Menge: 1, Name: "電撃", Setcode: "m11", Nummer: "1"},
				{Printing: "2", Menge: 1, Name: "嵐", Setcode: "m11", Nummer: "2"},
				{Printing: "3", Menge: 1, Name: "稲妻", Setcode: "m11", Nummer: "3"},
				{Printing: "4", Menge: 1, Name: "光", Setcode: "m11", Nummer: "4"},
			}, erwartet)
		})
	}
}

// pruefeOrdnung ordnet nach Bezeichnung in einer Sprache und vergleicht.
func pruefeOrdnung(
	t *testing.T, sprache string, exemplare []*domain.Exemplar, erwartet []string,
) {
	t.Helper()

	sicht := domain.Sammlungssicht{
		Sortierung: domain.SortierungBezeichnung,
		Sprache:    sprache,
	}

	gefiltert := sicht.Gesehen(exemplare)

	gelesen := make([]string, 0, len(gefiltert))
	for _, eines := range gefiltert {
		gelesen = append(gelesen, eines.Name)
	}

	for stelle, name := range erwartet {
		if stelle >= len(gelesen) || gelesen[stelle] != name {
			t.Fatalf("%q ordnete %v, erwartet %v", sprache, gelesen, erwartet)
		}
	}
}
