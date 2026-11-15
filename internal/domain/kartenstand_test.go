package domain_test

// **Der Stand einer Karte** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Derselbe Inhalt gibt denselben Stand.**
func TestDerselbeInhaltGibtDenselbenStand(t *testing.T) {
	t.Parallel()

	eine := domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "name", Text: "Blitzschlag"}, {Field: "text", Text: "3 Schaden"},
	}}

	// **Auch in anderer Reihenfolge** — die Felder einer Karte sind eine
	// Menge, und eine Liste, deren Reihenfolge zählte, meldete einen Wechsel,
	// wo keiner war.
	andere := domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "text", Text: "3 Schaden"}, {Field: "name", Text: "Blitzschlag"},
	}}

	if eine.Stand() != andere.Stand() {
		t.Errorf("%q gegen %q", eine.Stand(), andere.Stand())
	}
}

// **Ein geänderter Wert gibt einen anderen Stand.**
func TestEinGeaenderterWertGibtEinenAnderenStand(t *testing.T) {
	t.Parallel()

	vorher := domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "text", Text: "3 Schaden"},
	}}

	nachher := domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "text", Text: "4 Schaden"},
	}}

	if vorher.Stand() == nachher.Stand() {
		t.Error("die Änderung blieb im Stand unsichtbar")
	}
}

// **Und ein Feld, das dazukommt, ebenso.**
//
// **Die Gegenprobe zur Verkettung**: wer Namen und Werte einfach
// aneinanderhängte, gäbe für `{ab: c}` und `{a: bc}` denselben Stand.
func TestEinVerschobenesFeldGibtEinenAnderenStand(t *testing.T) {
	t.Parallel()

	eine := domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "ab", Text: "c"},
	}}

	andere := domain.CardRow{ID: 7, Values: []domain.Value{
		{Field: "a", Text: "bc"},
	}}

	if eine.Stand() == andere.Stand() {
		t.Error("zwei verschiedene Karten geben denselben Stand")
	}
}

// **Eine Karte ohne Felder hat einen Stand.**
//
// **Und keinen leeren**: ein leerer Stand sähe aus wie „nicht gerechnet", und
// der Vergleich in `service.KIAuftragswerk` ginge dann immer durch.
func TestEineLeereKarteHatEinenStand(t *testing.T) {
	t.Parallel()

	if stand := (domain.CardRow{ID: 7}).Stand(); stand == "" {
		t.Error("die leere Karte hat keinen Stand")
	}
}
