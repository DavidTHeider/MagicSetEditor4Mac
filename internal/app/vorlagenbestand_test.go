package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was der Bestand der Ausfuhr nicht gibt** —
// #660.
//
// `vorlagenausfuhr_test.go` daneben prüft den Weg. Hier stehen die beiden
// Auskünfte, die fehlen können: die Karten und der Kopf des Sets — und sie
// wiegen verschieden schwer. Ohne Karten gibt es nichts auszuführen; ohne Kopf
// fehlt nur das Setsymbol.

// **Ein Lesefehler beim Holen der Karten wird gemeldet.**
//
// Er steht zwischen dem Suchen der Vorlage und ihrem Lauf; ohne diesen Fall
// bliebe der Zweig eine Behauptung darüber, was im Ernstfall geschieht (§3).
func TestVorlageAusfuehrenMeldetEinenKartenfehler(t *testing.T) {
	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
	})

	anwendung.MitBestand(kartenfehlerprobe{})

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// kartenfehlerprobe scheitert beim Lesen der Karten.
type kartenfehlerprobe struct{ leseprobe }

func (kartenfehlerprobe) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, errors.New("die Karten sind nicht lesbar")
}

// **Ein Set ohne lesbaren Kopf hält die Ausfuhr nicht auf** —
// #660.
//
// Der Ort des Sets sagt `symbol_variation`, wo `set.symbol` liegt. Ist er nicht
// zu ermitteln, fehlt das Setsymbol — die Ausfuhr selbst läuft, und das ist
// mehr als gar keine (§8).
func TestEinUnlesbarerSetkopfHaeltDieAusfuhrNichtAuf(t *testing.T) {
	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
	}

	anwendung := mitVorlagen(t, probe)
	anwendung.MitBestand(kopfloserBestand{})

	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", ziel, nil, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if probe.setpaket != "" {
		t.Errorf("Setort %q — erwartet keinen", probe.setpaket)
	}
}

// kopfloserBestand gibt Karten, aber keine Setköpfe.
type kopfloserBestand struct{}

func (kopfloserBestand) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return nil, errors.New("die Kopfliste ist nicht lesbar")
}

func (kopfloserBestand) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

func (kopfloserBestand) Rohbloecke(context.Context, int64) ([]domain.RawBlock, error) {
	return nil, nil
}

// **Ein Beigabenort, der sich nicht anlegen lässt, bricht die Ausfuhr ab** —
// #732.
//
// Eine Ausfuhr, die ihre Beigaben nicht ablegen kann, ist keine (§8). Der Weg
// durch die Fassade wird hier geprüft; dass `beigabenort` selbst meldet, steht
// in `beigabenort_intern_test.go`.
func TestEinUnanlegbarerBeigabenortBrichtDieAusfuhrAb(t *testing.T) {
	vorlage := probevorlage()
	vorlage.Verzeichnis = true

	anwendung := mitVorlagen(t, &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{vorlage},
		ausgabe:  "fertig",
	})

	verzeichnis := t.TempDir()

	// **Eine Datei an der Stelle des Verzeichnisses** — dasselbe, was ein
	// voller oder schreibgeschützter Datenträger auslöst.
	if err := os.WriteFile(
		filepath.Join(verzeichnis, "liste-files"), []byte("keine Mappe"), 0o600,
	); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	antwort := anwendung.VorlageAusfuehren(t.Context(), "1",
		"probe.mse-export-template", filepath.Join(verzeichnis, "liste.txt"), nil, nil)
	if antwort.Fehler == nil {
		t.Fatal("die Ausfuhr ging durch")
	}
}
