package app

// **Jeder Schritt des Räumens fängt seine eigene Panik** — Befund 2 des
// Nachlaufs über Phase 1 von 0.9.0 (F02, #2341).

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinSchrittInPanikLaesstDieUebrigenLaufen(t *testing.T) {
	var gesammelt bytes.Buffer

	vorher := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&gesammelt, nil)))
	t.Cleanup(func() { slog.SetDefault(vorher) })

	gelaufen := []string{}

	verloren := schritteFahren([]raeumschritt{
		{"der erste", func() { gelaufen = append(gelaufen, "erste") }},
		{"das Gerätemodell", func() { panic("Probepanik im Gerätemodell") }},
		{"der dritte", func() { gelaufen = append(gelaufen, "dritte") }},
	})

	if !slices.Equal(gelaufen, []string{"erste", "dritte"}) {
		t.Errorf("gelaufen sind %v, erwartet beide ohne Panik", gelaufen)
	}

	if !slices.Equal(verloren, []string{"das Gerätemodell"}) {
		t.Errorf("verloren ist %v, erwartet der Schritt in Panik", verloren)
	}

	if zeile := gesammelt.String(); !strings.Contains(zeile, "Gerätemodell") {
		t.Errorf("die Panik steht nicht im Protokoll:\n%s", zeile)
	}
}

// werfenderGriff ist ein Tisch als Gast, dessen jeder Ruf in Panik endet —
// die eingebettete Schnittstelle ist leer.
type werfenderGriff struct{ service.Angenommene }

// **Durch den Schluss**: eine Panik beim Aufstehen vom fremden Tisch — der
// erste Schritt — lässt die Teilungen, den letzten, trotzdem räumen, und der
// Schluss nennt den Schritt.
func TestDerSchlussRaeumtHinterEinemSchrittInPanikWeiter(t *testing.T) {
	anwendung := gestartet(t)
	anwendung.partie.beitritt.griff = werfenderGriff{}
	vorher := anwendung.teilen.generation

	frist, aufhoeren := context.WithTimeout(context.Background(), 2*time.Second)
	defer aufhoeren()

	fehler := anwendung.Schliessen(frist, nil)

	if !errors.Is(fehler, ErrSchlussNachFrist) || !strings.Contains(fehler.Error(), "der Tisch als Gast") {
		t.Errorf("Schliessen gab %v, erwartet der benannte Schritt", fehler)
	}

	if anwendung.teilen.generation == vorher {
		t.Error("hinter der Panik wurde nicht weiter geräumt — die Teilungen blieben")
	}

	if anwendung.db != nil {
		t.Error("der Bestand ist nach dem Schluss noch offen")
	}
}
