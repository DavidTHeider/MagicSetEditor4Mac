package main

// **Die Spalten der Kartenliste und ihre Ordnung an der Kommandozeile** —
// Werkstück 3 von Phase 3.7
// (#1168).
//
//	mse kartenliste 1                       zeigt Spalten und Zeilen
//	mse kartenliste 1 --spalten=name,typ    setzt die sichtbaren Spalten
//	mse kartenliste 1 --sortieren=name      ordnet danach
//	mse kartenliste 1 --sortieren=          hebt die Ordnung auf
//
// **Die Wahl gilt für alle Sets desselben Spiels** — das ist die Zusage, und
// sie steht in der Hilfe.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kartenlistenbefehl beantwortet `mse kartenliste`.
func kartenlistenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 {
		an.klage("kartenliste braucht eine Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	var (
		spalten      string
		setztSpalten bool
		sortieren    string
		setztOrdnung bool
		absteigend   bool
	)

	for _, fahne := range rest[1:] {
		switch {
		case strings.HasPrefix(fahne, "--spalten="):
			spalten, setztSpalten = strings.TrimPrefix(fahne, "--spalten="), true
		case strings.HasPrefix(fahne, "--sortieren="):
			sortieren, setztOrdnung = strings.TrimPrefix(fahne, "--sortieren="), true
		case fahne == "--absteigend":
			absteigend = true
		default:
			an.klage("unbekannte Angabe %q bei kartenliste.\n%s", fahne, hilfe)

			return CodeBedienung
		}
	}

	return mitListe(ctx, bestandsort, an, func(ctx context.Context, naht *app.Fremdliste) int {
		if setztSpalten {
			if code := spaltenSchreiben(ctx, naht, an, setID, spalten); code != 0 {
				return code
			}
		}

		if setztOrdnung {
			fehler := naht.SortierungSetzen(ctx, setID, domain.Kartenordnung{
				Feld: sortieren, Aufsteigend: !absteigend,
			})
			if fehler != nil {
				an.klageUeber(fehler)

				return CodeFachlich
			}
		}

		return listeZeigen(ctx, naht, an, setID)
	})
}

// spaltenSchreiben setzt die sichtbaren Spalten aus einer Aufzählung.
//
// **Die genannten sind sichtbar, alle übrigen nicht** — und die Reihenfolge
// der Aufzählung ist die Reihenfolge der Spalten. Wer eine Spalte weglässt,
// blendet sie aus; das ist dieselbe Bedeutung wie am Fenster, wo eine ganze
// Wahl geschrieben wird.
func spaltenSchreiben(
	ctx context.Context, naht *app.Fremdliste, an *schreiber, setID int64, roh string,
) int {
	aufbau, fehler := naht.Aufbau(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	genannt := map[string]int{}

	for stelle, name := range strings.Split(roh, ",") {
		if name = strings.TrimSpace(name); name != "" {
			genannt[name] = stelle + 1
		}
	}

	wahl := make([]domain.Spaltenwahl, 0, len(aufbau.Spalten))

	for _, spalte := range aufbau.Spalten {
		stelle, sichtbar := genannt[spalte.Feld]
		if !sichtbar {
			// **Die ungenannten stehen hinter den genannten**, in der
			// Reihenfolge, die sie schon hatten.
			stelle = len(genannt) + len(wahl) + 1
		}

		wahl = append(wahl, domain.Spaltenwahl{
			Feld: spalte.Feld, Sichtbar: sichtbar, Stelle: stelle,
		})
	}

	if fehler := naht.SpaltenSetzen(ctx, setID, wahl); fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return 0
}

// listeZeigen schreibt Spalten, Ordnung und Zeilen.
func listeZeigen(
	ctx context.Context, naht *app.Fremdliste, an *schreiber, setID int64,
) int {
	aufbau, fehler := naht.Aufbau(ctx, setID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{
			"spalten":     listenspalten(aufbau.Spalten),
			"zeilen":      listenzeilen(aufbau),
			"sortierfeld": aufbau.Ordnung.Feld,
			"aufsteigend": aufbau.Ordnung.Aufsteigend,
		})
	}

	sichtbare := sichtbareSpalten(aufbau.Spalten)

	// **Die Kopfzeile nennt die sichtbaren Spalten**, mit Tabulatoren wie
	// `mse karten` — damit `cut` und `awk` damit umgehen können.
	an.sag("Nr")

	for _, spalte := range sichtbare {
		an.sag("\t%s", aufschriftOder(spalte))
	}

	an.sag("\n")

	for _, zeile := range aufbau.Zeilen {
		an.sag("%d", zeile.Nummer+1)

		for _, spalte := range sichtbare {
			an.sag("\t%s", zeile.Werte[spalte.Feld])
		}

		an.sag("\n")
	}

	return an.abschluss()
}

// listenspalten macht aus den Spalten eine Liste — **immer eine**.
func listenspalten(spalten []domain.Listenspalte) []map[string]any {
	liste := make([]map[string]any, 0, len(spalten))

	for _, spalte := range spalten {
		liste = append(liste, map[string]any{
			"feld":       spalte.Feld,
			"aufschrift": spalte.Aufschrift,
			"sichtbar":   spalte.Sichtbar,
		})
	}

	return liste
}

// listenzeilen macht aus den Zeilen eine Liste — ebenso.
func listenzeilen(aufbau domain.Kartenlistenaufbau) []map[string]any {
	liste := make([]map[string]any, 0, len(aufbau.Zeilen))

	for _, zeile := range aufbau.Zeilen {
		liste = append(liste, map[string]any{
			"karteID": zeile.KarteID,
			"nummer":  zeile.Nummer + 1,
			"werte":   zeile.Werte,
		})
	}

	return liste
}

// sichtbareSpalten gibt nur die, die zu sehen sind.
func sichtbareSpalten(spalten []domain.Listenspalte) []domain.Listenspalte {
	sichtbare := make([]domain.Listenspalte, 0, len(spalten))

	for _, spalte := range spalten {
		if spalte.Sichtbar {
			sichtbare = append(sichtbare, spalte)
		}
	}

	return sichtbare
}

// aufschriftOder gibt die Aufschrift der Spalte — oder ihren Feldnamen.
//
// **Der Rückfall steht hier und nicht im Dienst**: was ein Paket nicht
// beschriftet, beschriftet jede Fläche für sich, und eine Kommandozeile
// schreibt den rohen Namen.
func aufschriftOder(spalte domain.Listenspalte) string {
	if spalte.Aufschrift != "" {
		return spalte.Aufschrift
	}

	return spalte.Feld
}

// mitListe öffnet den Bestand und baut die Naht.
func mitListe(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Fremdliste) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.Listenflaechennaht(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
