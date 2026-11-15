package main

// Alles, was **nicht** der Prozesseinstieg ist.
//
// **Getrennt von `main.go`, damit die Abdeckungsausnahme nur den Einstieg
// deckt** (§3) — dieselbe Trennung wie bei `cmd/mse-mcp`.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Fristen des Servers.
//
// **Sie stehen hier und nicht in `internal/rest`**, weil sie zum Betrieb
// gehören und nicht zur Fachlichkeit. Ohne sie hält eine halb geschriebene
// Anfrage eine Verbindung beliebig lange — und das ist auf der Rückschleife
// zwar kein Angriff von außen, aber ein hängender Prozess.
const (
	lesefrist    = 10 * time.Second
	schreibfrist = 5 * time.Minute // ein Rastern über ein großes Set dauert
	leerlauf     = 2 * time.Minute
	abbaufrist   = 5 * time.Second
)

// lauf startet den Dienst und bedient, bis der Kontext endet.
//
// **Mit Argumenten und Ausgabestrom als Parameter**, obwohl `main` immer
// dieselben übergibt: sonst bliebe der Composition Root ungeprüft — und genau
// dort entstand in Phase 5 der Fehler, den kein Test sah (#130).
func lauf(ctx context.Context, argumente []string, aus io.Writer) error {
	genannt, port, zeugnisdateien, ansagedatei, fehler := portAus(argumente, aus)
	if fehler != nil {
		return fehler
	}

	// **Unter launchd hat die Fehlerausgabe eine Grenze** (SEC02, #2355): sie
	// geht in `dienst.log`, und launchd hängt an, solange es den Dienst gibt.
	// Über der Grenze beginnt sie beim Start neu.
	if ansagedatei != "" {
		if begrenzt := app.FehlerausgabeBegrenzen(os.Stderr, app.Fehlerausgabegrenze); begrenzt != nil {
			fmt.Fprintln(os.Stderr, begrenzt)
		}
	}

	pfad, err := app.BestandsortFuerFlaechen()
	if err != nil {
		return err
	}

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

	// **Migriert und gesichert wird auch hier** — dieselbe Zusage wie in den
	// beiden anderen Composition Roots. Wer den Dienst vor dem Fenster
	// startete, migrierte sonst denselben produktiven Bestand ohne Sicherung.
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

	schreiber, wegfehler := app.Schreibnaht(arbeit, domain.UrheberDienst)
	if wegfehler != nil {
		return wegfehler
	}

	marke, markenfehler := rest.NeueMarke()
	if markenfehler != nil {
		return markenfehler
	}

	// **Die Grenzen beim Start, samt `grenzen.kontingente`** (QA-08.AK-03):
	// Guthaben und Auskunft hängen an denselben.
	grenzen, grenzfehler := app.GrenzenUeber(ctx, db)
	if grenzfehler != nil {
		return grenzfehler
	}

	dienst := rest.NeuerDienst(marke, time.Now,
		rest.Naehte{
			Leser:            storage.NeuerSetLeser(arbeit),
			Schreiber:        schreiber,
			Ausfuhr:          ausfuhr{db: arbeit},
			Abgleich:         fremdabgleich{db: arbeit},
			Bilder:           fremdbilder{db: arbeit},
			Auskunft:         fremdauskunft{db: arbeit},
			Suche:            fremdsuche{db: arbeit},
			Katalog:          fremdkatalog{db: arbeit},
			Peer:             peerweg{db: arbeit},
			Kontakte:         kontaktweg{db: arbeit},
			Sichtfreigaben:   freigabeweg{db: arbeit},
			Tausch:           tauschweg{db: arbeit},
			Post:             postweg{db: arbeit},
			Verbuchung:       verbuchungsweg{db: arbeit},
			Leitung:          leitungsweg{db: arbeit},
			Links:            linkweg{db: arbeit},
			Nachbarn:         nachbarweg{db: arbeit},
			Grenzen:          grenzenauskunft{grenzen: rest.GrenzenVon(grenzen)},
			Gesundheit:       gesundheitsauskunft{db: arbeit},
			Metriken:         metrikauskunft{db: arbeit},
			Meta:             metaauskunft{},
			Vertraeglichkeit: vertraeglichkeitsauskunft{db: arbeit},
			Ruecksprung:      ruecksprungauskunft{db: arbeit, ort: pfad},
			Pflege:           pflegeauskunft{db: arbeit, ort: pfad},
			Ruecksetzung:     ruecksetzungsauskunft{db: arbeit},
			Index:            indexauskunft{db: arbeit, pfad: pfad},
			Recherche:        rechercheauskunft{db: arbeit},
			Kosten:           kostenauskunft{db: arbeit},
			Verweise:         fremdverweise{db: arbeit},
			Marken:           app.Markennaht(arbeit),
			Objektmarken:     app.ObjektmarkenUeber(arbeit),
			Ablage:           fremdablage{db: arbeit},
			Ersetzung:        fremdersetzung{db: arbeit},
			Liste:            fremdliste{db: arbeit},
			// **Die Leute bleiben an der schreibenden Verbindung**
			// (Werkstück 16, #1351): sonst fände niemand aus einem fremden
			// Profil wieder hinaus — der Weg hinaus schreibt den Zeiger.
			// **Nutzer und Freigaben an der schreibenden Verbindung** — die
			// benannte Ausnahme aus Werkstück 16 von Phase 4.12. Sie gehören dem
			// **Menschen** und nicht dem gezeigten Profil: sonst fände niemand
			// aus einem fremden Profil wieder hinaus, und wer auf seinem eigenen
			// etwas freigeben will, könnte es aus einem fremden heraus nicht.
			//
			// **Token, Vorgaben und die Sicherung hängen an `arbeit`**, und das
			// war bis zum 2026-09-06 anders: ein Leser konnte im fremden Profil
			// ein Token anlegen, eine Vorgabe setzen und den ganzen Bestand in
			// eine Datei schreiben. Angestrichen vom Nachlauf zum Zweitblick
			// über Phase 4.12.
			Nutzer:    nutzer{db: db},
			Token:     fremdtoken{db: arbeit},
			Vorgaben:  fremdvorgaben{db: arbeit},
			Freigaben: fremdfreigaben{db: db},

			// **Der Platz hängt am Ort und nicht an der Verbindung** —
			// Werkstück 4 von Phase 4.13 (#1352).
			Platz:          fremdplatz{ort: pfad},
			Deckuebernahme: deckuebernahme{db: arbeit},
			Sammeldaten:    fremdsammeldaten{db: arbeit},
			Anlage:         fremdanlage{db: arbeit},
			Saetze:         fremdsaetze{db: arbeit},
			Konsole:        skriptkonsole{db: arbeit},
			Werkstatt:      paketwerkstatt{},

			// **Die Sammlung samt Preisen** (#1165) — sie hängen an keinem
			// Set, und deshalb steht `/sammlung` neben `/sets`.
			Sammlung: sammlung{db: arbeit},
			Preise:   &preise{db: arbeit},

			Wortwahl:        wortlisten{db: arbeit},
			Woerter:         schluesselwoerter{db: arbeit},
			Parameterhilfen: parameterhilfen{db: arbeit},
			Pack:            pack{db: arbeit},
			Angaben:         setangaben{db: arbeit},
			Probleme:        probleme{db: arbeit},
			Stile:           stile{db: arbeit},
			Zusatz:          zusatz{db: arbeit},
			Pakete:          pakete{db: arbeit},
			Kartenbilder:    kartenbilder{db: arbeit},
			Vorlagen:        vorlagen{db: arbeit},
			Probelauf:       probelauf{db: arbeit},

			// **Die Listeneinfuhr** (#1292).
			Listeneinfuhr: listeneinfuhr{db: arbeit},
			Deckeinfuhr:   deckeinfuhr{db: arbeit},
			Konto:         kontoweg{db: arbeit},

			// **Und die Ausfuhr und die vier Fragen** (#1292).
			Sammlungsausfuhr: sammlungsausfuhr{db: arbeit},
			Sammlungssichten: sammlungssichten{db: arbeit},
			Kartenskripte:    kartenskripte{db: arbeit},
			Auszaehlung:      auszaehlung{db: arbeit},
			Autoersetzungen:  autoersetzungen{db: arbeit},

			// **Decks samt Formaten, Turniere und die Symbolbibliothek** (#1293,
			// #1294, DOM-02) — dieselben Werkstätten wie bei MCP.
			Decks:            app.NeueDeckwerkstatt(arbeit),
			Turniere:         app.NeueTurnierwerkstatt(arbeit),
			Symbolbibliothek: app.NeueSymbolbibliothekswerkstatt(arbeit, domain.UrheberDienst),
			Sprachpakete:     app.NeueSprachwerkstatt(arbeit),

			// **Und das Nachzeichnen eines Bildes** (#1340).
			Symbol: symbolnachzeichnung{db: arbeit},

			// **Und die Massenvorgänge** (#1292).
			Masse:         massenvorgaenge{db: arbeit},
			Zahlen:        zahlen{db: arbeit},
			Verlauf:       verlauf{db: arbeit},
			Einstellungen: einstellungen{db: arbeit},

			// **Der Austausch** — Werkstück O3 von Phase 5.2 (#1427). An
			// der **arbeitenden** Verbindung wie die Einstellungen selbst:
			// die Übernahme schreibt, und der Riegel dafür sitzt im Dienst.
			Austausch: einstellungen{db: arbeit},
			Anleger:   setanleger{db: arbeit},
			Loescher:  setloescher{db: arbeit},

			// **Die Sicherung des ganzen Bestands** (#1169) — ohne das
			// Zurückspielen; warum, steht in `internal/rest/bestandssicherung.go`.
			Sicherung: bestandssicherung{db: arbeit, ort: pfad},
			Dateien:   dateiansicht{},

			// **Und der portable Umschlag** (#1457) — an `arbeit`, weil er
			// ein Set aus **diesem** Profil schnürt und in dieses hereinnimmt.
			Umschlag: umschlag{db: arbeit},

			// **Und die Partien** (G1, #1458) — über dieselbe kopflose
			// Anwendung wie die Kommandozeile und der MCP-Server.
			Partien:        partien{db: arbeit},
			Partienverlauf: partien{db: arbeit},

			// **Und der Netzweg** (#1571) — er beschreibt diesen Prozess.
			Netzweg: netzweg{db: arbeit},

			// **Und die Schnittstellen** (OPS-09, #1907) — nur lesend.
			Schnittstellen: schnittstellen{},

			// **Und die Schalter** (OPS-10, #1908) — nur lesend.
			Schalter: schalter{db: arbeit},

			// **Und die Regelfreigabe** (G4, #1461).
			Regelfreigabe: regelfreigabe{db: arbeit},
			Fassungen:     fassungen{},

			// **Und die Aufträge** (FMT-03, #1692) — im Prozess, wie das
			// Auftragsbuch des Fensters.
			Auftraege: rest.NeuesAuftragsbuch(time.Now),
			// **Und die Vorschau davor** — über denselben Bestand.
			Setvorschau: pakete{db: arbeit},
			// **Und die Übertragung** (FMT-03.AK-03) — neben dem Bestand.
			Dateiablage: app.DienstablageNeben(arbeit, pfad),
		}).MitHaushalt(app.NeueKontingente(grenzen, time.Now), app.NeueZugangssperre(time.Now)).
		// **Der Pfadvertrag** (OPS-02.AK-04, #1719): kein Pfad aus einem
		// verborgenen Winkel des Zuhauses oder von außerhalb.
		MitPfadvertrag(app.Dienstpfadvertrag(arbeit)).
		MitFortschrittsnaht(app.MitFortschritt).
		MitMessung(rufMessen)

	if fehler := app.MessungEinrichten(ctx, db); fehler != nil {
		return fehler
	}

	// **Und die Grenze des Zwischenspeichers** — Nachlauf über Phase 6.4, N9.
	if fehler := app.ZwischenspeicherPflegen(ctx, db); fehler != nil {
		return fehler
	}

	// **Im gezeigten Profil** — die Einstellung `netz.lauschen` hängt an
	// einem, und ohne es starb der Dienst auf jedem Bestand, der noch keine
	// genannte Adresse hatte (gefunden an G09, #2273).
	imProfil, profilfehler := storage.MitGezeigtemProfil(ctx, db)
	if profilfehler != nil {
		return profilfehler
	}

	wahl, wahlfehler := app.DienstflaecheUeber(db).LauschwahlFuerDienst(imProfil, genannt)
	if wahlfehler != nil {
		return wahlfehler
	}

	return bediene(ctx, wahl.Bindungen(port), zeugnisdateien, marke, dienst, db, aus, ansagedatei, grenzen.Verbindungen)
}
