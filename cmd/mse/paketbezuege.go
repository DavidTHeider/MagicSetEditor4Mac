// **Die Stände der bezogenen Bäume** — G04
// (#2251).
//
// Dieselbe Auskunft und dieselben Schritte wie im Fenster (§16):
//
//	paketbezuege [--fragen]                          die Bäume und ihre Stände
//	paketbezug-vorschau <name> <commit>              was eine Aktualisierung täte
//	paketbezug-aktualisieren <name> <commit> [--bestaetigt]
//	paketbezug-zurueck <name>                        den aufgehobenen Stand zurückholen
//
// Einordnung nach §16: **erweiternd** — vier Befehle kommen hinzu, keiner
// ändert seine Bedeutung.

package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// fragenfahne fragt die Quellen — ohne sie geht nichts ins Netz.
const fragenfahne = "--fragen"

// bezuegeAntwort trägt die bezogenen Bäume.
type bezuegeAntwort struct {
	Bezuege []app.PaketbezugDTO `json:"bezuege"`
}

// paketbezugbefehle verteilt die vier Wege rund um die Bezugsstände.
func paketbezugbefehle(
	ctx context.Context, befehl string, rest []string, bestandsort string, an *schreiber,
) (int, bool) {
	switch befehl {
	case "paketbezuege":
		return paketbezuegebefehl(ctx, bestandsort, an, rest), true
	case "paketbezug-vorschau":
		return paketbezugvorschaubefehl(ctx, bestandsort, an, rest), true
	case "paketbezug-aktualisieren":
		return paketbezugaktualisierbefehl(ctx, bestandsort, an, rest), true
	case "paketbezug-zurueck":
		return paketbezugzurueckbefehl(ctx, bestandsort, an, rest), true
	}

	return 0, false
}

// fahnenUndWoerter trennt die bekannten Fahnen von den übrigen Wörtern.
func fahnenUndWoerter(an *schreiber, befehl, fahne string, rest []string) (gesetzt bool, woerter []string, gut bool) {
	for _, wort := range rest {
		switch {
		case wort == fahne:
			gesetzt = true
		case strings.HasPrefix(wort, "--"):
			an.klage("%s kennt %s nicht.\n%s", befehl, wort, hilfe)

			return false, nil, false
		default:
			woerter = append(woerter, wort)
		}
	}

	return gesetzt, woerter, true
}

// mitPaketweg baut den Weg zu den Paketen und reicht ihn weiter.
func mitPaketweg(ctx context.Context, bestandsort string, an *schreiber, tun func(context.Context, *app.Paketweg) int) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return tun(ctx, weg)
	})
}

func paketbezuegebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	fragen, woerter, gut := fahnenUndWoerter(an, "paketbezuege", fragenfahne, rest)
	if !gut || len(woerter) != 0 {
		if gut {
			an.klage("paketbezuege nimmt keine Argumente außer %s.\n%s", fragenfahne, hilfe)
		}

		return CodeBedienung
	}

	return mitPaketweg(ctx, bestandsort, an, func(ctx context.Context, weg *app.Paketweg) int {
		bezuege, err := weg.Bezuege(ctx, fragen)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, bezuegeAntwort{Bezuege: bezuege})
		}

		for _, bezug := range bezuege {
			an.sag("%s\t%s\tbezogen %s\tgeprüft %s\t%s", bezug.Name, bezug.Adresse,
				kurz(bezug.Stand), kurz(bezug.Geprueft), bezug.Lage)

			if bezug.Spitze != "" {
				an.sag("\tSpitze %s vom %s, %d voraus", kurz(bezug.Spitze), bezug.SpitzeZeit, bezug.Vorsprung)
			}

			if bezug.Fehler != nil {
				an.sag("\t%s", bezug.Fehler.Meldung)
			}

			if bezug.Vorige {
				an.sag("\taufgehoben %s", kurz(bezug.VorigerStand))
			}

			an.sag("\n")
		}

		return an.abschluss()
	})
}

func paketbezugvorschaubefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("paketbezug-vorschau braucht Namen und Commit.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPaketweg(ctx, bestandsort, an, func(ctx context.Context, weg *app.Paketweg) int {
		vorschau, err := weg.BezugVorschau(ctx, rest[0], rest[1])
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, vorschau)
		}

		an.sag("%s: %s → %s (%s, %d voraus, %d zurück)\n", vorschau.Name, kurz(vorschau.Von), kurz(vorschau.Zu),
			vorschau.Lage, vorschau.Vorsprung, vorschau.Rueckstand)

		for _, paket := range vorschau.Pakete {
			an.sag("  %s\t+%d ~%d -%d\n", paket.Paket, paket.Neu, paket.Geaendert, paket.Entfernt)
		}

		if len(vorschau.Sets) > 0 {
			an.sag("  Sets: %s\n", strings.Join(vorschau.Sets, ", "))
		}

		if vorschau.BestaetigungNoetig {
			an.sag("  Braucht %s: %s\n", bestaetigtfahne, strings.Join(vorschau.Gruende, ", "))
		}

		return an.abschluss()
	})
}

func paketbezugaktualisierbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	bestaetigt, woerter, gut := fahnenUndWoerter(an, "paketbezug-aktualisieren", bestaetigtfahne, rest)
	if !gut || len(woerter) != 2 {
		if gut {
			an.klage("paketbezug-aktualisieren braucht Namen und Commit.\n%s", hilfe)
		}

		return CodeBedienung
	}

	return mitPaketweg(ctx, bestandsort, an, func(ctx context.Context, weg *app.Paketweg) int {
		ergebnis, err := weg.BezugAktualisieren(ctx, woerter[0], woerter[1], bestaetigt)
		if ergebnis == nil && err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return bezugZeigen(an, ergebnis.Ort, ergebnis.Stand, ergebnis.Pruefsumme, ergebnis.Uebergangen, err)
	})
}

func paketbezugzurueckbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("paketbezug-zurueck braucht den Namen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitPaketweg(ctx, bestandsort, an, func(ctx context.Context, weg *app.Paketweg) int {
		rueckweg, err := weg.BezugZurueck(ctx, rest[0])
		if rueckweg == nil && err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return bezugZeigen(an, rueckweg.Ort, rueckweg.Stand, rueckweg.Pruefsumme, nil, err)
	})
}

// bezugZeigen gibt einen Baum nach dem Tausch aus — **dieselbe Gestalt wie
// `paket-beziehen`**. Ein Fehler beim Vermerken nimmt den Tausch nicht
// zurück (§7, Regel 3): er wird genannt, nachdem gesagt ist, was liegt.
func bezugZeigen(an *schreiber, ort, stand, pruefsumme string, uebergangen []string, vermerkfehler error) int {
	antwort := bezugAntwort{Ort: ort, Stand: stand, Pruefsumme: pruefsumme, Uebergangen: append([]string{}, uebergangen...)}

	if an.alsJSON {
		if code := alsJSONSchreiben(an, antwort); vermerkfehler == nil {
			return code
		}
	} else {
		an.sag("Bezogen nach %s (Stand %s, %s).\n", antwort.Ort, antwort.Stand, antwort.Pruefsumme)
	}

	if vermerkfehler != nil {
		an.klageUeber(vermerkfehler)

		return CodeFachlich
	}

	return an.abschluss()
}

// kurz gibt die ersten sieben Zeichen eines Commits — oder einen Strich.
func kurz(stand string) string {
	switch {
	case stand == "":
		return "—"
	case len(stand) > 7:
		return stand[:7]
	default:
		return stand
	}
}
