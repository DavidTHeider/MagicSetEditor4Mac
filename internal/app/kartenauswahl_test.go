package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Kartenauswahl an der Fassade** — Werkstück 4 von Phase 3.6
// (#1167).

// **Was die Seite anhakt, kommt beim Leger an.**
func TestDieAuswahlKommtBeimBogenlegerAn(t *testing.T) {
	t.Parallel()

	probe := &bogenprobe{lauf: &service.Bogenlauf{Pfad: "/tmp/b.pdf", Seiten: 1, Karten: 2}}

	antwort := mitBogenleger(probe).BogenLegen(
		probekontext(t), "1", "/tmp/b.pdf", "a4", 0, 300, false, []string{"7", "9"})
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(probe.auswahl.Karten) != 2 ||
		probe.auswahl.Karten[0] != 7 || probe.auswahl.Karten[1] != 9 {
		t.Errorf("angekommen ist %v", probe.auswahl.Karten)
	}
}

// **Ohne Auswahl kommt eine leere an** — und die heißt „das ganze Set".
func TestOhneAuswahlKommtEineLeereAn(t *testing.T) {
	t.Parallel()

	probe := &bogenprobe{lauf: &service.Bogenlauf{Pfad: "/tmp/b.pdf"}}

	antwort := mitBogenleger(probe).BogenLegen(
		probekontext(t), "1", "/tmp/b.pdf", "a4", 0, 300, false, nil)
	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if probe.auswahl.Filtert() {
		t.Errorf("die leere Auswahl filtert: %v", probe.auswahl.Karten)
	}
}

// **Eine unbrauchbare Kennung wird abgewiesen und nicht übergangen.**
//
// # Warum abgewiesen und nicht weggelassen
//
// Weil eine übergangene Kennung eine Auswahl ergäbe, die jemand anders gemeint
// hat — und bei genau einer unbrauchbaren wäre das die leere Auswahl, also das
// **ganze Set**. Aus einem Tippfehler würden dreihundert Seiten.
func TestEineUnbrauchbareKennungInDerAuswahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	probe := &bogenprobe{lauf: &service.Bogenlauf{Pfad: "/tmp/b.pdf"}}

	antwort := mitBogenleger(probe).BogenLegen(
		probekontext(t), "1", "/tmp/b.pdf", "a4", 0, 300, false, []string{"keine Zahl"})
	if antwort.Fehler == nil {
		t.Fatal("die unbrauchbare Kennung kam als Erfolg zurück")
	}

	if probe.pfad != "" {
		t.Error("der Leger wurde trotzdem gerufen")
	}
}

// **Und die Bindung reicht sie durch** — mit dem Kontext des Fensters.
func TestDieBogenbindungReichtDieAuswahlDurch(t *testing.T) {
	t.Parallel()

	probe := &bogenprobe{lauf: &service.Bogenlauf{Pfad: "/tmp/b.pdf"}}
	anwendung := mitBogenleger(probe)
	anwendung.MitDateiwahl(
		func(context.Context, app.Dateiauswahl) (string, error) { return "/tmp/b.pdf", nil })

	if antwort := app.NeueBindungen(anwendung).
		BogenLegen(t.Context(), "1", "a4", 0, 300, false, []string{"5"}); antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}

	if len(probe.auswahl.Karten) != 1 || probe.auswahl.Karten[0] != 5 {
		t.Errorf("angekommen ist %v", probe.auswahl.Karten)
	}
}

// **Auch die Vorlagenausfuhr nimmt eine Auswahl** — und weist eine
// unbrauchbare Kennung ab, bevor sie etwas schreibt.
func TestDieVorlagenausfuhrWeistEineUnbrauchbareAuswahlAb(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").VorlageAusfuehren(
		probekontext(t), "1", "p", "/tmp/ziel.txt", nil, []string{"keine Zahl"})
	if antwort.Fehler == nil {
		t.Fatal("die unbrauchbare Kennung kam als Erfolg zurück")
	}
}

// **Eine Auswahl, die keine Karte des Sets trifft, schreibt keine Ausfuhr.**
//
// # Warum das hier steht und nicht bei den Vorlagentests
//
// Weil die Zusage der **Auswahl** gehört und nicht der Vorlage: sie gilt für
// beide Ausgabewege, und wer sie bei einem prüfte, ließe den anderen offen.
func TestEineAuswahlOhneTrefferSchreibtKeineAusfuhr(t *testing.T) {
	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
	}

	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	antwort := mitVorlagen(t, probe).VorlageAusfuehren(probekontext(t), "1",
		"probe.mse-export-template", ziel, nil, []string{"999"})
	if antwort.Fehler == nil {
		t.Fatal("die leere Auswahl kam als Erfolg zurück")
	}

	if _, err := os.Stat(ziel); err == nil {
		t.Error("es wurde trotzdem geschrieben")
	}
}

// **Und die Bildausgabe weist eine unbrauchbare Kennung ab** — Befund 2 des
// Zweitblicks über Phase 3.6.
func TestDieBildausgabeWeistEineUnbrauchbareAuswahlAb(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Bildnamen(
		probekontext(t), "1", "{name}", "png", "nummerierenneu", []string{"keine Zahl"})
	if antwort.Fehler == nil {
		t.Fatal("die unbrauchbare Kennung kam als Erfolg zurück")
	}
}
