package main

// **Was der Dienst beim Start liest und jeder Anfrage mitgibt** — aus
// `dienst.go` herausgelöst, als G08 und G09 es über 400 Zeilen trugen (§2).
// Der Schnitt folgt der Frage: dort steht, **wie** der Dienst läuft; hier,
// **womit** er beginnt (Port, Adressen, Zeugnis) und was jede Anfrage trägt
// (Profil, Modus, Vorgangskennung).

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// portAus liest Port, genannte Adressen und Zeugnisdateien.
//
// **Vorgabe 0**: das Betriebssystem sucht einen freien, und der Dienst sagt
// ihn an. Ein fester Vorgabeport wäre eine Verabredung, die niemand getroffen
// hat — und zwei gleichzeitig gestartete Dienste stritten sich darum.
//
// **Die Adresse ist neu seit N6**
// (#1572),
// **wiederholbar seit G08** (#2272): jede genannte bekommt ihren Lauscher.
// **Ohne eine gilt die Einstellung `netz.lauschen`** und ohne sie die
// Schleife ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)) —
// gelesen wird sie erst am offenen Bestand, in [lauf]. Jenseits der Schleife
// sind TLS und ein Zugang Pflicht; was das durchsetzt, steht in
// `internal/rest`.
func portAus(argumente []string, aus io.Writer) (
	genannt []string, port int, dateien leitung.Zeugnisdateien, ansage string, fehler error,
) {
	fahne := flag.NewFlagSet("mse-rest", flag.ContinueOnError)
	fahne.SetOutput(aus)

	portwert := fahne.Int("port", 0, "der Port; 0 heißt: einen freien suchen")
	genannte := app.Lauschwahlflagge(fahne,
		"eine Schnittstelle, wiederholbar; ohne sie gilt "+domain.NetzLauschen+
			" und sonst "+domain.Schleifenadresse+" — jenseits davon mit TLS")
	zeugnis := app.Zeugnisflaggen(fahne)
	ansagedatei := fahne.String("ansage", "", "die Ansage zusätzlich in diese Datei, je Start ersetzt (0600) — für den Dienst unter launchd (SEC02)")

	if gelesen := fahne.Parse(argumente); gelesen != nil {
		return nil, 0, leitung.Zeugnisdateien{}, "", fmt.Errorf("die Argumente lesen: %w", gelesen)
	}

	if gepruefte := domain.PruefeBindung(domain.Bindung{Port: *portwert}); gepruefte != nil {
		return nil, 0, leitung.Zeugnisdateien{}, "", gepruefte
	}

	// **Geprüft vor dem Bestand**: eine krumme Adresse soll nicht erst eine
	// Migration auslösen.
	if len(genannte()) > 0 {
		if _, gewaehlt := app.LauschwahlAusGenannten(genannte()); gewaehlt != nil {
			return nil, 0, leitung.Zeugnisdateien{}, "", gewaehlt
		}
	}

	return genannte(), *portwert, zeugnis(), *ansagedatei, nil
}

// imProfil hängt an jede Anfrage das Profil, in dem gearbeitet wird.
//
// # Warum je Anfrage und nicht einmal beim Start
//
// **Weil der Dienst lange lebt** und ein `PUT /profil` den Stand ändert, aus
// dem die nächste Anfrage liest (#1351). Ein Kontext, der beim Start
// ausgestattet wurde, trüge danach das **alte** Profil, und `GET /sets` gäbe
// die Sets von vorhin.
//
// **Gefunden hat das der Flächenlauf gegen den gebauten Dienst**, nicht ein
// Test dieses Pakets — dort war die Sache grün.
//
// **Und es ist zugleich die Stelle, an der eine Anmeldung ankommt**
// (Werkstück 4): dann sagt nicht der Bestand, wer arbeitet, sondern die
// Anfrage — und der Träger ist schon der richtige.
func imProfil(db *sql.DB, weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		// **Die Fläche nennt sich** — Werkstück O1 von Phase 5.2 (#1425): ein
		// Nachweis wird nur am Fenster ausgestellt, und die Wache innen
		// hängt den Umfang eines Tokens daneben.
		// **Und die Vorgangskennung entsteht je Anfrage** — Werkstück O9 von
		// Phase 5.2 (#1433). Sie geht als Kopfzeile zurück, und zwar an
		// **jeder** Antwort: der Rumpf einer geglückten ist die Nutzlast, und
		// wer eine langsame Anfrage nachsehen will, braucht den Griff genauso
		// wie bei einer gescheiterten. Gesetzt wird hier, weil ein Kopf nach
		// dem Statuscode nicht mehr ankommt.
		//
		// **Oder sie kommt mit** — OPS-05.AK-02 (#1722): wer den Dienst ruft,
		// darf seine eigene Kennung in derselben Kopfzeile mitgeben, und zwar
		// nur in der Gestalt, die dieses Programm selbst vergibt. Eine andere
		// wird nicht übernommen — sie landete sonst in jeder Protokollzeile.
		mitgebracht := anfrage.Context()
		if kennung, gut := domain.VorgangskennungAus(anfrage.Header.Get(rest.Vorgangskopf)); gut {
			mitgebracht = app.MitVorgang(mitgebracht, kennung)
		}

		gezeichnet := app.MitNeuemVorgang(mitgebracht)
		schreiben.Header().Set(rest.Vorgangskopf, string(app.VorgangAus(gezeichnet)))

		ausgestattet, fehler := storage.MitGezeigtemProfil(
			app.MitDienstmodus(gezeichnet, domain.ModusDienst), db)
		if fehler != nil {
			// **Ohne Profil geht gar nichts**, und das ist kein Fehler der
			// Anfrage: der Bestand hat kein Profil, obwohl die Wanderung eines
			// anlegt. `500` sagt das; `400` schöbe es dem Aufrufer zu.
			schreiben.Header().Set("Content-Type", "application/json")
			schreiben.WriteHeader(http.StatusInternalServerError)

			// **Der Schreibfehler wird verworfen und nicht gemeldet.** Wem
			// soll man melden, dass die Antwort nicht ankam? Der Empfänger ist
			// weg; ein Protokolleintrag wäre die einzige Wirkung, und dafür
			// gibt es hier keine Senke.
			_, _ = schreiben.Write([]byte( //nolint:errcheck // siehe oben
				`{"code":"unbekannt","meldung":"das gezeigte Profil ließ sich nicht lesen"}`))

			return
		}

		weiter.ServeHTTP(schreiben, anfrage.WithContext(ausgestattet))
	})
}
