package app

// **Das Fernhorchen gehört dem Profil, das es geöffnet hat** — Befund N2
// des Nachlaufs über Phase 6.7 (NET-Teil).
//
// Das gemeinsame Räumen beim Wechsel, Sperren und Abmelden beendete das
// Horchen im eigenen Netz und die Dienste, das Fernhorchen nicht: es lief mit
// dem Profil seines Starts weiter und nahm für es an.

import (
	"testing"
	"time"
)

func TestDerWechselBeendetDasFernhorchen(t *testing.T) {
	t.Parallel()

	horcher := &probefernhorcher{einladung: "mse4mac:1:tc1abc"}
	anwendung, ctx := eineFernhorchende(t, horcher, nil)

	t.Cleanup(anwendung.fernhorchenBeenden)

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	// **Unter der Umschaltsperre**, wie beim Wechsel — und ohne zu verklemmen.
	fertig := make(chan struct{})

	go func() {
		anwendung.umschaltung.Lock()
		defer anwendung.umschaltung.Unlock()

		anwendung.profilgebundenesRaeumen()
		close(fertig)
	}()

	select {
	case <-fertig:
	case <-time.After(5 * time.Second):
		t.Fatal("das Räumen verklemmte")
	}

	if stand := anwendung.fernstand(); stand.Laeuft || stand.Einladung != "" {
		t.Errorf("nach dem Räumen: %+v", stand)
	}
}
