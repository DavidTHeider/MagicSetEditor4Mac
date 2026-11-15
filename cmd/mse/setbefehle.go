// **Die Befehle, die ein ganzes Set betreffen** — anlegen und löschen.
//
// **Eine eigene Datei, weil `lauf.go` sonst über die Zeilengrenze ginge**
// (§2). Der Schnitt folgt der Frage: dort steht, **welcher** Befehl gerufen
// wird, hier, **was** seine Argumente bedeuten müssen.

package main

import (
	"context"
	"database/sql"
)

// setneubefehl legt ein leeres Set an (#891).
//
// **Name und Spiel sind Pflicht, der Stil ist wahlfrei** — wie in MSE2s
// Dialog.
func setneubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 || len(rest) > 3 {
		an.klage("set-neu braucht Name und Spiel, wahlweise einen Stil.\n%s", hilfe)

		return CodeBedienung
	}

	stil := ""
	if len(rest) == 3 {
		stil = rest[2]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return setAnlegen(ctx, db, an, rest[0], rest[1], stil)
	})
}

// setloeschbefehl nimmt ein Set aus dem Bestand
// (#1218).
//
// **Kennung und Name sind beide Pflicht.** Der Name ist die Bestätigung für
// einen unumkehrbaren Vorgang; ohne ihn gäbe es keine.
func setloeschbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("set-loeschen braucht die Kennung und den heutigen Namen des Sets.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return setLoeschen(ctx, db, an, rest[0], rest[1])
	})
}
