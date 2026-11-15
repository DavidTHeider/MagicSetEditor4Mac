package app

// **Wer erreichbar machen darf** — Werkstück N7 von Phase 5.4
// (#1568).
//
// # Was hier gemessen wird
//
// **Dass die Mindestrolle aus `docs/AUTORISIERUNG.md` wirklich gilt.** Sie
// stand dort und wurde von keinem Ruf geprüft — eine Tabelle, die eine
// Schranke nennt, die nichts durchsetzt, verspricht mehr, als ihr Rumpf hält
// (§10).
//
// **Und dass ein Mitarbeiter weiterhin darf**: die neue Klasse steht in der
// Reihe **vor** `externSenden`, und das Issue verlangt ausdrücklich, dass
// sich am Verhalten nichts ändert.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// riegelprobe antwortet auf die Frage nach der Wirkung.
type riegelprobe struct {
	gefragt    domain.Wirkung
	verweigern bool
}

func (r *riegelprobe) Pruefe(_ context.Context, wirkung domain.Wirkung) error {
	r.gefragt = wirkung

	if r.verweigern {
		return domain.ErrErlaubnis
	}

	return nil
}

// **Alle drei Wege fragen — und sie fragen nach derselben Klasse.**
func TestJederWegZurErreichbarkeitFragtNachDerKlasse(t *testing.T) {
	for name, fahre := range map[string]func(*App, context.Context) *Fehler{
		"Horchposten": func(a *App, ctx context.Context) *Fehler {
			return a.HorchStarten(ctx, ":0").Fehler
		},
		"Ankündigung": func(a *App, ctx context.Context) *Fehler {
			return a.SichtbarWerden(ctx, "Probe").Fehler
		},
		"Ferneinladung": func(a *App, ctx context.Context) *Fehler {
			return a.FernHorchStarten(ctx).Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

			riegel := &riegelprobe{verweigern: true}
			anwendung.MitWache(riegel)

			if fehler := fahre(anwendung, t.Context()); fehler == nil {
				t.Error("der Riegel hat nichts aufgehalten")
			}

			if riegel.gefragt != domain.WirkungExternEmpfangen {
				t.Errorf("gefragt wurde nach %q", riegel.gefragt)
			}
		})
	}
}

// **Die Gegenprobe: ohne Verweigerung kommt der Weg weiter.**
//
// **Ohne sie sagte der Lauf oben nichts** — ein Weg, der aus einem anderen
// Grund scheitert, sähe genauso aus (§10: „eine Gegenprobe, die nichts
// ändert, ist keine").
func TestOhneVerweigerungGehtDerWegWeiter(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	riegel := &riegelprobe{}
	anwendung.MitWache(riegel)

	ctx := t.Context()

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	antwort := anwendung.HorchStarten(ctx, "127.0.0.1:0")
	if antwort.Fehler != nil {
		t.Fatalf("horchen: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { anwendung.HorchBeenden(ctx) })

	if !antwort.Stand.Laeuft {
		t.Error("der Horchposten läuft nicht")
	}

	if riegel.gefragt != domain.WirkungExternEmpfangen {
		t.Errorf("gefragt wurde nach %q", riegel.gefragt)
	}
}
