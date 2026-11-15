package service_test

// **Ein Deck als Liste ausgeben** — NET-04
// (#1708):
// wer ein Deck teilen will, braucht es als Text. Bis hierher ließ es sich nur
// **einlesen** (`DecklisteVorschau`), nicht ausgeben.
//
// **In der Gestalt, die die Einfuhr versteht**: Hauptdeck und Sideboard unter
// ihren Überschriften, jede Zeile wie `sammelliste.SchreibeListe` sie
// schreibt. Der Rundlauf durch `sammelliste.Lies` ist der Beleg.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinDeckWirdZurListeUndZurueck(t *testing.T) {
	t.Parallel()

	deck := domain.Deck{Name: "Rot", Plaetze: []domain.Deckplatz{
		{Printing: "bolt-m11", Menge: 4, Bereich: domain.BereichHauptdeck, Bezeichnung: "Lightning Bolt"},
		{Printing: "berg-m11", Menge: 20, Bereich: domain.BereichHauptdeck, Bezeichnung: "Mountain"},
		{Printing: "pyro-m11", Menge: 2, Bereich: domain.BereichSideboard, Bezeichnung: "Pyroclasm"},
		{KarteID: 7, Menge: 1, Bereich: domain.BereichSideboard, Bezeichnung: "Eigene Karte"},
	}}

	liste := service.Deckliste(deck)

	erwartet := "Deck\n4 Lightning Bolt\n20 Mountain\n\nSideboard\n2 Pyroclasm\n1 Eigene Karte\n"
	if liste != erwartet {
		t.Fatalf("die Liste ist\n%s\nerwartet\n%s", liste, erwartet)
	}

	zeilen, befunde := sammelliste.Lies(liste)
	if len(befunde) != 0 || len(zeilen) != 4 {
		t.Fatalf("gelesen: %+v, Befunde %+v", zeilen, befunde)
	}

	if zeilen[2].Abschnitt != "sideboard" || !strings.Contains(zeilen[2].Name, "Pyroclasm") {
		t.Errorf("die dritte Zeile ist %+v", zeilen[2])
	}
}

// **Ein leeres Deck ist eine leere Liste**, und ein Deck ohne Sideboard hat
// keine Überschrift dafür.
func TestEinDeckOhneSideboardNenntKeines(t *testing.T) {
	t.Parallel()

	if liste := service.Deckliste(domain.Deck{}); liste != "" {
		t.Errorf("ein leeres Deck gab %q", liste)
	}

	nurHaupt := domain.Deck{Plaetze: []domain.Deckplatz{
		{Printing: "bolt-m11", Menge: 1, Bereich: domain.BereichHauptdeck, Bezeichnung: "Lightning Bolt"},
	}}

	if liste := service.Deckliste(nurHaupt); strings.Contains(liste, "Sideboard") {
		t.Errorf("ohne Sideboard: %q", liste)
	}
}
