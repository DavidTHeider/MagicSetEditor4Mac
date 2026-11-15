// **Ein Paket aus einem Gerüst anlegen — an allen vier Flächen**
// (#1350).
//
// # Warum hier und nicht in jeder Fläche
//
// Wie jede Fähigkeit dieses Programms (§16): das Fenster, die Kommandozeile,
// der MCP-Server und der Dienst rufen **dieselbe** Fassade. Eine zweite
// Zusammensetzung wäre eine zweite Fachlogik, und zwei Fachlogiken laufen
// auseinander.

package app

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PaketAnlegen legt ein neues Paket aus einem Gerüst an.
//
// **Vier Zeichenketten und kein DTO**: die Wails-Bindung nimmt sie so
// entgegen, und ein Verbund mit vier Feldern wäre auf der Frontendseite ein
// Typ, den niemand außerhalb dieses einen Aufrufs braucht.
func (a *App) PaketAnlegen(ctx context.Context, art, name, vollname, spiel string) PaketAntwort {
	ctx = a.imProfil(ctx)

	// **Der Vollname fällt auf den Namen zurück.** Ein Paket ohne `full name`
	// steht in der Liste als leere Zeile; MSE2 setzt dann den Dateinamen ein
	// (`Packaged::validate`, `src/util/io/package.cpp`, Zeilen 621–631,
	// Fassung f9d9356 vom 2021-07-06), und dasselbe tun wir — sichtbar hier
	// statt unsichtbar dort.
	if strings.TrimSpace(vollname) == "" {
		vollname = name
	}

	paket, err := a.pakete.Anlegen(ctx, service.Neuanlage{
		Art: art, Name: name, Vollname: vollname, Spiel: spiel,
	})
	if err != nil {
		return PaketAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := a.mitUebersicht(ctx, paket)
	a.arbeitVermerken(ctx, domain.ArbeitsartPaket, gelesen.Name, domain.VorgangEingefuehrt)

	return PaketAntwort{Paket: &gelesen}
}

// Anlegbar nennt die Arten, die sich anlegen lassen.
//
// **Vier von sieben, und das ist eine Entscheidung.** Include und
// Exportvorlage kamen mit FMT-02 (#1691) dazu. Ein `.mse-symbol-font`
// oder eine `.mse-locale` leer anzulegen ergäbe eine Datei, die nichts kann
// und die niemand von Hand füllen möchte; ein `.mse-set` legt
// [App.SetAnlegen] an. Die Liste steht hier, damit das Fenster sie anbieten
// kann, ohne sie zu erfinden.
func Anlegbar() []Anlegbarkeit {
	return []Anlegbarkeit{
		{Art: "mse-game", BrauchtSpiel: false},
		{Art: "mse-style", BrauchtSpiel: true},
		// **Include und Exportvorlage seit FMT-02** (#1691) — die
		// Exportvorlage gehört wie ein Stil zu einem Spiel.
		{Art: "mse-include", BrauchtSpiel: false},
		{Art: "mse-export-template", BrauchtSpiel: true},
	}
}

// Anlegbarkeit ist eine anlegbare Paketart.
type Anlegbarkeit struct {
	Art string `json:"art"`

	// BrauchtSpiel sagt, ob ein Spielname mitzugeben ist — ein Stil legt die
	// Felder **seines** Spiels aus und kann ohne es nichts.
	BrauchtSpiel bool `json:"brauchtSpiel"`
}
