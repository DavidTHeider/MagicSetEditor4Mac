// **Zugriffstoken an der Kommandozeile** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// **Vier Befehle, alle erweiternd** (§16): sie kommen hinzu, keiner ändert
// einen bestehenden.
//
// # Der Klartext geht nach stdout und nirgendwo sonst
//
// **Damit er sich weiterreichen lässt** — `mse token-anlegen … > .mse-token`
// oder in eine Variable. Und deshalb steht **nur** er dort, wenn nicht
// `--json` verlangt ist: eine Zeile mit Kennung, Name und Klartext ließe sich
// nicht ohne Schneiden benutzen.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// tokenbefehl führt die Zugriffstoken eines Profils.
func tokenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("token braucht genau eine Profilkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Token(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, eines := range antwort.Token {
			an.sag("%s\t%s\t%s\t%s\n",
				eines.ID, eines.Scope, gueltigkeit(eines), eines.Bezeichnung)
		}

		return an.abschluss()
	})
}

// gueltigkeit sagt in einem Wort, woran ein Token gerade ist.
func gueltigkeit(eines app.TokenDTO) string {
	if eines.Abgelaufen {
		return "abgelaufen"
	}

	if eines.Ablauf == "" {
		return "unbefristet"
	}

	return eines.Ablauf
}

// tokenanlegenbefehl legt eines an und gibt seinen Klartext **einmal**.
func tokenanlegenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("token-anlegen braucht genau eine Profilkennung; die Angaben "+
			"kommen als JSON von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	angaben, code := tokenangabenVonDerEingabe(an, "token-anlegen")
	if code != CodeGut {
		return code
	}

	angaben.ProfilID = rest[0]

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).TokenAnlegen(ctx, angaben)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		// **Nur der Klartext, ohne Beiwerk.** Er ist das einzige, was es
		// gleich nicht mehr gibt — und das einzige, was jemand weiterreicht.
		an.sag("%s\n", antwort.Klartext)

		return an.abschluss()
	})
}

// tokenaendernbefehl schreibt Bezeichnung, Erlaubnis und Ablauf fort.
func tokenaendernbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("token-aendern braucht genau eine Kennung; die Angaben kommen "+
			"als JSON von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	angaben, code := tokenangabenVonDerEingabe(an, "token-aendern")
	if code != CodeGut {
		return code
	}

	angaben.ID = rest[0]

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).TokenAendern(ctx, angaben)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Token.ID, antwort.Token.Bezeichnung)

		return an.abschluss()
	})
}

// tokenloeschenbefehl nimmt eines heraus.
func tokenloeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("token-loeschen braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).TokenLoeschen(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		return an.abschluss()
	})
}

// tokenangabenVonDerEingabe liest die Angaben eines Tokens als JSON.
//
// **Dieselbe Gestalt, die `mse token --json` ausgibt** — eine zweite wäre eine
// zweite Stelle, an der die Feldnamen stimmen müssen.
func tokenangabenVonDerEingabe(an *schreiber, befehl string) (
	angaben app.TokenangabenDTO, code int,
) {
	return jsonVonDerEingabe(an, befehl, angaben)
}

// tokenrotierenbefehl gibt einem Token ein neues Geheimnis — Werkstück O2 von
// Phase 5.2 (#1426).
//
// **Der Klartext geht nach stdout und sonst nirgendwohin** — wie beim
// Anlegen: nicht ins Protokoll, nicht in eine Meldung.
func tokenrotierenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("token-rotieren braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).TokenRotieren(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\n", antwort.Klartext)

		return an.abschluss()
	})
}

// tokenallewiderrufenbefehl zieht alle Token eines Profils zurück.
func tokenallewiderrufenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("token-alle-widerrufen braucht genau eine Profilkennung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).TokenAlleWiderrufen(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%d\n", antwort.Gezogen)

		return an.abschluss()
	})
}

// tokenablaufwahlenbefehl nennt die Vorschläge für den Ablauf.
func tokenablaufwahlenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("token-ablaufwahlen nimmt keine Angaben.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Tokenablaufwahlen(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, wahl := range antwort.Wahlen {
			an.sag("%s\t%s\n", wahl.Name, wahl.Ablauf)
		}

		return an.abschluss()
	})
}
