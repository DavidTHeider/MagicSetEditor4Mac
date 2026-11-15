// **Der Dateibaum der Werkstatt** — FMT-02.AK-01
// (#1691).
//
// Vier Bindungen: anlegen, einfügen, umbenennen, entfernen. Was erlaubt ist,
// steht in `service/paketdateibaum.go`; hier wird übersetzt (§2).
//
// **Einfügen fragt selbst nach der Datei.** Das Fenster reicht keinen Pfad
// herein, nur den Ordner im Paket — so kann die Oberfläche keine beliebige
// Datei des Rechners in ein Paket legen lassen.

package app

import (
	"context"
	"path"
	"path/filepath"
)

// Paketdateieinfuegung sagt, unter welchem Namen eine Datei ins Paket kam.
type Paketdateieinfuegung struct {
	Datei string `json:"datei"`

	// Abgebrochen ist wahr, wenn niemand eine Datei gewählt hat — **kein
	// Fehler**, wer den Dialog schließt, hat es sich anders überlegt.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// PaketdateiAnlegen legt eine neue Vorlage an und gibt ihren Stand.
func (a *App) PaketdateiAnlegen(ctx context.Context, paket, datei, inhalt string) Paketdateistand {
	ctx = a.imProfil(ctx)

	if a.paketdateien == nil {
		return Paketdateistand{Fehler: Uebersetze(ctx, keinePaketwerkstatt())}
	}

	stand, err := a.paketdateien.Anlegen(ctx, paket, datei, inhalt)
	if err != nil {
		return Paketdateistand{Fehler: Uebersetze(ctx, err)}
	}

	return Paketdateistand{Stand: stand}
}

// PaketdateiEinfuegen fragt nach einer Datei und legt sie in `ordner` des
// Pakets — leer heißt oben.
func (a *App) PaketdateiEinfuegen(ctx context.Context, paket, ordner string) Paketdateieinfuegung {
	ctx = a.imProfil(ctx)

	if a.paketdateien == nil {
		return Paketdateieinfuegung{Fehler: Uebersetze(ctx, keinePaketwerkstatt())}
	}

	texte := DialogtexteFuer(a.sprache)

	quelle, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  texte.PaketdateiEinfuegen,
		Filter: texte.PaketdateiFilter,
		Muster: "*.png;*.jpg;*.jpeg;*.gif;*.ttf;*.otf",
	})
	if fehler != nil {
		return Paketdateieinfuegung{Fehler: Uebersetze(ctx, fehler)}
	}

	if quelle == "" {
		return Paketdateieinfuegung{Abgebrochen: true}
	}

	datei := path.Join(ordner, filepath.Base(quelle))

	if err := a.paketdateien.Einfuegen(ctx, paket, quelle, datei); err != nil {
		return Paketdateieinfuegung{Fehler: Uebersetze(ctx, err)}
	}

	return Paketdateieinfuegung{Datei: datei}
}

// PaketdateiUmbenennen gibt einer Datei des Pakets einen neuen Namen.
func (a *App) PaketdateiUmbenennen(ctx context.Context, paket, alt, neu, stand string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.paketdateien == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, keinePaketwerkstatt())}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.paketdateien.Umbenennen(ctx, paket, alt, neu, stand))}
}

// PaketdateiEntfernen nimmt eine Datei aus dem Paket — in die Sicherung.
func (a *App) PaketdateiEntfernen(ctx context.Context, paket, datei, stand string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.paketdateien == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, keinePaketwerkstatt())}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.paketdateien.Entfernen(ctx, paket, datei, stand))}
}
