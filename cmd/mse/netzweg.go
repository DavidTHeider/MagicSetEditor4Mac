package main

// **Wie dieses Programm hinausgeht, an der Kommandozeile** — Werkstück N5
// von Phase 5.4
// (#1571).
//
// **Rein lesend.** Geändert wird mit `mse einstellungen setzen netz.proxy.https …`
// — demselben Weg wie jede andere Einstellung. Ein eigener Schreibbefehl wäre
// eine zweite Fassung derselben Frage.
//
// **Und er ist der Beleg für Werkstück 2 des Issues**: `mse` ist ein Klient
// wie das Fenster, und was er hier zeigt, ist der Weg, den er selbst nimmt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// netzwegbefehl nennt den wirksamen Netzweg samt Herkunft.
func netzwegbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("netzweg nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NetzflaecheUeber(db).Netzstand(ctx)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("proxy.http\t%s\t%s\n", antwort.ProxyHTTP.Wert, antwort.ProxyHTTP.Herkunft)
		an.sag("proxy.https\t%s\t%s\n", antwort.ProxyHTTPS.Wert, antwort.ProxyHTTPS.Herkunft)
		an.sag("proxy.aus\t%s\t%s\n", antwort.ProxyAus.Wert, antwort.ProxyAus.Herkunft)
		an.sag("frist.verbindung\t%d\n", antwort.FristVerbindung)
		an.sag("frist.kopf\t%d\n", antwort.FristKopf)
		an.sag("frist.stillstand\t%d\n", antwort.FristStillstand)

		return an.abschluss()
	})
}
