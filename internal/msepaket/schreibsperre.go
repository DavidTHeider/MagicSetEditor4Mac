// **Ein Schreiber je Paket** — der Befund des Zweitblicks durch Codex über
// Phase 3.3.
//
// # Der Befund
//
// [DateiSchreiben] liest den Stand, vergleicht ihn mit der erwarteten
// Prüfsumme und tauscht dann die Datei. Zwischen dem Vergleich und dem Tausch
// lag ein Fenster: zwei Fenster oder Prozesse lesen denselben alten Inhalt,
// beide bestehen die Prüfung, und beide schreiben. Der zweite schiebt dabei
// die **neue** Datei des ersten in die Sicherung, setzt seinen eigenen Stand
// ein — und beide Aufrufe melden Erfolg.
//
// Damit war die Zusage falsch, ein zweites Fenster werde „abgewiesen statt
// überschrieben". Gemessen: von sechs gleichzeitig startenden Prozessen kamen
// **zwei bis vier** durch (`schreibwettlauf_test.go`).
//
// `useVorgang` im Fenster half dagegen nicht: es hält einen `useRef` **einer**
// React-Instanz.
//
// # Die Antwort ist dieselbe wie beim Bestand
//
// Eine Nachbardatei neben dem Paket, ausschließlich gehalten über **Lesen,
// Prüfen und Tauschen** — der Kniff steht in `internal/dateisperre` und wird
// von `storage`, `netz` und `scryfall` schon benutzt.
//
// **Sie liegt neben dem Paket und nicht darin.** Ein Archiv wird beim Tausch
// umbenannt, und eine Sperrdatei, die mitwandert, sperrt danach etwas
// anderes; in einem Verzeichnispaket wäre sie außerdem eine Datei **des
// Pakets** (dieselbe Überlegung wie bei der Sicherung, `schreibweg.go`).
//
// # Warum das den zweiten Schreiber abweist statt ihn zu übermalen
//
// Wer die Sperre bekommt, liest den Stand **hinter** ihr. Der zweite liest
// deshalb, was der erste geschrieben hat — und dessen Prüfsumme ist nicht
// mehr die erwartete. Er bekommt [domain.ErrConflict], also genau die
// Meldung, die sagt, was zu tun ist.
//
// # Was sie nicht kann
//
//   - **Sie ist beratend.** Wer die Datei mit einem Texteditor anfasst, wird
//     nicht aufgehalten — kein Unix-Dateisystem hält jemanden auf. Dagegen
//     steht weiter die Prüfsumme.
//   - **Auf Windows gibt es sie nicht.** `internal/dateisperre` ist dort ohne
//     Wirkung (`dateisperre_andere.go`), und damit bleibt dort das Fenster
//     zwischen Prüfung und Tausch offen. Windows wird beschrieben und nicht
//     gebaut (§12); der Satz steht hier, damit niemand die Zusage dorthin
//     mitnimmt.
//   - **Sie überlebt keinen Absturz** — das Betriebssystem gibt sie mit dem
//     Prozess frei, und das ist richtig.

package msepaket

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schreibsperrendung hängt an den Paketpfad.
const Schreibsperrendung = ".schreibsperre"

// Schreibfrist ist, wie lange auf einen laufenden Schreibvorgang gewartet wird.
//
// **Gewartet und nicht sofort abgewiesen**: ein Tausch dauert zwei
// Umbenennungen, und wer in dieser Zeit dazukommt, soll den fertigen Stand
// sehen und daran scheitern — nicht an einer Sperre, die gleich wieder fällt.
//
// **Zehn Sekunden**, dieselbe Frist wie beim Bestand: ein Schreibvorgang, der
// länger dauert, ist keiner mehr.
var Schreibfrist = 10 * time.Second

// Schreibtakt ist der Abstand zwischen zwei Versuchen.
var Schreibtakt = 20 * time.Millisecond

// beimSchreiben hält die ausschließliche Sperre, solange geschrieben wird.
//
// **Ein Fehlschlag hält hier auf**, anders als beim Öffnen des Bestands. Dort
// ist die Sperre eine Vorsichtsmaßnahme neben einer zweiten Probe; hier ist
// sie die **einzige** Absprache zwischen zwei Schreibern, und ohne sie wäre
// genau der Verlust wieder da, um dessentwillen es sie gibt.
func beimSchreiben(ctx context.Context, ort string) (fertig func(), fehler error) {
	griff, err := dateisperre.Oeffnen(ort + Schreibsperrendung)
	if err != nil {
		return nil, fmt.Errorf("die Schreibsperre zu %q nehmen: %w", ort, err)
	}

	if wartefehler := griff.AlleinBis(ctx, Schreibfrist, Schreibtakt); wartefehler != nil {
		_ = griff.Schliessen() //nolint:errcheck // der Griff geht ohnehin weg

		return nil, domain.MitGrund(domain.ErrConflict, wartefehler,
			"an %q schreibt gerade jemand anderes — versuche es gleich noch einmal",
			ort)
	}

	return func() { _ = griff.Schliessen() }, nil //nolint:errcheck // die Sperre fällt mit dem Griff
}
