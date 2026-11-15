package main

import (
	"context"
	"database/sql"
	"slices"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// grenzenbefehl nennt die Grenzen — Werkstück O5 von Phase 5.2 (#1429). Was
// ein Ruf kostet und wie viele Arbeiten zugleich laufen, gilt dem Programm;
// **ob die Guthaben gelten, sagt seit QA-08.AK-03 der Bestand**
// (`grenzen.kontingente`), und genannt wird, was ein jetzt gestarteter Dienst
// über ihm anwendete.
func grenzenbefehl(ctx context.Context, ort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("grenzen nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, ort, an, func(ctx context.Context, db *sql.DB) int {
		haushalt, fehler := app.GrenzenUeber(ctx, db)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return grenzenSagen(an, app.GrenzenauskunftVon(haushalt))
	})
}

// grenzenSagen schreibt die Auskunft — als JSON oder als Zeilen.
func grenzenSagen(an *schreiber, grenzen app.GrenzenDTO) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, grenzen)
	}

	an.sag("parallel\n")

	for _, klasse := range sortiert(grenzen.Parallel) {
		an.sag("\t%s\t%d\n", klasse, grenzen.Parallel[klasse])
	}

	an.sag("kosten\n")

	for _, klasse := range sortiert(grenzen.Kosten) {
		an.sag("\t%s\t%d\n", klasse, grenzen.Kosten[klasse])
	}

	an.sag("principal\t%d Einheiten, %d je %d s\n",
		grenzen.Principal.Burst, grenzen.Principal.Nachfuellen, grenzen.Principal.FensterSekunden)
	an.sag("global\t%d Einheiten, %d je %d s, Reserve %d\n",
		grenzen.Global.Burst, grenzen.Global.Nachfuellen, grenzen.Global.FensterSekunden, grenzen.Reserve)
	an.sag("antwort\t%d Bytes\n", grenzen.AntwortBytes)
	an.sag("kontingente\t%s\n", grenzen.Kontingente)
	an.sag("verbindungen\t%d\n", grenzen.Verbindungen)
	an.sag("auftraege\t%d laufend, %d behalten, %d s\n",
		grenzen.Auftraege.Laufend, grenzen.Auftraege.Behalten, grenzen.Auftraege.AufbewahrungSekunden)

	// **Die Kostentafel mit ihren Wegen an jeder Fläche** (#2003): was
	// `mse zeichnen` hier ist, kostet über MCP und REST dasselbe.
	an.sag("kostentafel\n")

	for _, zeile := range grenzen.Kostentafel {
		an.sag("\t%s\t%s\t%d\t%s\n", zeile.Faehigkeit, zeile.Klasse, zeile.Kosten,
			strings.Join(slices.Concat(zeile.CLI, zeile.MCP, zeile.REST), " · "))
	}

	return an.abschluss()
}

// sortiert gibt die Schlüssel einer Tabelle in fester Reihenfolge.
func sortiert(tabelle map[string]int) []string {
	schluessel := make([]string, 0, len(tabelle))
	for eines := range tabelle {
		schluessel = append(schluessel, eines)
	}

	sort.Strings(schluessel)

	return schluessel
}
