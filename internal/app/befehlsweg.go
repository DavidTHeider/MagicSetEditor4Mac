package app

// **Die Kommandozeile aus den Einstellungen** — OPS-03
// (#1720),
// Marke `new`: MSE2 hat keine Kommandozeile, die sich einrichten ließe.
//
// # Was es tut
//
// **Verknüpfungen in einem Verzeichnis des Nutzers** — `~/.local/bin` —, die
// auf `mse` und `mse-mcp` im Bündel zeigen. **Ohne Administratorrechte**, und
// **ohne das Shellprofil anzufassen**: steht das Verzeichnis nicht im `PATH`,
// nennt die Vorschau die Zeile, die hineingehört, und der Mensch trägt sie
// selbst ein (OPS-03.AK-02). Ein Programm, das Profile umschreibt, hätte eine
// Rücknahme zu verantworten, die es nicht sehen kann.
//
// # Was es nie tut
//
// **Etwas überschreiben, das es nicht selbst angelegt hat.** Eigen ist eine
// Verknüpfung, die in ein Bündel `MagicSetEditor4Mac*.app` zeigt; alles
// andere ist fremd und bleibt (AK-01, AK-03). Eine eigene, die ins Leere oder
// in ein anderes Bündel zeigt — nach dem Verschieben oder Erneuern der
// Anwendung —, ist veraltet: Installieren ersetzt sie, Entfernen nimmt sie.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// befehlsnamen sind die Befehle, die sich einrichten lassen — der Dienst
// `mse-rest` gehört nicht dazu: er wird gestartet, nicht getippt.
var befehlsnamen = []string{"mse", "mse-mcp"}

// Die Stände eines Befehls im Zielverzeichnis.
const (
	befehlFrei     = "frei"
	befehlEigen    = "eigen"
	befehlVeraltet = "veraltet"
	befehlFremd    = "fremd"
)

// Befehlseintrag ist ein Befehl im Zielverzeichnis.
type Befehlseintrag struct {
	Name   string `json:"name"`
	Ziel   string `json:"ziel"`
	Quelle string `json:"quelle"`

	// Stand ist frei, eigen, veraltet oder fremd.
	Stand string `json:"stand"`

	// Verdeckt nennt einen gleichnamigen Befehl, der im PATH vorn steht.
	Verdeckt string `json:"verdeckt"`
}

// Befehlsvorschau sagt, was eine Einrichtung täte — und wie es jetzt steht.
type Befehlsvorschau struct {
	Verzeichnis string `json:"verzeichnis"`
	ImPfad      bool   `json:"imPfad"`

	// Pfadzeile gehört ins Shellprofil, wenn das Verzeichnis nicht im PATH
	// steht — leer, wenn es dort steht.
	Pfadzeile string `json:"pfadzeile"`

	// Pruefbefehl ist eine ungefährliche Abfrage (AK-04).
	Pruefbefehl string           `json:"pruefbefehl"`
	Eintraege   []Befehlseintrag `json:"eintraege"`
	Fehler      *Fehler          `json:"fehler"`

	// pfad ist der PATH, gegen den gelesen wurde — für das Nachlesen.
	pfad string
}

// befehlsvorschau liest, wie es im Verzeichnis steht.
func befehlsvorschau(verzeichnis, ort, pfad string) Befehlsvorschau {
	eintraege := make([]Befehlseintrag, 0, len(befehlsnamen))
	for _, name := range befehlsnamen {
		quelle := filepath.Join(ort, name)
		ziel := filepath.Join(verzeichnis, name)
		eintraege = append(eintraege, Befehlseintrag{
			Name: name, Ziel: ziel, Quelle: quelle,
			Stand: befehlsstand(ziel, quelle), Verdeckt: verdecktVon(name, verzeichnis, pfad),
		})
	}

	v := Befehlsvorschau{
		Verzeichnis: verzeichnis, Pruefbefehl: "mse meta", Eintraege: eintraege,
		ImPfad: slices.Contains(filepath.SplitList(pfad), verzeichnis), pfad: pfad,
	}
	if !v.ImPfad {
		v.Pfadzeile = fmt.Sprintf("export PATH=%q:\"$PATH\"", verzeichnis)
	}

	return v
}

// befehlsstand ordnet ein, was am Ziel liegt.
func befehlsstand(ziel, quelle string) string {
	angaben, err := os.Lstat(ziel)
	if err != nil {
		return befehlFrei
	}

	verweis, err := os.Readlink(ziel)
	if angaben.Mode()&os.ModeSymlink == 0 || err != nil || !eigeneVerknuepfung(verweis) {
		return befehlFremd
	}

	if _, err := os.Stat(verweis); err != nil || verweis != quelle {
		return befehlVeraltet
	}

	return befehlEigen
}

// eigeneVerknuepfung sagt, ob ein Verweis in ein Bündel dieser Anwendung zeigt.
//
// **Gibt es das Bündel, entscheidet seine Kennung** ([service.Buendelzustand],
// Zweitblick über Phase 4, Befund 8, und sein Nachlauf) — umbenannt darf es
// sein, wie bei den Einträgen der Klienten. **Gibt es es nicht mehr**, bleibt
// nur die Gestalt `MagicSetEditor4Mac*.app/Contents/MacOS`: dann ist die
// Anwendung umgezogen, und „Alles widerrufen" soll ihre Verknüpfungen
// trotzdem nehmen (OPS03).
func eigeneVerknuepfung(verweis string) bool {
	if vorhanden, eigen := service.Buendelzustand(verweis); vorhanden {
		return eigen
	}

	macos := filepath.Dir(verweis)
	buendel := filepath.Base(filepath.Dir(filepath.Dir(macos)))

	return filepath.Base(macos) == "MacOS" &&
		strings.HasPrefix(buendel, "MagicSetEditor4Mac") && strings.HasSuffix(buendel, ".app")
}

// verdecktVon nennt einen gleichnamigen Befehl, der im PATH vor dem
// Zielverzeichnis steht — oder irgendwo, wenn das Ziel dort gar nicht steht.
func verdecktVon(name, verzeichnis, pfad string) string {
	for _, stelle := range filepath.SplitList(pfad) {
		if stelle == verzeichnis {
			return ""
		}

		kandidat := filepath.Join(stelle, name)
		//nolint:gosec // G703: Orte aus dem PATH werden angesehen, nicht geöffnet
		// #nosec G703 -- dieselbe Begründung für den eigenständigen gosec
		if angaben, err := os.Stat(kandidat); err == nil && !angaben.IsDir() && angaben.Mode()&0o111 != 0 {
			return kandidat
		}
	}

	return ""
}

// befehleInstallieren legt die Verknüpfungen an, die frei oder veraltet sind.
func befehleInstallieren(v Befehlsvorschau) (Befehlsvorschau, error) {
	if err := os.MkdirAll(v.Verzeichnis, 0o750); err != nil {
		return v, domain.MitGrund(domain.ErrValidation, err, "das Verzeichnis für die Befehle anlegen")
	}

	var fehler error
	for _, eintrag := range v.Eintraege {
		if eintrag.Stand == befehlVeraltet {
			fehler = errors.Join(fehler, os.Remove(eintrag.Ziel))
		}

		if eintrag.Stand == befehlFrei || eintrag.Stand == befehlVeraltet {
			fehler = errors.Join(fehler, os.Symlink(eintrag.Quelle, eintrag.Ziel))
		}
	}

	return neuGelesen(v), fehler
}

// befehleEntfernen nimmt nur Verknüpfungen, die nachweislich eigene sind.
func befehleEntfernen(v Befehlsvorschau) (Befehlsvorschau, error) {
	var fehler error
	for _, eintrag := range v.Eintraege {
		if eintrag.Stand == befehlEigen || eintrag.Stand == befehlVeraltet {
			fehler = errors.Join(fehler, os.Remove(eintrag.Ziel))
		}
	}

	return neuGelesen(v), fehler
}

// neuGelesen liest die Vorschau nach einer Änderung neu — mit demselben Ort
// und demselben PATH.
func neuGelesen(v Befehlsvorschau) Befehlsvorschau {
	return befehlsvorschau(v.Verzeichnis, filepath.Dir(v.Eintraege[0].Quelle), v.pfad)
}

// Befehlszeile richtet die Kommandozeile ein. **Leer liest die Vorschau**;
// »installieren« und »entfernen« schreiben — nur in `~/.local/bin` des
// Nutzers, nie außerhalb, und nie über etwas Fremdes.
func (a *App) Befehlszeile(ctx context.Context, handlung string) Befehlsvorschau {
	ctx = a.imProfil(ctx)

	// **Der PATH der Shell** (Befund 9 des Zweitblicks über Phase 6.4) —
	// der des Fensters ist der von `launchd`.
	pfad := shellpfad(ctx, (*exec.Cmd).Output, os.Getenv("PATH"))

	return befehlszeileNeben(ctx, handlung, Werkzeugort, os.UserHomeDir, pfad)
}

func befehlszeileNeben(
	ctx context.Context,
	handlung string,
	programm, heim func() (string, error),
	pfad string,
) Befehlsvorschau {
	ort, err := programm()
	if err == nil {
		var zuhause string
		if zuhause, err = heim(); err == nil {
			v := befehlsvorschau(filepath.Join(zuhause, ".local", "bin"), ort, pfad)

			switch handlung {
			case "":
			case "installieren":
				v, err = befehleInstallieren(v)
			case "entfernen":
				v, err = befehleEntfernen(v)
			default:
				err = fmt.Errorf("%w: unbekannte Handlung %q", domain.ErrValidation, handlung)
			}

			if err != nil {
				v.Fehler = Uebersetze(ctx, err)
			}

			return v
		}
	}

	return Befehlsvorschau{Eintraege: []Befehlseintrag{}, Fehler: Uebersetze(ctx, err)}
}
