// **Die vier Kontoarten an der Kommandozeile** — Werkstück 6 von Phase 4.12
// (#1351).
//
// # Das Kennwort steht nicht im Aufruf
//
// **Weil ein Argument öffentlich ist.** Es landet in `~/.zsh_history` und
// steht, solange der Aufruf läuft, für jeden Prozess desselben Rechners in
// `ps aux`. Beides überlebt den Aufruf, und beides sieht niemand.
//
// **Gelesen wird deshalb von der Standardeingabe** — dieselbe Naht wie bei
// `ablage-einfuegen` (#1168):
//
//	printf '%s' "$KENNWORT" | mse konto-neu wer@example.org
//
// **Was das nicht ist**: eine Eingabeaufforderung ohne Echo. Die bräuchte ein
// Terminal, und die Kommandozeile soll in einem Skript laufen.

package main

import (
	"context"
	"database/sql"
	"io"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kontoneubefehl legt ein Konto unter einer Adresse oder einer Rufnummer an.
func kontoneubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 || len(rest) > 2 {
		an.klage("konto-neu braucht eine Kennung und darf einen Namen tragen.\n%s", hilfe)

		return CodeBedienung
	}

	kennwort, code := kennwortVonDerEingabe(an, "konto-neu")
	if code != CodeGut {
		return code
	}

	name := ""
	if len(rest) > 1 {
		name = rest[1]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return kontoAntwortSchreiben(an,
			app.NutzerflaecheUeber(db).KontoEintragen(ctx, rest[0], kennwort, name))
	})
}

// kontoanmeldebefehl lässt einen Nutzer unter seiner Kennung herein.
func kontoanmeldebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 1 || len(rest) > 2 {
		an.klage("konto-anmelden braucht eine Kennung und darf einen Nachweis "+
			"tragen.\n%s", hilfe)

		return CodeBedienung
	}

	kennwort, code := kennwortVonDerEingabe(an, "konto-anmelden")
	if code != CodeGut {
		return code
	}

	// **Der Nachweis steht im Argument, das Kennwort nicht** — Werkstück 7
	// (#1351). Das ist kein Widerspruch, sondern die Folge daraus, wie lange
	// beides gilt: ein Einmalkennwort ist nach neunzig Sekunden eine Zahl,
	// ein Kennwort bleibt eines.
	nachweis := ""
	if len(rest) == 2 {
		nachweis = rest[1]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return kontoAntwortSchreiben(an, app.NutzerflaecheUeber(db).
			KontoAnmelden(ctx, rest[0], kennwort, nachweis))
	})
}

// kontoanonymbefehl legt ein Konto ohne jede Angabe an.
func kontoanonymbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("konto-anonym nimmt keine weiteren Wörter.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return kontoAntwortSchreiben(an, app.NutzerflaecheUeber(db).KontoAnonym(ctx))
	})
}

// kennwortVonDerEingabe liest das Kennwort von der Standardeingabe.
//
// **Ohne den Rand**: eine Pipeline hängt fast immer einen Zeilenumbruch an
// (`echo` tut es, `printf '%s'` nicht), und ein Kennwort, das sich danach
// unterscheidet, wäre der Fehler, den niemand findet.
func kennwortVonDerEingabe(an *schreiber, befehl string) (kennwort string, code int) {
	if an.eingabe == nil {
		an.klage("%s liest das Kennwort von der Standardeingabe.\n%s", befehl, hilfe)

		return "", CodeBedienung
	}

	gelesen, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageMitSatz(fehler, "die Standardeingabe lesen: %v", fehler)

		return "", CodeFachlich
	}

	return strings.TrimRight(string(gelesen), "\r\n"), CodeGut
}

// kontoAntwortSchreiben gibt einen Nutzer aus — für alle drei Kontobefehle.
func kontoAntwortSchreiben(an *schreiber, antwort app.EinNutzerAntwort) int {
	if antwort.Fehler != nil {
		an.klageUeberFehler(antwort.Fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\t%s\n", antwort.Nutzer.ID, antwort.Nutzer.Name)

	return an.abschluss()
}
