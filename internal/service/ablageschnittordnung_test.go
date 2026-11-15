// **Nichts wird gelöscht, was nicht abgelegt ist** — Befund 2 des Nachlaufs
// über Phase 3.7
// (#1168).
//
// # Der Fall
//
// `Ausschneiden` baute den Inhalt, löschte die Karten — und **erst danach**
// versuchte die Fassade, ihn ins Systemklemmbrett zu schreiben. Scheitert
// `ClipboardSetText`, sind die Karten fort und in der Ablage steht weiter der
// Inhalt von vorhin. Der Kommentar „kopiert wird zuerst, gelöscht danach"
// stimmte für die Zeichenkette und nicht für die beobachtbare Wirkung.
//
// **MSE2 macht es andersherum**: `doCut` ruft das wirkliche `doCopy()`
// einschließlich `wxTheClipboard->SetData` und löscht nur bei dessen Erfolg
// (`src/gui/control/item_list.cpp`, Zeilen 67–72; `card_list.cpp`,
// Zeilen 146–156; Fassung `f9d9356` vom 2021-07-06).
package service_test

import (
	"errors"
	"testing"
)

// **Scheitert das Ablegen, bleibt jede Karte stehen.**
func TestOhneAblageWirdNichtGeloescht(t *testing.T) {
	t.Parallel()

	loescher := &ablageloescher{}
	ablage := kartenablage(zweiSets()).MitLoescher(loescher)

	abgelehnt := errors.New("das Klemmbrett nimmt nichts an")

	_, err := ablage.Ausschneiden(t.Context(), 1, []int64{11},
		func(string) error { return abgelehnt })

	if !errors.Is(err, abgelehnt) {
		t.Fatalf("Fehler = %v, erwartet den des Ablegens", err)
	}

	if len(loescher.geloescht) != 0 {
		t.Fatalf("gelöscht wurden %v, erwartet nichts", loescher.geloescht)
	}
}

// **Und geglückt ist geglückt**: erst die Ablage, dann der Schnitt.
func TestMitAblageWirdGeloescht(t *testing.T) {
	t.Parallel()

	loescher := &ablageloescher{}
	ablage := kartenablage(zweiSets()).MitLoescher(loescher)

	var abgelegt string

	inhalt, err := ablage.Ausschneiden(t.Context(), 1, []int64{11},
		func(text string) error {
			abgelegt = text

			return nil
		})
	if err != nil {
		t.Fatalf("ausschneiden: %v", err)
	}

	if abgelegt != inhalt {
		t.Errorf("abgelegt wurde %q, herausgegeben %q", abgelegt, inhalt)
	}

	if len(loescher.geloescht) != 1 {
		t.Fatalf("gelöscht wurden %v, erwartet die eine", loescher.geloescht)
	}
}
