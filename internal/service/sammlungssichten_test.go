package service_test

// **Die Fragen an eine geordnete Sammlung** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk 0057).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// geordnet gibt eine Ablage mit Besitz, Wunsch und zwei Lagerorten.
func geordnet(t *testing.T) *sammlungsdoppel {
	t.Helper()

	ablage := &sammlungsdoppel{}

	for _, exemplar := range []*domain.Exemplar{
		{
			Printing: probeprinting, Menge: 4,
			Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
			Lagerort: "Kiste 4",
		},
		{
			// **Derselbe Druck als Wunsch** — genau der Fall, für den die
			// Eindeutigkeit `wunsch` aufnimmt (0057).
			Printing: probeprinting, Menge: 2,
			Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
			Wunsch: true,
		},
		{
			Printing: "22222222-2222-4222-8222-222222222222", Menge: 1,
			Zustand: domain.ZustandPlayed, Sprache: domain.Vorgabesprache,
			Lagerort: "Ordner Rot",
		},
	} {
		if _, fehler := ablage.Anlegen(t.Context(), exemplar); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	return ablage
}

// **Besitz und Wunsch werden getrennt gezählt.**
func TestBesitzUndWunschWerdenGetrenntGezaehlt(t *testing.T) {
	t.Parallel()

	stand, fehler := service.NeueSammlungssichten(geordnet(t)).
		Besitz(t.Context(), probeprinting)
	if fehler != nil {
		t.Fatalf("Besitz: %v", fehler)
	}

	if stand.Stueck != 4 {
		t.Errorf("%d Stück, erwartet 4", stand.Stueck)
	}

	if stand.Gewuenscht != 2 {
		t.Errorf("%d gewünscht, erwartet 2", stand.Gewuenscht)
	}

	if !stand.Besitzt() {
		t.Error("vier Stück gelten als kein Besitz")
	}
}

// **Ein Druck, den es nur als Wunsch gibt, ist kein Besitz.**
//
// Das ist die Aussage aus 0057, Punkt 2 — und sie ist die Naht zum Deckbau:
// „besitze ich diese Karte?" muss Nein sagen, wenn sie auf der Wunschliste
// steht.
func TestEinReinerWunschIstKeinBesitz(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
		Printing: probeprinting, Menge: 3,
		Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
		Wunsch: true,
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	stand, fehler := service.NeueSammlungssichten(ablage).
		Besitz(t.Context(), probeprinting)
	if fehler != nil {
		t.Fatalf("Besitz: %v", fehler)
	}

	if stand.Besitzt() || stand.Stueck != 0 {
		t.Errorf("ein Wunsch gilt als Besitz: %+v", stand)
	}

	if stand.Gewuenscht != 3 {
		t.Errorf("%d gewünscht, erwartet 3", stand.Gewuenscht)
	}
}

// **Die Lagerorte stehen nach Namen — und ohne Wünsche.**
func TestDieLagerorteStehenNachNamen(t *testing.T) {
	t.Parallel()

	orte, fehler := service.NeueSammlungssichten(geordnet(t)).Lagerorte(t.Context())
	if fehler != nil {
		t.Fatalf("Lagerorte: %v", fehler)
	}

	if len(orte) != 2 {
		t.Fatalf("%d Lagerorte, erwartet 2: %+v", len(orte), orte)
	}

	if orte[0].Lagerort != "Kiste 4" || orte[1].Lagerort != "Ordner Rot" {
		t.Errorf("die Reihenfolge ist %q, %q", orte[0].Lagerort, orte[1].Lagerort)
	}

	if orte[0].Stueck != 4 || orte[0].Zeilen != 1 {
		t.Errorf("Kiste 4 hat %d Stück in %d Zeilen", orte[0].Stueck, orte[0].Zeilen)
	}
}

// **Ein Wunsch liegt nirgends** — er taucht in keinem Lagerort auf.
func TestEinWunschLiegtNirgends(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
		Printing: probeprinting, Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
		Wunsch: true, Lagerort: "Traumkiste",
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	orte, fehler := service.NeueSammlungssichten(ablage).Lagerorte(t.Context())
	if fehler != nil {
		t.Fatalf("Lagerorte: %v", fehler)
	}

	if len(orte) != 0 {
		t.Errorf("%d Lagerorte, erwartet 0: %+v", len(orte), orte)
	}
}

// **Dubletten hängen an der Zahl, die die fragende Person mitbringt.**
func TestDublettenHaengenAnDerBehaltenenZahl(t *testing.T) {
	t.Parallel()

	sichten := service.NeueSammlungssichten(geordnet(t))

	for _, fall := range []struct {
		behalten int
		zeilen   int
		ueber    int
	}{
		{behalten: 1, zeilen: 1, ueber: 3},
		{behalten: 3, zeilen: 1, ueber: 1},
		{behalten: 4, zeilen: 0},
		{behalten: 0, zeilen: 2, ueber: 4},
	} {
		dubletten, fehler := sichten.Dubletten(t.Context(), fall.behalten)
		if fehler != nil {
			t.Fatalf("Dubletten(%d): %v", fall.behalten, fehler)
		}

		if len(dubletten) != fall.zeilen {
			t.Errorf("behalten %d: %d Zeilen, erwartet %d",
				fall.behalten, len(dubletten), fall.zeilen)

			continue
		}

		if fall.zeilen > 0 && dubletten[0].Ueberzaehlig != fall.ueber {
			t.Errorf("behalten %d: %d überzählig, erwartet %d",
				fall.behalten, dubletten[0].Ueberzaehlig, fall.ueber)
		}
	}
}

// **Eine Zahl unter null ist keine, die man behält.**
func TestEineBehalteneZahlUnterNull(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeueSammlungssichten(&sammlungsdoppel{}).
		Dubletten(t.Context(), -1)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
	}
}

// **Die Wunschliste gibt nur Wünsche.**
func TestDieWunschlisteGibtNurWuensche(t *testing.T) {
	t.Parallel()

	wuensche, fehler := service.NeueSammlungssichten(geordnet(t)).Wuensche(t.Context())
	if fehler != nil {
		t.Fatalf("Wuensche: %v", fehler)
	}

	if len(wuensche) != 1 {
		t.Fatalf("%d Wünsche, erwartet 1: %+v", len(wuensche), wuensche)
	}

	if wuensche[0].Menge != 2 {
		t.Errorf("Menge %d, erwartet 2", wuensche[0].Menge)
	}
}

// **Das Portfolio zählt Wünsche nicht — auch nicht als „ohne Preis".**
//
// 0057, Punkt 1: was man nicht hat, ist nichts wert. Wer sich zwanzig Karten
// wünscht, hätte sonst zwanzig Exemplare „ohne Preis" in einer Übersicht über
// seinen **Besitz**.
func TestDasPortfolioZaehltKeineWuensche(t *testing.T) {
	t.Parallel()

	pflege := service.NeuePreispflege(&preisdoppel{}, geordnet(t),
		quellendoppel{}, probeuhr())

	uebersicht, fehler := pflege.Portfolio(t.Context())
	if fehler != nil {
		t.Fatalf("Portfolio: %v", fehler)
	}

	// **Fünf Stück Besitz** — vier plus eines. Die zwei gewünschten zählen
	// nicht mit.
	if uebersicht.OhnePreis != 5 {
		t.Errorf("%d ohne Preis, erwartet 5", uebersicht.OhnePreis)
	}
}

// **Ein Lesefehler kommt auf jedem der vier Wege durch.**
func TestDieSichtenReichenLesefehlerDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")
	sichten := service.NeueSammlungssichten(&sammlungsdoppel{fehler: gestolpert})

	if _, fehler := sichten.Besitz(t.Context(), probeprinting); !errors.Is(fehler, gestolpert) {
		t.Errorf("Besitz: %v", fehler)
	}

	if _, fehler := sichten.Lagerorte(t.Context()); !errors.Is(fehler, gestolpert) {
		t.Errorf("Lagerorte: %v", fehler)
	}

	if _, fehler := sichten.Dubletten(t.Context(), 1); !errors.Is(fehler, gestolpert) {
		t.Errorf("Dubletten: %v", fehler)
	}

	if _, fehler := sichten.Wuensche(t.Context()); !errors.Is(fehler, gestolpert) {
		t.Errorf("Wuensche: %v", fehler)
	}
}
