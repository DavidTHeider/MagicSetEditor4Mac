package service

// **Pakete als `.mse-installer` weitergeben** — FMT-02.AK-04
// (#1691),
// Marke **ported**.
//
// Wie `magicseteditor --create-installer` (`src/main.cpp`, Zeilen 143–156,
// Fassung f9d9356 vom 2021-07-06): die genannten Pakete, in dieser
// Reihenfolge, und ohne Ziel der Name des ersten (`prefered_filename`,
// `Installer::addPackage`, `src/data/installer.cpp`, Zeilen 177–186). Den
// Inhalt schreibt `msepaket`; hier steht, **welche** Pakete und **wohin**.
//
// # Was es nicht tut
//
// **Es nimmt die Abhängigkeiten nicht von selbst mit** — MSE2 tut es auch
// nicht; sie stehen in der Beschreibung (`depends_on:`), und wer den
// Installer liest, sieht, was fehlt. Ein Stil ohne sein Spiel ist oft
// gewollt: das Spiel hat der Empfänger längst.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Installerendung ist die Endung eines Installers.
const Installerendung = ".mse-installer"

// Weitergeben packt Pakete in einen Installer und gibt den Pfad, an dem er
// liegt.
//
// **`ziel` ist ein Ordner oder eine Datei auf `.mse-installer`.** Eine
// vorhandene Datei wird ersetzt — atomar, wie jede Ausfuhr; gefragt hat
// vorher der Dialog.
func (v *Paketverwalter) Weitergeben(ctx context.Context, namen []string, ziel string) (string, error) {
	pfad, err := Installerziel(namen, ziel)
	if err != nil {
		return "", err
	}

	pakete := make([]*msepaket.Paket, 0, len(namen))

	defer func() {
		for _, paket := range pakete {
			_ = paket.Schliessen() //nolint:errcheck // nur gelesen
		}
	}()

	for nummer, name := range namen {
		if err := Schritt(ctx, domain.Fortschritt{
			Erledigt: nummer, Gesamt: len(namen), Einheit: domain.EinheitPaket,
		}); err != nil {
			return "", err
		}

		paket, err := v.weiterzugeben(name)
		if err != nil {
			return "", err
		}

		pakete = append(pakete, paket)
	}

	// **Die letzte Frage vor dem Schreiben** (FMT-03) — danach entsteht der
	// Installer atomar, und ein Abbruch ließe nichts Halbes zurück.
	if err := Schritt(ctx, domain.Fortschritt{
		Erledigt: len(namen), Gesamt: len(namen), Einheit: domain.EinheitPaket,
	}); err != nil {
		return "", err
	}

	if err := schreibeAtomarMit(pfad, func(ziel io.Writer) error {
		return msepaket.SchreibeInstaller(ziel, pakete)
	}); err != nil {
		return "", fmt.Errorf("den Installer %q schreiben: %w", pfad, err)
	}

	return pfad, nil
}

// weiterzugeben sucht ein Paket nach seinem Dateinamen und öffnet es.
func (v *Paketverwalter) weiterzugeben(dateiname string) (*msepaket.Paket, error) {
	art := msepaket.ArtAusName(dateiname)
	if art == msepaket.ArtUnbekannt || art == msepaket.ArtSet {
		return nil, domain.NewFieldError("pakete", domain.SchluesselValidierung,
			"%q ist kein Vorlagenpaket — genannt wird es mit Endung, etwa probe.mse-game.", dateiname)
	}

	gefunden, _, err := v.gefunden(string(art), strings.TrimSuffix(dateiname, "."+string(art)))
	if err != nil {
		return nil, err
	}

	return msepaket.Oeffne(gefunden.Pfad)
}

// Installerziel macht aus der Angabe den Pfad der Datei, die entsteht.
//
// **Exportiert, weil MCP und REST ihn vorher reservieren**: sie überschreiben
// nichts (0044), und ob ein Ordner genannt war, entscheidet, welche Datei es
// wird — die Regel steht hier einmal und nicht an jeder Fläche.
func Installerziel(namen []string, ziel string) (string, error) {
	if len(namen) == 0 {
		return "", domain.NewFieldError("pakete", domain.SchluesselPflichtfeld,
			"Welche Pakete sollen hinaus? Es ist keines genannt.")
	}

	erstes := namen[0]
	ziel = strings.TrimSpace(ziel)
	if ziel == "" {
		return "", domain.NewFieldError("ziel", domain.SchluesselPflichtfeld,
			"Wohin soll der Installer? Es ist kein Ort genannt.")
	}

	if angaben, err := os.Stat(ziel); err == nil && angaben.IsDir() {
		name := strings.TrimSuffix(erstes, filepath.Ext(erstes))

		return filepath.Join(ziel, name+Installerendung), nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", domain.MitGrund(domain.ErrValidation, err, "das Ziel ansehen")
	}

	if !strings.HasSuffix(ziel, Installerendung) {
		return "", domain.NewFieldError("ziel", domain.SchluesselValidierung,
			"%q ist weder ein Ordner noch eine Datei auf %s.", ziel, Installerendung)
	}

	return ziel, nil
}
