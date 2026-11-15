package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einordnen bildet einen SQLite-Fehler auf eine Fehlerart der Domäne ab.
//
// **Über den Fehlercode, nicht über den Text** (AGENTS.md §8). Ein
// Zeichenkettenvergleich überlebt die erste Fassung des Treibers nicht, in der
// jemand die Meldung umformuliert — und dann gilt ein Konflikt als unbekannter
// Fehler, das Frontend lädt nicht nach, und niemand versteht, warum.
//
// Bis zum Zweitblick über Phase 2 reichte `storage` seine SQLite-Fehler
// unverändert nach oben: ein doppelter Feldwert endete am Primärschlüssel und
// kam als „SQL logic error" an, statt als Konflikt.
//
// Abgebildet wird **nur**, was hier wirklich vorkommt: eine verletzte
// Bedingung. Weitere Codes aufzunehmen, ohne einen Fall dafür zu haben, hieße
// eine Zuordnung zu behaupten, die niemand geprüft hat.
//
// **Nicht exportiert, und das ist die Grenze selbst.** Ein SQLite-Fehler
// entsteht nirgends außerhalb dieses Pakets; wäre die Funktion exportiert, wäre
// sie eine Einladung an `service`, einen zu deuten — und damit an die Schicht,
// die von SQLite nichts wissen soll, ihn zu kennen.
//
// # Der Fremdschlüssel ist etwas anderes als der Primärschlüssel
//
// Bis Phase 5 galt **jede** verletzte Bedingung als Konflikt, und #89 hielt
// fest, dass das zu breit ist — auseinandergelegt werde es „an einem Fall, der
// ihn erzeugt". Der Fall ist da: `Kartenschreiber.FeldSetzen` gegen eine Karte,
// die es nicht gibt, verletzt den **Fremdschlüssel**. Das heißt „nicht
// gefunden", nicht „überholt", und der Unterschied entscheidet, ob das Frontend
// nachlädt oder eine Meldung zeigt (§8).
//
// Getrennt wird deshalb nach dem **erweiterten** Code:
//
//   - `CONSTRAINT_FOREIGNKEY` → `ErrNotFound`, denn die verwiesene Zeile fehlt
//   - `CONSTRAINT_PRIMARYKEY`, `CONSTRAINT_UNIQUE` → `ErrDoppelt`, mit Feld
//     ([doppelt], seit ENG-05)
//   - `CHECK`, `NOT NULL`, Auslöser, Typ → `ErrValidation` (seit ENG-05, mit
//     je einem Fall in `eindeutigkeit_intern_test.go`)
//
// # Und seit dem 2026-08-23: der belegte Bestand
//
// `SQLITE_BUSY` steht hier aus demselben Grund wie die beiden darüber — es
// gibt einen **gemessenen** Fall
// (#886).
// Hält das Fenster eine Schreibsperre länger als `busy_timeout`, endete
// `mse feld …` mit „Migration beginnen: database is locked (5) (SQLITE_BUSY)":
// eine rohe SQLite-Meldung an der Oberfläche, die §8 verbietet, und noch dazu
// eine, die auf eine kaputte Migration zeigt statt auf einen Nachbarn.
//
// # Nach dem Code fragen, nicht nach der Hülle
//
// Der Treiber `ncruces` gibt denselben Fehler in **drei** Gestalten, und zwei
// davon tragen den Code blank, ohne Hülle darum — ein `errors.As` auf
// `*sqlite3.Error` träfe nur eine. Was `sqlitecode` daraus macht, steht dort.
//
// # Und der Treibertext bleibt im Protokoll, nicht in der Meldung
//
// **Nicht `fmt.Errorf("%w: %w", …)`**: damit wanderte „database is locked (5)
// (SQLITE_BUSY)" in `Error()`, von dort in `app.Fehler.Meldung` und an
// Kommandozeile, MCP und REST unübersetzt hinaus — genau das verbietet §8.
// [domain.MitGrund] trennt beides: der Satz sagt, was
// zu tun ist, der Treibertext bleibt in der Kette und geht ins Protokoll
// (Werkstück 1 von Phase 4.13, #1352).
//
// **Der Basiscode, nicht der erweiterte.** SQLite kennt `SQLITE_BUSY`,
// `_SNAPSHOT`, `_RECOVERY` und `_TIMEOUT`; für den Aufrufer heißen alle
// dasselbe — gleich nochmal. Die Fallunterscheidung darüber trennt nach
// erweitertem Code, deshalb wird hier ausdrücklich maskiert.
func einordnen(fehler error) error {
	if fehler == nil {
		return nil
	}

	erweitert, ist := sqlitecode(fehler)
	if !ist {
		return fehler
	}

	// Der erweiterte Code trägt den Basiscode in seinen unteren acht Bit.
	if erweitert.Code() == sqlite.BUSY {
		return domain.MitGrund(domain.ErrBelegt, fehler,
			"ein anderer Zugriff schreibt gerade — versuche es gleich noch einmal")
	}

	// **Ein abgebrochener Lauf ist ein Abbruch und kein Fehler** —
	// angestrichen vom Nahtlauf am 2026-09-07.
	//
	// # Der Befund
	//
	// Zwei Kartenvorschauen zugleich: die ältere wird abgebrochen (§7, die
	// Sonderregel für die Vorschau). Der Nahtlauf erwartet dafür den Code
	// `abgebrochen` und bekam **`unbekannt`** — im Fenster steht dann „Es ist
	// ein unerwarteter Fehler aufgetreten", wo gar nichts unerwartet ist.
	//
	// **Der Grund stand im Protokoll**: `Karte lesen: sqlite3: interrupted`.
	// SQLite bricht eine laufende Abfrage über `sqlite3_interrupt` ab und gibt
	// **seinen eigenen** Fehler zurück, nicht `context.Canceled`.
	//
	// # Warum das keine Probe fand
	//
	// **Ein Kontext, der schon abgebrochen ist, erreicht den Treiber gar
	// nicht**: `database/sql` prüft ihn vorher und gibt `context.Canceled`
	// selbst zurück. Nur ein Abbruch **während** der Abfrage geht diesen Weg —
	// und den stellt kein Test her, der den Kontext vorher schließt. Der erste
	// Versuch, ihn nachzustellen, war deshalb grün und sagte nichts (§10).
	//
	// # Was hier steht und was nicht
	//
	// **`context.Canceled` und nicht `ErrBelegt`**: `internal/app` übersetzt
	// ihn nach `CodeAbgebrochen`, und die Übersetzung steht dort schon ganz
	// oben — „der Abbruch zuerst". Es fehlte nur der Weg dorthin.
	//
	// **Der Grund reist mit** (§8): wer im Protokoll nachsieht, findet
	// weiterhin `sqlite3: interrupted`.
	if erweitert.Code() == sqlite.INTERRUPT {
		return fmt.Errorf("%w: %w", context.Canceled, fehler)
	}

	// **Der Nur-Lese-Bestand** — Werkstück 16 von Phase 4.12.
	//
	// **Der Basiscode, nicht der erweiterte**, aus demselben Grund wie bei
	// `BUSY` darüber: SQLite kennt `READONLY_RECOVERY`, `_CANTLOCK`,
	// `_ROLLBACK`, `_DBMOVED`, `_CANTINIT` und `_DIRECTORY`, und für den
	// Aufrufer heißen alle dasselbe — hier wird nicht geschrieben.
	//
	// **Gemessen am 2026-09-06**: eine Verbindung mit `query_only(1)` gibt für
	// `INSERT`, `DELETE`, `CREATE TABLE` und schon für `BeginTx` mit
	// `_txlock=immediate` den Code 8.
	if erweitert.Code() == sqlite.READONLY {
		return domain.MitGrund(domain.ErrNurLesend, fehler,
			"wechsle in ein eigenes Profil oder bitte um eine Rolle, die schreiben darf")
	}

	// **Der volle Datenträger** — #1812.
	// Er kam als „unerwarteter Fehler" an, obwohl der Rat klar ist; dass der
	// alte Stand dabei stehen bleibt, belegt
	// `TestEinVollerDatentraegerLaesstDenAltenStandStehen`.
	if erweitert.Code() == sqlite.FULL {
		return domain.MitGrund(domain.ErrVoll, fehler,
			"schaffe Platz und versuche es noch einmal — es wurde nichts geschrieben")
	}

	return startlage(erweitert, fehler)
}

// startlage ordnet ein, was beim Öffnen und Heben scheitern kann — F05
// (#2344),
// Befunde A-05 und A-21.
//
// **Drei Lagen und drei Ratschläge**, und nur einer davon ersetzt die Datei:
// eine beschädigte Datei — oder eine, die gar kein Bestand ist — braucht
// eine Sicherung; ein Fehler des Datenträgers sagt nichts darüber, ob die
// Datei heil ist; ein Ort, den SQLite nicht öffnen darf, braucht eine
// Freigabe. Bis hierhin kamen alle drei roh heraus, und der Start riet zu
// allen dreien dasselbe: zurückspielen.
//
// **Der Basiscode, nicht der erweiterte**, wie bei `BUSY` darüber: SQLite
// kennt ein Dutzend `IOERR_…`, und für den Aufrufer heißen alle dasselbe.
func startlage(erweitert sqlite.ExtendedErrorCode, fehler error) error {
	switch erweitert.Code() {
	case sqlite.CORRUPT, sqlite.NOTADB:
		return domain.MitGrund(domain.ErrBeschaedigt, fehler,
			"die Datei lässt sich nicht lesen — spiele eine Sicherung zurück")
	case sqlite.IOERR:
		return domain.MitGrund(domain.ErrEA, fehler,
			"prüfe den Datenträger und versuche es noch einmal")
	case sqlite.CANTOPEN, sqlite.PERM:
		return domain.MitGrund(domain.ErrGesperrt, fehler,
			"gib den Ort frei oder wähle einen anderen")
	}

	return bedingung(erweitert, fehler)
}

// bedingung ordnet eine verletzte Bedingung ein — ENG-05.AK-03 (#1730).
//
// **Eindeutigkeit ist ein doppelter Eintrag** mit Feld, wo eines bekannt ist
// ([doppelt]). **`CHECK`, `NOT NULL`, ein Auslöser und ein falscher Typ sind
// eine Validierung**: sie stehen als zweite Reihe hinter den Prüfungen der
// Domäne (`docs/INVARIANTEN.md`), und wer sie erreicht, hat eine Angabe, die
// der Bestand nicht annimmt. Bis zum 2026-09-23 gingen sie unverändert durch
// und kamen als „unerwarteter Fehler" an; der Treibertext bleibt im
// Protokoll, damit die Lücke in der Domäne zu finden ist.
func bedingung(erweitert sqlite.ExtendedErrorCode, fehler error) error {
	switch erweitert {
	case sqlite.CONSTRAINT_FOREIGNKEY:
		return domain.MitGrund(domain.ErrNotFound, fehler,
			"ein Eintrag, auf den das hier zeigt, liegt nicht im Bestand")

	case sqlite.CONSTRAINT_PRIMARYKEY, sqlite.CONSTRAINT_UNIQUE:
		return doppelt(fehler)

	case sqlite.CONSTRAINT_CHECK, sqlite.CONSTRAINT_NOTNULL,
		sqlite.CONSTRAINT_TRIGGER, sqlite.CONSTRAINT_DATATYPE:
		return domain.MitGrund(domain.ErrValidation, fehler,
			"der Bestand nimmt diese Angabe nicht an — eine seiner Regeln steht dagegen")
	}

	return fehler
}

// uebersetze umschreibt einen Fehler und ordnet ihn zugleich ein.
func uebersetze(fehler error, umschreibung string, werte ...any) error {
	if fehler == nil {
		return nil
	}

	return fmt.Errorf(umschreibung+": %w", append(werte, einordnen(fehler))...)
}

// betroffen zählt die Zeilen, die ein Schreiben getroffen hat.
//
// **Der Satz bleibt beim Aufrufer.** „Das Set führt kein Schlüsselwort" und
// „das Exemplar gibt es nicht" sind zwei verschiedene Auskünfte; eine
// gemeinsame Meldung wäre für beide die falsche. Geteilt wird die Übersetzung
// des Zählfehlers, nicht die Aussage.
// eingefuegt gibt die Kennung der eben eingefügten Zeile.
//
// **Wie [betroffen] daneben**: der Treiberfehler steht im Rückgabewert und
// nicht in einem `if`. Ein Zweig, den kein Lauf betreten kann, sähe aus wie
// eine geprüfte Zusage (§10) — und `LastInsertId` scheitert bei diesem
// Treiber nicht.
func eingefuegt(ergebnis sql.Result) (int64, error) {
	kennung, err := ergebnis.LastInsertId()

	return kennung, uebersetze(err, "die Kennung der neuen Zeile lesen")
}

func betroffen(ergebnis sql.Result) (int64, error) {
	zeilen, err := ergebnis.RowsAffected()

	return zeilen, uebersetze(err, "die betroffenen Zeilen zählen")
}

// sqlitecode holt Basis- und erweiterten Fehlercode aus einem Treiberfehler.
//
// # Warum nicht `errors.As` auf `*sqlite3.Error`
//
// **Weil dieser Treiber denselben Fehler in drei Gestalten gibt** — gemessen
// am 2026-09-06 beim Wechsel auf `ncruces/go-sqlite3` (Werkstück 14 von Phase
// 4.12, [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)):
//
//	INSERT mit NULL            → *sqlite3.Error
//	BeginTx auf Nichtdatenbank → sqlite3.ExtendedErrorCode, blank
//
// Ein `errors.As` auf den **Zeiger** greift an der zweiten daneben, und dann
// kam „die Datei ist keine Datenbank" als „von einem anderen Programm
// geöffnet" an. Beides sind Auskünfte, an denen ein Mensch etwas tut (§8), und
// sie sagen Gegenteiliges: die eine „ersetze die Datei", die andere „schließe
// das andere Fenster".
//
// # Gefragt wird deshalb nach dem Code, und das genügt für alle drei
//
// Der Treiber gibt **jeder** Gestalt ein `As` auf `*ExtendedErrorCode`
// (`error.go`, Zeilen 77, 110 und 140 in v0.35.4) — der Hülle, dem blanken
// Basiscode und dem blanken erweiterten Code. Ein einziges `errors.As` deckt
// damit alle ab, ohne dass hier ein Typschalter stünde, der bei der nächsten
// Gestalt wieder danebengriffe.
//
// # Und warum nur der erweiterte Code herauskommt
//
// **Weil `sqlite3.ExtendedErrorCode` selbst ein `error` ist.** Ein Aufrufer,
// der den Basiscode nicht braucht, schriebe `basis, _, ist := …` — und für
// `errcheck` ist das ein weggeworfener Fehler. Der Treiber rechnet den
// Basiscode ohnehin selbst aus (`ExtendedErrorCode.Code()`, dieselbe
// Maskierung auf die unteren acht Bit, die SQLite anstellt); ihn hier ein
// zweites Mal herauszureichen wäre eine Bequemlichkeit mit einem Wächter
// dagegen.
func sqlitecode(fehler error) (erweitert sqlite.ExtendedErrorCode, ist bool) {
	if !errors.As(fehler, &erweitert) {
		return 0, false
	}

	return erweitert, true
}
