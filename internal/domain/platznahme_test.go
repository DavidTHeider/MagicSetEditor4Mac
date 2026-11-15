package domain_test

// **Der Gast setzt sich mit seinem Deck** — P9b
// (#2377).
// Er bringt genau eines mit: einen eigenen Abzug oder die Kennung einer
// Leihe des Wirts. Gesetzt wird nur, wer eingeladen ist, nur an seinen Sitz
// und nur einmal.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestMitgebrachtIstGenauEines(t *testing.T) {
	t.Parallel()

	deck := einDeckabzug()

	faelle := map[string]struct {
		mit  domain.Mitgebracht
		gilt bool
	}{
		"ein Deck":            {domain.Mitgebracht{Deck: &deck, Spiel: "magic"}, true},
		"ein Deck ohne Spiel": {domain.Mitgebracht{Deck: &deck}, false},
		"eine Leihe":          {domain.Mitgebracht{Leihe: "l1"}, true},
		"beides":              {domain.Mitgebracht{Deck: &deck, Leihe: "l1", Spiel: "magic"}, false},
		"nichts":              {domain.Mitgebracht{}, false},
		"ein leeres Deck":     {domain.Mitgebracht{Deck: &domain.Partiedeck{}, Spiel: "magic"}, false},
		"überlange Leihe":     {domain.Mitgebracht{Leihe: strings.Repeat("l", 200)}, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.mit.Pruefe()
			if (fehler == nil) != fall.gilt {
				t.Errorf("gilt %v, Prüfung sagt %v", fall.gilt, fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("kein Validierungsfehler: %v", fehler)
			}
		})
	}
}

func eineBibliothek(sitz string) []domain.Spielobjekt {
	return []domain.Spielobjekt{{
		Kennung: sitz + "/k1/0", Art: domain.ObjektKarte, Beherrscher: sitz,
		Eigner: sitz, Zone: domain.ZoneBibliothek, Karte: "k1", Ordnung: 1,
	}}
}

func TestDerEingeladeneSetztSichMitSeinemDeck(t *testing.T) {
	t.Parallel()

	deck := einDeckabzug()
	deck.Quelle = "peer-bo"

	danach, fehler := eineEinladung().Platznehmen("sitz-2", "peer-bo", deck, eineBibliothek("sitz-2"))
	if fehler != nil {
		t.Fatalf("Platznehmen: %v", fehler)
	}

	sitz := danach.SitzVon("sitz-2")
	if sitz.Wartet() || sitz.Deck.Abdruck != "abcd" || sitz.Deck.Quelle != "peer-bo" {
		t.Errorf("der Sitz: %+v", sitz)
	}

	if len(danach.Objekte) != 1 || danach.Objekte[0].Eigner != "sitz-2" {
		t.Errorf("die Bibliothek: %+v", danach.Objekte)
	}

	if fehler := danach.Pruefe(); fehler != nil {
		t.Errorf("danach keine gültige Partie: %v", fehler)
	}
}

func TestWerSichNichtSetzenDarf(t *testing.T) {
	t.Parallel()

	gesetzt, _ := eineEinladung().Platznehmen("sitz-2", "peer-bo", einDeckabzug(), nil) //nolint:errcheck // geprüft im Lauf darüber

	faelle := map[string]struct {
		partie     domain.Partie
		sitz, peer string
		deck       domain.Partiedeck
		art        error
	}{
		"ein anderer Peer":      {eineEinladung(), "sitz-2", "peer-cora", einDeckabzug(), domain.ErrConflict},
		"an einen fremden Sitz": {eineEinladung(), "sitz-1", "peer-bo", einDeckabzug(), domain.ErrConflict},
		"an keinen Sitz":        {eineEinladung(), "sitz-9", "peer-bo", einDeckabzug(), domain.ErrValidation},
		"ein zweites Mal":       {gesetzt, "sitz-2", "peer-bo", einDeckabzug(), domain.ErrConflict},
		"mit einem leeren Deck": {eineEinladung(), "sitz-2", "peer-bo", domain.Partiedeck{}, domain.ErrValidation},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorher := len(fall.partie.Objekte)

			danach, fehler := fall.partie.Platznehmen(fall.sitz, fall.peer, fall.deck, eineBibliothek(fall.sitz))
			if !errors.Is(fehler, fall.art) {
				t.Fatalf("es kam %v", fehler)
			}

			if len(danach.Objekte) != 0 || len(fall.partie.Objekte) != vorher {
				t.Error("trotz Abweisung liegt etwas auf dem Tisch")
			}
		})
	}
}
