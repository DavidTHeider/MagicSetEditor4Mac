// **Den Lauscher aufmachen und bedienen** — herausgelöst aus `dienst.go`, als
// die Datei mit der Bindung jenseits der Schleife (N6,
// #1572)
// über die Zeilengrenze aus §2 wuchs.
//
// **Der Schnitt folgt der Frage**: dort wird zusammengesetzt, **was** der
// Dienst bedient; hier steht, **wie** er lauscht und wie er wieder aufhört.

package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// bediene macht den Lauscher auf, sagt die Marke an und bedient —
// **verbindungen** ist die Verbindungsgrenze aus den Grenzen beim Start.
func bediene(
	ctx context.Context, bindungen []domain.Bindung, zeugnisdateien leitung.Zeugnisdateien,
	marke string, dienst *rest.Dienst, db *sql.DB, aus io.Writer, ansagedatei string, verbindungen int,
) error {
	// **Und die Wache über die Zugriffstoken** — Werkstück 10 (#1351).
	//
	// **Sie wird gesetzt und nicht abgefragt**: ob ein Bestand technische
	// Profile führt, entscheidet sich erst, wenn jemand eines anlegt. Ohne
	// Token gibt es nichts zu prüfen, und die Prozessmarke bleibt der einzige
	// Nachweis — dieselbe Antwort wie vorher.
	dienst = dienst.MitTokenwache(app.TokenwacheUeber(db))

	// **Und wer schreibt, darf es bei jeder Anfrage neu** — ID-03.AK-02
	// (#1702).
	// Die Verbindung steht seit dem Start; ob **diese** Anfrage schreiben
	// darf, hängt am Profil und an den Freigaben von jetzt — und mit Token am
	// Umfang von jetzt. Der Akteur entsteht dafür je Anfrage aus dem Kontext,
	// den `imProfil` und die Tokenwache gerichtet haben.
	autorisierung := app.AutorisierungAus(db)

	if abgeschaltetVorDemStart(ctx, db, domain.FlaecheREST, "dienst") {
		return nil
	}

	// **Das Zeugnis nur jenseits der Schleife** (0115): dort ist TLS Pflicht,
	// hier wäre es eine Hürde ohne Gewinn — wer auf der Schleife mithören
	// kann, hat den Rechner schon.
	var (
		zeugnis *tls.Config
		angabe  app.Dienstzeugnisangabe
	)

	if draussen := app.NachAussen(bindungen); len(draussen) > 0 {
		gebaut, gemessen, fehler := zeugnisFuerDenDienst(ctx, db, draussen, zeugnisdateien)
		if fehler != nil {
			return fehler
		}

		zeugnis, angabe = gebaut, gemessen
	}

	// **Jede gewählte Adresse bekommt ihren Lauscher** (G08, #2272); eine,
	// die es nicht mehr gibt, wird angesagt und protokolliert, und die
	// übrigen laufen.
	roh, befunde, fehler := rest.LauschenAufAllen(ctx, bindungen, zeugnis)
	if fehler != nil {
		return fehler
	}

	for _, befund := range befunde {
		slog.WarnContext(ctx, "auf einer gewählten Adresse lauscht der Dienst nicht",
			"adresse", befund.Adresse, "grund", befund.Grund)
	}

	adressen := rest.Adressen(roh)

	// **Die Zahl der Verbindungen kommt aus den Grenzen** (#2003) und stand
	// bis dahin als Konstante hier: `mse-mcp` über HTTP nimmt jetzt dieselbe,
	// und `grenzen` nennt sie an jeder Fläche.
	lauscher := rest.HoechstensGleichzeitig(roh, verbindungen)

	// **Die eigenen Namen jenseits der Schleife** — Werkstück N6 von
	// Phase 5.4
	// (#1572).
	//
	// **Ohne sie weist der Dienst jede Anfrage über die LAN-Adresse ab**,
	// obwohl Marke und Zeugnis stimmen — gemessen am 2026-09-15. Der Grund
	// steht bei [rest.HerkunftTaugtUnter]; hier wird nur gerechnet, was
	// dieser Rechner ist.
	//
	// **Ein Befund hier hält nichts auf**: der Port steht, und die zwei
	// Vorgabenamen gelten weiter. Wer dann über die LAN-Adresse fragt,
	// bekommt eine klare Absage statt eines Dienstes, der nicht startet.
	//
	// **Und die Namen des eigenen Zeugnisses** (NET-03). Gerechnet wird an
	// **einer** Stelle für beide Dienste (SEC-11, `app.EigeneDienstnamen`).
	//
	// **Und jede Adresse, auf der er lauscht** (G08): ein Dienst auf `::1`
	// wies seit N6 jede eigene Anfrage ab — er kannte nur `127.0.0.1`.
	namen := app.Rechnerteile(adressen)

	// **Gepaart über die gelungenen Bindungen** (Befund 3 des Zweitblicks
	// über Phase G): fiel eine aus, bekäme jede danach sonst die Adresse
	// einer anderen und deren Namen.
	for stelle, bindung := range domain.Gelungene(bindungen, befunde) {
		if stelle < len(adressen) {
			namen = append(namen, app.EigeneDienstnamen(ctx, bindung, adressen[stelle], angabe.Namen)...)
		}
	}

	dienst = dienst.MitEigenenNamen(namen...)

	// **Die Marke geht auf die Standardausgabe** — für wen den Dienst von
	// Hand oder aus dem Fenster startet: dort steht sie genau so lange, wie
	// der Dienst läuft, und ein Skript nimmt sie sich mit `read -r _ marke`.
	//
	// **Und unter launchd zusätzlich in die Datei aus `-ansage`** (SEC02,
	// #2355, Vermerk 0192): je Start ersetzt, mit 0600 — dort hängte launchd
	// bis dahin jede Ansage an dieselbe Datei an, und nach zehn Neustarts
	// standen zehn Marken darin. Gültig ist immer nur die letzte: eine Marke
	// lebt so lange wie ihr Prozess.
	ansage := app.Dienstansage{
		MitTLS:        zeugnis != nil,
		Adresse:       adressen[0],
		Weitere:       adressen[1:],
		Fehlende:      befunde,
		Fingerabdruck: angabe.Fingerabdruck,
		Aussteller:    angabe.Aussteller,
		GueltigBis:    angabe.GueltigBis,
		Marke:         marke,
	}

	if ansagedatei != "" {
		if fehler := app.AnsageInDatei(ansagedatei, ansage); fehler != nil {
			return errors.Join(fehler, lauscher.Close())
		}
	}

	if _, fehler := fmt.Fprint(aus, ansage.Text()); fehler != nil {
		// **Ohne Ausgabe kein Dienst.** Wer die Marke nicht bekommt, kann ihn
		// nicht benutzen — dann ist ein Fehlschlag ehrlicher als ein Server,
		// der auf eine Marke wartet, die niemand kennt.
		return errors.Join(fmt.Errorf("die Marke ansagen: %w", fehler), lauscher.Close())
	}

	// **Die Elternwache** — Werkstück N6 von Phase 5.4
	// (#1572).
	//
	// **Sie greift nur, wenn ein Fenster das Rohr hält.** Wer den Dienst von
	// Hand startet, hat ein Endgerät an der Eingabe, und dort kommt kein EOF.
	// Der Grund steht ausführlich in `elternwache.go`.
	laufender, aufhoeren := context.WithCancel(ctx)
	defer aufhoeren()

	_ = elternwacheAufziehen(laufender, aufhoeren)

	ctx = laufender

	// **Der Dienst hat einen eigenen Schalter** — OPS-10.AK-03 (#1908): ist
	// er aus, endet dieser Prozess, statt beim Einschalten von selbst wieder
	// zu antworten. Dahinter fragt jeder Endpunkt seinen Bereich.
	dienst.MitSchaltern(app.SchalterwacheUeber(db).Tor(domain.FlaecheREST, "dienst", aufhoeren))

	// **Und vor allem anderen, ob er den Bestand noch kennt** — M02 (#2346):
	// jede Anfrage hält die Freigabe, solange sie läuft, und hat eine neuere
	// Fassung den Bestand unter diesem Prozess gewandert, antwortet er nicht
	// mehr und endet über denselben Weg wie der Schalter.
	var veraltet atomic.Bool

	fassungswache := app.Fassungswache(db, func() {
		veraltet.Store(true)
		aufhoeren()
	})

	dienst = dienst.MitSchemafreigabe(fassungswache).MitSchreibwache(func(anfrage context.Context) error {
		return autorisierung.Pruefe(anfrage, domain.WirkungSchreiben)
	})

	server := &http.Server{
		// **Die Freigabe vor dem Profil** (Nachlauf über Phase 2 von 0.9.0,
		// Befund 3): `imProfil` liest den Bestand schon.
		Handler: dienst.UnterFreigabe(imProfil(db, dienst.Behandler())),

		// **Der Wurzelkontext gehört auch in jede Anfrage.**
		//
		// `net/http` baut den Kontext einer Anfrage aus `BaseContext` und
		// nicht aus dem, mit dem jemand `Serve` gerufen hat. Ohne diese Zeile
		// erreichte ein Abbau des Dienstes eine laufende Anfrage nicht.
		BaseContext: func(net.Listener) context.Context { return ctx },

		ReadHeaderTimeout: lesefrist,
		ReadTimeout:       lesefrist,
		WriteTimeout:      schreibfrist,
		IdleTimeout:       leerlauf,

		// **Eine ruhende Verbindung hält keinen Platz, den jemand braucht**
		// (QA-08, #1739): sonst sperrte ein Pool mit sechzehn
		// `keepAlive`-Verbindungen jeden anderen für den Leerlauf aus.
		ConnState: lauscher.Zustand,
	}

	fertig := make(chan error, 1)
	go func() { fertig <- server.Serve(lauscher) }()

	// **Und die Aufträge enden vor dem Bestand** (BK-07b, #2116): sie sind vom
	// Ruf gelöst, `Shutdown` kennt sie nicht, und `main` schließt den Bestand,
	// sobald dieser Ruf zurückkehrt. **In derselben Frist wie der Server** —
	// es ist ein Ende und nicht zwei.
	select {
	case fehler := <-fertig:
		schluss, ende := context.WithTimeout(context.Background(), abbaufrist)
		defer ende()

		if errors.Is(fehler, http.ErrServerClosed) {
			fehler = nil
		}

		return mitWiederanlauf(&veraltet, errors.Join(fehler, dienst.AuftraegeSchliessen(schluss)))

	case <-ctx.Done():
		// **Sanft schließen und nicht abschneiden.** Ein laufendes Rastern
		// soll seine Antwort noch loswerden; wer länger braucht als die Frist,
		// wird abgeschnitten.
		aufhoeren, ende := context.WithTimeout(context.Background(), abbaufrist)
		defer ende()

		// **Server und Aufträge enden zugleich, nicht nacheinander** — sie
		// hängen nicht voneinander ab, und nacheinander bekam das Auftragsbuch
		// nur, was `Shutdown` von der Frist übrig ließ.
		auftraege := make(chan error, 1)
		go func() { auftraege <- dienst.AuftraegeSchliessen(aufhoeren) }()

		beendet := make(chan error, 1)
		go func() { beendet <- server.Shutdown(aufhoeren) }()

		// **Und wer nichts sagt, geht sofort** ([rest.Gedraenge.StummeSchliessen]):
		// `Shutdown` wartete sonst fünf Sekunden auf eine vorgewählte
		// Verbindung — die ganze Frist.
		lauscher.StummeSchliessen()

		return mitWiederanlauf(&veraltet, errors.Join(<-beendet, <-auftraege))
	}
}
