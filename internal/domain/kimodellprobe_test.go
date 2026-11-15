package domain_test

// **Eine gemessene Fähigkeit** — Werkstück A9 von Phase 5.3
// (#1448).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineProbe baut eine vollständige Messung.
func eineProbe(teile ...func(*domain.KIModellprobe)) domain.KIModellprobe {
	probe := domain.KIModellprobe{
		Anbieter: "openai", Alias: "gpt-probe",
		Faehigkeit: domain.KIStrukturierteGenerierung,
		Datum:      "2026-09-12",
		Beleg:      "Probeauftrag gegen /v1/responses",
		Laufzeit:   "Responses-API /v1/responses",
		Modell:     "gpt-probe-2026-09-01",
	}

	for _, teil := range teile {
		teil(&probe)
	}

	return probe
}

// **Eine vollständige Messung ist gültig.**
func TestEineVollstaendigeMessungIstGueltig(t *testing.T) {
	t.Parallel()

	if befunde := eineProbe().Validate(); len(befunde) != 0 {
		t.Fatalf("die Befunde sind %v", befunde)
	}
}

// **Und eine unvollständige nennt, was fehlt.**
//
// **Jedes Feld einzeln**, denn eine Messung ohne Modell ist etwas anderes
// als eine ohne Datum: die erste sagt nicht, **worüber** sie spricht, die
// zweite nicht, **wann**.
func TestEineUnvollstaendigeMessungNenntWasFehlt(t *testing.T) {
	t.Parallel()

	for name, leeren := range map[string]func(*domain.KIModellprobe){
		"ohne Anbieter": func(p *domain.KIModellprobe) { p.Anbieter = "" },
		"ohne Alias":    func(p *domain.KIModellprobe) { p.Alias = "" },
		"ohne Laufzeit": func(p *domain.KIModellprobe) { p.Laufzeit = "" },
		"ohne Modell":   func(p *domain.KIModellprobe) { p.Modell = "  " },
		"ohne Datum":    func(p *domain.KIModellprobe) { p.Datum = "" },
		"ohne Beleg":    func(p *domain.KIModellprobe) { p.Beleg = "" },
		"mit erfundener Fähigkeit": func(p *domain.KIModellprobe) {
			p.Faehigkeit = "hellsehen"
		},
		"mit unlesbarem Datum": func(p *domain.KIModellprobe) {
			p.Datum = "neulich"
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if befunde := eineProbe(leeren).Validate(); len(befunde) == 0 {
				t.Error("die unvollständige Messung gilt als gültig")
			}
		})
	}
}

// **Eine Messung ergänzt den Katalog — und ersetzt ihn nicht.**
func TestEineMessungErgaenztDenKatalog(t *testing.T) {
	t.Parallel()

	modell := domain.KIModell{
		Alias: "gpt-probe", Anbieter: "openai", Ort: domain.KIOrtCloud,
		Laufzeit: "Responses-API /v1/responses",
	}

	// **Vorher kann es nichts** — das ist die Gegenprobe, ohne die der Lauf
	// darunter nur sagte, dass irgendetwas wahr ist.
	if modell.Kann(domain.KIStrukturierteGenerierung) {
		t.Fatal("ein Modell ohne Beleg gilt als fähig")
	}

	ergaenzt := domain.MitProben([]domain.KIModell{modell},
		[]domain.KIModellprobe{eineProbe()})

	if len(ergaenzt) != 1 ||
		!ergaenzt[0].Kann(domain.KIStrukturierteGenerierung) {
		t.Fatalf("die Messung kam nicht an: %+v", ergaenzt)
	}

	// **Und der Beleg reist mit.** „Geprüft" ohne das Wie ist eine
	// Behauptung, die niemand nachvollziehen kann (§10).
	if ergaenzt[0].Geprueft[0].Beleg == "" ||
		ergaenzt[0].Geprueft[0].Datum != "2026-09-12" {
		t.Errorf("der Beleg ist %+v", ergaenzt[0].Geprueft[0])
	}
}

// **Eine Messung zu einem anderen Anbieter greift nicht.**
//
// **Zwei Anbieter dürfen denselben Namen führen** — genau deshalb steht der
// Anbieter seit A8 überall daneben.
func TestEineMessungZuEinemAnderenAnbieterGreiftNicht(t *testing.T) {
	t.Parallel()

	modell := domain.KIModell{
		Alias: "gpt-probe", Anbieter: "einanderer", Ort: domain.KIOrtCloud,
		Laufzeit: "x",
	}

	ergaenzt := domain.MitProben([]domain.KIModell{modell},
		[]domain.KIModellprobe{eineProbe()})

	if ergaenzt[0].Kann(domain.KIStrukturierteGenerierung) {
		t.Error("die fremde Messung wurde übernommen")
	}
}

// **Und eine unvollständige Messung greift ebenso wenig.**
//
// **Sonst käme ein Eintrag in den Katalog, den [service.NeuerModellkatalog]
// abweist** — und mit ihm der ganze Katalog, auch die Modelle der anderen
// Anbieter (A8).
func TestEineUnvollstaendigeMessungGreiftNicht(t *testing.T) {
	t.Parallel()

	modell := domain.KIModell{
		Alias: "gpt-probe", Anbieter: "openai", Ort: domain.KIOrtCloud,
		Laufzeit: "x",
	}

	ergaenzt := domain.MitProben([]domain.KIModell{modell},
		[]domain.KIModellprobe{eineProbe(func(p *domain.KIModellprobe) {
			p.Beleg = ""
		})})

	if ergaenzt[0].Kann(domain.KIStrukturierteGenerierung) {
		t.Error("die unvollständige Messung wurde übernommen")
	}

	if len(ergaenzt[0].Geprueft) != 0 {
		t.Errorf("sie steht trotzdem am Modell: %+v", ergaenzt[0].Geprueft)
	}
}
