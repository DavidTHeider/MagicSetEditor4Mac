package storage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	// Der Treiber ohne cgo — **seit Werkstück 14 von Phase 4.12
	// `ncruces/go-sqlite3` und nicht mehr `modernc.org/sqlite`**
	// ([0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)).
	//
	// Die Wahl gegen cgo steht in
	// [0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md) und gilt
	// unverändert: mit cgo bräuchte jede Zielplattform eine eigene
	// Werkzeugkette, und die Kreuzübersetzung nach Windows und Linux in der
	// Stufe `types` gäbe es nicht. **Beide Treiber erfüllen sie**; nur einer
	// kann verschlüsseln.
	//
	// **Kein `…/embed` daneben.** Der Baustein liegt seit v0.35 im Treiber
	// selbst; das Paket ist abgekündigt und druckt in seinem `init()` eine
	// Zeile auf die Standardausgabe. Sie hat am 2026-09-06 **fünfzehn**
	// Flächenläufe rot gemacht: der REST-Dienst sagt seine Adresse auf der
	// ersten Zeile an, und die war dann das Banner.
	sqlite3 "github.com/ncruces/go-sqlite3"
	treiber "github.com/ncruces/go-sqlite3/driver"

	// **Die Volltextsuche als Erweiterung.** Sie ist im Standardbau nicht
	// enthalten; Wanderung 0003 legt einen FTS5-Index an, und ohne die
	// Anmeldung je Verbindung endet schon die dritte Wanderung mit „no such
	// module: fts5". **Gemessen am 2026-09-06**: ohne den Haken fielen 509
	// Läufe der Speicherschicht, mit ihm elf.
	"github.com/ncruces/go-sqlite3/ext/fts5"
)

// oeffner macht die Datenbank auf — mit der Volltextsuche je Verbindung.
//
// # Warum nicht `sql.Open`
//
// **Weil FTS5 bei diesem Treiber eine Erweiterung ist** und je **Verbindung**
// angemeldet werden muss. `database/sql` führt einen Pool; eine Anmeldung nach
// dem Öffnen träfe eine Verbindung von vielen, und die nächste käme ohne — das
// ist dieselbe Falle, die bei den Pragmas schon einmal zugeschnappt ist
// (siehe [mitPragmas]).
//
// # Als Variable und nicht als Ruf an Ort und Stelle
//
// **Damit der Fehlerweg erreichbar bleibt.** Ein Öffnen scheitert sonst nur an
// einer unbrauchbaren Datenquelle, und die prüft der Treiber erst bei der
// ersten Verbindung. Ohne diese Naht wäre die Fehlerbehandlung darunter
// unerreichbar — und ein Fehlerpfad, den kein Test betritt, ist eine
// Behauptung darüber, was im Ernstfall geschieht.
//
// # Und in einer Hülle, die den Zusammenbruch des Treibers fängt
//
// **BK-03b** (#2128,
// Vermerk 0175): ein `panic` im Treiber ließ die einzige Verbindung des
// Vorrats ausgeliehen stehen. Deshalb baut dieser Weg den Verbinder selbst,
// statt `treiber.Open` zu rufen — dasselbe, nur mit der Hülle dazwischen
// ([fangverbinder]). Der Rückruf läuft wie dort nach den `_pragma=` der
// Datenquelle.
var oeffner = func(quelle, schluessel string) (*sql.DB, error) {
	innen, err := (&treiber.SQLite{}).OpenConnector(quelle)
	if err != nil {
		return nil, err
	}

	return sql.OpenDB(fangverbinder{innen: innen, vorbereiten: vorbereitenMit(schluessel)}), nil
}

// vorbereitenMit setzt je Verbindung Schlüssel, `synchronous` und die
// Volltextsuche — in dieser Reihenfolge.
func vorbereitenMit(schluessel string) func(driver.Conn) error {
	return func(treiberverbindung driver.Conn) error {
		roh, ok := treiberverbindung.(interface{ Raw() *sqlite3.Conn })
		if !ok {
			return fmt.Errorf("die Verbindung %T ist keine von SQLite", treiberverbindung)
		}

		verbindung := roh.Raw()

		// **Der Schlüssel zuerst und erst dann alles andere.** Die Hülle
		// verlangt das `PRAGMA` „immediately after opening a connection";
		// jede Anweisung davor liefe gegen eine Datei, die noch als Rauschen
		// gelesen wird.
		if schluessel != "" {
			if fehler := verbindung.Exec(
				`PRAGMA hexkey='` + schluessel + `'`); fehler != nil {
				return fehler
			}
		}

		// **Und `synchronous` hier und nicht in der Datenquelle** — Werkstück
		// 7 von Phase 4.13
		// (#1352).
		//
		// Es gehört zu [mitPragmas], und dort stand es zuerst. Der Treiber
		// führt `_pragma=` aber **vor** diesem Rückruf aus, also vor dem
		// Schlüssel — und `synchronous` fasst den Seitenverwalter der
		// Hauptdatenbank an. Auf einem verschlüsselten Bestand kam
		// „invalid _pragma: unable to open database file"; gefunden hat es der
		// Lauf über die Entschlüsselung, nicht der Wächter über die Pragmas.
		//
		// Warum FULL und was es kostet, steht bei [mitPragmas].
		if fehler := verbindung.Exec(`PRAGMA synchronous=FULL`); fehler != nil {
			return fehler
		}

		return fts5.Register(verbindung)
	}
}
