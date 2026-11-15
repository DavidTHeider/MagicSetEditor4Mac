package app

// **Ein Lauf, der startet, während sein Schalter ausgeht, endet selbst** —
// Z2 des Zweitblicks über Phase 6.12
// (#1905).
//
// Das Tor fragt vor dem Start, das Abschalten beendet, was eingetragen ist.
// Zwischen beiden liegt der Start selbst — ein Lauscher, ein Tunnel, ein
// Prozess —, und wer in dieser Zeit abschaltete, fand nichts zu beenden.
// Danach trug der Start seinen Lauf ein, und er lief. **Deshalb fragt jeder
// Start nach dem Eintragen noch einmal**: entweder sieht das Abschalten den
// Lauf, oder der Lauf sieht den Schalter.
//
// Die Läufe hier rufen die Methoden der Anwendung unmittelbar, mit dem
// Schalter schon aus — das ist der Stand, den ein Start vorfindet, dessen
// Tor vor dem Abschalten durchließ.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schalterAus gibt Einstellungen, in denen diese Schalter aus sind.
func schalterAus(kennungen ...string) *merkendeEinstellungen {
	werte := map[string]string{}
	for _, kennung := range kennungen {
		werte[domain.SchalterEinstellung(kennung)] = domain.SchalterAus
	}

	return &merkendeEinstellungen{werte: werte}
}

func TestEinHorchpostenNachDemAbschaltenEndetSelbst(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	t.Cleanup(anwendung.horchenBeenden)
	anwendung.MitEinstellungen(schalterAus("netz"))

	anwendung.HorchStarten(ctx, ":0")

	if anwendung.horchstand().Laeuft {
		t.Error("der Horchposten läuft nach `netz` aus")
	}
}

func TestEinFernhorchenNachDemAbschaltenEndetSelbst(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	anwendung.MitFernhorcher(func(peer.Identitaet) (Fernhorcher, error) {
		return &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil
	})
	t.Cleanup(anwendung.fernhorchenBeenden)
	anwendung.MitEinstellungen(schalterAus("netz"))

	anwendung.FernHorchStarten(ctx)

	if anwendung.fernstand().Laeuft {
		t.Error("das Fernhorchen läuft nach `netz` aus")
	}
}

func TestEineAnkuendigungNachDemAbschaltenEndetSelbst(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	ankuendigung := &probeankuendigung{}
	anwendung.MitEntdeckung(func(entdeckung.Angebot) (Ankuendigung, error) {
		return ankuendigung, nil
	}, nil)
	t.Cleanup(anwendung.horchenBeenden)

	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})
	anwendung.HorchStarten(ctx, ":0")

	anwendung.MitEinstellungen(schalterAus("netz"))
	anwendung.SichtbarWerden(ctx, "Davids Rechner")

	if anwendung.horchstand().Sichtbar || !ankuendigung.beendet {
		t.Error("die Ankündigung läuft nach `netz` aus")
	}
}

func TestEinDienstNachDemAbschaltenEndetSelbst(t *testing.T) {
	t.Parallel()

	anwendung, _, _ := einePostenhalterin(t, "http://127.0.0.1:8123", "geheim", "", nil)
	t.Cleanup(anwendung.diensteBeenden)
	anwendung.MitEinstellungen(schalterAus("dienst"))

	if _, fehler := anwendung.dienstStarten(context.Background(),
		domain.DienstREST, domain.Bindung{}); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	if anwendung.dienststand(domain.DienstREST).Laeuft {
		t.Error("der Dienst läuft nach `dienst` aus")
	}
}

func TestEinBildauftragNachDemAbschaltenEndetSelbst(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(schalterAus("ki"))

	auftrag, fertig, fehler := anwendung.bildlaufAnmelden(context.Background(), 7)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(fertig)

	if auftrag.Err() == nil {
		t.Error("der Bildauftrag läuft nach `ki` aus")
	}
}

// abbrechenderAufbau sagt, ob sein Kontext beim Aufbau schon abgebrochen war.
type abbrechenderAufbau struct {
	Retrievalfuehrer

	abgebrochen bool
}

func (a *abbrechenderAufbau) Aufbauen(ctx context.Context, _ service.Aufbauauftrag) (service.Aufbaubericht, error) {
	a.abgebrochen = ctx.Err() != nil

	return service.Aufbaubericht{}, ctx.Err()
}

func (a *abbrechenderAufbau) Abbrechen() bool { return false }

// **Ein Indexaufbau bekommt seinen Abbruch, bevor er beginnt** — und `ki`
// aus erreicht ihn, auch wenn der Führer seinen eigenen noch nicht hat.
func TestEinIndexaufbauNachDemAbschaltenEndetSelbst(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(schalterAus("ki"))

	aufbau := &abbrechenderAufbau{}
	anwendung.MitRetrieval(aufbau)

	if antwort := anwendung.IndexAufbauen(context.Background(), "m", "q", "r", "", false); antwort.Fehler == nil || !aufbau.abgebrochen {
		t.Errorf("der Aufbau lief nach `ki` aus: %+v", antwort.Fehler)
	}
}

// **Und mit eingeschaltetem Schalter läuft er** — die Gegenprobe, dass die
// Frage nach dem Start nicht jeden Lauf beendet.
func TestEinLaufMitEingeschaltetemSchalterLaeuft(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	t.Cleanup(anwendung.horchenBeenden)
	anwendung.MitEinstellungen(schalterAus("spielen"))

	anwendung.HorchStarten(ctx, ":0")

	if !anwendung.horchstand().Laeuft {
		t.Error("ein fremder Schalter beendete den Horchposten")
	}
}
