package service_test

// **Wer Bilder zeichnet, sagt es** — Nachlauf zum Zweitblick über Phase 7,
// Rest R4. Eine Zeile im eigenen Bestand ist kein Fähigkeitsnachweis: der
// Wirt kann an Bos Sitz eine Leihe **von Bo** vorlegen, und Bo verbindet sich
// mit einer Fassung vor 0.9.0, die `bildwunsch` nicht kennt. Deshalb erklärt
// ein Gast in jedem Rückkehrwunsch, was er kann, und der Wirt fragt nur, wer
// es erklärt hat.

import (
	"context"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerGastErklaertBeiJederRueckkehrDassErBilderZeichnet(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})
	gast.MitStandhalter(&standdoppel{folge: 4, abdruck: "vier"})

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	if wunsch := derWunsch(t, naht); !slices.Contains(wunsch.Kann, domain.FaehigkeitBild) {
		t.Errorf("der Wunsch ist %+v", wunsch)
	}
}

// **Ohne Erklärung fragt der Wirt niemanden** — auch wenn der Bestand des
// Gastes die Zeile hält.
func TestOhneErklaerungFragtDerWirtKeinenGast(t *testing.T) {
	t.Parallel()

	wirt, naht, cora, _, runde := einWeiterreichenderWirt(t)
	runde.Kann("cora", nil)

	if antwort := erfrage(t, wirt, naht, "coras"); antwort.Grund == "" {
		t.Errorf("die Antwort: %+v", antwort)
	}

	if len(cora.alle()) != 0 {
		t.Error("Cora wurde gefragt")
	}
}

// **Der Wirt lernt es aus dem Rückkehrwunsch** — der Wirt von Coras Sitz
// trägt ihre Erklärung in die Runde ein, und erst danach wird sie gefragt.
func TestDerWirtLerntDieFaehigkeitAusDerRueckkehr(t *testing.T) {
	t.Parallel()

	_, _, _, _, runde := einWeiterreichenderWirt(t)
	runde.Kann("cora", nil)

	coraswirt := service.NeuerPartiewirt("p1", "cora", &nehmerdoppel{},
		service.NeuePartiesitzung("p1", "wirt", &nahtdoppel{}, service.NeueReihenfolge(1))).MitRunde(runde)

	if fehler := coraswirt.Angekommen(context.Background(), eineKontrollnachricht(t,
		domain.KontrolleRueckkehr, domain.Wiederaufnahmewunsch{
			Partie: "p1", Kann: []string{domain.FaehigkeitBild},
		})); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if !runde.Zeichnet("cora") {
		t.Error("die Erklärung kam nicht an")
	}
}
