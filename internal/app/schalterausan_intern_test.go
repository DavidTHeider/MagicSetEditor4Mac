package app

// **Aus und wieder an, an jedem Start mit einer Naht** — N2 des Nachlaufs
// über Phase 6.12
// (#1905),
// über den Dienst hinaus.
//
// Die Marke wird **vor** der blockierenden Arbeit genommen; nähme ein Start
// sie danach, sähe er das Abschalten nicht, und die Läufe mit Schalter schon
// aus aus Z2 blieben trotzdem grün. Deshalb hier je Start derselbe Ablauf:
// in der Naht festhalten, `netz` aus und wieder an, weiter.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ausUndAn schaltet `netz` aus und wieder an, während ein Start in seiner
// Naht steht, und wartet, bis er zurück ist.
func ausUndAn(anwendung *App, drin, weiter chan struct{}, starten func()) {
	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		starten()
	}()

	<-drin

	ctx := context.Background()
	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("netz"), domain.SchalterAus)
	anwendung.EinstellungSetzen(ctx, domain.SchalterEinstellung("netz"), domain.SchalterAn)

	close(weiter)
	<-fertig
}

func TestEinHorchpostenEndetNachAusUndAn(t *testing.T) {
	t.Parallel()

	drin, weiter := make(chan struct{}), make(chan struct{})

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	anwendung.MitHorcher(func(context.Context, string, peer.Identitaet) (service.Hereinkommend, error) {
		close(drin)
		<-weiter

		return &probelauscher{}, nil
	})
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})
	t.Cleanup(anwendung.horchenBeenden)

	ausUndAn(anwendung, drin, weiter, func() { anwendung.HorchStarten(ctx, ":0") })

	if anwendung.horchstand().Laeuft {
		t.Error("der Horchposten läuft, obwohl `netz` während des Starts aus war")
	}
}

func TestEinFernhorchenEndetNachAusUndAn(t *testing.T) {
	t.Parallel()

	drin, weiter := make(chan struct{}), make(chan struct{})

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	anwendung.MitFernhorcher(func(peer.Identitaet) (Fernhorcher, error) {
		close(drin)
		<-weiter

		return &probefernhorcher{einladung: "mse4mac:1:tc1abc"}, nil
	})
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})
	t.Cleanup(anwendung.fernhorchenBeenden)

	ausUndAn(anwendung, drin, weiter, func() { anwendung.FernHorchStarten(ctx) })

	if anwendung.fernstand().Laeuft {
		t.Error("das Fernhorchen läuft, obwohl `netz` während des Starts aus war")
	}
}

// **Hier halten zwei Riegel**, und eine Gegenprobe muss beide lösen: die
// Marke und der eigene Zähler der Ankündigung (`sichtbarkeit`, Nachlauf zum
// Zweitblick über Phase 5.4), den `sichtbarkeitBeenden` auch dann erhöht,
// wenn noch nichts läuft. Ohne die Marke allein bleibt dieser Lauf grün —
// gemessen, nicht vermutet.
func TestEineAnkuendigungEndetNachAusUndAn(t *testing.T) {
	t.Parallel()

	drin, weiter := make(chan struct{}), make(chan struct{})
	ankuendigung := &probeankuendigung{}

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)
	anwendung.MitEntdeckung(func(entdeckung.Angebot) (Ankuendigung, error) {
		close(drin)
		<-weiter

		return ankuendigung, nil
	}, nil)
	anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})
	t.Cleanup(anwendung.horchenBeenden)
	anwendung.HorchStarten(ctx, ":0")

	ausUndAn(anwendung, drin, weiter, func() { anwendung.SichtbarWerden(ctx, "Davids Rechner") })

	if anwendung.horchstand().Sichtbar || !ankuendigung.beendet {
		t.Error("die Ankündigung läuft, obwohl `netz` während des Starts aus war")
	}
}
