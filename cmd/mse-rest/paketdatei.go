// **Die Paketwerkstatt für den Dienst auf der Rückschleife** —
// #1164.
package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// paketwerkstatt reicht die drei Züge an die Naht durch.
type paketwerkstatt struct{}

func (paketwerkstatt) DateienUndBeiwerk(ctx context.Context, paket string) (dateien, beiwerk []string, fehler error) {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return nil, nil, err
	}

	return naht.DateienUndBeiwerk(ctx, paket)
}

// DateiAnlegen legt eine neue Vorlage an (FMT-02, #1691).
func (paketwerkstatt) DateiAnlegen(ctx context.Context, paket, datei, inhalt string) (string, error) {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return "", err
	}

	return naht.Anlegen(ctx, paket, datei, inhalt)
}

// DateiUmbenennen gibt einer Datei einen neuen Namen.
func (paketwerkstatt) DateiUmbenennen(ctx context.Context, paket, alt, neu, erwartet string) error {
	naht, err := app.WerkstattnahtUeber()
	if err != nil {
		return err
	}

	return naht.Umbenennen(ctx, paket, alt, neu, erwartet)
}

// DateiEntfernen nimmt eine Datei aus dem Paket.
func (paketwerkstatt) DateiEntfernen(ctx context.Context, paket, datei, erwartet string) error {
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
// **Ohne Naht über das Paketverzeichnis** — der Text kommt mit. Dieselbe
// Begründung wie im MCP-Server.
func (paketwerkstatt) Pruefen(inhalt string) (zeile, spalte int, meldung string) {
	return app.PaketSkripteGeprueft(inhalt)
}
