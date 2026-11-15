package main

// **Die bekannten Gegenüber an der Kommandozeile** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
//	mse kontakte                              wer bekannt ist und was er darf
//	mse kontakt-aufnehmen <kennung> <name>    ein Gegenüber eintragen
//	mse kontakt-bestaetigen <kennung> <abdruck>   den Abdruck vergleichen
//	mse kontakt-sperren <kennung>             ausschließen
//	mse kontakt-entsperren <kennung>          auf unbestätigt zurück
//	mse zweck-erlauben <kennung> <zweck>      einen Zweck freigeben
//	mse zweck-entziehen <kennung> <zweck>     ihn zurücknehmen
//	mse kontakt-adresse <kennung> <adresse>   wo es erreichbar ist
//	mse kontakt-notiz <kennung> <text>        was man sich notiert
//	mse nachbarn                              wer sich im selben Netz meldet
//
// **Der Abdruck ist ein Argument und kein Schalter.** Wer bestätigt, tippt
// ihn ab — das ist der Vergleich, und ein `--ja` wäre ein Klick mit anderen
// Mitteln.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kontaktebefehl nennt die bekannten Gegenüber.
func kontaktebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakte", 0,
		func(ctx context.Context, anwendung *app.App, _ []string) app.PeerkontakteAntwort {
			return anwendung.Peerkontakte(ctx)
		})
}

// kontaktaufnehmenbefehl trägt ein Gegenüber ein.
func kontaktaufnehmenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakt-aufnehmen", 2,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerkontaktAufnehmen(ctx, teile[0], teile[1])
		})
}

// kontaktbestaetigenbefehl hält den verglichenen Abdruck gegen die Kennung.
func kontaktbestaetigenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakt-bestaetigen", 2,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerkontaktBestaetigen(ctx, teile[0], teile[1])
		})
}

// kontaktsperrenbefehl schließt ein Gegenüber aus.
func kontaktsperrenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakt-sperren", 1,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerkontaktSperren(ctx, teile[0])
		})
}

// kontaktentsperrenbefehl nimmt die Sperre zurück.
func kontaktentsperrenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakt-entsperren", 1,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerkontaktEntsperren(ctx, teile[0])
		})
}

// zweckerlaubenbefehl gibt einen Zweck frei.
func zweckerlaubenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "zweck-erlauben", 2,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerzweckErlauben(ctx, teile[0], teile[1])
		})
}

// zweckentziehenbefehl nimmt eine Freigabe zurück.
func zweckentziehenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "zweck-entziehen", 2,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerzweckEntziehen(ctx, teile[0], teile[1])
		})
}

// kontaktadressebefehl trägt ein, wo ein Gegenüber erreichbar ist.
//
// **Eine leere Adresse löscht die eingetragene** — und sie steht als leeres
// Argument da (`mse kontakt-adresse <kennung> ""`), damit das Löschen eine
// Angabe bleibt und kein vergessenes Argument.
func kontaktadressebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakt-adresse", 2,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerkontaktAdresse(ctx, teile[0], teile[1])
		})
}

// kontaktnotizbefehl schreibt, was man sich zu einem Gegenüber notiert.
//
// **Ein leerer Text löscht die Notiz** — dieselbe Form wie bei der Adresse.
func kontaktnotizbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	return kontaktlauf(ctx, bestandsort, an, rest, "kontakt-notiz", 2,
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			return anwendung.PeerkontaktNotiz(ctx, teile[0], teile[1])
		})
}

// kontaktlauf fährt einen der acht Wege und schreibt die Liste.
func kontaktlauf(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
	name string, argumente int,
	fahre func(context.Context, *app.App, []string) app.PeerkontakteAntwort,
) int {
	if len(rest) != argumente {
		an.klage("%s nimmt %d Argumente.\n%s", name, argumente, hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := fahre(ctx, app.PeerflaecheUeber(db), rest)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		nachladenMelden(an, antwort.Nachladefehler)

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, kontakt := range antwort.Kontakte {
			an.sag("%s\t%s\t%s\t%s\n", kontakt.Stand, kontakt.Kennung,
				kontakt.Fingerabdruck, kontakt.SelbstgewaehlterName)

			// **Nur wo eine steht** — eine Zeile „unter: " wäre eine
			// Frage ohne Antwort.
			if kontakt.Adresse != "" {
				an.sag("\tunter\t%s\n", kontakt.Adresse)
			}

			// **Und nur, wo eine Notiz steht** (T3, #1468) — aus
			// demselben Grund wie bei der Adresse.
			if kontakt.Notiz != "" {
				an.sag("\tnotiz\t%s\n", kontakt.Notiz)
			}

			// **Der Code, solange es etwas zu vergleichen gibt**, und der
			// geänderte Schlüssel, weil ihn der Stand allein nicht zeigt
			// (NET-01, #1705).
			if kontakt.Vergleichscode != "" {
				an.sag("\tvergleich\t%s\n", kontakt.Vergleichscode)
			}

			if kontakt.Lage == string(domain.LageSchluesselGeaendert) {
				an.sag("\tlage\t%s\n", kontakt.Lage)
			}

			for _, zweck := range kontakt.Rechte {
				an.sag("\tdarf\t%s\n", zweck)
			}
		}

		return an.abschluss()
	})
}
