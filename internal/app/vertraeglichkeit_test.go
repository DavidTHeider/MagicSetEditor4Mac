package app_test

// **Kann die neueste Fassung diesen Bestand öffnen?** — Werkstück O6 von
// Phase 5.2 (#1430).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ledgerdoppel gibt einen festen Ledger.
type ledgerdoppel struct{ stand map[string]string }

func (l ledgerdoppel) Ledger(context.Context) (map[string]string, error) {
	return l.stand, nil
}

// neuereFreigabe ist eine Freigabe mit einer Fassung über der laufenden.
func neuereFreigabe() aktualisierung.Freigabe {
	return aktualisierung.Freigabe{
		Fassung: "99.0.0",
		Abbild:  aktualisierung.Beigabe{Name: "x.dmg", Groesse: 10},
	}
}

// einstellungsdoppel gibt feste Einstellungen.
type einstellungsdoppel struct{ werte map[string]string }

func (e einstellungsdoppel) Alle(context.Context) (map[string]string, error) {
	return e.werte, nil
}

// **Die zwei anderen Verfahren gehören zur Naht und nicht zu dieser Frage.**
// Sie stehen hier, weil Go es verlangt, und tun nichts.
func (einstellungsdoppel) Setzen(context.Context, string, string) error { return nil }

func (einstellungsdoppel) Stand(context.Context) (map[string]service.Wahl, error) {
	return map[string]service.Wahl{}, nil
}

// mitPruefung baut eine Anwendung über einem Anleger und einem Prüfer.
//
// **Mit eingeschalteter Aktualisierung**, denn ohne sie geht dieser Weg
// gar nicht erst ins Netz — die Vorgabe ist „aus" (§16, Regel 6), und ein
// eigener Lauf hält das fest.
func mitPruefung(t *testing.T, anleger *freigabeprobe, pruefer *service.Vorpruefer) *app.App {
	t.Helper()

	anwendung := app.NeueApp("")
	anwendung.MitFassungen(anleger)
	anwendung.MitVorpruefer(pruefer)
	anwendung.MitEinstellungen(einstellungsdoppel{werte: map[string]string{
		domain.EinstellungAktualisierung: domain.AktualisierungMelden,
	}})

	return anwendung
}

// **Ohne eingeschaltete Aktualisierung wird nicht gefragt** — und zwar
// wirklich nicht: der Anleger wird gar nicht gerufen.
func TestOhneEingeschalteteAktualisierungWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{freigabe: neuereFreigabe()}

	anwendung := app.NeueApp("")
	anwendung.MitFassungen(anleger)
	anwendung.MitVorpruefer(service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"))
	anwendung.MitEinstellungen(einstellungsdoppel{werte: map[string]string{}})

	antwort := anwendung.Vertraeglichkeit(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case antwort.Urteil != "":
		t.Errorf("Urteil = %q", antwort.Urteil)
	case anleger.gefragt != 0:
		t.Errorf("der Anleger wurde %d-mal gefragt", anleger.gefragt)
	}
}

// **Der ganze Weg: Freigabe, signierte Liste, Manifest, Ledger.**
func TestEinePassendeFassungIstBereit(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{
		freigabe: neuereFreigabe(),
		summen:   []byte("egal"),
		manifest: domain.Fassungsmanifest{
			Fassung: "99.0.0", Plattform: "probe/probe",
			Einstellungen: domain.AustauschFassung, Protokoll: domain.Protokollfassung,
			Wanderungen: []domain.Wanderungsangabe{{Name: "0001.sql", Summe: "aaa"}},
		},
	}

	pruefer := service.NeuerVorpruefer(
		ledgerdoppel{stand: map[string]string{"0001.sql": "aaa"}}, "probe/probe").
		MitPlatz(platzdoppel{frei: 1 << 30}).
		MitVorgaengen(vorgangsdoppel{}).
		MitSicherungsprobe(sicherungsdoppel{})

	antwort := mitPruefung(t, anleger, pruefer).Vertraeglichkeit(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Vertraeglichkeit: %+v", antwort.Fehler)
	case antwort.Urteil != string(domain.UrteilVertraeglich):
		t.Errorf("Urteil = %q", antwort.Urteil)
	case !antwort.Bereit:
		t.Errorf("nicht bereit: %+v", antwort)
	case antwort.Fassung != "99.0.0":
		t.Errorf("Fassung = %q", antwort.Fassung)
	}
}

// **Eine unbekannte Wanderung weist ab, und die Antwort nennt sie** — die
// erste Abnahme aus dem Issue, am Fenster.
func TestEineUnbekannteWanderungWeistDenWechselAb(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{
		freigabe: neuereFreigabe(),
		summen:   []byte("egal"),
		manifest: domain.Fassungsmanifest{
			Fassung: "99.0.0", Plattform: "probe/probe",
			Einstellungen: domain.AustauschFassung, Protokoll: domain.Protokollfassung,
			Wanderungen: []domain.Wanderungsangabe{{Name: "0001.sql", Summe: "aaa"}},
		},
	}

	pruefer := service.NeuerVorpruefer(
		ledgerdoppel{stand: map[string]string{"0002_neu.sql": "bbb"}}, "probe/probe").
		MitPlatz(platzdoppel{frei: 1 << 30}).
		MitVorgaengen(vorgangsdoppel{}).
		MitSicherungsprobe(sicherungsdoppel{})

	antwort := mitPruefung(t, anleger, pruefer).Vertraeglichkeit(t.Context())

	switch {
	case antwort.Bereit:
		t.Fatalf("der Wechsel war bereit: %+v", antwort)
	case len(antwort.Hindernisse) != 1:
		t.Fatalf("Hindernisse = %+v", antwort.Hindernisse)
	case antwort.Hindernisse[0].Art != string(domain.HindernisWanderungFehlt):
		t.Errorf("Art = %q", antwort.Hindernisse[0].Art)
	case antwort.Hindernisse[0].Stelle != "0002_neu.sql":
		t.Errorf("Stelle = %q", antwort.Hindernisse[0].Stelle)
	}
}

// **Ohne Manifest bleibt es ungeklärt** — die zweite Abnahme: das ist kein
// Nein zur Fassung, sondern eines zum automatischen Weg.
func TestEineFreigabeOhneManifestBleibtUngeklaert(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{freigabe: neuereFreigabe(), summen: []byte("egal")}

	pruefer := service.NeuerVorpruefer(
		ledgerdoppel{stand: map[string]string{"0001.sql": "aaa"}}, "probe/probe")

	antwort := mitPruefung(t, anleger, pruefer).Vertraeglichkeit(t.Context())

	switch {
	case antwort.Urteil != string(domain.UrteilUngeklaert):
		t.Errorf("Urteil = %q", antwort.Urteil)
	case antwort.Bereit:
		t.Error("ungeklärt war bereit")
	case len(antwort.Hindernisse) != 1 ||
		antwort.Hindernisse[0].Art != string(domain.HindernisKeinManifest):
		t.Errorf("Hindernisse = %+v", antwort.Hindernisse)
	}
}

// **Ohne Prüfer bleibt es auch ungeklärt** — eine Anwendung, die den Ledger
// nicht lesen kann, weiß es nicht.
func TestOhnePrueferBleibtEsUngeklaert(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Vertraeglichkeit(t.Context())

	if antwort.Urteil != string(domain.UrteilUngeklaert) || antwort.Bereit {
		t.Errorf("Antwort = %+v", antwort)
	}
}

// **Und ein Fehler auf jedem der drei Wege kommt an** (§8).
func TestJederFehlerAufDemWegKommtAn(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("es klemmt")

	faelle := map[string]*freigabeprobe{
		"die Freigabe":   {fehler: klemmt},
		"die Prüfsummen": {freigabe: neuereFreigabe(), summenfehler: klemmt},
		"das Manifest": {
			freigabe: neuereFreigabe(), summen: []byte("egal"), manifestfehler: klemmt,
		},
	}

	pruefer := service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe")

	for name, anleger := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := mitPruefung(t, anleger, pruefer).Vertraeglichkeit(t.Context())
			if antwort.Fehler == nil {
				t.Errorf("der Fehler ging verloren: %+v", antwort)
			}
		})
	}
}

// **Wer schon die neueste hat, bekommt kein Urteil** — über einen Wechsel,
// den es nicht gibt, ist nichts zu sagen.
func TestOhneNeuereFassungGibtEsKeinUrteil(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{
		freigabe: aktualisierung.Freigabe{Fassung: "0.0.1"},
	}

	antwort := mitPruefung(t, anleger,
		service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe")).Vertraeglichkeit(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case antwort.Urteil != "":
		t.Errorf("Urteil = %q", antwort.Urteil)
	case antwort.Hindernisse == nil:
		t.Error("die Hindernisse kamen als nil")
	}
}

// platzdoppel, vorgangsdoppel und sicherungsdoppel sind die drei Fragen des
// Augenblicks, jede mit einer festen Antwort.
type platzdoppel struct{ frei int64 }

func (p platzdoppel) FreierPlatz() (int64, error) { return p.frei, nil }

type vorgangsdoppel struct{ laufen int }

func (v vorgangsdoppel) LaufendeVorgaenge(context.Context) (int, error) {
	return v.laufen, nil
}

type sicherungsdoppel struct{}

func (sicherungsdoppel) SichereUndPruefe(context.Context) error { return nil }
