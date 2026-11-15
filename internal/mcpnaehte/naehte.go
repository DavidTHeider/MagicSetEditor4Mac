// Package mcpnaehte füllt die Nähte des MCP-Servers aus dem Bestand — für den
// Server und für die Kommandozentrale im Fenster (U12, #1423).
package mcpnaehte

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die Nähte des MCP-Servers — als Paket, nicht als Hauptprogramm** —
// Werkstück U12 von Phase 5.1
// (#1423).
//
// # Warum sie umgezogen sind
//
// Bis U12 lagen die 42 Adapter, die `mcp.Naehte` aus dem Bestand füllen, in
// `cmd/mse-mcp` — erreichbar nur für den MCP-Server. Die Kommandozentrale im
// Fenster soll **dieselben** Werkzeuge rufen, mit derselben Fachantwort: kein
// zweiter Katalog, keine Schattenbefehle, keine Shell. Dafür müssen die Nähte
// dort liegen, wo zwei Prozesse sie erreichen.
//
// **Was gleich bleibt**: jede Naht ist ein dünner Adapter auf
// `internal/app` — die Fachlogik liegt weiter in `service`. Das Paket kennt
// `app`, `mcp`, `domain`, `service` und `storage`; wer es umgekehrt aus `app`
// importieren will, hat einen Kreis und muss den Weg über eine Schnittstelle
// nehmen (siehe die Kommandozentrale).

// Alle füllt die Nähte für einen Bestand: `db` ist die Verbindung, die den
// Bestand gehört, `arbeit` die, über die gearbeitet wird (dieselbe, solange
// kein Profil gezeigt wird), `pfad` der Ort des Bestands.
func Alle(db, arbeit *sql.DB, pfad string, leser mcp.Leser, schreiber mcp.Schreiber) mcp.Naehte {
	return mcp.Naehte{
		Leser:            leser,
		Schreiber:        schreiber,
		Ausfuhr:          ausfuhr{db: arbeit},
		Abgleich:         fremdabgleich{db: arbeit},
		Bilder:           fremdbilder{db: arbeit},
		Auskunft:         fremdauskunft{db: arbeit},
		Suche:            fremdsuche{db: arbeit},
		Katalog:          fremdkatalog{db: arbeit},
		Grenzen:          grenzenauskunft{db: db},
		Peer:             peerauskunft{db: arbeit},
		Kontakte:         kontaktauskunft{db: arbeit},
		Gesundheit:       gesundheitsauskunft{db: arbeit},
		Metriken:         metrikauskunft{db: arbeit},
		Meta:             metaauskunft{},
		Vertraeglichkeit: vertraeglichkeitsauskunft{db: arbeit},
		Ruecksprung:      ruecksprungauskunft{db: arbeit, ort: pfad},
		Datenpflege:      pflegeauskunft{db: arbeit, ort: pfad},
		Ruecksetzung:     ruecksetzungsauskunft{db: arbeit, ort: pfad},
		Kosten:           kostenauskunft{db: arbeit},
		Index:            indexauskunft{db: arbeit},
		Recherche:        rechercheauskunft{db: arbeit},
		Verweise:         fremdverweise{db: arbeit},
		Marken:           app.Markennaht(arbeit),
		Objektmarken:     app.ObjektmarkenUeber(arbeit),
		Ablage:           fremdablage{db: arbeit},
		Ersetzung:        fremdersetzung{db: arbeit},
		Liste:            fremdliste{db: arbeit},
		// **Die Leute bleiben an der schreibenden Verbindung** (Werkstück 16,
		// #1351): sonst fände niemand aus einem fremden Profil wieder hinaus
		// — der Weg hinaus schreibt den Zeiger. Dieselbe benannte Ausnahme
		// wie am Fenster (`internal/app/verdrahtung.go`) und an der
		// Kommandozeile (`cmd/mse/leutebefehle.go`).
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
		Token:     token{db: arbeit},
		Vorgaben:  vorgaben{db: arbeit},
		Freigaben: freigaben{db: db},

		// **Der Platz hängt am Ort und nicht an der Verbindung** —
		// Werkstück 4 von Phase 4.13 (#1352).
		Platz:           platz{ort: pfad},
		Probleme:        probleme{db: arbeit},
		Deckuebernahme:  deckuebernahme{db: arbeit},
		Sammeldaten:     fremdsammeldaten{db: arbeit},
		Anlage:          fremdanlage{db: arbeit},
		Saetze:          fremdsaetze{db: arbeit},
		Konsole:         skriptkonsole{db: arbeit},
		Wortlisten:      wortlisten{db: arbeit},
		Woerter:         schluesselwoerter{db: arbeit},
		Parameterhilfen: parameterhilfen{db: arbeit},
		Pack:            pack{db: arbeit},
		Anleger:         setanleger{db: arbeit},
		Loescher:        setloescher{db: arbeit},

		// **Die Sicherung des ganzen Bestands** (#1169) — ohne das
		// Zurückspielen; warum, steht in `internal/mcp/bestandssicherung.go`.
		// **Und die Sicherung ebenso** — sie schreibt in eine **andere** Datei
		// und liest die ganze Bestandsdatei, nicht ein Profil. Wer sie
		// abschnitte, nähme einem Leser die Sicherung, ohne ihm etwas zu
		// verwehren, das er nicht ohnehin könnte: die Datei liegt offen
		// ([0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)).
		Sicherung: bestandssicherung{db: arbeit, ort: pfad},
		Dateien:   dateiansicht{},

		// **Und die Vorschau vor dem Einfahren** (FMT-03, #1692) — über
		// denselben Bestand wie das Einfahren selbst.
		Setvorschau: pflege{db: arbeit},

		// **Und der portable Umschlag** (#1457) — an `arbeit`, weil er ein
		// Set aus **diesem** Profil schnürt und in dieses hereinnimmt.
		Umschlag: umschlag{db: arbeit},
		Links:    marktlinke{db: arbeit},

		// **Und die Partien** (G1, #1458) — über dieselbe kopflose
		// Anwendung wie die Kommandozeile.
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

		Angaben:      setangaben{db: arbeit},
		Stile:        stile{db: arbeit},
		Zusatz:       zusatz{db: arbeit},
		Bestand:      bestand{db: arbeit},
		Pflege:       pflege{db: arbeit},
		Kartenbilder: kartenbilder{db: arbeit},

		// **Die Listeneinfuhr** (#1292) — ohne sie braucht jedes Exemplar
		// eine UUID von Hand.
		Listeneinfuhr: listeneinfuhr{db: arbeit},
		Deckeinfuhr:   deckeinfuhr{db: arbeit},
		Konto:         marktkonto{db: arbeit},
		Handel:        handelsstand{db: arbeit},

		// **Und die Ausfuhr** (#1292).
		Sammlungsausfuhr: sammlungsausfuhr{db: arbeit},

		// **Und die vier Fragen** (#1292).
		Sammlungssichten: sammlungssichten{db: arbeit},
		Kartenskripte:    kartenskripte{db: arbeit},
		Auszaehlung:      auszaehlung{db: arbeit},
		Autoersetzungen:  autoersetzungen{db: arbeit},

		// **Und die Decks samt Formaten** (#1293, Phase 4.6).
		Decks:    app.NeueDeckwerkstatt(arbeit),
		Turniere: app.NeueTurnierwerkstatt(arbeit),

		// **Und das Nachzeichnen eines Bildes** (#1340).
		Symbolnachzeichnung: symbolnachzeichnung{db: arbeit},

		// **Und die Massenvorgänge** (#1292).
		Massenvorgaenge: massenvorgaenge{db: arbeit},
		Vorlagen:        vorlagen{db: arbeit},
		Probelauf:       probelauf{db: arbeit},
		Werkstatt:       paketwerkstatt{},
		Symbolwerkstatt: symbolwerkstatt{db: arbeit},

		// **Und die Symbolbibliothek** (DOM-02, #1697) — geschrieben wird als
		// Server, wie bei den Setangaben.
		Symbolbibliothek: app.NeueSymbolbibliothekswerkstatt(arbeit, domain.UrheberServer),

		// **Und die eigenen Übersetzungen** (LOC-02, #1688).
		Sprachpakete: app.NeueSprachwerkstatt(arbeit),

		// **Und die Sammlung samt Preisen** (#1165). Sie hängt an keinem Set:
		// was jemand besitzt, ist ein **Druck** und kein Entwurf.
		Sammlung: sammlung{db: arbeit},
		Preise:   &preise{db: arbeit},
	}
}
