package domain_test

// **Wunsch und Fassung** — Werkstück A4 von Phase 5.3
// (#1443).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein leerer Wunsch lässt alles durch.**
func TestEinLeererWunschLaesstAllesDurch(t *testing.T) {
	t.Parallel()

	angebot := domain.Marktangebot{
		Finish: domain.FinishFoil, Zustand: "Near Mint", Sprache: "de",
	}

	if !(domain.Angebotswunsch{}).Passt(angebot) {
		t.Error("der leere Wunsch hat gefiltert")
	}
}

// **Jeder Teil des Wunsches filtert für sich.**
//
// **Und der Vergleich achtet nicht auf Groß und Klein**: ein Anbieter, der
// `EN` statt `en` schickt, meint dieselbe Sprache — und ein Filter, der das
// nicht wüsste, lieferte eine leere Liste ohne Begründung.
func TestJederTeilDesWunschesFiltert(t *testing.T) {
	t.Parallel()

	angebot := domain.Marktangebot{
		Finish: domain.FinishNonfoil, Zustand: "Near Mint", Sprache: "en",
	}

	for _, fall := range []struct {
		name   string
		wunsch domain.Angebotswunsch
		passt  bool
	}{
		{"dieselbe Sprache", domain.Angebotswunsch{Sprache: "en"}, true},
		{"dieselbe Sprache anders geschrieben", domain.Angebotswunsch{Sprache: "EN"}, true},
		{"eine andere Sprache", domain.Angebotswunsch{Sprache: "de"}, false},
		{"dieselbe Ausführung", domain.Angebotswunsch{Finish: domain.FinishNonfoil}, true},
		{"eine andere Ausführung", domain.Angebotswunsch{Finish: domain.FinishFoil}, false},
		{
			"ein Zustand darunter",
			domain.Angebotswunsch{Zustaende: []string{"Played", "near mint"}},
			true,
		},
		{
			"kein Zustand darunter",
			domain.Angebotswunsch{Zustaende: []string{"Played"}},
			false,
		},
		{
			"alles zusammen",
			domain.Angebotswunsch{
				Sprache: "en", Finish: domain.FinishNonfoil,
				Zustaende: []string{"Near Mint"},
			},
			true,
		},
		{
			"alles zusammen, und eines passt nicht",
			domain.Angebotswunsch{
				Sprache: "en", Finish: domain.FinishFoil,
				Zustaende: []string{"Near Mint"},
			},
			false,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if fall.wunsch.Passt(angebot) != fall.passt {
				t.Errorf("%+v passt %v, erwartet %v",
					fall.wunsch, !fall.passt, fall.passt)
			}
		})
	}
}

// **Eine Fassung erkennt ihre Kennung — auch anders geschrieben.**
func TestEineFassungErkenntIhreKennung(t *testing.T) {
	t.Parallel()

	fassung := domain.Fremdfassung{
		Kennung: "89", Name: "Beast // Plant",
		Kennungen: []domain.Druckkennung{
			{System: domain.SystemCardTrader, Semantik: domain.SemantikDruck, Wert: "89"},
			{
				System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
				Wert: "381854e2-7369-473e-a604-4dd7c010fc89",
			},
		},
	}

	gross := domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
		Wert: "381854E2-7369-473E-A604-4DD7C010FC89",
	}

	if !fassung.Nennt(gross) {
		t.Error("dieselbe Kennung in Großschreibung wurde nicht erkannt")
	}

	// **Und die Semantik gehört dazu**: dieselbe Zeichenkette unter einer
	// anderen Semantik ist eine andere Aussage (0086).
	anders := domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikKonzept,
		Wert: "381854e2-7369-473e-a604-4dd7c010fc89",
	}

	if fassung.Nennt(anders) {
		t.Error("eine Kennung mit anderer Semantik galt als dieselbe")
	}

	if fassung.Nennt(domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: "andere",
	}) {
		t.Error("eine fremde Kennung galt als eigene")
	}
}

// **Und auch die Fassung selbst wird normalisiert.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil die Gegenprobe es verlangt hat.** Im Lauf darüber steht die Kennung
// der Fassung schon klein, und nur die gesuchte ist groß — wer die
// Normalisierung **in der Schleife** entfernte, sah einen grünen Lauf. Eine
// Absicherung, die man folgenlos entfernen kann, ist keine (§10).
//
// Der Fall ist nicht erfunden: `internal/cardtrader` normalisiert beim Lesen,
// ein zweiter Marktplatz muss es nicht tun — und die Fassung kommt von dort.
func TestAuchDieFassungWirdNormalisiert(t *testing.T) {
	t.Parallel()

	fassung := domain.Fremdfassung{
		Kennung: "89",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
			Wert: "  381854E2-7369-473E-A604-4DD7C010FC89 ",
		}},
	}

	if !fassung.Nennt(domain.Druckkennung{
		System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
		Wert: "381854e2-7369-473e-a604-4dd7c010fc89",
	}) {
		t.Error("eine groß geschriebene Kennung der Fassung wurde nicht erkannt")
	}
}
