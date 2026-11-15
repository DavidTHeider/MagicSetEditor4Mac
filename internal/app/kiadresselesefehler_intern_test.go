package app

// **Ein Lesefehler ist keine fehlende Adresse** — Rest von Befund 3 des
// Zweitblicks über Phase D, angestrichen im Nachlauf.
//
// [App.dienstadresse] gab bei einem Lesefehler dasselbe wie ohne
// Einstellung: „". Die Freigabe eines KI-Auftrags nannte dann den Anbieter
// und seine Wolke — obwohl ein eingestelltes Gateway das Ziel sein konnte.
// Wer nicht weiß, wohin es geht, gibt keine Freigabe aus.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stolperstand scheitert am Lesen der Einstellungen.
type stolperstand struct{ Einstellungsfuehrer }

func (stolperstand) Stand(context.Context) (map[string]service.Wahl, error) {
	return nil, errors.New("der Bestand ist zu")
}

func TestEinLesefehlerGibtKeineFreigabeMitFalschemZiel(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(stolperstand{})
	anwendung.MitRecherche(&stummeRecherche{})

	antwort := anwendung.KontextZusammenstellen(t.Context(), "a", "",
		string(domain.FallRegeln), "b", 5, string(domain.DienstOpenAI),
		"gpt", 4096, 1024, 0, nil)

	if antwort.Fehler == nil {
		t.Fatalf("eine Freigabe ohne gelesene Adresse: %+v", antwort.Freigabe)
	}
}

// zielOrt gibt den Ort eines Anbieters — und bricht ab, wenn das Lesen
// scheitert: ein Test über den Ort soll nicht still an einem Lesefehler
// vorbeiprüfen.
func zielOrt(t *testing.T, a *App, ctx context.Context, anbieter string) domain.KIOrt {
	t.Helper()

	ort, _, fehler := a.zielFuer(ctx, anbieter)
	if fehler != nil {
		t.Fatalf("das Ziel von %s: %v", anbieter, fehler)
	}

	return ort
}
