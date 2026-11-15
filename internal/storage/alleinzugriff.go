// **Hält ein zweiter Verbraucher die Bestandsdatei?**
// (#961)
//
// Eine eigene Datei, weil es eine eigene Frage ist: nicht „was steht darin"
// ([PruefeSicherung]), sondern „darf ich sie unter jemandem wegziehen".

package storage

import (
	"context"
	"fmt"
	"path/filepath"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// AlleinAmBestand sagt, ob niemand sonst die Datei geöffnet hat.
//
// # Warum das eine Frage an SQLite ist und nicht ans Betriebssystem
//
// [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md) macht den
// MCP-Server zu einem **eigenen Prozess**. Zwei Prozesse auf einer Datei sind
// der Normalfall dieses Projekts — solange beide nur lesen und schreiben.
// Beim **Zurückspielen** ist es anders: dort wird die Datei umbenannt und ihre
// Journaldateien entfernt, und SQLite nennt genau das undefiniert und
// korruptionsgefährlich (https://www.sqlite.org/howtocorrupt.html,
// „Unlinking or renaming a database file while in use", nachgelesen am
// 2026-08-23).
//
// Nach laufenden Prozessen zu suchen wäre die falsche Frage: es zählt nicht,
// **wer** die Datei hält, sondern **dass** sie jemand hält. Das weiß SQLite,
// und es sagt es über seine Sperren.
//
// # Wie gemessen wird
//
// `locking_mode=exclusive` verlangt **alle** Sperren der Datei, auch die der
// Leser; `_txlock=immediate` nimmt sie mit dem `BEGIN` statt beim ersten
// Schreiben. Gelingt beides, hält sie niemand sonst — und die Rücknahme gibt
// sie sofort wieder her.
//
// **Nachgemessen am 2026-08-23**, mit zwei Verbindungen auf einer Datei:
//
//	bei offenem Leser  → database is locked (5) (SQLITE_BUSY)
//	ohne Leser         → nil
//	danach gewöhnlich öffnen → nil
//
// **Ohne `journal_mode`**, und das ist der Punkt: der erste Anlauf setzte es
// wie [Oeffnen] und scheiterte **auch ohne** zweiten Verbraucher — ein
// `journal_mode`-Wechsel braucht selbst die exklusive Sperre und geriet mit
// ihr aneinander. Eine Prüfung, die immer „belegt" sagt, ist keine.
//
// # Was sie nicht sagt
//
//   - **Wer** die Datei hält. Es kann `mse-mcp` sein, ein zweites Fenster oder
//     ein `sqlite3` in einem Terminal.
//   - **Ob** in der Zwischenzeit jemand hinzukommt. Zwischen der Prüfung und
//     dem Tausch liegt eine Lücke; sie ist klein und nicht null. Wer sie
//     schließen wollte, müsste die Sperre über den Tausch halten — und dann
//     hielte sie eine Verbindung auf eine Datei, die gerade verschwindet.
//   - **Etwas über andere Dateien.** Sie fragt genau diese eine.
func AlleinAmBestand(ctx context.Context, pfad string) error {
	// **Eine kurze Wartezeit und keine lange.** Diese Prüfung soll eine
	// Antwort geben, nicht auf eine warten: wer den Server laufen hat, hat ihn
	// auch in fünf Sekunden noch laufen.
	// **Auch diese Prüfung braucht den Schlüssel** — sonst läse SQLite einen
	// verschlüsselten Bestand als Rauschen, meldete `SQLITE_NOTADB`, und
	// `keinBestand` weiter unten machte daraus „die Datei ist frei".
	//
	// **Ein fehlender Schlüssel hält sie aber nicht auf**, und das ist der
	// Unterschied zu [Oeffnen]. Dort will jemand **an** seine Daten, und die
	// Auskunft „der Schlüssel fehlt" ist die, die er braucht. Hier lautet die
	// Frage „hält jemand anderes diese Datei?", und darauf ist „ich konnte sie
	// nicht aufmachen" die alte, richtige Antwort: **frei**.
	//
	// Ohne diese Zeile galt eine gewöhnliche Textdatei plötzlich als
	// verschlüsselter Bestand ohne Schlüssel — denn wiedererkennen lässt sich
	// eine Chiffre nicht, nur ein Klartextbestand.
	schluessel, _ := schluesselFuer(ctx, pfad) //nolint:errcheck // siehe oben

	quelle := AlsURI(pfad) +
		"?_pragma=busy_timeout(200)&_pragma=locking_mode(exclusive)&_txlock=immediate"

	db, err := oeffner(mitHuelle(quelle, schluessel), schluessel)
	if err != nil {
		return fmt.Errorf("den Bestand %q prüfen: %w", pfad, err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // die Sperre fällt mit der Verbindung

	// **Eine Verbindung, nicht ein Vorrat.** Eine zweite aus demselben Pool
	// wäre der zweite Verbraucher, den die Prüfung sucht.
	db.SetMaxOpenConns(1)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		// **Eine Datei, die kein Bestand ist, hält keine Sperre.**
		//
		// Hier stand bis zum 2026-08-25 nur der Konfliktfall, und damit
		// beantwortete diese Prüfung eine Frage, die sie nicht gestellt hat:
		// `BeginTx` scheitert auch, wenn die Datei **gar keine Datenbank**
		// ist — und dann meldete sie „von einem anderen Programm geöffnet".
		//
		// Das ist nicht nur die falsche Auskunft, sondern die teuerste: der
		// Weg, der diese Prüfung durchläuft, ist das **Zurückspielen einer
		// Sicherung**. Wer eine unbrauchbare Bestandsdatei hat, bekam sie
		// dadurch nicht ersetzt — die Prüfung, die den Bestand schützen soll,
		// verhinderte seine Rettung. Nachgemessen an einer laufenden
		// Anwendung (#962).
		//
		// **Für SQLite gibt es hier nichts zu sperren**: Sperren stehen in
		// der Datei, und diese ist keine. Ob ein anderes Programm sie offen
		// hält, ist auf diesem Weg nicht zu beantworten — und für den
		// Aufrufer ohne Belang, denn er ersetzt sie gleich.
		if keinBestand(err) {
			return nil
		}

		return fmt.Errorf("%w: %s ist von einem anderen Programm geöffnet — "+
			"schließe es und versuche es erneut (%v)",
			domain.ErrConflict, filepath.Base(pfad), err)
	}

	// **Zurückgenommen und nicht festgeschrieben.** Die Transaktion hat nichts
	// getan; sie war die Frage, nicht die Änderung.
	//
	// **Ohne eigenen Zweig für den Rücknahmefehler** — derselbe Grund wie in
	// [Sichern]: eine Transaktion, die nichts ausgeführt hat, scheitert beim
	// Rücknehmen nicht auf eine Art, die ein Test herbeiführen könnte, und ein
	// Zweig, den kein Test betritt, ist eine Behauptung darüber, was im
	// Ernstfall geschieht (§3). Das `defer` darüber gibt die Sperre ohnehin
	// mit der Verbindung zurück.
	_ = tx.Rollback() //nolint:errcheck // siehe darüber

	return nil
}

// keinBestand sagt, ob SQLite die Datei gar nicht als Datenbank ansieht.
//
// **Über `sqlitecode` und nicht über ein eigenes `errors.As`.** Genau hier lief
// beim Treiberwechsel eine der beiden Abfragen desselben Fehlers auf die falsche
// Gestalt — nebenan steht, welche.
//
// **`SQLITE_NOTADB` und nichts sonst.** Ein Zugriffsfehler, ein fehlendes
// Recht, eine belegte Datei — die bleiben Konflikte; nur „das ist keine
// Datenbank" ist die Auskunft, dass hier nichts zu sperren war.
func keinBestand(fehler error) bool {
	erweitert, ist := sqlitecode(fehler)

	return ist && erweitert.Code() == sqlite.NOTADB
}
