package app

// **Ein Deck durch den Umschlag und zurück** — FMT-04.AK-04
// (#1693):
// Karten aus **mehreren Sets**, **mehrfach** vorhandene Karten (derselbe
// Druck in Haupt- und Nebendeck) und was dabei nicht mitkommt.
//
// **Was mitkommt**: jeder Druck mit Menge und Bereich, auch zweimal. **Was
// nicht**: eigene Karten — ihre Kennung gilt im Bestand des Absenders
// (`docs/DECKUNDSAMMLUNG.md`). Das Deck steht trotzdem da, und die Meldung
// nennt beide.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinDeckAusZweiSetsUndDoppeltenDruckenGehtDurchDenUmschlag(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	const druck = "a1b2c3d4-0000-4000-8000-00000000000a"

	deck := anwendung.DeckAnlegen(ctx, "Rundlauf", "probe", "")
	if deck.Fehler != nil {
		t.Fatalf("Deck: %+v", deck.Fehler)
	}

	for _, setname := range []string{"Erstes Set", "Zweites Set"} {
		set := anwendung.SetAnlegen(ctx, setname, "probe", "probe-schlicht")
		karte := anwendung.KarteAnlegen(ctx, set.ID)

		if gelegt := anwendung.DeckplatzLegen(ctx, deck.Deck.ID, karte.KarteID, "", 1, "hauptdeck"); gelegt.Fehler != nil {
			t.Fatalf("eigene Karte aus %s: %+v", setname, gelegt.Fehler)
		}
	}

	for bereich, menge := range map[string]int{"hauptdeck": 4, "sideboard": 2} {
		if gelegt := anwendung.DeckplatzLegen(ctx, deck.Deck.ID, "", druck, menge, bereich); gelegt.Fehler != nil {
			t.Fatalf("Druck im %s: %+v", bereich, gelegt.Fehler)
		}
	}

	ziel := einUmschlagsziel(t, "rundlauf")
	if geschnuert := anwendung.UmschlagSchnueren(ctx, "deck", deck.Deck.ID, ziel); geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	angenommen := anwendung.UmschlagDeck(ctx, ziel)
	if angenommen.Deck == nil {
		t.Fatalf("kein Deck: %+v", angenommen.Fehler)
	}

	drucke := map[string]int{}

	for _, platz := range angenommen.Deck.Plaetze {
		if platz.KarteID != "" {
			t.Errorf("eine eigene Karte des Absenders liegt im Deck: %+v", platz)
		}

		drucke[platz.Bereich] += platz.Menge
	}

	if drucke["hauptdeck"] != 4 || drucke["sideboard"] != 2 {
		t.Errorf("die Drucke kamen als %v an, erwartet Hauptdeck 4 und Sideboard 2", drucke)
	}

	if angenommen.Fehler == nil || angenommen.Fehler.Schluessel != domain.SchluesselFremdeEigeneKarten {
		t.Fatalf("die Meldung fehlt: %+v", angenommen.Fehler)
	}

	// **Eine Karte ohne Namen heißt nach ihrer Kennung** — sonst stünde in
	// der Meldung ein leerer Eintrag.
	if karten := angenommen.Fehler.Werte["karten"]; !strings.HasPrefix(karten, "#") || strings.Contains(karten, ", ,") {
		t.Errorf("die Meldung nennt %q", karten)
	}
}
