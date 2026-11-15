// **Die Paketwerkstatt für den MCP-Server** —
// #1164.

package mcpnaehte

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// paketwerkstatt reicht die drei Züge an die Naht durch.
//
// **Ohne Bestand.** Eine Paketdatei liegt auf der Platte; ein Griff nach der
// Datenbank wäre eine Abhängigkeit ohne Anlass.
type paketwerkstatt struct{}

func (paketwerkstatt) DateienUndBeiwerk(ctx context.Context, paket string) (dateien, beiwerk []string, fehler error) {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return nil, nil, err
	}

	return naht.DateienUndBeiwerk(ctx, paket)
}

// Anlegen legt eine neue Vorlage an (FMT-02, #1691).
func (paketwerkstatt) Anlegen(ctx context.Context, paket, datei, inhalt string) (string, error) {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return "", err
	}

	return naht.Anlegen(ctx, paket, datei, inhalt)
}

// Umbenennen gibt einer Datei einen neuen Namen.
func (paketwerkstatt) Umbenennen(ctx context.Context, paket, alt, neu, erwartet string) error {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return err
	}

	return naht.Umbenennen(ctx, paket, alt, neu, erwartet)
}

// Entfernen nimmt eine Datei aus dem Paket.
func (paketwerkstatt) Entfernen(ctx context.Context, paket, datei, erwartet string) error {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return err
	}

	return naht.Entfernen(ctx, paket, datei, erwartet)
}

func (paketwerkstatt) Lesen(ctx context.Context, paket, datei string) (inhalt, stand string, fehler error) {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return "", "", err
	}

	return naht.Lesen(ctx, paket, datei)
}

func (paketwerkstatt) Schreiben(
	ctx context.Context, paket, datei, inhalt, erwartet string,
) (string, error) {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return "", err
	}

	return naht.Schreiben(ctx, paket, datei, inhalt, erwartet)
}

// Pruefen nennt die Stelle des ersten Syntaxfehlers in einem Paketdateitext.
//
// **Ohne Naht über das Paketverzeichnis.** Die drei Nachbarn oben lesen und
// schreiben installierte Dateien und brauchen es; hier kommt der Text mit,
// und ein Griff nach dem Verzeichnis wäre eine Abhängigkeit ohne Anlass —
// samt ihrem Fehlerweg, den niemand betreten könnte.
func (paketwerkstatt) Pruefen(inhalt string) (zeile, spalte int, meldung string) {
	return app.PaketSkripteGeprueft(inhalt)
}
