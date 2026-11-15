// Diese Datei schreibt eine Sicherung des Bestands.
//
// Sie steht für sich, weil sie eine eigene Frage beantwortet: nicht wie der
// Bestand entsteht oder gelesen wird, sondern wie man ihn **wegtragen** kann,
// ohne ihn dafür zu schließen.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sichern schreibt den Bestand als **eine** Datei nach `ziel`.
//
// # Warum `VACUUM INTO` und nicht eine Dateikopie
//
// Der Bestand läuft im WAL-Modus (`datenbank.go`). Neben `mse.db` liegen dann
// `mse.db-wal` und `mse.db-shm`, und alles, was seit dem letzten Checkpoint
// geschrieben wurde, steht **im `-wal`**, nicht in der Hauptdatei. Wer nur
// diese kopiert, bekommt einen Bestand, der sich anstandslos öffnen lässt und
// stillschweigend älter ist, als er aussieht. Genau das tut jemand, der „die
// Datei sichert".
//
// `VACUUM INTO` läuft dagegen auf der **offenen** Datenbank, hält dabei eine
// Lesetransaktion — sieht also einen in sich stimmigen Stand — und schreibt
// **eine** Datei ohne Nebendateien. Das ist die Sicherung, die jemand erwartet,
// wenn er „die Datei" sagt.
//
// Der Test daneben zeigt beides: dass die Sicherung vollständig ist, und dass
// eine nackte Kopie es nicht ist.
//
// # Warum über eine Nachbardatei und einen harten Verweis
//
// SQLite weist ein vorhandenes Ziel ab — **aber nur eines mit Inhalt.**
// Nachgemessen am 2026-08-22: eine Datei mit einem Byte gibt `file is not a
// database`, eine **leere** wird angenommen und gefüllt. Deshalb stand hier
// ein `os.Stat` und danach eine Reservierung des endgültigen Namens mit
// `O_CREATE|O_EXCL` (Zweitblick über Phase N).
//
// **Die Reservierung hatte selbst ein Absturzfenster** (BK-06a, #2130):
// `VACUUM INTO` schrieb in den **endgültigen** Namen, und endete der Prozess
// dazwischen, stand dort eine leere oder halbe Datei. Der nächste Start sah
// sie als „schon da", fand sie untauglich und hielt an — „lege sie beiseite",
// von Hand. Gemessen am 2026-10-07 mit einem zweiten Prozess, der nach der
// Reservierung und mitten im Kopieren endete (`sicherungsabbruch_test.go`).
//
// Jetzt entsteht die Sicherung in einer **Nachbardatei** im selben
// Verzeichnis, wird festgeschrieben (`fsync`) und erst dann unter dem
// endgültigen Namen **veröffentlicht** — über einen harten Verweis:
// `link(2)` legt den Namen an oder scheitert, wenn er schon da ist, und
// beides ist atomar. Damit gilt, was die Reservierung wollte: **nichts wird
// überschrieben**, auch keine Sicherung, die ein zweiter Lauf eben
// fertigstellte. Und ein Absturz lässt höchstens die Nachbardatei liegen,
// nie eine halbe unter dem Namen, den der nächste Start prüft.
//
// | Ausgang | endgültiger Name | Nachbardatei |
// | --- | --- | --- |
// | geglückt | die Sicherung | entfernt |
// | der Name ist schon vergeben | unverändert | entfernt, `ErrConflict` |
// | Kopieren oder Festschreiben scheitert | frei | entfernt |
// | Absturz vor dem Veröffentlichen | frei | bleibt liegen (`.<name>.entsteht-…`) |
// | Absturz nach dem Veröffentlichen | die Sicherung | bleibt liegen, zweiter Name derselben Datei |
//
// **Eine liegengebliebene Nachbardatei räumt niemand von selbst weg**: ob sie
// zu einem Lauf gehört, der gerade noch kopiert, weiß hier keiner. Sie trägt
// einen Punkt vorn und `.entsteht-` im Namen.
func Sichern(ctx context.Context, db *sql.DB, ziel string) error {
	// **Ein vergebener Name wird gleich abgewiesen** — nicht erst nach dem
	// Kopieren eines Bestands von zwei Gigabyte. Entscheidend bleibt der
	// Verweis unten; dies ist nur der schnelle Weg zum selben Satz.
	if _, fehler := os.Lstat(ziel); fehler == nil {
		return schonDa(ziel)
	}

	// #nosec G304 -- das Verzeichnis ist das der Sicherung, die der Aufrufer
	// benennt; das ist der Zweck dieser Funktion und keine Einschleusung.
	platz, fehler := os.CreateTemp(filepath.Dir(ziel), "."+filepath.Base(ziel)+".entsteht-*")
	if fehler != nil {
		return fmt.Errorf("Sicherung nach %s anlegen: %w", ziel, fehler)
	}

	zwischen := platz.Name()

	// **Gleich wieder zu**: die Datei beschreibt SQLite, nicht dieses
	// Programm. Ein Griff auf eine eben angelegte, nie beschriebene Datei
	// scheitert beim Schließen nicht auf eine Art, die ein Test herbeiführen
	// könnte; der Fehler geht deshalb denselben Weg wie die übrigen.
	//
	// **Ausdrücklich in die Klarhülle** — Werkstück 14 von Phase 4.12
	// ([0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md)).
	// `VACUUM INTO` ohne `vfs=` erbt das VFS der Hauptdatenbank, bei einem
	// verschlüsselten Bestand die Chiffre — und eine Sicherung, die nur auf
	// dem Rechner aufgeht, auf dem sie entstand, ist keine. Der Vermerk
	// entscheidet, dass eine Sicherung **Klartext** trägt.
	if fehler = platz.Close(); fehler == nil {
		sicherungsschritt("vor dem Kopieren", zwischen)
		fehler = mustern(ctx, db, "VACUUM INTO ?", AlsURI(zwischen)+"?vfs="+Klarhuelle)
	}

	if fehler == nil {
		fehler = festschreiben(zwischen)
	}

	// **Eingeordnet, und mit dem Journal geräumt** — M02 (#2346): auf einem
	// vollen Datenträger kam hier der Treibertext an die Kommandozeile, und
	// das Journal blieb liegen ([halbeRaeumen]).
	if fehler != nil {
		return errors.Join(fmt.Errorf("Sicherung nach %s schreiben: %w", ziel, einordnen(fehler)), halbeRaeumen(zwischen))
	}

	// **Geprüft, bevor sie ihren Namen bekommt** — F05 (#2344): eine
	// Sicherung, die ihre Prüfung nicht besteht, ist keine, und unter dem
	// gewählten Namen täuschte sie eine vor.
	sicherungsschritt("vor dem Prüfen", zwischen)

	if fehler := geschriebeneSicherungPruefen(ctx, zwischen); fehler != nil {
		return errors.Join(domain.MitGrund(domain.ErrSicherungVerworfen, fehler,
			"versuche es noch einmal, oder sichere an einen anderen Ort"), halbeRaeumen(zwischen))
	}

	sicherungsschritt("vor dem Veröffentlichen", zwischen)

	return veroeffentlichen(zwischen, ziel)
}

// halbeRaeumen entfernt eine Sicherung, die nichts geworden ist — **samt
// ihrem Journal**. `VACUUM INTO` schreibt die neue Datei mit einem Journal
// daneben, und scheitert es am Schreiben, bleibt das Journal stehen:
// gemessen auf einem vollen Abbild (`wanderung09`, 2026-10-14) und unter
// einer Grenze für die Dateigröße (`sicherungsschreibfehler_test.go`).
func halbeRaeumen(zwischen string) error {
	fehler := os.Remove(zwischen)

	if weg := os.Remove(zwischen + "-journal"); !errors.Is(weg, fs.ErrNotExist) {
		fehler = errors.Join(fehler, weg)
	}

	return fehler
}

// veroeffentlichen gibt der fertigen Nachbardatei den endgültigen Namen —
// oder lässt ihn, wem er schon gehört.
func veroeffentlichen(zwischen, ziel string) error {
	fehler := os.Link(zwischen, ziel)

	// **Die Nachbardatei geht in jedem Fall** — nach einem geglückten Verweis
	// ist sie ein zweiter Name derselben Datei. Scheitert das Entfernen dann,
	// ist die Sicherung trotzdem da und vollständig; der Fehler bleibt nur
	// stehen, wenn auch der Verweis scheiterte.
	weg := os.Remove(zwischen)

	if errors.Is(fehler, fs.ErrExist) {
		return errors.Join(schonDa(ziel), weg)
	}

	if fehler != nil {
		return errors.Join(fmt.Errorf("Sicherung nach %s veröffentlichen: %w", ziel, fehler), weg)
	}

	// **Und das Verzeichnis wird festgeschrieben**: erst dann übersteht der
	// neue Name einen Stromausfall. Gemessen am 2026-10-07: `fsync` auf ein
	// Verzeichnis geht unter APFS.
	return festschreiben(filepath.Dir(ziel))
}

// festschreiben bringt eine Datei oder ein Verzeichnis auf die Platte — unter
// macOS über `F_FULLFSYNC`, das `(*os.File).Sync` dort nimmt.
func festschreiben(pfad string) error {
	griff, fehler := os.Open(pfad) // #nosec G304 -- der eigene Pfad dieser Sicherung
	if fehler != nil {
		return fehler
	}

	return errors.Join(griff.Sync(), griff.Close())
}

// schonDa ist der Satz für einen vergebenen Namen.
//
// **`domain.ErrConflict` und nicht `ErrBestand`** (#939): „das Ziel liegt
// schon da" ist ein Konflikt und keine Aussage über die Form des Bestands.
// Erst damit lässt sich der Fall von außen unterscheiden — `internal/app`
// darf `domain` kennen und `storage` nicht (§2).
func schonDa(ziel string) error {
	return fmt.Errorf("%w: %s liegt schon da — eine Sicherung "+
		"überschreibt nichts, wähle einen anderen Namen",
		domain.ErrConflict, filepath.Base(ziel))
}

// sicherungsschritt meldet, wo `Sichern` gerade steht, und welche Datei es
// dabei schreibt — eine Naht für die Absturzprobe (BK-06a, #2130). Im Betrieb
// tut sie nichts.
var sicherungsschritt = func(_, _ string) {}

// Sicherungsbedarf **schätzt**, wie viele Bytes eine Sicherung braucht — die
// **belegten** Seiten, denn `VACUUM INTO` schreibt die freien nicht mit. Das
// Journal daneben zählt nicht mit; es hält während des Kopierens nur wenige
// Seiten (Zweitblick über Phase 2, Befund 10: eine obere Grenze ist es nicht).
// **0, wenn es sich nicht erfahren lässt**: der Bedarf ist eine Auskunft für
// die Einordnung eines Fehlers (M02, #2346), kein eigener.
func Sicherungsbedarf(ctx context.Context, db *sql.DB) int64 {
	var bytes int64

	if db.QueryRowContext(ctx, `SELECT (p.page_count - f.freelist_count) * s.page_size
		FROM pragma_page_count AS p, pragma_freelist_count AS f, pragma_page_size AS s`).Scan(&bytes) != nil {
		return 0
	}

	return bytes
}
