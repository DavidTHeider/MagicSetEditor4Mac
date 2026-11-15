// **Plan, Ersetzen und Zurück** — DOM-01.AK-02 und AK-03
// (#1696).
//
// **Eine eigene Verteilung**, weil `lauf.go` und `laufweiter.go` an der
// Zeilengrenze aus §2 stehen — und weil die Frage eine eigene ist: nicht
// „einbinden oder entfernen", sondern „was geschähe, und wie komme ich
// zurück".
//
// Einordnung nach §16: **erweiternd** — drei Befehle kommen hinzu, keiner
// ändert seine Bedeutung.

package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bestaetigtfahne erlaubt das Ersetzen eines lokal geänderten Pakets.
const bestaetigtfahne = "--bestaetigt"

// planAntwort trägt einen Plan.
type planAntwort struct {
	Plan app.PaketplanDTO `json:"plan"`
}

// paketplanbefehle verteilt die vier Wege rund um den Plan.
func paketplanbefehle(
	ctx context.Context, befehl string, rest []string, bestandsort string, an *schreiber,
) (int, bool) {
	switch befehl {
	case "paket-plan":
		return paketplanbefehl(ctx, bestandsort, an, rest), true
	case "paket-ersetzen":
		return paketersatzbefehl(ctx, bestandsort, an, rest), true
	case "paket-zurueck":
		return paketzurueckbefehl(ctx, bestandsort, an, rest), true
	case "paketbedarf":
		return paketbedarfbefehl(ctx, bestandsort, an, rest), true
	case "set-vorschau":
		return setvorschaubefehl(ctx, bestandsort, an, rest), true
	}

	return 0, false
}

// paketplanbefehl prüft die Wörter von `paket-plan`.
//
//	paket-plan entfernen <art> <name>
//	paket-plan zurueck   <art> <name>
//	paket-plan einbinden <quelle>
func paketplanbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) == 0 {
		an.klage("paket-plan braucht einen Vorgang: entfernen, zurueck oder einbinden.\n%s", hilfe)

		return CodeBedienung
	}

	vorgang, art, name, quelle := rest[0], "", "", ""

	switch {
	case vorgang == "einbinden" && len(rest) == 2:
		quelle = rest[1]
	case (vorgang == "entfernen" || vorgang == "zurueck") && len(rest) == 3:
		art, name = rest[1], rest[2]
	default:
		an.klage("paket-plan entfernen und zurueck brauchen Art und Namen, "+
			"einbinden braucht die Quelle.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return paketPlanen(ctx, db, an, vorgang, art, name, quelle)
	})
}

// paketPlanen rechnet den Plan und zeigt ihn.
func paketPlanen(ctx context.Context, db *sql.DB, an *schreiber, vorgang, art, name, quelle string) int {
	weg, err := app.PaketnahtUeber(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	plan, fehler := weg.Plan(ctx, vorgang, art, name, quelle)
	if fehler != nil {
		an.klageUeber(fehler)

		return paketCode(fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, planAntwort{Plan: plan})
	}

	an.sag("%s\t%s\t%s → %s\n", plan.Vorgang, plan.Paket, plan.Von, plan.Zu)

	if !plan.Moeglich {
		an.sag("nicht möglich\n")
	}

	if plan.Geaendert {
		an.sag("lokal geändert — ersetzen nur mit %s\n", bestaetigtfahne)
	}

	for _, set := range plan.Sets {
		an.sag("set\t%s\n", set)
	}

	for _, konflikt := range plan.Konflikte {
		an.sag("konflikt\t%s braucht %s (%s)\n", konflikt.Paket, konflikt.Braucht, konflikt.Art)
	}

	return an.abschluss()
}

// paketersatzbefehl prüft die Wörter von `paket-ersetzen`.
func paketersatzbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	bestaetigt, quellen := false, []string{}

	for _, wort := range rest {
		switch {
		case wort == bestaetigtfahne:
			bestaetigt = true
		case strings.HasPrefix(wort, "--"):
			an.klage("paket-ersetzen kennt %s nicht.\n%s", wort, hilfe)

			return CodeBedienung
		default:
			quellen = append(quellen, wort)
		}
	}

	if len(quellen) != 1 {
		an.klage("paket-ersetzen braucht genau eine Quelle.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return paketZeigen(an)(weg.Ersetzen(ctx, quellen[0], bestaetigt))
	})
}

// paketzurueckbefehl prüft die Wörter von `paket-zurueck`.
func paketzurueckbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 {
		an.klage("paket-zurueck braucht Art und Namen.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return paketZeigen(an)(weg.Zurueck(ctx, rest[0], rest[1]))
	})
}

// paketZeigen zeigt ein Paket nach dem Tausch — **dieselbe Ausgabe wie
// `paket-installieren`**: eine Zeile aus Art, Name, Fassung und Pfad.
func paketZeigen(an *schreiber) func(domain.Paket, error) int {
	return func(paket domain.Paket, fehler error) int {
		if fehler != nil {
			an.klageUeber(fehler)

			return paketCode(fehler)
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, einbindungAntwort{Paket: alsPaketDTO(paket)})
		}

		an.sag("%s\t%s\t%s\t%s\n", paket.Art, paket.Name, paket.Version, paket.Pfad)

		return an.abschluss()
	}
}

// bedarfAntwort trägt, was einem Set an Paketen fehlt.
type bedarfAntwort struct {
	Befunde []app.AbhaengigkeitDTO `json:"befunde"`
}

// paketbedarfbefehl nennt, was einem Set an Paketen fehlt — DOM-01.AK-04.
//
// **Eine Zeile je Befund**: die Art, dann der Satz „wer braucht was" — so
// lässt sich mit `grep fehlt` herausziehen, was zu beziehen ist.
func paketbedarfbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 1 {
		an.klage("paketbedarf braucht genau eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		befunde, fehler := weg.Setbedarf(ctx, setID)
		if fehler != nil {
			an.klageUeber(fehler)

			return paketCode(fehler)
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, bedarfAntwort{Befunde: befunde})
		}

		for _, befund := range befunde {
			an.sag("%s\t%s braucht %s\n", befund.Art, befund.Paket, befund.Braucht)
		}

		return an.abschluss()
	})
}
