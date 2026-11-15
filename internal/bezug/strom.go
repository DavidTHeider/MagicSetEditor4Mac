// **Ein Archiv wird geströmt, nicht gesammelt** — Befund Z7-01, letztes
// Werkstück (#1163).
//
// # Der Befund
//
// `laden` gab `[]byte` zurück — bis zu zwei Gigabyte, vollständig im
// Arbeitsspeicher, **bevor** das Auspacken begann. Danach lief `sha256` ein
// zweites Mal darüber, und `gzip` ein drittes Mal.
//
// Ein Bezug braucht das nicht: die Prüfsumme entsteht **während** des Lesens,
// und tar liest ohnehin von vorn nach hinten. Was bleibt, sind die Puffer von
// gzip und tar — Kilobyte statt Gigabyte.
//
// # Was dabei gleich bleibt
//
// **Die Prüfsumme geht über den ganzen Rumpf**, nicht über das, was tar
// gelesen hat. Deshalb wird der Rest nach dem Auspacken **leergelesen**: ein
// Archiv mit Anhängsel ergäbe sonst eine andere Summe als vorher, und die
// Prüfsumme steht in `paketkatalog` und wird verglichen.

package bezug

import (
	"context"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// gedeckelterStrom liest bis zur Grenze und meldet ein Byte darüber.
//
// **Ein eigener Leser und kein `io.LimitReader`**: der endet bei Erreichen der
// Grenze mit `io.EOF`, und ein `EOF` ist für gzip und tar ein **gültiges**
// Ende. Ein Archiv über der Grenze käme damit als „unerwartetes Dateiende"
// heraus — oder, schlimmer, als halb ausgepacktes Paket ohne Fehler.
type gedeckelterStrom struct {
	unten   io.ReadCloser
	adresse string
	grenze  int64
	gelesen int64
}

// Read reicht durch und zählt mit.
func (s *gedeckelterStrom) Read(nach []byte) (int, error) {
	gelesen, err := s.unten.Read(nach)

	s.gelesen += int64(gelesen)

	if s.gelesen > s.grenze {
		return gelesen, domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"Die Antwort von %s ist größer als %d Byte.", s.adresse, s.grenze)
	}

	return gelesen, err
}

// Close gibt die Verbindung frei.
func (s *gedeckelterStrom) Close() error { return s.unten.Close() }

// stroemen öffnet ein Archiv, ohne es zu lesen.
//
// **Ohne selbsttätige Wiederholung**, und das bleibt eine Entscheidung: ein
// Bezug beginnt mit einer ausdrücklichen Zustimmung (#56) und lädt bis zu zwei
// Gigabyte. Eine Wiederholung im Hintergrund machte daraus etwas, das
// weiterläuft — und die Gegenstelle sagt mit 429 gerade, dass sie das nicht
// will. **Deshalb steht hier keine Drossel** (`nil`), obwohl `netz.Strom` eine
// annähme: sie gehört zu einem Dienst, der viele kleine Anfragen stellt.
func stroemen(ctx context.Context, adresse, kenner string) (io.ReadCloser, error) {
	// **Ein 429 wird als das benannt, was es ist** (#899). „antwortet mit 429
	// Too Many Requests" reicht eine Statuszeile weiter; wer sie liest, weiß
	// nicht, ob er etwas falsch gemacht hat oder nur warten muss.
	antwort, err := netz.Strom(ctx, klient(), nil, kennung(kenner), adresse)
	if err != nil {
		if fehler := alsZielbefund(err, "adresse", adresse); fehler != nil {
			return nil, fehler
		}

		if fehler := alsDrosselbefund(err, "adresse", adresse); fehler != nil {
			return nil, fehler
		}

		return nil, fmt.Errorf("Paket beziehen: %w", err)
	}

	if antwort.Status != netz.StatusGut {
		_ = antwort.Rumpf.Close() //nolint:errcheck // der Rumpf eines Fehlers trägt nichts

		return nil, domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"%s antwortet mit %d.", adresse, antwort.Status)
	}

	return &gedeckelterStrom{unten: antwort.Rumpf, adresse: adresse, grenze: hoechstens}, nil
}
