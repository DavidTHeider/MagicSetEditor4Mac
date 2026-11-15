package domain_test

// **Eine Turnierpaarung, an eine Partie gebunden** — Werkstück G7 von Phase
// 5.5 (#1464).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineBindung gibt eine gültige Bindung mit den genannten Abweichungen.
func eineBindung(aendere ...func(*domain.Matchbindung)) domain.Matchbindung {
	bindung := domain.Matchbindung{
		TurnierID: 1, PaarungID: 7, Fassung: 3, PartieID: "p1",
		Hoechstens: 3,
		Sitze: []domain.Matchsitz{
			{TeilnehmerID: 10, Sitz: "anna", Deckabdruck: "abdruckA"},
			{TeilnehmerID: 11, Sitz: "bert", Deckabdruck: "abdruckB"},
		},
	}

	for _, aenderung := range aendere {
		aenderung(&bindung)
	}

	return bindung
}

func TestMatchbindungPruefe(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		aendere func(*domain.Matchbindung)
		gilt    bool
	}{
		"so, wie sie entsteht": {aendere: func(*domain.Matchbindung) {}, gilt: true},
		"ohne Turnier": {
			aendere: func(b *domain.Matchbindung) { b.TurnierID = 0 },
		},
		"ohne Paarung": {
			aendere: func(b *domain.Matchbindung) { b.PaarungID = 0 },
		},
		"ohne Fassung": {
			aendere: func(b *domain.Matchbindung) { b.Fassung = 0 },
		},
		"ohne Partie": {
			aendere: func(b *domain.Matchbindung) { b.PartieID = "" },
		},
		"mit einem Sitz": {
			aendere: func(b *domain.Matchbindung) {
				b.Sitze = b.Sitze[:1]
			},
		},
		"mit drei Sitzen": {
			aendere: func(b *domain.Matchbindung) {
				b.Sitze = append(b.Sitze, domain.Matchsitz{
					TeilnehmerID: 12, Sitz: "cem", Deckabdruck: "abdruckC",
				})
			},
		},
		"mit zweimal demselben Sitz": {
			aendere: func(b *domain.Matchbindung) {
				b.Sitze[1].Sitz = b.Sitze[0].Sitz
			},
		},
		"mit zweimal demselben Teilnehmer": {
			aendere: func(b *domain.Matchbindung) {
				b.Sitze[1].TeilnehmerID = b.Sitze[0].TeilnehmerID
			},
		},
		"ohne Deckabdruck": {
			aendere: func(b *domain.Matchbindung) {
				b.Sitze[0].Deckabdruck = ""
			},
		},
		"ohne Sitzkennung": {
			aendere: func(b *domain.Matchbindung) { b.Sitze[0].Sitz = "" },
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := eineBindung(fall.aendere).Pruefe()

			if fall.gilt && fehler != nil {
				t.Fatalf("erwartet gültig, bekam %v", fehler)
			}

			if !fall.gilt && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}
		})
	}
}

// eineMeldung gibt eine Meldung mit den genannten Abweichungen.
func eineMeldung(aendere ...func(*domain.Matchmeldung)) domain.Matchmeldung {
	meldung := domain.Matchmeldung{
		PaarungID: 7, Fassung: 3, PartieID: "p1",
		Ergebnis: domain.Ergebnis{SpieleA: 2, SpieleB: 1},
		Kette:    "letztesglied", Von: "anna",
	}

	for _, aenderung := range aendere {
		aenderung(&meldung)
	}

	return meldung
}

// **Eine Meldung, die nicht zu dieser Bindung gehört, wird abgewiesen.**
func TestEineFremdeMeldungPasstNicht(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		aendere func(*domain.Matchmeldung)
		passt   bool
	}{
		"so, wie sie kommt": {aendere: func(*domain.Matchmeldung) {}, passt: true},
		"eine andere Paarung": {
			aendere: func(m *domain.Matchmeldung) { m.PaarungID = 9 },
		},
		"eine andere Fassung": {
			aendere: func(m *domain.Matchmeldung) { m.Fassung = 4 },
		},
		"eine andere Partie": {
			aendere: func(m *domain.Matchmeldung) { m.PartieID = "p9" },
		},
		"von einem fremden Sitz": {
			aendere: func(m *domain.Matchmeldung) { m.Von = "cem" },
		},
		"ohne Nachweis": {
			aendere: func(m *domain.Matchmeldung) { m.Kette = "" },
		},
		"ohne ein einziges Spiel": {
			aendere: func(m *domain.Matchmeldung) {
				m.Ergebnis = domain.Ergebnis{}
			},
		},
		"mit mehr Spielen als erlaubt": {
			aendere: func(m *domain.Matchmeldung) {
				m.Ergebnis = domain.Ergebnis{SpieleA: 3, SpieleB: 3}
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := eineBindung().Nimmt(eineMeldung(fall.aendere))

			if fall.passt && fehler != nil {
				t.Fatalf("erwartet passend, bekam %v", fehler)
			}

			if !fall.passt && fehler == nil {
				t.Fatal("erwartet eine Abweisung")
			}
		})
	}
}

// **Ein Ergebnis wird erst übernommen, wenn beide Seiten dasselbe melden.**
func TestDieUebernahmeBrauchtBeideSeiten(t *testing.T) {
	t.Parallel()

	vonAnna := eineMeldung()
	vonBert := eineMeldung(func(m *domain.Matchmeldung) { m.Von = "bert" })

	faelle := map[string]struct {
		meldungen []domain.Matchmeldung
		fertig    bool
		streit    bool
	}{
		"keine Meldung":  {},
		"nur eine Seite": {meldungen: []domain.Matchmeldung{vonAnna}},
		"zweimal dieselbe Seite": {
			meldungen: []domain.Matchmeldung{vonAnna, vonAnna},
		},
		"beide Seiten, dasselbe": {
			meldungen: []domain.Matchmeldung{vonAnna, vonBert}, fertig: true,
		},
		"beide Seiten, andere Zahlen": {
			meldungen: []domain.Matchmeldung{vonAnna, eineMeldung(
				func(m *domain.Matchmeldung) {
					m.Von, m.Ergebnis = "bert", domain.Ergebnis{SpieleB: 2}
				})},
			streit: true,
		},
		"beide Seiten, anderer Nachweis": {
			meldungen: []domain.Matchmeldung{vonAnna, eineMeldung(
				func(m *domain.Matchmeldung) {
					m.Von, m.Kette = "bert", "einanderesglied"
				})},
			streit: true,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stand, fehler := eineBindung().Uebernahme(fall.meldungen)

			switch {
			case fall.streit:
				if !errors.Is(fehler, domain.ErrConflict) {
					t.Fatalf("erwartet ErrConflict, bekam %v", fehler)
				}
			case fall.fertig:
				if fehler != nil {
					t.Fatalf("erwartet ein Ergebnis, bekam %v", fehler)
				}

				if stand.SpieleA != 2 || stand.SpieleB != 1 {
					t.Fatalf("der Stand ist %+v", stand)
				}
			default:
				if fehler == nil {
					t.Fatal("erwartet, dass es noch nicht reicht")
				}

				if errors.Is(fehler, domain.ErrConflict) {
					t.Fatalf("erwartet keinen Streit, bekam %v", fehler)
				}
			}
		})
	}
}

// **Eine Meldung, die nicht passt, hält die ganze Übernahme auf.**
//
// **Und nicht nur sich selbst**: wer sie überginge, rechnete ein Ergebnis aus
// den übrigen — und das wäre eines, über das nicht beide Seiten gesprochen
// haben.
func TestEineUnpassendeMeldungHaeltDieUebernahmeAuf(t *testing.T) {
	t.Parallel()

	_, fehler := eineBindung().Uebernahme([]domain.Matchmeldung{
		eineMeldung(),
		eineMeldung(func(m *domain.Matchmeldung) { m.Von = "cem" }),
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}
