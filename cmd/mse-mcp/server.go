package main

// Alles, was **nicht** der Prozesseinstieg ist.
//
// **Getrennt von `main.go`, damit die Abdeckungsausnahme nur den Einstieg
// deckt** (§3). Der Zweitblick über Phase 9 hat es angestrichen: die Ausnahme
// stand auf der ganzen Datei und blendete damit den Composition Root mit aus —
// obwohl `main_test.go` ihn ausführt.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcpnaehte"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// lauf öffnet den Bestand und bedient das Protokoll.
//
// **Mit Strömen als Parameter**, obwohl `main` immer stdio übergibt: ein Test
// kann ihn dann ausführen. Ohne das bliebe der Composition Root des zweiten
// Programms ungeprüft — und genau dort entstand in Phase 5 der Fehler, den kein
// Test sah (#130).
func lauf(
	ctx context.Context, argumente []string, ein io.Reader, aus io.Writer,
) error {
	bindung, zeugnisdateien, fehler := bindungAus(argumente, aus)
	if fehler != nil {
		return fehler
	}

	pfad, err := app.BestandsortFuerFlaechen()
	if err != nil {
		return err
	}

	// **Das Verzeichnis anlegen, nicht voraussetzen.** Wer den Server startet,
	// ohne das Fenster je geöffnet zu haben, hat noch keines — und bekäme sonst
	// einen Fehler über eine Datei, die nirgends liegen kann.
	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
		return fmt.Errorf("Bestandsverzeichnis anlegen: %w", fehler)
	}

	db, err := storage.Oeffnen(ctx, pfad)
	if err != nil {
		// **Und wenn er nicht aufgeht, steht daneben, wie es weitergeht** —
		// Werkstück 5 von Phase 4.13 (#1352).
		return app.MitWegZurueck(ctx, pfad, err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // beim Beenden ist nichts mehr zu retten

	// **Migriert wird auch hier.** Wer den Server startet, ohne das Fenster je
	// geöffnet zu haben, hat eine Datei ohne Schema — und bekäme sonst Fehler
	// über fehlende Tabellen statt eines leeren Bestands.
	//
	// **Und gesichert wird auch hier** (Zweitblick über Phase S). Hier stand
	// `nil`, und der Migrator überspringt damit die Vorbereitung: wer den
	// Server vor dem Fenster startete, migrierte denselben produktiven Bestand
	// **ohne** Sicherung — das Fenster fand danach nichts mehr zu tun und holte
	// sie nicht nach. Zwei Composition Roots, dieselbe Zusage.
	if fehler := storage.Migriere(ctx, db,
		app.SicherungVorMigration(db, filepath.Dir(pfad), app.Fassung)); fehler != nil {
		// **Und wenn er nicht aufgeht, steht daneben, wie es weitergeht** —
		// Werkstück 5 von Phase 4.13 (#1352).
		return app.MitWegZurueck(ctx, pfad, fehler)
	}

	// **Und wer hier nur lesen darf, arbeitet über eine zweite Verbindung** —
	// Werkstück 16 von Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// **Nach der Migration**, denn `Migriere` beginnt unbedingt eine
	// Transaktion — und die scheitert nur lesend (gemessen am 2026-09-06).
	//
	// **Die erste bleibt offen**: sie trägt die Nutzernähte, und ohne sie
	// fände niemand aus einem fremden Profil wieder hinaus. Dieselbe benannte
	// Ausnahme wie am Fenster.
	arbeit := db

	nurLesend, lesefehler := app.NurLesendGefragt(ctx, db)
	if lesefehler != nil {
		return lesefehler
	}

	if nurLesend {
		zweite, oeffnungsfehler := storage.OeffnenNurLesend(ctx, pfad)
		if oeffnungsfehler != nil {
			return oeffnungsfehler
		}

		defer func() { _ = zweite.Close() }() //nolint:errcheck // beim Beenden ist nichts mehr zu retten

		arbeit = zweite
	}

	leser := storage.NeuerSetLeser(arbeit)

	// **Die Grenzen beim Start, samt `grenzen.kontingente`** (QA-08.AK-03):
	// Guthaben und `lesen.grenzen` hängen an denselben — wie beim Dienst.
	grenzen, grenzfehler := app.GrenzenUeber(ctx, db)
	if grenzfehler != nil {
		return grenzfehler
	}

	// **Der Server sagt sich an** (#52). Was er ändert, steht unter seinem Namen
	// im Verlauf; eine Rücknahme, die nicht sagt, wessen Änderung sie umkehrt,
	// ist im Zwei-Schreiber-Betrieb eine Vermutung (0010).

	// **Und über dieselbe fachliche Naht wie das Fenster**
	// (#710).
	//
	// Hier stand der Bestandsschreiber unmittelbar, und damit entstand über MCP
	// ein roher `text`-Wert, den das Paket nie liest — während dasselbe Feld
	// über das Fenster seit #696 in `rule text` und `flavor text` geht.
	// **Derselbe Aufruf, zwei Ergebnisse.** Zusammengesetzt wird deshalb an
	// **einer** Stelle: `app.Schreibnaht`.
	// **Über `arbeit` und nicht über `db`** — Werkstück 16 von Phase 4.12.
	//
	// Hier stand `db`, und damit schrieb der MCP-Server auch dann über die
	// **schreibende** Verbindung, wenn der Zeiger in einem Leserprofil stand:
	// die zweite Verbindung war geöffnet, der Leser las über sie, und das
	// Kartenschreibwerkzeug ging daran vorbei. Angestrichen vom Zweitblick
	// über Phase 4.12 — REST machte dieselbe Naht von Anfang an richtig.
	schreiber, wegfehler := app.Schreibnaht(arbeit, domain.UrheberServer)
	if wegfehler != nil {
		return wegfehler
	}

	// **Die Fassung kommt aus dem Composition Root** (#896): `internal/mcp`
	// ist ein Blatt und kennt die Anwendung nicht.
	// **Und die drei Werkzeuge des Arbeitsablaufs**
	// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
	// einfahren, sichern, zeichnen. Ihre Wache — nichts überschreiben — steht
	// in `ausfuhr.go`, denn sie ist eine Aussage über die Platte, und
	// `internal/mcp` ist ein Blatt.
	server := mcp.NeuerServer(mcp.Werkzeuge(mcpnaehte.MitGrenzenVomStart(
		mcpnaehte.Alle(db, arbeit, pfad, leser, schreiber), grenzen))).
		MitFassung(app.Fassung).
		// **Das Profil hängt an jedem Ruf** und nicht am Wurzelkontext
		// (#1351). Warum, steht an `mcp.Server.ausstatten`: eine Sitzung lebt
		// lange, und `schreiben.profil.wechseln` ändert den Stand, aus dem der
		// nächste Ruf liest.
		MitAusstattung(func(ruf context.Context) (context.Context, error) {
			// **Und die Fläche nennt sich** — Werkstück O1 von Phase 5.2
			// (#1425): ein Nachweis wird nur am Fenster ausgestellt.
			// **Und die Vorgangskennung entsteht je Ruf** — Werkstück O9
			// von Phase 5.2 (#1433). Hier und nicht am Wurzelkontext: eine
			// Sitzung lebt lange, und eine Kennung für alle Rufe wäre keine.
			return storage.MitGezeigtemProfil(
				app.MitNeuemVorgang(app.MitDienstmodus(ruf, domain.ModusMCP)), db)
		}).
		MitKontingent(mcpnaehte.KontingentFuer(app.NeueKontingente(grenzen, time.Now))).
		// **Der Griff ins Protokoll am Fehlertext** (#1433): `internal/mcp`
		// ist ein Blatt und weiß nicht, wo die Kennung liegt.
		MitKennung(func(ruf context.Context) string {
			return string(app.VorgangAus(ruf))
		}).
		// **Und der Code eines Fehlers maschinenlesbar** — F03 (#2342):
		// dieselbe Übersetzung wie im Fenster und im Dienst.
		MitFehlerform(func(ruf context.Context, fehler error) any {
			return app.Uebersetze(ruf, fehler)
		}).
		// **Und wer um Fortschritt bittet, bekommt ihn** — FMT-03 (#1692).
		MitFortschrittsnaht(app.MitFortschritt).
		// **Und jeder Ruf wird gemessen** — OPS-05 (#1722): Verteilung
		// immer, Spur nur, wenn sie eingeschaltet ist.
		MitMessung(func(ruf context.Context, messung mcp.Messung) {
			app.ErfasseRuf(ruf, app.Rufmessung{
				Flaeche: domain.ModusMCP, Weg: messung.Weg, Schreibend: messung.Schreibend,
				Dauer: messung.Dauer, Ausgang: messung.Ausgang,
			})
		})

	// **Die Spur nach den Einstellungen des Geräts** — gelesen beim Start,
	// wie das Protokoll am Fenster (OPS-05, #1722).
	if fehler := app.MessungEinrichten(ctx, db); fehler != nil {
		return fehler
	}

	// **Und die Grenze des Zwischenspeichers** — Nachlauf über Phase 6.4, N9.
	if fehler := app.ZwischenspeicherPflegen(ctx, db); fehler != nil {
		return fehler
	}

	// **Der Ausweis dieses Prozesses** — Werkstück O2 von Phase 5.2 (#1426):
	// aus Datei oder Umgebung, nie aus einem Argument. Mit Token prüft eine
	// Wache **jeden** Ruf — Widerruf, Ablauf und Umfang wirken beim nächsten
	// Werkzeug, nicht erst beim nächsten Start. Ohne Token bleibt der
	// Altbetrieb: das gezeigte Profil, ohne Wache.
	klartext, tokenfehler := mcpnaehte.TokenAusUmgebung(os.Getenv, os.ReadFile, mcpnaehte.Dateirechte)
	if tokenfehler != nil {
		return tokenfehler
	}

	if klartext != "" {
		wache := app.TokenwacheUeber(db)

		server.MitWache(func(ruf context.Context, art mcp.Art) (context.Context, error) {
			// **Die Maschinenidentität geht in den Kontext, nicht hier
			// heraus** (Befund 5 des Zweitblicks über Phase 5.2): der
			// Haushalt dieser Sitzung liest sie von dort, und die Wache des
			// MCP-Servers kennt nur Kontext und Fehler.
			//
			// **Und seit Befund 3 des Nachlaufs liest er sie wirklich.** Die
			// Kontingentnaht bekam vorher keinen Kontext und zählte nach dem
			// Wort `mcp`; dieser Satz stand daneben und war eine Zusage ohne
			// Deckung.
			ausgestattet, _, fehler := wache.Pruefe(ruf, klartext, art == mcp.Schreibend)

			return ausgestattet, fehler
		})
	}

	// **Und wer schreibt, darf es bei jedem Ruf neu** — ID-03.AK-02
	// (#1702).
	//
	// Die Verbindung steht seit dem Start; ob **dieser** Ruf schreiben darf,
	// hängt am Profil, das der Bestand jetzt zeigt, an den Freigaben von
	// jetzt und — mit Token — am Umfang von jetzt. Der Akteur entsteht dafür
	// je Ruf aus dem Kontext, den Ausstattung und Wache gerichtet haben.
	autorisierung := app.AutorisierungAus(db)

	// **Und ob sein Bereich an ist** — OPS-10 (#1908): vor jedem Werkzeug,
	// mit demselben Nein wie an jeder anderen Fläche.
	server.MitSchaltern(app.SchalterwacheUeber(db).Tor(domain.FlaecheMCP, "", nil))

	// **Und vor allem anderen, ob er den Bestand noch kennt** — M02 (#2346):
	// jeder Ruf hält die Freigabe, solange er läuft, und hat eine neuere
	// Fassung den Bestand unter diesem Prozess gewandert, antwortet er nicht
	// mehr und endet — wie bei einem Signal, ein reguläres Ende.
	ctx, beenden := context.WithCancel(ctx)
	defer beenden()

	fassungswache := app.Fassungswache(db, beenden)

	server.MitSchemafreigabe(fassungswache).MitSchreibwache(func(ruf context.Context) error {
		return autorisierung.Pruefe(ruf, domain.WirkungSchreiben)
	})

	// **Und über HTTP, wenn eine Bindung genannt ist** — Werkstück N6 von
	// Phase 5.4
	// (#1572).
	//
	// **Eine erweiternde Änderung** nach
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// ohne die Fahnen bleibt alles, wie es war — stdio, Zeile für Zeile. Wer
	// sie setzt, bekommt einen zweiten Transport an derselben Werkzeugliste.
	if bindung != nil {
		return bedieneUeberHTTP(ctx, server, bindung, zeugnisdateien, db, aus, grenzen.Verbindungen)
	}

	// **Ein Signal ist ein reguläres Ende** — wie über HTTP (BK-07c, #2117).
	// `main` hängt SIGINT und SIGTERM an diesen Kontext; sein Abbruch ist
	// gewollt und kein Fehler.
	if fehler := server.Bediene(ctx, ein, aus); fehler != nil &&
		(!errors.Is(fehler, context.Canceled) || ctx.Err() == nil) {
		return fehler
	}

	return nil
}
