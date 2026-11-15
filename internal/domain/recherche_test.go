package domain_test

// **Die Recherche** — Werkstück A16 von Phase 5.3
// (#1515).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Fall sucht in seinen Korpora.**
//
// **Und keiner in allen**: eine Regelfrage, die in Kartentexten sucht, findet
// Karten mit dem Wort — und das ist eine andere Frage als die gestellte.
func TestJederRecherchefallSuchtInSeinenKorpora(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		fall     domain.Recherchefall
		erwartet []domain.Retrievalquelle
	}{
		"Karten":   {domain.FallKarten, []domain.Retrievalquelle{domain.QuelleOracle, domain.QuelleEigeneKarte}},
		"Ähnliche": {domain.FallAehnlich, []domain.Retrievalquelle{domain.QuelleOracle, domain.QuelleEigeneKarte}},
		"Regeln":   {domain.FallRegeln, []domain.Retrievalquelle{domain.QuelleRegelwerk}},
		"Rulings":  {domain.FallRulings, []domain.Retrievalquelle{domain.QuelleRuling}},

		// **Der Regelhinweis sucht in dreien** — A13.2a (#1452): eine
		// Aussage über eine Karte stützt sich auf das Regelwerk, auf
		// datierte Rulings **und** auf Oracle-Texte.
		"Regelhinweis": {domain.FallRegelhinweis, []domain.Retrievalquelle{
			domain.QuelleRegelwerk, domain.QuelleRuling, domain.QuelleOracle,
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quellen := fall.fall.Quellen()
			if len(quellen) != len(fall.erwartet) {
				t.Fatalf("die Quellen sind %v", quellen)
			}

			for stelle, quelle := range fall.erwartet {
				if quellen[stelle] != quelle {
					t.Fatalf("die Quellen sind %v", quellen)
				}
			}
		})
	}

	// **Ein Fall, den es nicht gibt, sucht nirgends.** Eine leere Liste
	// hieße „überall", und dann suchte ein Tippfehler im ganzen Bestand.
	if quellen := domain.Recherchefall("erfunden").Quellen(); quellen != nil {
		t.Errorf("ein erfundener Fall sucht in %v", quellen)
	}

	// **Fünf seit A13.2a** (#1452). Die Zahl steht hier, damit ein sechster
	// nicht still dazukommt: jeder Fall ist eine Zusage über Korpora, und
	// eine, die niemand aufgeschrieben hat, ist keine.
	if len(domain.Recherchefaelle()) != 5 {
		t.Errorf("es sind %d Fälle", len(domain.Recherchefaelle()))
	}
}

// **Eine Ähnlichkeitsfrage braucht ihre Ausgangskarte samt Stand.**
//
// **Der Stand ist der Punkt**: eine Karte, die sich seit der Anfrage geändert
// hat, ist eine andere Frage — und die Antwort sähe aus wie die auf die
// gestellte.
func TestEineAehnlichkeitsfrageBrauchtIhrenAusgang(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anfrage domain.Rechercheanfrage
		haelt   bool
	}{
		"mit Karte und Stand": {
			domain.Rechercheanfrage{
				Fall: domain.FallAehnlich, AusgangskarteID: 7,
				Ausgangsstand: "rev-1", Grenze: 10,
			}, true,
		},
		"ohne Karte": {
			domain.Rechercheanfrage{
				Fall: domain.FallAehnlich, Ausgangsstand: "rev-1", Grenze: 10,
			}, false,
		},
		// **Ein leerer Stand ist kein fehlender** (A16, seit dem
		// 2026-09-12): er heißt „der, auf dem sie jetzt steht", und die
		// Antwort nennt ihn. Die Pflicht, die hier stand, war an keiner
		// Fläche einzulösen — den gerechneten Stand einer Karte kennt nur
		// der Bestand.
		"ohne Stand": {
			domain.Rechercheanfrage{
				Fall: domain.FallAehnlich, AusgangskarteID: 7, Grenze: 10,
			}, true,
		},
		"mit Frage statt Karte": {
			domain.Rechercheanfrage{
				Fall: domain.FallAehnlich, Frage: "etwas wie Blitzschlag",
				Grenze: 10,
			}, false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.anfrage.Pruefe()
			if fall.haelt != (fehler == nil) {
				t.Fatalf("Pruefe: %v", fehler)
			}

			if fehler != nil && !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("das Wahrzeichen fehlt: %v", fehler)
			}
		})
	}
}

// **Die anderen drei Fälle brauchen eine Frage.**
func TestDieAnderenFaelleBrauchenEineFrage(t *testing.T) {
	t.Parallel()

	for _, fall := range []domain.Recherchefall{
		domain.FallKarten, domain.FallRegeln, domain.FallRulings,
	} {
		t.Run(string(fall), func(t *testing.T) {
			t.Parallel()

			if fehler := (domain.Rechercheanfrage{
				Fall: fall, Grenze: 10,
			}).Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("die leere Frage kam durch: %v", fehler)
			}

			if fehler := (domain.Rechercheanfrage{
				Fall: fall, Frage: "lightning", Grenze: 10,
			}).Pruefe(); fehler != nil {
				t.Fatalf("Pruefe: %v", fehler)
			}
		})
	}

	// **Und ein Fall, den es nicht gibt, wird abgewiesen.**
	if fehler := (domain.Rechercheanfrage{
		Fall: "erfunden", Frage: "x", Grenze: 10,
	}).Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der erfundene Fall kam durch: %v", fehler)
	}
}

// **Die Trefferzahl hat eine Grenze nach oben und nach unten.**
func TestDieTrefferzahlHatZweiGrenzen(t *testing.T) {
	t.Parallel()

	for name, grenze := range map[string]int{
		"null":     0,
		"negativ":  -3,
		"zu viele": domain.Recherchehoechstens + 1,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := (domain.Rechercheanfrage{
				Fall: domain.FallKarten, Frage: "x", Grenze: grenze,
			}).Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%s kam durch: %v", name, fehler)
			}
		})
	}

	// **Und dieselbe Grenze gilt für die Ähnlichkeitsfrage** — sie geht
	// durch einen anderen Zweig und käme sonst ungeprüft durch.
	if fehler := (domain.Rechercheanfrage{
		Fall: domain.FallAehnlich, AusgangskarteID: 1, Ausgangsstand: "r",
		Grenze: domain.Recherchehoechstens + 1,
	}).Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die zu große Grenze kam durch: %v", fehler)
	}
}

// **Es gibt sechs Gründe für eine leere Antwort.**
func TestEsGibtSechsGruendeFuerEineLeereAntwort(t *testing.T) {
	t.Parallel()

	gruende := domain.Rechercheleeren()

	if len(gruende) != 6 {
		t.Fatalf("es sind %d Gründe", len(gruende))
	}

	for _, grund := range gruende {
		if grund == domain.LeereKeine {
			t.Error("die Aufzählung führt »nicht leer« als Grund")
		}
	}
}
