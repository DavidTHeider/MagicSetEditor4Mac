package service_test

// **Ein Tischbild aus einem gelesenen Stand** — #2481 (Zweitblick über
// Phase 7 von 0.9.0, Befund 8): Sicht, Erlaubnis und Handlungen kamen aus
// drei Lesevorgängen und konnten drei Fassungen tragen.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// erlaubtVon fragt das Tischbild nach der Erlaubnis — die Frage, die es bis
// #2481 als eigenen Weg gab.
func erlaubtVon(
	l *service.Partieleitung, ctx context.Context, partie, fuer string,
) ([]string, error) {
	bild, fehler := l.Tischbild(ctx, partie, fuer)

	return bild.Erlaubt, fehler
}

// handlungenVon fragt es nach den Handlungen je Objekt.
func handlungenVon(
	l *service.Partieleitung, ctx context.Context, partie, fuer string,
) ([]domain.Objekthandlungen, error) {
	bild, fehler := l.Tischbild(ctx, partie, fuer)

	return bild.Handlungen, fehler
}

// wanderndeProbe zieht nach dem ersten Lesen weiter: jedes spätere sieht die
// Partie angehalten — als hätte ein anderer Zug dazwischen geschrieben.
type wanderndeProbe struct {
	*partieprobe

	gelesen int
}

func (w *wanderndeProbe) Partie(ctx context.Context, kennung string) (domain.Partie, error) {
	partie, fehler := w.partieprobe.Partie(ctx, kennung)

	w.gelesen++
	if w.gelesen > 1 {
		partie.Stand = domain.StandPausiert
		partie.Fassung++
	}

	return partie, fehler
}

// bildfaenger ist ein Fenster, das sich das letzte Tischbild merkt.
type bildfaenger struct {
	bild domain.Tischbild
}

func (*bildfaenger) Ereignis(context.Context, domain.Ereignis) error { return nil }

func (b *bildfaenger) Tischbild(_ context.Context, bild domain.Tischbild) error {
	b.bild = bild

	return nil
}

// **Ein Tischbild trägt Sicht, Erlaubnis und Handlungen derselben Fassung.**
func TestEinTischbildStammtAusEinemStand(t *testing.T) {
	t.Parallel()

	probe := &wanderndeProbe{partieprobe: &partieprobe{partien: map[string]domain.Partie{}}}
	leitung := service.NeuePartieleitung(probe, service.NeuerSandkasten(service.NeuerZufall()))
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	probe.gelesen = 0
	fenster := &bildfaenger{}

	if fehler := service.NeuerTischfunk("p1", leitung, nil).MitFenster("sitz-1", fenster).
		Verkuende(ctx, "sitz-2", service.Befehlsausgang{}); fehler != nil {
		t.Fatalf("Verkuende: %v", fehler)
	}

	bild := fenster.bild

	switch {
	case bild.Sicht.Stand != domain.StandLaeuft:
		t.Fatalf("die Sicht zeigt %q, erwartet die laufende Partie", bild.Sicht.Stand)
	case len(bild.Erlaubt) == 0:
		t.Error("die Sicht ist die einer laufenden Partie, die Erlaubnis die einer angehaltenen")
	case len(bild.Handlungen) == 0:
		t.Error("die Sicht ist die einer laufenden Partie, die Handlungen die einer angehaltenen")
	}
}
