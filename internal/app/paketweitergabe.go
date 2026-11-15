package app

// **Ein Paket als `.mse-installer` weitergeben und einen einlesen — am
// Fenster** — FMT-02.AK-04
// (#1691),
// über `service.Paketverwalter.Weitergeben` und `Einlesen`.
//
// **Beide fragen selbst nach dem Ort**, wie das Einfügen einer Datei in ein
// Paket: das Fenster reicht keinen Pfad herein. Kommandozeile, REST und MCP
// nennen ihn — dort gibt es keinen Dialog.

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Paketweitergabe sagt, wohin ein Paket hinausging.
type Paketweitergabe struct {
	Datei string `json:"datei"`

	// Abgebrochen ist wahr, wenn niemand einen Ort gewählt hat — **kein
	// Fehler**, wer den Dialog schließt, hat es sich anders überlegt.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// PaketWeitergeben fragt nach einem Ort und schreibt das Paket als
// Installer dorthin — `paket` mit Endung, etwa `probe.mse-game`.
func (a *App) PaketWeitergeben(ctx context.Context, paket string) Paketweitergabe {
	ctx = a.imProfil(ctx)
	texte := DialogtexteFuer(a.sprache)

	ziel, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Speichern: true,
		Titel:     texte.PaketWeitergeben,
		// **Der Name, den MSE2 wählte**: das Paket ohne seine Endung
		// (`prefered_filename`, `src/data/installer.cpp`, Zeilen 177–186,
		// Fassung f9d9356 vom 2021-07-06).
		Vorschlag: strings.TrimSuffix(paket, filepath.Ext(paket)) + service.Installerendung,
		Filter:    texte.InstallerFilter,
		Muster:    "*" + service.Installerendung,
	})
	if fehler != nil {
		return Paketweitergabe{Fehler: Uebersetze(ctx, fehler)}
	}

	if ziel == "" {
		return Paketweitergabe{Abgebrochen: true}
	}

	return imDateilauf(a, ctx, "weitergabe", ziel, func(lauf context.Context) Paketweitergabe {
		datei, fehler := a.pakete.Weitergeben(lauf, []string{paket}, ziel)
		if fehler != nil {
			return Paketweitergabe{Fehler: Uebersetze(ctx, fehler)}
		}

		return Paketweitergabe{Datei: datei}
	}, func(antwort Paketweitergabe) *Fehler { return antwort.Fehler })
}

// InstallerEinlesen fragt nach einem Installer und legt seine Pakete ins
// Paketverzeichnis.
func (a *App) InstallerEinlesen(ctx context.Context) UebernahmeAntwort {
	ctx = a.imProfil(ctx)
	texte := DialogtexteFuer(a.sprache)

	leer := UebernahmeAntwort{Uebernommen: []PaketDTO{}, Uebergangen: []string{}, Befunde: []string{}}

	quelle, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  texte.InstallerEinlesen,
		Filter: texte.InstallerFilter,
		Muster: "*" + service.Installerendung,
	})
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	if quelle == "" {
		leer.Abgebrochen = true

		return leer
	}

	return imDateilauf(a, ctx, "installer", quelle, func(lauf context.Context) UebernahmeAntwort {
		ergebnis, fehler := a.pakete.Einlesen(lauf, quelle)

		return a.uebernahmeAntwort(ctx, leer, ergebnis, fehler)
	}, func(antwort UebernahmeAntwort) *Fehler { return antwort.Fehler })
}
