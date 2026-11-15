package app

// **Eine Panik im Schluss beendet den Prozess nicht** — F02
// (#2341),
// Befund A-02.
//
// # Warum
//
// [innerhalb] führt jeden Schritt des Schlusses in einer eigenen Goroutine
// aus, damit eine Frist ihn begrenzen kann. Eine Panik dort fängt niemand:
// warf das Merken der Fenstergröße, endete der ganze Prozess — mitten im
// Schluss und **bevor der Bestand geschlossen war**. Wails sieht diese
// Goroutine nicht; sein Panikweg hilft hier nicht.
//
// # Was die Prüfungen halten
//
// Eine Panik im Schritt ist ein Verlust wie eine Fristüberschreitung: der
// Schritt gilt als nicht erledigt, die Zeile steht im Protokoll, und der
// Schluss räumt danach den Bestand ab.

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestEinePanikImBestenfallsSchrittIstEinVerlust(t *testing.T) {
	var gesammelt bytes.Buffer

	vorher := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&gesammelt, nil)))
	t.Cleanup(func() { slog.SetDefault(vorher) })

	if Bestenfalls(t.Context(), "die Fenstergröße", func() { panic("Probepanik im Merken") }) {
		t.Error("ein Schritt, der in Panik endete, galt als erledigt")
	}

	if zeile := gesammelt.String(); !strings.Contains(zeile, "Probepanik im Merken") {
		t.Errorf("die Panik steht nicht im Protokoll:\n%s", zeile)
	}
}

func TestDerSchlussRaeumtTrotzPanikDenBestandAb(t *testing.T) {
	anwendung := gestartet(t)

	frist, aufhoeren := context.WithTimeout(context.Background(), 2*time.Second)
	defer aufhoeren()

	angefangen := time.Now()
	fehler := anwendung.Schliessen(frist, func(context.Context) { panic("Probepanik im Merken") })

	if dauer := time.Since(angefangen); dauer > 5*time.Second {
		t.Errorf("der Schluss brauchte %v", dauer)
	}

	if !errors.Is(fehler, ErrSchlussNachFrist) {
		t.Errorf("Schliessen gab %v, erwartet war der benannte Verlust", fehler)
	}

	// **Geschlossen heißt hier: abgeräumt** — `bestandSchliessen` lässt die
	// Datenbank fallen, nachdem sie zu ist.
	if anwendung.db != nil {
		t.Error("der Bestand ist nach dem Schluss noch offen")
	}
}
