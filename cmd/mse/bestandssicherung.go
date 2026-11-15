// **Den Bestand sichern und zurückspielen von der Kommandozeile** —
// #1169,
// Werkstück 2.
//
// # Warum es das erst jetzt gibt
//
// Der Grund im Katalog war ✋ — „mindestens eine Bindung öffnet einen nativen
// Dateidialog". **Der Dialog war nie der Kern**: er beantwortet *welche
// Datei*, und hier steht sie im Argument. Was wirklich fehlte, war die
// Trennung von Weg und Lebenszyklus, und die hat #1175 gebracht.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitBestandsgriff ist [mitBestand] für einen Befehl, der die Verbindung
// **austauscht**.
//
// # Der Befund, aus dem sie entstanden ist
//
// `mitBestand` reicht ein `*sql.DB` **als Wert** weiter. Wer davon die Adresse
// nimmt, nimmt die Adresse seiner eigenen Parametervariablen — und ein
// Zurückspielen, das darüber eine neue Verbindung eintrug, änderte damit
// nichts an der, auf der das `defer` hier liegt.
//
// **Was das kostete:** die alte Verbindung war nach dem Tausch schon
// geschlossen, das `defer` schloss sie ein zweites Mal (harmlos), und die
// **neue** blieb offen bis zum Prozessende. Bei einem Befehl fällt das nicht
// auf; der Kommentar an der Naht behauptete trotzdem das Gegenteil.
// Angestrichen als Befund 2 des Zweitblicks über Phase 3.8.
//
// **Eine eigene Funktion und keine geänderte**: vierzig Aufrufer tauschen
// nichts aus, und ein `**sql.DB` in jedem von ihnen wäre eine Erschwernis für
// den einen Fall, der sie braucht.
func mitBestandsgriff(
	ctx context.Context, ort string, an *schreiber, arbeit func(context.Context, **sql.DB) int,
) int {
	return mitBestandsgriffWie(ctx, ort, an, false, arbeit)
}

// mitBestandsgriffUeberNeuerem ist [mitBestandsgriff] für Ansehen und
// Zurückspielen einer Sicherung — auch über einem Bestand einer neueren
// Fassung, dann ohne Wanderung ([mitBestandWie]).
func mitBestandsgriffUeberNeuerem(
	ctx context.Context, ort string, an *schreiber, arbeit func(context.Context, **sql.DB) int,
) int {
	return mitBestandsgriffWie(ctx, ort, an, true, arbeit)
}

func mitBestandsgriffWie(
	ctx context.Context, ort string, an *schreiber, ueberNeuerem bool, arbeit func(context.Context, **sql.DB) int,
) int {
	return mitBestandWie(ctx, ort, an, ueberNeuerem, func(ctx context.Context, db *sql.DB) int {
		griff := db

		// **Geschlossen wird hier und nicht draußen.** `mitBestand` schließt
		// seine eigene Variable; die zeigt nach einem Tausch auf die alte
		// Datei. Ein zweites `Close` darauf ist harmlos — was zählt, ist,
		// dass die **neue** auch zugeht.
		defer func() {
			if griff != db {
				_ = griff.Close() //nolint:errcheck // beim Beenden ist nichts mehr zu retten
			}
		}()

		return arbeit(ctx, &griff)
	})
}

// klartextschalter sagt ausdrücklich „ohne Kennwort", auch wenn der Bestand
// verschlüsselt ist (OPS-06.AK-02).
const klartextschalter = "--klartext"

// mitSicherungswunsch legt das Kennwort aus der Umgebung und den
// Klartextschalter in den Kontext und gibt die Argumente ohne ihn.
//
// **Das Kennwort kommt aus der Umgebung und nie aus einem Argument** — ein
// Argument steht in der Prozessliste und im Verlauf der Shell.
func mitSicherungswunsch(ctx context.Context, rest []string) (mit context.Context, ohne []string) {
	klartext := slices.Contains(rest, klartextschalter)
	rest = slices.DeleteFunc(slices.Clone(rest), func(wort string) bool { return wort == klartextschalter })

	return app.MitSicherungswunsch(ctx, domain.FlaecheCLI, os.Getenv(app.Sicherungskennwortumgebung), klartext), rest
}

// bestandSichernBefehl schreibt Bestand und Bilder als eine `.msesicherung`.
func bestandSichernBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	ctx, rest = mitSicherungswunsch(ctx, rest)

	if len(rest) != 1 {
		an.klage("bestand-sichern braucht genau ein Ziel.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestandsgriff(ctx, bestandsort, an, func(ctx context.Context, db **sql.DB) int {
		naht := app.BestandsnahtUeber(db, bestandsort)

		if fehler := naht.Sichern(ctx, rest[0]); fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		an.sag("Gesichert nach %s.\n", rest[0])

		return an.abschluss()
	})
}

// bestandZurueckspielenBefehl ersetzt den Bestand durch eine Sicherung.
//
// # Warum eine Bestätigung
//
// **Weil er den ganzen Bestand ersetzt.** Dieselbe Gestalt wie bei
// `set-loeschen`, wo der Name bestätigt: wer den Befehl aus der Verlaufsliste
// wiederholt, soll nicht versehentlich einen Tagesstand wegwerfen. Bestätigt
// wird mit dem Zeitpunkt, den die Sicherung selbst nennt — er steht in
// `bestand-schauen`, und wer ihn abtippt, hat hingesehen.
func bestandZurueckspielenBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	ctx, rest = mitSicherungswunsch(ctx, rest)

	if len(rest) != 2 {
		an.klage("bestand-zurueckspielen braucht die Sicherung und ihren "+
			"Zeitpunkt aus »mse bestand-schauen«.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestandsgriffUeberNeuerem(ctx, bestandsort, an, func(ctx context.Context, db **sql.DB) int {
		naht := app.BestandsnahtUeber(db, bestandsort)

		schau, fehler := naht.Beschauen(ctx, rest[0])
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if schau.AngelegtAm != rest[1] {
			an.klageMitSatz(fmt.Errorf("%w: die Sicherung ist von %s, bestätigt wurde %s",
				domain.ErrConflict, schau.AngelegtAm, rest[1]),
				"Die Sicherung ist von %s, bestätigt wurde %s. Nichts geändert.\n", schau.AngelegtAm, rest[1])

			return CodeFachlich
		}

		beiseite, rueckfehler := naht.Zurueckspielen(ctx, rest[0])
		if rueckfehler != nil {
			an.klageUeber(rueckfehler)

			return CodeFachlich
		}

		an.sag("Zurückgespielt. Der alte Bestand liegt unter %s.\n", beiseite)

		return an.abschluss()
	})
}

// bestandSchauenBefehl sagt, was in einer Sicherung liegt.
//
// **Ohne etwas zu ändern.** Wer zurückspielen will, hat vorher gesehen, was er
// bekommt — und er bekommt hier den Zeitpunkt, mit dem er bestätigt.
func bestandSchauenBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	ctx, rest = mitSicherungswunsch(ctx, rest)

	if len(rest) != 1 {
		an.klage("bestand-schauen braucht genau eine Sicherung.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestandsgriffUeberNeuerem(ctx, bestandsort, an, func(ctx context.Context, db **sql.DB) int {
		naht := app.BestandsnahtUeber(db, bestandsort)

		schau, fehler := naht.Beschauen(ctx, rest[0])
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{
				"angelegtAm":  schau.AngelegtAm,
				"schemastand": schau.Schemastand,
				"sets":        schau.Sets,
				"karten":      schau.Karten,
				"bilder":      schau.Bilder,

				"verschluesselt":  schau.Verschluesselt,
				"paketeEnthalten": schau.PaketeEnthalten,
				"paketeFehlen":    schau.PaketeFehlen,
			})
		}

		an.sag("angelegt\t%s\n", schau.AngelegtAm)
		an.sag("schema\t%s\n", schau.Schemastand)
		an.sag("sets\t%d\n", schau.Sets)
		an.sag("karten\t%d\n", schau.Karten)
		an.sag("bilder\t%s\n", schau.Bilder)
		an.sag("verschluesselt\t%t\n", schau.Verschluesselt)

		// **Was fehlt, steht mit Namen da** — vor dem Zurückspielen und nicht
		// danach (OPS-06.AK-03).
		an.sag("pakete enthalten\t%s\n", strings.Join(schau.PaketeEnthalten, ", "))
		an.sag("pakete fehlen\t%s\n", strings.Join(schau.PaketeFehlen, ", "))

		return an.abschluss()
	})
}
