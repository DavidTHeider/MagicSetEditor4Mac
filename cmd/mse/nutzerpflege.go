// **Nutzer und Profile ändern und löschen an der Kommandozeile** —
// Werkstück 5 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// nutzeraendernbefehl schreibt die Angaben eines Nutzers fort.
//
// # Die Angaben kommen von der Standardeingabe
//
// **Weil es elf sind** — Werkstück 8 (#1351). Elf Wörter in einer Zeile
// zählte niemand richtig ab, und elf Fahnen wären elf Namen, die man sich
// merken muss. Die Gestalt ist die, die `mse nutzer --json` ausgibt:
//
//	mse nutzer --json | jq '.nutzer[0]' | mse nutzer-aendern 1
//
// **Die Kennung steht trotzdem im Aufruf.** Sie ist das, wonach jemand in
// einem Skript sucht — und sie sticht, was in der Eingabe steht.
//
// **Die »fassung« gehört dazu.** Sie kommt aus derselben Ausgabe; ohne sie
// wird abgewiesen, statt eine gleichzeitige Änderung zu überschreiben.
func nutzeraendernbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("nutzer-aendern braucht genau eine Kennung; die Angaben kommen "+
			"als JSON von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	angaben, code := angabenVonDerEingabe(an)
	if code != CodeGut {
		return code
	}

	angaben.ID = rest[0]

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).NutzerAendern(ctx, angaben)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Nutzer.ID, antwort.Nutzer.Name)

		return an.abschluss()
	})
}

// profilaendernbefehl schreibt die Angaben eines Profils fort.
//
// # Die Angaben kommen von der Standardeingabe
//
// **Seit Werkstück 9 (#1351), und das ist eine brechende Änderung an der
// Kommandozeile** (§16). Vorher standen Kennung, Name und Fassung als drei
// Wörter im Aufruf; mit den vier Angaben aus den Punkten 10 bis 13 wären es
// **sieben**, und sieben Wörter in einer Zeile zählt niemand richtig ab.
//
// **Der Preis ist tragbar, weil der Befehl drei Tage alt ist**: er entstand
// in Werkstück 5 derselben, unveröffentlichten Phase. Die Gestalt ist die,
// die `mse nutzer --json` ausgibt — dieselbe wie bei `nutzer-aendern`:
//
//	mse nutzer --json | jq '.nutzer[0].profile[0]' | mse profil-aendern 1
func profilaendernbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("profil-aendern braucht genau eine Kennung; die Angaben kommen "+
			"als JSON von der Standardeingabe.\n%s", hilfe)

		return CodeBedienung
	}

	angaben, code := profilangabenVonDerEingabe(an)
	if code != CodeGut {
		return code
	}

	angaben.ID = rest[0]

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).ProfilAendern(ctx, angaben)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Profil.ID, antwort.Profil.Name)

		return an.abschluss()
	})
}

// nutzerloeschenbefehl nimmt einen Nutzer samt seinen Profilen heraus.
//
// **Der Name bestätigt** — dieselbe Form wie `set-loeschen` (#1218).
func nutzerloeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("nutzer-loeschen braucht Kennung und Namen — der Name bestätigt.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).NutzerLoeschen(ctx, rest[0], rest[1])
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

// profilloeschenbefehl nimmt ein Profil heraus.
func profilloeschenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("profil-loeschen braucht Kennung und Namen — der Name bestätigt.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).ProfilLoeschen(ctx, rest[0], rest[1])
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

// profilangabenVonDerEingabe liest die Angaben eines Profils als JSON.
//
// **Eine eigene Fassung neben [angabenVonDerEingabe]**, weil sie einen
// anderen Verbund füllt. Der Rumpf ist derselbe — und er steht in
// [jsonVonDerEingabe], damit nicht zweimal dasselbe schiefgehen kann.
func profilangabenVonDerEingabe(an *schreiber) (angaben app.ProfilangabenDTO, code int) {
	return jsonVonDerEingabe(an, "profil-aendern", angaben)
}

// angabenVonDerEingabe liest die Angaben als JSON.
//
// **Dieselbe Gestalt, die `mse nutzer --json` ausgibt** — eine zweite wäre
// eine zweite Stelle, an der die Feldnamen stimmen müssen.
func angabenVonDerEingabe(an *schreiber) (angaben app.NutzerangabenDTO, code int) {
	return jsonVonDerEingabe(an, "nutzer-aendern", angaben)
}

// jsonVonDerEingabe liest einen Verbund als JSON von der Standardeingabe.
//
// **Ein Rumpf für beide Befehle** — sie unterscheiden sich in nichts als dem
// Verbund und dem Namen in der Meldung, und zwei Fassungen liefen beim ersten
// Sonderfall auseinander.
func jsonVonDerEingabe[T any](an *schreiber, befehl string, leer T) (
	gelesenerVerbund T, code int,
) {
	if an.eingabe == nil {
		an.klage("%s liest die Angaben von der Standardeingabe.\n%s", befehl, hilfe)

		return leer, CodeBedienung
	}

	gelesen, fehler := io.ReadAll(an.eingabe)
	if fehler != nil {
		an.klageMitSatz(fehler, "die Standardeingabe lesen: %v", fehler)

		return leer, CodeFachlich
	}

	// **Unlesbare Angaben sind eine falsche Eingabe** — mit `--json` heißt
	// die Form `validierung`, und der Satz bleibt, wie er war.
	if err := json.Unmarshal(gelesen, &gelesenerVerbund); err != nil {
		an.klageMitSatz(fmt.Errorf("%w: die Angaben lesen: %w", domain.ErrValidation, err), "die Angaben lesen: %v", err)

		return leer, CodeFachlich
	}

	return gelesenerVerbund, CodeGut
}
