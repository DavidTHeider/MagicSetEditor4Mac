package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Sechs Posten, und genau zwei dürfen weg** — Werkstück 4 von Phase 4.13
// (#1352),
// erweitert in Werkstück 6.
//
// **Die Zahl steht hier, damit eine leere Aufzählung nicht grün ist**: eine
// Liste, aus der jemand einen Posten nimmt, sähe sonst aus wie eine
// vollständige.
//
// **Vier waren es bis zum 2026-09-07.** Dann kamen das Protokoll und der Rest
// dazu — der Anwendungsordner wog 13 GB, und die vier nannten davon 1,5.
func TestDieSechsPostenUndDieZweiRaeumbaren(t *testing.T) {
	t.Parallel()

	alle := domain.PostenAlle()
	if len(alle) != 6 {
		t.Fatalf("es sind %d Posten: %v", len(alle), alle)
	}

	// **Der Bestand zuerst**, weil er die Frage beantwortet, mit der jemand
	// herkommt — **und der Rest zuletzt**, weil er das ist, was übrig bleibt.
	if alle[0] != domain.PostenBestand {
		t.Errorf("der erste Posten ist %q", alle[0])
	}

	if alle[len(alle)-1] != domain.PostenSonstiges {
		t.Errorf("der letzte Posten ist %q", alle[len(alle)-1])
	}

	raeumbar := map[domain.Platzposten]bool{}

	for _, posten := range alle {
		if posten.Raeumbar() {
			raeumbar[posten] = true
		}
	}

	erwartet := map[domain.Platzposten]bool{
		domain.PostenVorsicherungen: true, domain.PostenProtokoll: true,
	}

	for posten := range raeumbar {
		if !erwartet[posten] {
			t.Errorf("%q gilt als räumbar", posten)
		}
	}

	for posten := range erwartet {
		if !raeumbar[posten] {
			t.Errorf("%q gilt nicht als räumbar", posten)
		}
	}
}

// **Und was kein Posten ist, wird nicht geführt.**
func TestNurGefuehrtePostenGeltenAlsSolche(t *testing.T) {
	t.Parallel()

	for _, posten := range domain.PostenAlle() {
		if !domain.PostenGefuehrt(string(posten)) {
			t.Errorf("%q ist ein Posten und gilt nicht als einer", posten)
		}
	}

	for _, kein := range []string{"", "bestände", "zwischenspeicher"} {
		if domain.PostenGefuehrt(kein) {
			t.Errorf("%q gilt als Posten", kein)
		}
	}
}
