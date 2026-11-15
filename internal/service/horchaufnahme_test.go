package service_test

// **Was mit einer angenommenen Sitzung geschieht** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Die Lücke, die N2 gelassen hat
//
// **Der Horchposten nahm Sitzungen an und ließ sie fallen.** Er trug sie in
// die Übersicht ein, und dort standen sie — niemand las von ihnen, niemand
// schickte über sie. Für einen Handschlag reicht das; für eine Partie nicht.
//
// **Deshalb bekommt er eine Aufnahme**: einen Ruf je entstandener Sitzung, der
// sagt, was mit ihr geschehen soll. **Welche** das ist, entscheidet er nicht —
// das steht am Zweck der Sitzung, und ihn kennt der Aufrufer.

import (
	"context"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einmaligerLauscher gibt genau eine Verbindung und dann nichts mehr.
//
// **Über `horchdoppel`**, damit dieser Lauf die Frage stellt, für die es ihn
// gibt — und nicht ein zweites Doppel danebenstellt, das beim ersten neuen
// Feld abweicht.
func einmaligerLauscher(verbindung service.Verbindung) service.Hereinkommend {
	return &horchdoppel{
		folge: []func() (service.Verbindung, error){
			func() (service.Verbindung, error) { return verbindung, nil },
		},
	}
}

// **Eine angenommene Sitzung erreicht die Aufnahme — mit ihrem Zweck.**
func TestEineAngenommeneSitzungErreichtDieAufnahme(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	var (
		schloss    sync.Mutex
		aufgenomme []service.Sitzungsstand
	)

	posten := service.NeuerHorchposten(paar.dort, nil).MitAufnahme(
		func(_ context.Context, griff service.Angenommene) {
			schloss.Lock()
			defer schloss.Unlock()

			aufgenomme = append(aufgenomme, griff.Stand())
		})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), einmaligerLauscher(paar.dortigeLeitung))
	}()

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckPartie,
	); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	<-fertig

	schloss.Lock()
	defer schloss.Unlock()

	switch {
	case len(aufgenomme) != 1:
		t.Fatalf("%d Aufnahmen, erwartet 1", len(aufgenomme))
	case aufgenomme[0].Zweck != domain.ZweckPartie:
		t.Errorf("Zweck %q", aufgenomme[0].Zweck)
	case aufgenomme[0].Gegenueber != paar.hiesige.Kennung():
		t.Errorf("das Gegenüber stimmt nicht")
	}
}

// **Ohne Aufnahme geschieht nichts, und das ist kein Fehler.**
//
// **Die Gegenprobe** (§10): so lief der Posten bis hierher, und für einen
// Handschlag ohne Folgeverkehr ist das der richtige Betrieb.
func TestOhneAufnahmeNimmtDerPostenTrotzdemAn(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	posten := service.NeuerHorchposten(paar.dort, nil)

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), einmaligerLauscher(paar.dortigeLeitung))
	}()

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckPartie,
	); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	<-fertig

	if len(paar.dort.Alle()) != 1 {
		t.Errorf("%d Sitzungen in der Übersicht", len(paar.dort.Alle()))
	}
}
