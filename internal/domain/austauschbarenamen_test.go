// **Austauschbare Namen zählen als derselbe Name** —
// #1375.
//
// # Die Regel
//
// CR 100.2a endet mit einem Satz, den die Höchstzahl einlösen muss:
//
//	For the purposes of deck construction, cards with interchangeable names
//	have the same English name (see rule 201.3).
//
// Und CR 201.3b: für Deckbau und Formatzulässigkeit gelten Karten mit
// austauschbaren Namen als dieselbe Karte. **Nachgelesen** in der
// offiziellen Fassung vom 2026-08-19 (`MagicCompRules 20260819.txt`).
//
// Gemeint sind die Paare aus „Universes Beyond" und „Universes Within":
// dieselbe Karte, zwei englische Namen.
//
// # Warum es jetzt geht
//
// **Weil Scryfall sie doch verbindet — über `oracle_id`.** Das Issue schloss
// das aus, und die Messung dahinter traf die falsche Karte: „Aisha of Sparks
// and Smoke" heißt in `sld` **und** `slx` gleich und ist gar kein Paar. Am
// 2026-09-04 gegen `api.scryfall.com` nachgemessen:
//
//	Mind Flayer, the Shadow  ↔ Arvinox, the Mind Flail        9abecd23…
//	Eleven, the Mage         ↔ Cecily, Haunted Mage           a726c27d…
//	Dustin, Gadget Genius    ↔ Hargilde, Kindly Runechanter   747009e7…
//
// Drei Paare, drei verschiedene Namen je Paar, **eine** `oracle_id`. Damit
// braucht dieses Programm keine eingebaute Liste, die mit dem nächsten Secret
// Lair veraltete (§15).
package domain_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Vier von einem und vier vom anderen Namen sind acht derselben Karte.**
//
// Das ist die Abnahme aus dem Issue, und sie steht auf einem echten Paar aus
// der Messung — nicht auf erfundenen Namen, die auch dann durchgingen, wenn
// die Gruppierung etwas anderes täte.
func TestAustauschbareNamenZaehlenAlsDerselbeName(t *testing.T) {
	posten := []domain.Deckposten{
		{
			Platz: domain.Deckplatz{
				Printing: "sld-1", Menge: 4,
				Bereich: domain.BereichHauptdeck,
			},
			Name:          "Mind Flayer, the Shadow",
			Orakelkennung: "9abecd23-a2d9-43c7-8d01-f8afdc53026e",
		},
		{
			Platz: domain.Deckplatz{
				Printing: "slx-1", Menge: 4,
				Bereich: domain.BereichHauptdeck,
			},
			Name:          "Arvinox, the Mind Flail",
			Orakelkennung: "9abecd23-a2d9-43c7-8d01-f8afdc53026e",
		},
	}

	befunde := domain.PruefeGroessen(posten, domain.Format{HoechstzahlGleich: 4})

	gleiche := 0

	for _, befund := range befunde {
		if befund.Feld == "gleiche" {
			gleiche++

			if !strings.Contains(befund.Meldung, "8") {
				t.Errorf("die Meldung nennt nicht acht: %q", befund.Meldung)
			}

			// **Beide Namen, nicht der erste.** Acht Karten liegen hier unter
			// zwei Namen; wer nur einen liest, sucht acht davon und findet
			// vier.
			for _, name := range []string{"Mind Flayer", "Arvinox"} {
				if !strings.Contains(befund.Meldung, name) {
					t.Errorf("die Meldung nennt %q nicht: %q", name, befund.Meldung)
				}
			}
		}
	}

	if gleiche != 1 {
		t.Errorf("%d Befunde zur Höchstzahl, erwartet war einer: %v", gleiche, befunde)
	}
}

// **Zwei wirklich verschiedene Karten bleiben zwei.**
//
// Die Gegenprobe zum Test darüber: ohne sie wäre eine Gruppierung, die
// **alles** zusammenwirft, ebenso grün — und ein Deck mit vier „Lightning
// Bolt" und vier „Shock" wäre plötzlich unzulässig.
func TestVerschiedeneOrakelkennungenBleibenVerschieden(t *testing.T) {
	posten := []domain.Deckposten{
		{
			Platz:         domain.Deckplatz{Printing: "a", Menge: 4, Bereich: domain.BereichHauptdeck},
			Name:          "Lightning Bolt",
			Orakelkennung: "4457ed35-7c10-48c8-9776-456485fdf070",
		},
		{
			Platz:         domain.Deckplatz{Printing: "b", Menge: 4, Bereich: domain.BereichHauptdeck},
			Name:          "Shock",
			Orakelkennung: "3ba00fbf-a1cc-4a3f-b0e9-4b12b4c2c2b0",
		},
	}

	if befunde := domain.PruefeGroessen(
		posten, domain.Format{HoechstzahlGleich: 4}); len(befunde) != 0 {
		t.Errorf("%v", befunde)
	}
}

// **Ohne Orakelkennung entscheidet weiter der Name.**
//
// Ein Deck aus eigenen Entwürfen fragt Scryfall gar nicht; dort gibt es keine
// Kennung, und vier gleichnamige Karten müssen trotzdem vier sein. Ohne diesen
// Fall hätte die Umstellung die Höchstzahl für jedes Set ohne Netz abgeschaltet
// — ein falsches Ja an genau der Stelle, die dieses Issue schließen soll.
func TestOhneOrakelkennungEntscheidetWeiterDerName(t *testing.T) {
	posten := []domain.Deckposten{
		{
			Platz: domain.Deckplatz{KarteID: 1, Menge: 3, Bereich: domain.BereichHauptdeck},
			Name:  "Feuerball",
		},
		{
			Platz: domain.Deckplatz{KarteID: 2, Menge: 2, Bereich: domain.BereichHauptdeck},
			Name:  "Feuerball",
		},
	}

	if befunde := domain.PruefeGroessen(
		posten, domain.Format{HoechstzahlGleich: 4}); len(befunde) != 1 {
		t.Errorf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}
}
