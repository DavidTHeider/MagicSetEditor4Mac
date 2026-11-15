package app_test

// **Einen Fassungswechsel vorbereiten** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Eine eigene Datei**, weil `vertraeglichkeit_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **ob
// eine Fassung passt**, hier, **ob der Tausch jetzt vorbereitet werden darf**
// — und das sind zwei Fragen, die auch im Quelltext getrennt sind.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Wechsel wird vorbereitet — und „bereit" heißt gesichert und
// zurückgelesen.**
func TestDerWechselWirdVorbereitet(t *testing.T) {
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

	antwort := mitPruefung(t, anleger, pruefer).WechselVorbereiten(t.Context())

	if !antwort.Bereit || antwort.Fehler != nil {
		t.Fatalf("Antwort = %+v", antwort)
	}
}

// **Ohne Sicherungsprobe ist er nicht bereit — und sagt warum.**
func TestOhneSicherungIstDerWechselNichtBereit(t *testing.T) {
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
		MitVorgaengen(vorgangsdoppel{})

	antwort := mitPruefung(t, anleger, pruefer).WechselVorbereiten(t.Context())

	switch {
	case antwort.Bereit:
		t.Fatalf("ohne Sicherung war er bereit: %+v", antwort)
	case len(antwort.Hindernisse) != 1 ||
		antwort.Hindernisse[0].Art != string(domain.HindernisKeineSicherung):
		t.Errorf("Hindernisse = %+v", antwort.Hindernisse)
	}
}

// **Und jeder Fehler auf dem Weg zum Vorbereiten kommt an** — dieselben drei
// Stellen wie beim Urteil, plus die Prüfung selbst.
func TestJederFehlerBeimVorbereitenKommtAn(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("es klemmt")

	faelle := map[string]struct {
		anleger *freigabeprobe
		pruefer *service.Vorpruefer
	}{
		"die Freigabe": {
			anleger: &freigabeprobe{fehler: klemmt},
			pruefer: service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"),
		},
		"die Prüfsummen": {
			anleger: &freigabeprobe{freigabe: neuereFreigabe(), summenfehler: klemmt},
			pruefer: service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"),
		},
		"das Manifest": {
			anleger: &freigabeprobe{
				freigabe: neuereFreigabe(), summen: []byte("egal"), manifestfehler: klemmt,
			},
			pruefer: service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"),
		},
		"der Ledger": {
			anleger: &freigabeprobe{
				freigabe: neuereFreigabe(), summen: []byte("egal"),
				manifest: domain.Fassungsmanifest{
					Fassung: "99.0.0", Plattform: "probe/probe",
					Einstellungen: domain.AustauschFassung, Protokoll: domain.Protokollfassung,
					Wanderungen: []domain.Wanderungsangabe{{Name: "0001.sql", Summe: "aaa"}},
				},
			},
			pruefer: service.NeuerVorpruefer(stolperledger{fehler: klemmt}, "probe/probe"),
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := mitPruefung(t, fall.anleger, fall.pruefer).WechselVorbereiten(t.Context())
			if antwort.Fehler == nil {
				t.Errorf("der Fehler ging verloren: %+v", antwort)
			}
		})
	}
}

// **Ohne eingeschaltete Aktualisierung wird auch nicht vorbereitet.**
func TestOhneEingeschalteteAktualisierungWirdNichtVorbereitet(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{freigabe: neuereFreigabe()}

	anwendung := app.NeueApp("")
	anwendung.MitFassungen(anleger)
	anwendung.MitVorpruefer(service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"))
	anwendung.MitEinstellungen(einstellungsdoppel{werte: map[string]string{}})

	if antwort := anwendung.WechselVorbereiten(t.Context()); antwort.Urteil != "" ||
		anleger.gefragt != 0 {
		t.Errorf("Antwort = %+v, gefragt %d", antwort, anleger.gefragt)
	}
}

// **Und ein Lesefehler der Einstellungen ist ein Fehler** — nicht „aus".
func TestEinLesefehlerDerEinstellungenKommtAn(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitFassungen(&freigabeprobe{})
	anwendung.MitVorpruefer(service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"))
	anwendung.MitEinstellungen(stolpereinstellungen{})

	if antwort := anwendung.Vertraeglichkeit(t.Context()); antwort.Fehler == nil {
		t.Error("der Lesefehler ging verloren")
	}
}

// stolperledger und stolpereinstellungen klemmen auf Ansage.
type stolperledger struct{ fehler error }

func (s stolperledger) Ledger(context.Context) (map[string]string, error) {
	return nil, s.fehler
}

type stolpereinstellungen struct{}

func (stolpereinstellungen) Alle(context.Context) (map[string]string, error) {
	return nil, errors.New("die Einstellungen klemmen")
}

func (stolpereinstellungen) Setzen(context.Context, string, string) error { return nil }

func (stolpereinstellungen) Stand(context.Context) (map[string]service.Wahl, error) {
	return map[string]service.Wahl{}, nil
}

// **Eine unlesbare Marke ist ein Fehler und kein Nein** — sie sagt nicht
// „nichts Neues", sondern „diese Freigabe ist nicht einzuordnen".
func TestEineUnlesbareMarkeIstEinFehler(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{
		freigabe: aktualisierung.Freigabe{Fassung: "keine-zahl"},
	}

	pruefer := service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe")
	anwendung := mitPruefung(t, anleger, pruefer)

	if antwort := anwendung.Vertraeglichkeit(t.Context()); antwort.Fehler == nil {
		t.Error("die unlesbare Marke kam als Nein durch")
	}
}

// **Ein klemmender Ledger ist ein Fehler, kein Urteil** — auch auf dem
// kurzen Weg. Ohne diesen Lauf wäre der Zweig eine Behauptung über den
// Ernstfall (§3).
func TestEinKlemmenderLedgerMeldetSichBeimUrteil(t *testing.T) {
	t.Parallel()

	anleger := &freigabeprobe{
		freigabe: neuereFreigabe(), summen: []byte("egal"),
		manifest: domain.Fassungsmanifest{
			Fassung: "99.0.0", Plattform: "probe/probe",
			Einstellungen: domain.AustauschFassung, Protokoll: domain.Protokollfassung,
			Wanderungen: []domain.Wanderungsangabe{{Name: "0001.sql", Summe: "aaa"}},
		},
	}

	pruefer := service.NeuerVorpruefer(
		stolperledger{fehler: errors.New("der Bestand klemmt")}, "probe/probe")

	if antwort := mitPruefung(t, anleger, pruefer).Vertraeglichkeit(
		t.Context()); antwort.Fehler == nil {
		t.Error("der klemmende Ledger kam als Urteil durch")
	}
}

// **Ohne Prüfer wird auch nicht vorbereitet**, und ein Lesefehler der
// Einstellungen ist dort ebenso ein Fehler.
func TestOhnePrueferWirdNichtVorbereitet(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").WechselVorbereiten(t.Context()); antwort.Urteil !=
		string(domain.UrteilUngeklaert) {
		t.Errorf("Urteil = %q", antwort.Urteil)
	}

	anwendung := app.NeueApp("")
	anwendung.MitFassungen(&freigabeprobe{})
	anwendung.MitVorpruefer(service.NeuerVorpruefer(ledgerdoppel{}, "probe/probe"))
	anwendung.MitEinstellungen(stolpereinstellungen{})

	if antwort := anwendung.WechselVorbereiten(t.Context()); antwort.Fehler == nil {
		t.Error("der Lesefehler beim Vorbereiten ging verloren")
	}
}
