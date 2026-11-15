// **Ein Protokoll über Prozessgrenzen hinweg** — Befund Z6-02 des
// Audits für 0.3.0 (Entwicklungsarchiv), Phase 3.2
// (#1163,
// #1174).
//
// # Der Befund
//
// [AlleinAmBestand] fragt SQLite, ob jemand anderes die Datei hält — und gibt
// die Sperre **vor** dem Tausch wieder her. Zwischen der Antwort und dem
// Umbenennen liegt ein Fenster; wer es trifft, arbeitet auf einem Inode, den
// es nicht mehr gibt. SQLite nennt genau das undefiniert und
// korruptionsgefährlich.
//
// Der Quelltext dort hat die Lücke seit je benannt und auch gesagt, warum sie
// mit SQLite-Mitteln nicht zu schließen ist: „dann hielte sie eine Verbindung
// auf eine Datei, die gerade verschwindet."
//
// # Die Antwort ist eine Nachbardatei
//
// `bestand.db.sperre` liegt neben dem Bestand und wird **nie** umbenannt.
//
//   - Wer den Bestand **öffnet**, nimmt sie geteilt — und gibt sie zurück,
//     sobald die Datenbank offen ist. **Gibt es die Datei noch nicht**,
//     nimmt er sie ausschließlich: wer sie anlegt, stellt sie auf WAL um
//     ([beimOeffnen]).
//   - Wer **tauscht**, nimmt sie ausschließlich und hält sie über den ganzen
//     Tausch.
//
// # Warum das zusammen genügt, und warum es kurz ist
//
// **Zwei Fälle sind zu trennen.** Ein Prozess, der die Datenbank schon
// **offen hält**, wird von [AlleinAmBestand] gesehen: seine Verbindung hält
// Sperren auf `-wal` und `-shm`, und die exklusive Probe scheitert daran.
// Nachgemessen als Gegenprobe.
//
// Was die Probe **nicht** sehen kann, ist ein Prozess, der genau **jetzt**
// öffnet: zwischen ihrer Antwort und dem Umbenennen. Genau dieses Fenster
// deckt die Nachbardatei ab — der Öffner hält sie, solange er öffnet, und der
// Tauscher hält sie, solange er tauscht. Beide können nicht zugleich.
//
// **Deshalb wird sie nicht über die Lebensdauer gehalten**, und das ist keine
// Sparsamkeit: ein Griff je Bestandspfad, der nie zurückgegeben wird, ist ein
// Leck. Bei einer Grenze von 256 offenen Dateien scheiterte der 250.
// Bestand — mit „unable to open database file", also einer Meldung, die auf
// SQLite zeigt statt auf die Ursache.
//
// # Was sie nicht kann
//
//   - **Ein `sqlite3` im Terminal nimmt sie nicht.** Dafür bleibt
//     [AlleinAmBestand]: die eine fragt die Teilnehmer, die andere die Datei.
//   - **Sie ist beratend.** Wer die Datei ohne diese Nachbardatei anfasst,
//     wird nicht aufgehalten — kein Unix-Dateisystem hält jemanden auf.
//   - **Sie überlebt keinen Absturz**, und das ist richtig: das
//     Betriebssystem gibt sie mit dem Prozess frei.

package storage

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sperrendung hängt an den Bestandspfad.
const Sperrendung = ".sperre"

// Sperrfrist ist, wie lange auf einen laufenden Tausch gewartet wird.
//
// # Warum überhaupt gewartet wird
//
// Ein Tausch dauert zwei Umbenennungen. Wer in dieser Zeit öffnet, soll den
// **fertigen** Stand bekommen und nicht eine Datei, die gleich verschwindet —
// warten ist hier die richtige Antwort und nicht eine Abweisung.
//
// # Und warum nicht unbegrenzt
//
// Ein Tausch, der hängenbleibt, hielte sonst **jedes** Öffnen für immer auf.
// Nach der Frist geht es ohne Sperre weiter: das ist der Stand vor Phase 3.2,
// und der ist schlechter als mit — aber besser als eine Anwendung, die nicht
// mehr startet.
//
// **Zehn Sekunden**: ein Tausch, der länger dauert, ist keiner mehr.
var Sperrfrist = 10 * time.Second

// Sperrtakt ist der Abstand zwischen zwei Versuchen.
var Sperrtakt = 20 * time.Millisecond

// sperrpfad gibt die Nachbardatei — leer, wo es keine geben kann.
//
// **Ohne Datei keine Nachbardatei.** Eine Datenbank im Arbeitsspeicher hat
// keinen Ort, und eine `file:`-URI mit Parametern ist kein Pfad — beides
// kommt aus [Oeffnen] herein, und beides braucht keine Sperre: was nur dieser
// Prozess sieht, tauscht niemand unter ihm aus.
func sperrpfad(bestandspfad string) string {
	if bestandspfad == "" || bestandspfad == ImArbeitsspeicher ||
		strings.HasPrefix(bestandspfad, "file:") {
		return ""
	}

	return bestandspfad + Sperrendung
}

// beimOeffnen hält die Sperre, solange der Aufrufer öffnet — geteilt, und
// ausschließlich für eine Datei, die es noch nicht gibt.
//
// **Ein Fehler hält niemanden auf.** Wer sie nicht nehmen kann, öffnet ohne
// sie; sie schützt den **Tausch**, und der ist selten. Eine Anwendung, die
// wegen einer Nachbardatei nicht startet, tauscht einen kleinen Mangel gegen
// einen großen (§8).
func beimOeffnen(ctx context.Context, bestandspfad string) (fertig func()) {
	pfad := sperrpfad(bestandspfad)
	if pfad == "" {
		return func() {}
	}

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		return func() {}
	}

	// **Eine Datei, die es noch nicht gibt, nimmt sie ausschließlich.** Wer
	// sie anlegt, stellt sie auf WAL um, und dieser Wechsel braucht die
	// exklusive Sperre von SQLite — die meldet einem zweiten Prozess, der
	// dieselbe neue Datei im selben Augenblick öffnet, sofort „belegt", ohne
	// `busy_timeout` abzuwarten. Ausschließlich genommen, öffnet der zweite
	// erst, wenn der erste fertig ist, und findet die Datei schon im WAL
	// (`grundlinienwettlauf_intern_test.go`, Fenster und Dienst beim ersten
	// Start).
	nehmen := griff.TeilenBis
	if _, fehlt := os.Stat(bestandspfad); errors.Is(fehlt, fs.ErrNotExist) {
		nehmen = griff.AlleinBis
	}

	if fehler := nehmen(ctx, Sperrfrist, Sperrtakt); fehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // ohne Sperre geht es weiter

		return func() {}
	}

	return func() { _ = griff.Schliessen() } //nolint:errcheck // die Sperre fällt mit dem Griff
}

// Alleinanspruch nimmt die Nachbardatei ausschließlich.
//
// **Ohne Warten — bis auf einen Erben.** Wer tauschen will, will eine Antwort
// und keine Wartezeit: ein zweiter Verbraucher hört nicht von selbst auf.
// Abgewartet wird nur die [dateisperre.Erbfrist] (#2401).
//
// Die zurückgegebene Funktion gibt den Anspruch zurück. Sie muss laufen, bevor
// der Bestand neu geöffnet wird — sonst stünde der eigene Anspruch dem eigenen
// Öffnen im Weg.
func Alleinanspruch(bestandspfad string) (zurueck func() error, fehler error) {
	pfad := sperrpfad(bestandspfad)
	if pfad == "" {
		// **Ohne Sperre geht es weiter.** Sie ist eine Verbesserung des
		// Schutzes und keine Voraussetzung des Tauschs; was ohne sie bleibt,
		// ist der Stand vor Phase 3.2.
		return func() error { return nil }, nil
	}

	griff, err := dateisperre.Oeffnen(pfad)
	if err != nil {
		// **Ohne Nachbardatei geht es weiter** — dieselbe Antwort wie oben,
		// und derselbe Grund: der Anspruch ist eine Verbesserung des Schutzes
		// und keine Voraussetzung des Tauschs. `nilerr` sähe hier gern eine
		// Weitergabe; sie machte aus einer fehlenden Verbesserung einen
		// Fehlschlag.
		return func() error { return nil }, nil //nolint:nilerr // siehe darüber
	}

	// **Ein Erbe ist kein anderes Programm** (#2401): ein Kind, das dieser
	// Prozess gerade startet, hält die Nachbardatei bis zu seinem `exec`.
	if fehler := griff.AlleinOhneErbe(context.Background()); fehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // nichts zu retten

		return nil, fmt.Errorf("%w: %s wird von einem anderen Programm "+
			"benutzt — schließe es und versuche es erneut",
			domain.ErrConflict, filepath.Base(bestandspfad))
	}

	return griff.Schliessen, nil
}
