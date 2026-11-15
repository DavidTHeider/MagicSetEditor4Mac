package service_test

// **Die Vorprüfung vor einem Fassungswechsel** — Werkstück O6 von Phase 5.2
// (#1430).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

type ledgerprobe struct {
	stand  map[string]string
	fehler error
}

func (l ledgerprobe) Ledger(context.Context) (map[string]string, error) {
	return l.stand, l.fehler
}

type platzprobe struct {
	frei   int64
	fehler error
}

func (p platzprobe) FreierPlatz() (int64, error) { return p.frei, p.fehler }

type vorgangsprobe struct {
	laufen int
	fehler error
}

func (v vorgangsprobe) LaufendeVorgaenge(context.Context) (int, error) {
	return v.laufen, v.fehler
}

type sicherungsprobe struct{ fehler error }

func (s sicherungsprobe) SichereUndPruefe(context.Context) error { return s.fehler }

// passendesManifest kennt genau die eine verbuchte Wanderung.
func passendesManifest() domain.Fassungsmanifest {
	return domain.Fassungsmanifest{
		Fassung: "0.5.0", Plattform: "darwin/arm64",
		Einstellungen: domain.AustauschFassung, Protokoll: domain.Protokollfassung,
		Wanderungen: []domain.Wanderungsangabe{{Name: "0001_anfang.sql", Summe: "aaa"}},
	}
}

// vollstaendig baut einen Vorprüfer, dem nichts fehlt.
func vollstaendig() *service.Vorpruefer {
	return service.NeuerVorpruefer(
		ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
		MitPlatz(platzprobe{frei: 1 << 30}).
		MitVorgaengen(vorgangsprobe{}).
		MitSicherungsprobe(sicherungsprobe{})
}

func TestEinVollstaendigerVorpruefLaufIstBereit(t *testing.T) {
	t.Parallel()

	befund, fehler := vollstaendig().Pruefe(t.Context(), passendesManifest(), 100<<20)
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if !befund.Bereit || len(befund.Hindernisse) != 0 {
		t.Fatalf("Befund = %+v", befund)
	}
}

// **Jede der vier Fragen des Augenblicks nimmt die Bereitschaft.**
func TestJedeFrageDesAugenblicksNimmtDieBereitschaft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		pruefer *service.Vorpruefer
		art     domain.Hindernisart
	}{
		"zu wenig Platz": {
			pruefer: service.NeuerVorpruefer(
				ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
				MitPlatz(platzprobe{frei: 1 << 20}).
				MitVorgaengen(vorgangsprobe{}).
				MitSicherungsprobe(sicherungsprobe{}),
			art: domain.HindernisPlatz,
		},
		"es läuft noch etwas": {
			pruefer: service.NeuerVorpruefer(
				ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
				MitPlatz(platzprobe{frei: 1 << 30}).
				MitVorgaengen(vorgangsprobe{laufen: 2}).
				MitSicherungsprobe(sicherungsprobe{}),
			art: domain.HindernisVorgaengeLaufen,
		},
		"die Sicherung lässt sich nicht lesen": {
			pruefer: service.NeuerVorpruefer(
				ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
				MitPlatz(platzprobe{frei: 1 << 30}).
				MitVorgaengen(vorgangsprobe{}).
				MitSicherungsprobe(sicherungsprobe{fehler: errors.New("das Archiv ist kaputt")}),
			art: domain.HindernisSicherungUnlesbar,
		},
		// **Eine Frage, die niemand stellen kann, ist nicht beantwortet.**
		"gar keine Sicherungsprobe": {
			pruefer: service.NeuerVorpruefer(
				ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
				MitPlatz(platzprobe{frei: 1 << 30}).
				MitVorgaengen(vorgangsprobe{}),
			art: domain.HindernisKeineSicherung,
		},
		"gar keine Platzquelle": {
			pruefer: service.NeuerVorpruefer(
				ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
				MitVorgaengen(vorgangsprobe{}).
				MitSicherungsprobe(sicherungsprobe{}),
			art: domain.HindernisPlatz,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			befund, fehler := fall.pruefer.Pruefe(t.Context(), passendesManifest(), 100<<20)
			if fehler != nil {
				t.Fatalf("prüfen: %v", fehler)
			}

			switch {
			case befund.Bereit:
				t.Fatalf("der Lauf war bereit: %+v", befund)
			case len(befund.Hindernisse) != 1 || befund.Hindernisse[0].Art != fall.art:
				t.Errorf("Hindernisse = %+v, erwartet %q", befund.Hindernisse, fall.art)
			}
		})
	}
}

// **Bei „ungeklärt" wird der Augenblick nicht mehr geprüft** — eine Liste aus
// einem unbekannten Urteil und vier Nebenbefunden liest niemand.
func TestOhneManifestBleibtEsBeiEinemHindernis(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerVorpruefer(
		ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64")

	befund, fehler := pruefer.Pruefe(t.Context(), domain.Fassungsmanifest{}, 100<<20)
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	switch {
	case befund.Urteil != domain.UrteilUngeklaert:
		t.Errorf("Urteil = %q", befund.Urteil)
	case len(befund.Hindernisse) != 1:
		t.Errorf("Hindernisse = %+v", befund.Hindernisse)
	case befund.Bereit:
		t.Error("ungeklärt war bereit")
	}
}

// **Ein Fehler beim Fragen ist ein Fehler, kein Hindernis** — wer den Ledger
// nicht lesen kann, hat keine Antwort und soll keine erfinden.
func TestEinFehlerBeimFragenKommtAn(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("der Bestand klemmt")

	faelle := map[string]*service.Vorpruefer{
		"der Ledger": service.NeuerVorpruefer(
			ledgerprobe{fehler: klemmt}, "darwin/arm64"),
		"der Platz": service.NeuerVorpruefer(
			ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
			MitPlatz(platzprobe{fehler: klemmt}),
		"die Vorgänge": service.NeuerVorpruefer(
			ledgerprobe{stand: map[string]string{"0001_anfang.sql": "aaa"}}, "darwin/arm64").
			MitPlatz(platzprobe{frei: 1 << 30}).
			MitVorgaengen(vorgangsprobe{fehler: klemmt}),
	}

	for name, pruefer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := pruefer.Pruefe(
				t.Context(), passendesManifest(), 100<<20); !errors.Is(fehler, klemmt) {
				t.Errorf("Fehler = %v", fehler)
			}
		})
	}
}
