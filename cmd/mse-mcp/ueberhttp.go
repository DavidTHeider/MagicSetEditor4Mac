// **Der zweite Transport des Werkzeugservers** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Was hier steht und was in `internal/mcp`
//
// **Dort das Protokoll, hier der Composition Root.** `internal/mcp` weiß, wie
// „Streamable HTTP" aussieht; dieses Verzeichnis weiß, wo der Bestand liegt,
// welches Zeugnis gilt und wie sich der Dienst ansagt.
//
// # Und warum die Ansage dieselbe ist wie bei `mse-rest`
//
// **Weil ein Posten im Fenster beide liest.** Zwei Ansagen wären zwei
// Leser — und der eine liefe beim ersten Zusatz dem anderen davon.

package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Die Fristen eines hereinkommenden Rufs.
//
// **Dieselben wie bei `mse-rest`**, und aus demselben Grund: ein Ruf ohne
// Frist hält eine Verbindung beliebig lange fest.
const (
	lesefrist    = 10 * time.Second
	schreibfrist = 5 * time.Minute
	leerlauf     = 2 * time.Minute
)

// bindungAus liest die Fahnen — `nil`, wenn keine Adresse genannt ist.
//
// **`nil` und nicht eine Vorgabebindung**: „keine Fahne" heißt hier stdio, und
// eine Vorgabe machte aus dem Altbetrieb still einen Netzdienst.
//
// **`--adresse` ist wiederholbar seit G08** (#2272): jede genannte bekommt
// ihren Lauscher. Die Einstellung `netz.lauschen` liest dieser Prozess
// **nicht** — sonst würde ein Klient, der ihn über die Standardeingabe
// startet, still zum Netzdienst; das Fenster gibt sie beim Start mit.
func bindungAus(argumente []string, aus io.Writer) ([]domain.Bindung, leitung.Zeugnisdateien, error) {
	fahne := flag.NewFlagSet("mse-mcp", flag.ContinueOnError)
	fahne.SetOutput(aus)

	genannt := app.Lauschwahlflagge(fahne,
		"eine Schnittstelle für HTTP, wiederholbar; ohne sie: über die Standardeingabe")
	port := fahne.Int("port", 0, "der Port; 0 heißt: einen freien suchen")
	zeugnis := app.Zeugnisflaggen(fahne)

	if fehler := fahne.Parse(argumente); fehler != nil {
		return nil, leitung.Zeugnisdateien{}, fmt.Errorf("die Argumente lesen: %w", fehler)
	}

	if len(genannt()) == 0 {
		return nil, leitung.Zeugnisdateien{}, nil
	}

	if fehler := domain.PruefeBindung(domain.Bindung{Port: *port}); fehler != nil {
		return nil, leitung.Zeugnisdateien{}, fehler
	}

	wahl, fehler := app.LauschwahlAusGenannten(genannt())
	if fehler != nil {
		return nil, leitung.Zeugnisdateien{}, fehler
	}

	return wahl.Bindungen(*port), zeugnis(), nil
}

// bedieneUeberHTTP macht den Lauscher auf und bedient, bis der Kontext endet.
//
// **verbindungen** ist die Verbindungsgrenze aus den Grenzen beim Start —
// dieselbe Zahl, die `lesen.grenzen` nennt.
func bedieneUeberHTTP(
	ctx context.Context, server *mcp.Server, bindungen []domain.Bindung,
	zeugnisdateien leitung.Zeugnisdateien, db *sql.DB, aus io.Writer, verbindungen int,
) error {
	// **Über HTTP ist er ein offener Server** (OPS-02.AK-04, #1719) — und
	// bekommt denselben Pfadvertrag wie der Dienst.
	server.MitPfadvertrag(app.Dienstpfadvertrag(db))

	if abgeschaltetVorDemStart(ctx, db, domain.FlaecheMCP, "werkzeugserver") {
		return nil
	}

	marke, fehler := rest.NeueMarke()
	if fehler != nil {
		return fehler
	}

	// **Das Zeugnis nur jenseits der Schleife** (0115) — auf ihr wäre es eine
	// Hürde ohne Gewinn.
	var (
		zeugnis *tls.Config
		angabe  app.Dienstzeugnisangabe
	)

	if draussen := app.NachAussen(bindungen); len(draussen) > 0 {
		// **Mit dem gezeigten Profil** — Identität und Einstellungen hängen
		// an ihm; der Grund steht ausführlich in `cmd/mse-rest/zeugnis.go`.
		imProfil, profilfehler := storage.MitGezeigtemProfil(ctx, db)
		if profilfehler != nil {
			return profilfehler
		}

		// **Dieselbe Wahl wie bei `mse-rest`** — Identität oder eigene PKI,
		// an einer Stelle (NET-03, #1707).
		if zeugnis, angabe, fehler = app.DienstflaecheUeber(db).Dienstzeugnis(
			imProfil, draussen, zeugnisdateien, func(befund error) {
				slog.WarnContext(ctx, "das neue Zeugnis gilt nicht — das bisherige bleibt",
					"befund", befund)
			}); fehler != nil {
			return fehler
		}
	}

	// **Jede gewählte Adresse bekommt ihren Lauscher** (G08) — dieselbe Regel
	// wie bei `mse-rest`: eine fehlende wird angesagt, die übrigen laufen.
	lauscher, befunde, fehler := rest.LauschenAufAllen(ctx, bindungen, zeugnis)
	if fehler != nil {
		return fehler
	}

	for _, befund := range befunde {
		slog.WarnContext(ctx, "auf einer gewählten Adresse lauscht der Werkzeugserver nicht",
			"adresse", befund.Adresse, "grund", befund.Grund)
	}

	adressen := rest.Adressen(lauscher)

	// **Dieselbe Ansage wie bei `mse-rest`** — und zwar wörtlich dieselbe:
	// sie steht als [app.Dienstansage] an **einer** Stelle, neben dem Leser
	// im Fenster.
	if _, fehler := fmt.Fprint(aus, app.Dienstansage{
		MitTLS:        zeugnis != nil,
		Adresse:       adressen[0],
		Weitere:       adressen[1:],
		Fehlende:      befunde,
		Pfad:          mcp.Endpunkt,
		Fingerabdruck: angabe.Fingerabdruck,
		Aussteller:    angabe.Aussteller,
		GueltigBis:    angabe.GueltigBis,
		Marke:         marke,
	}.Text()); fehler != nil {
		return errors.Join(
			fmt.Errorf("die Marke ansagen: %w", fehler), lauscher.Close())
	}

	// **Die Elternwache erst jetzt** — vorher gehörte die Standardeingabe dem
	// Protokoll. Über HTTP ist sie frei, und das Fenster hält sie offen,
	// damit dieser Prozess sein Ende mitbekommt.
	laufender, aufhoeren := context.WithCancel(ctx)
	defer aufhoeren()

	_ = elternwacheAufziehen(laufender, aufhoeren)

	// **Über HTTP hat der Werkzeugserver einen eigenen Schalter** — OPS-10.AK-03
	// (#1908): ist er aus, endet dieser Prozess, statt beim Einschalten von
	// selbst wieder zu antworten.
	server.MitSchaltern(app.SchalterwacheUeber(db).Tor(domain.FlaecheMCP, "werkzeugserver", aufhoeren))

	// **Dieselben eigenen Namen wie beim Dienst** (SEC-11, #2140): ohne sie
	// ließ der Werkzeugserver jede Anfrage ohne `Origin` durch, auch mit
	// fremdem `Host`.
	// **Und jede Adresse, auf der er lauscht** (G08): über `::1` wies er
	// seit N6 jede Anfrage als fremde Herkunft ab.
	namen := app.Rechnerteile(adressen)

	// **Gepaart über die gelungenen Bindungen** (Befund 3 des Zweitblicks
	// über Phase G): fiel eine aus, bekäme jede danach sonst die Adresse
	// einer anderen und deren Namen.
	for stelle, bindung := range domain.Gelungene(bindungen, befunde) {
		if stelle < len(adressen) {
			namen = append(namen, app.EigeneDienstnamen(ctx, bindung, adressen[stelle], angabe.Namen)...)
		}
	}

	return bediene(laufender, lauscher, mitVorgangskopf(mcp.UeberHTTPMit(server, marke, namen...)), verbindungen)
}

// bediene fährt den Server, bis der Kontext endet.
//
// **Mit derselben Verbindungsgrenze wie `mse-rest`** (#2003) — und mit
// demselben Lauscher, nicht mit einer Abschrift: bis zum 2026-10-02 nahm der
// Werkzeugserver über HTTP beliebig viele Verbindungen an. Eine darüber
// wartet, bis ein Platz frei wird, und eine ruhende gibt ihren Platz her,
// sobald jemand wartet (`internal/rest/gedraenge.go`, QA-08).
func bediene(
	ctx context.Context, lauscher net.Listener, behandler http.Handler, verbindungen int,
) error {
	gedraenge := rest.HoechstensGleichzeitig(lauscher, verbindungen)

	server := &http.Server{
		Handler:           behandler,
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: lesefrist,
		ReadTimeout:       lesefrist,
		WriteTimeout:      schreibfrist,
		IdleTimeout:       leerlauf,
		ConnState:         gedraenge.Zustand,
	}

	abgebaut := make(chan error, 1)

	// **Der Abbaukontext hängt bewusst nicht am Ruf.** Der des Postens ist
	// gerade abgebrochen worden; ein `Shutdown` daran bräche sofort ab und
	// ließe laufende Anfragen fallen. `gosec` sieht nur, dass hier
	// `context.Background` steht (G118) — was fehlt, ist der Grund, und der
	// steht hier.
	//#nosec G118 -- der Grund steht darüber
	go func() {
		<-ctx.Done()

		abbau, fertig := context.WithTimeout(context.Background(), lesefrist)
		defer fertig()

		beendet := make(chan error, 1)
		go func() { beendet <- server.Shutdown(abbau) }()

		// **Wer nichts sagt, geht sofort** ([rest.Gedraenge.StummeSchliessen])
		// — `Shutdown` wartete sonst fünf Sekunden auf eine vorgewählte
		// Verbindung, dieselbe Klasse wie im Dienst.
		gedraenge.StummeSchliessen()

		abgebaut <- <-beendet
	}()

	if fehler := server.Serve(gedraenge); fehler != nil &&
		!errors.Is(fehler, http.ErrServerClosed) {
		return fmt.Errorf("MCP über HTTP bedienen: %w", fehler)
	}

	return <-abgebaut
}

// mitVorgangskopf übernimmt eine mitgebrachte Vorgangskennung — dieselbe
// Regel wie beim Dienst (OPS-05.AK-02, Befund 7 des Zweitblicks über Phase
// 6.4): nur in der Gestalt, die dieses Programm selbst vergibt. Die
// Ausstattung jedes Rufs behält sie, statt eine eigene zu würfeln.
func mitVorgangskopf(weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		if kennung, gut := domain.VorgangskennungAus(anfrage.Header.Get(rest.Vorgangskopf)); gut {
			anfrage = anfrage.WithContext(app.MitVorgang(anfrage.Context(), kennung))
		}

		weiter.ServeHTTP(schreiben, anfrage)
	})
}
