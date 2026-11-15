// **Einstellungen ausführen und einlesen, an der Kommandozeile** — Werkstück
// O3 von Phase 5.2
// (#1427).
//
// # Drei Verben unter einem Befehl
//
// `mse einstellungen ausfuhr`, `… vorschau <datei>`, `… einfuhr <datei>
// <revision>` — dieselbe Bauform wie `… setzen` daneben, und aus demselben
// Grund: es ist **eine** Sache, und drei Befehle im Hilfetext wären drei
// Einträge für eine Frage.
//
// # Warum die Vorschau eine Revision ausgibt und die Einfuhr sie verlangt
//
// **Weil zwischen beiden jemand geschrieben haben kann** (§7, Regel 2). Am
// Fenster liegen Sekunden dazwischen, hier ein Aufruf — der Unterschied ist
// keiner: der MCP-Server schreibt auf dieselbe Datei.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// austauschbefehl führt ein Verb des Austauschs aus.
//
// **Ein unbekanntes Verb ist ein Bedienfehler und keine Datei.** Eine
// Weiche, die alles außer »ausfuhr« für »einfuhr« hielte, übernähme bei
// `mse einstellungen quatsch datei.json`, was jemand nur ansehen wollte.
func austauschbefehl(
	ctx context.Context, bestandsort string, an *schreiber, verb string, rest []string,
) int {
	if verb == "ausfuhr" {
		return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
			return ausfuhrZeigen(ctx, an, app.EinstellungsnahtUeber(db), rest)
		})
	}

	if verb != "vorschau" && verb != "einfuhr" {
		an.klage("einstellungen kennt %q nicht.\n%s", verb, hilfe)

		return CodeBedienung
	}

	if len(rest) == 0 {
		an.klage("%s braucht eine Datei.\n%s", verb, hilfe)

		return CodeBedienung
	}

	//#nosec G304 -- ein Pfad, den der Aufrufende selbst genannt hat
	roh, lesefehler := os.ReadFile(rest[0])
	if lesefehler != nil {
		an.klageUeber(fmt.Errorf("die Datei lesen: %w", lesefehler))

		return CodeFachlich
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg := app.EinstellungsnahtUeber(db)

		if verb == "vorschau" {
			return vorschauZeigen(ctx, an, weg, roh)
		}

		return einfuhrAusfuehren(ctx, an, weg, roh, rest[1:])
	})
}

// ausfuhrZeigen schreibt den Austausch auf die Standardausgabe.
//
// **Immer als JSON, auch ohne `--json`.** Der Austausch **ist** eine Datei;
// ihn in Spalten zu zerlegen hieße, ihn unbrauchbar zu machen — wer ihn
// mitnehmen will, schreibt ihn mit `>` fort. Dieselbe Entscheidung wie bei
// `mse personendaten`.
func ausfuhrZeigen(
	ctx context.Context, an *schreiber, weg *app.Einstellungsweg, abschnitte []string,
) int {
	inhalt, fehler := weg.Ausfuhr(ctx, abschnitte)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	an.sag("%s\n", inhalt)

	return an.abschluss()
}

// vorschauZeigen schreibt, was eine Datei ändern würde.
func vorschauZeigen(
	ctx context.Context, an *schreiber, weg *app.Einstellungsweg, roh []byte,
) int {
	sicht, fehler := weg.Einfuhrvorschau(ctx, roh)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, sicht)
	}

	an.sag("revision\t%s\n", sicht.Revision)

	for _, zeile := range sicht.Zeilen {
		// **Die Wirkung als sechste Spalte und nicht dazwischen** (§16,
		// erweiternd): wer die ersten fünf liest, liest sie weiter.
		an.sag("%s\t%s\t%s\t%s\t%s\t%s\n", zeile.Ebene, zeile.Schluessel,
			zeile.Vorher, zeile.Nachher, zeile.Herkunft, zeile.Wirkung)
	}

	befundeZeigen(an, sicht.Befunde)

	return an.abschluss()
}

// einfuhrAusfuehren übernimmt die gewählten Zeilen.
func einfuhrAusfuehren(
	ctx context.Context, an *schreiber, weg *app.Einstellungsweg,
	roh []byte, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("einfuhr braucht die Revision aus der Vorschau.\n%s", hilfe)

		return CodeBedienung
	}

	bericht, fehler := weg.Einfuhruebernahme(ctx, roh, rest[0], rest[1:])
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, bericht)
	}

	an.sag("uebernommen\t%d\nrevision\t%s\n", bericht.Uebernommen, bericht.Revision)

	befundeZeigen(an, bericht.Befunde)

	return an.abschluss()
}

// befundeZeigen schreibt jeden Befund in eine Zeile.
//
// **Sie stehen auf der Standardausgabe und nicht auf der Fehlerausgabe**: ein
// Befund ist kein Fehler (§8), und die Zeilen daneben sind das Ergebnis
// desselben Laufs.
func befundeZeigen(an *schreiber, befunde []app.EinfuhrbefundDTO) {
	for _, befund := range befunde {
		an.sag("befund\t%s\t%s\t%s\n", befund.Ebene, befund.Schluessel, befund.Grund)
	}
}
