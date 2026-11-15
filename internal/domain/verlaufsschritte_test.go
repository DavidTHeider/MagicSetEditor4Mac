package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verlauf als Schritte** — Werkstück U6 von Phase 5.1 (#1417): eine
// Massenänderung ist **ein** Schritt, so wie sie in einem Zug zurückgeht;
// eine Einzeländerung ist einer für sich. Die Reihenfolge bleibt die des
// Verlaufs — der jüngste Schritt zuletzt, wie die Einträge.
func TestVerlaufsschritteFassenEineGruppeZusammen(t *testing.T) {
	eintraege := []domain.Verlaufseintrag{
		{ID: 1, KarteID: 7, Feld: "name", Nachher: "A"},
		{ID: 2, KarteID: 7, Feld: "typ", Nachher: "X", Gruppe: "g1"},
		{ID: 3, KarteID: 8, Feld: "typ", Nachher: "X", Gruppe: "g1"},
		{ID: 4, KarteID: 9, Feld: "typ", Nachher: "X", Gruppe: "g1"},
		{ID: 5, KarteID: 7, Feld: "name", Nachher: "B"},
		{ID: 6, KarteID: 8, Feld: "typ", Nachher: "Y", Gruppe: "g2"},
	}

	schritte := domain.Verlaufsschritte(eintraege)

	if len(schritte) != 4 {
		t.Fatalf("%d Schritte statt 4: %+v", len(schritte), schritte)
	}

	// Der erste Eintrag eines Schritts vertritt ihn — wie bei der Rücknahme
	// (#2396; bis dahin der jüngste).
	soll := []struct {
		id     int64
		umfang int
	}{{1, 1}, {2, 3}, {5, 1}, {6, 1}}

	for i, s := range soll {
		if schritte[i].ID != s.id || schritte[i].Umfang != s.umfang {
			t.Errorf("Schritt %d: ID %d, Umfang %d — erwartet ID %d, Umfang %d",
				i, schritte[i].ID, schritte[i].Umfang, s.id, s.umfang)
		}
	}
}

// **Eine Gruppe, die nicht am Stück steht, ist trotzdem eine.** Der Verlauf
// schreibt sie am Stück; sollte je etwas dazwischenliegen, gehört es nicht
// dazu — und die Gruppe bleibt ein Schritt.
func TestVerlaufsschritteOhneEintraegeSindLeer(t *testing.T) {
	if schritte := domain.Verlaufsschritte(nil); len(schritte) != 0 {
		t.Errorf("%+v", schritte)
	}
}

// **Ein Schritt zeigt das Feld, das bearbeitet wurde, und zählt Karten**
// (#2396). Seit der Schreibweg nachzieht, wer ein Feld liest, trägt schon eine
// einzige Bearbeitung mehrere Einträge einer Karte: das gesetzte Feld zuerst,
// dann seine Leser. Der jüngste vertrat den Schritt — die Liste nannte das
// zuletzt nachgezogene Feld mit dessen Werten und „2 Karten", wo eine
// bearbeitet war.
func TestEinSchrittNenntDasBearbeiteteFeldUndZaehltKarten(t *testing.T) {
	schritte := domain.Verlaufsschritte([]domain.Verlaufseintrag{
		{ID: 1, KarteID: 7, Feld: "eins", Gruppe: "g1"},
		{ID: 2, KarteID: 7, Feld: "leser", Gruppe: "g1"},
		{ID: 3, KarteID: 7, Feld: "name", Gruppe: "g2"},
		{ID: 4, KarteID: 7, Feld: "leser", Gruppe: "g2"},
		{ID: 5, KarteID: 8, Feld: "name", Gruppe: "g2"},
		{ID: 6, KarteID: 8, Feld: "leser", Gruppe: "g2"},
	})

	if len(schritte) != 2 {
		t.Fatalf("%d Schritte statt 2: %+v", len(schritte), schritte)
	}

	soll := []struct {
		id     int64
		feld   string
		umfang int
	}{{1, "eins", 1}, {3, "name", 2}}

	for i, s := range soll {
		if schritte[i].ID != s.id || schritte[i].Feld != s.feld || schritte[i].Umfang != s.umfang {
			t.Errorf("Schritt %d: ID %d, Feld %q, Umfang %d — erwartet ID %d, Feld %q, Umfang %d",
				i, schritte[i].ID, schritte[i].Feld, schritte[i].Umfang, s.id, s.feld, s.umfang)
		}
	}
}
