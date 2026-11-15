// **Ausschneiden** — Werkstück 1 von Phase 3.7
// (#1168).
//
// **Eigene Datei**, aus demselben Grund wie `ablagebefunde_test.go` — und der
// Schnitt folgt derselben Frage: Ausschneiden ist Kopieren **und** Löschen,
// und was dabei schiefgehen kann, ist etwas anderes als beim Kopieren allein.
package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ausschneiden kopiert zuerst und löscht danach** — Werkstück 1 von
// Phase 3.7.
//
//	case ID_EDIT_CUT: current_panel->doCut(); break;
//
// MSE2 kennt `ID_EDIT_CUT` neben `ID_EDIT_COPY` über Karten
// (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06).
func TestAusschneidenKopiertUndLoescht(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	loescher := &ablageloescher{}

	inhalt, err := kartenablage(doppel).MitLoescher(loescher).
		Ausschneiden(t.Context(), 1, []int64{11}, ohneAblage)
	if err != nil {
		t.Fatalf("ausschneiden: %v", err)
	}

	if !strings.Contains(inhalt, "name: Erste") {
		t.Errorf("Inhalt:\n%s", inhalt)
	}

	if len(loescher.geloescht) != 1 || loescher.geloescht[0] != 11 {
		t.Errorf("gelöscht wurde %v", loescher.geloescht)
	}
}

// **Ohne Löscher wird nicht ausgeschnitten** — statt still zu kopieren.
func TestAusschneidenOhneLoescherWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := kartenablage(zweiSets()).Ausschneiden(t.Context(), 1, []int64{11}, ohneAblage); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", fehler)
	}
}

// **Eine gescheiterte Kopie löscht nichts.**
//
// Das ist der Grund für die Reihenfolge: andersherum wäre die Karte fort und
// die Ablage leer.
func TestEineGescheiterteKopieLoeschtNichts(t *testing.T) {
	t.Parallel()

	loescher := &ablageloescher{}

	if _, fehler := kartenablage(zweiSets()).MitLoescher(loescher).
		Ausschneiden(t.Context(), 1, []int64{99}, ohneAblage); fehler == nil {
		t.Fatal("die fremde Kennung kam als Erfolg zurück")
	}

	if len(loescher.geloescht) != 0 {
		t.Errorf("gelöscht wurde %v", loescher.geloescht)
	}
}

// **Und ein gescheitertes Löschen gibt den Inhalt trotzdem heraus.**
//
// Er steht dann in der Zwischenablage, und die Karte auch — das ist
// wiederholbar und kein Verlust.
func TestEinGescheitertesLoeschenGibtDenInhaltHeraus(t *testing.T) {
	t.Parallel()

	loescher := &ablageloescher{fehler: domain.ErrBelegt}

	inhalt, fehler := kartenablage(zweiSets()).MitLoescher(loescher).
		Ausschneiden(t.Context(), 1, []int64{11}, ohneAblage)
	if fehler == nil {
		t.Fatal("der Fehler kam als Erfolg zurück")
	}

	if !strings.Contains(inhalt, "name: Erste") {
		t.Errorf("Inhalt:\n%q", inhalt)
	}
}

// ablageloescher merkt sich, was gelöscht wurde.
type ablageloescher struct {
	geloescht []int64
	fehler    error
}

// **Ganz oder gar nicht** — der Doppelgänger hält dieselbe Zusage wie der
// echte Weg (Befund 4 des Zweitblicks über Phase 3.7): scheitert er, hat er
// nichts gelöscht.
func (l *ablageloescher) KartenLoeschen(
	_ context.Context, _ int64, karteIDs []int64, _ bool,
) error {
	if l.fehler != nil {
		return l.fehler
	}

	l.geloescht = append(l.geloescht, karteIDs...)

	return nil
}

// ohneAblage ist das Ablegen, das immer glückt — für Läufe, die etwas anderes
// prüfen als die Reihenfolge zwischen Ablage und Schnitt.
//
// **Die Reihenfolge selbst steht in `ablageschnittordnung_test.go`**, und sie
// ist die Zusage: nichts wird gelöscht, was nicht abgelegt ist (Befund 2 des
// Nachlaufs über Phase 3.7).
func ohneAblage(string) error { return nil }
