package domain_test

// **Was geteilt wird, und was dabei mitgeht** — NET-04.AK-02
// (#1708):
// „Vor dem Teilen zeigt die App Zielinhalt, Umfang, enthaltene Zusatzdaten und
// mögliche Verluste."
//
// Die Antwort steht **hier** und nicht im Fenster: welche Angaben eine
// Sammlung als CSV mitnimmt, ist eine fachliche Aussage, und vier Flächen
// sollen dieselbe machen.

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeTeilartNenntWasMitgehtUndWasFehlt(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Teilarten() {
		if len(art.Mitgehendes()) == 0 {
			t.Errorf("%s nennt nichts, was mitgeht", art)
		}

		if gelesen, fehler := domain.PruefeTeilart(string(art)); fehler != nil || gelesen != art {
			t.Errorf("%s: %v, %v", art, gelesen, fehler)
		}
	}

	// **Eine Sammlung als CSV trägt Private** — Erwerbspreise, Lagerorte,
	// Notizen. Das muss dastehen, bevor sie hinausgeht.
	if !slices.Contains(domain.TeilartSammlung.Mitgehendes(), "erwerb") {
		t.Error("die Sammlung verschweigt die Erwerbsangaben")
	}

	// **Ein Deck als Liste verliert Zuordnungen und Format.**
	if !slices.Contains(domain.TeilartDeck.Verluste(), "zuordnungen") {
		t.Error("das Deck verschweigt die Zuordnungen")
	}

	if _, fehler := domain.PruefeTeilart("brief"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("eine fremde Art: %v", fehler)
	}
}

// **Eine Kennung aus dem Fenster ist ein Pfadteil** — sie wird geprüft,
// bevor sie an die Platte geht.
func TestEineTeilkennungIstHexUndSonstNichts(t *testing.T) {
	t.Parallel()

	for kennung, gut := range map[string]bool{
		"0123456789abcdef0123456789abcdef": true,
		"../../etc":                        false,
		"0123456789ABCDEF0123456789ABCDEF": false,
		"":                                 false,
		"0123":                             false,
	} {
		if fehler := domain.PruefeTeilkennung(kennung); (fehler == nil) != gut {
			t.Errorf("%q: %v", kennung, fehler)
		}
	}
}

// **Ein Titel wird ein Dateiname** — ohne Schrägstrich, ohne Doppelpunkt,
// ohne Punkt vorn.
func TestEinTitelWirdEinSichererDateiname(t *testing.T) {
	t.Parallel()

	for titel, erwartet := range map[string]string{
		"Mein Set":               "Mein Set.mse-set",
		"Rot/Grün: Deck":         "Rot-Grün- Deck.mse-set",
		"..versteckt":            "versteckt.mse-set",
		"":                       "teilen.mse-set",
		strings.Repeat("x", 300): strings.Repeat("x", 120) + ".mse-set",
	} {
		if name := domain.Teildateiname(titel, ".mse-set"); name != erwartet {
			t.Errorf("%q wurde %q, erwartet %q", titel, name, erwartet)
		}
	}
}

// **Über der Mailgrenze empfiehlt die Vorschau einen anderen Weg** (AK-03).
func TestEinGrossesTeilgutIstZuGrossFuerMail(t *testing.T) {
	t.Parallel()

	if (domain.Teilgut{Groesse: domain.Mailgrenze}).ZuGrossFuerMail() {
		t.Error("genau an der Grenze gilt als zu groß")
	}

	if !(domain.Teilgut{Groesse: domain.Mailgrenze + 1}).ZuGrossFuerMail() {
		t.Error("über der Grenze gilt als passend")
	}
}
