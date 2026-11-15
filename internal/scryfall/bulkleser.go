package scryfall

// **Die Sammeldatei Karte für Karte** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum das hier steht und nicht beim Suchen
//
// **Weil eine Sammeldatei zwei Fragen bekommt.** „Steht diese Karte darin?"
// beantwortet [BulkKarteNachName]; „gib mir alle" beantwortet der Index in
// A15. Beide gehen denselben Weg durch dieselben JSONL-Zeilen — zwei Leser
// wären zwei Deutungen desselben Formats, und eine von ihnen liefe irgendwann
// hinterher.
//
// # Eine kaputte Zeile kostet ihre Zeile
//
// **Dieselbe Regel wie überall** (§8): in einer Datei mit 400.000 Karten darf
// eine unleserliche nicht die übrigen unbrauchbar machen. Sie kommt als
// [domain.Befund] mit ihrer Zeilennummer zurück, und der Leser geht weiter.

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bulkleser geht die Karten einer Sammeldatei durch.
type Bulkleser struct {
	leser  *bufio.Scanner
	zeile  int
	offene []io.Closer
}

// NeuerBulkleser liest aus einem offenen Strom.
func NeuerBulkleser(quelle io.Reader) *Bulkleser {
	leser := bufio.NewScanner(quelle)
	leser.Buffer(make([]byte, 0, 64<<10), zeilengrenze)

	return &Bulkleser{leser: leser}
}

// OeffneBulk öffnet eine gepackte Sammeldatei.
func OeffneBulk(pfad string) (*Bulkleser, error) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- aus dem Bulk-Ordner oder, über den Dienst, durch den Pfadvertrag (#2073)
	if fehler != nil {
		return nil, fmt.Errorf("die Sammeldatei öffnen: %w", fehler)
	}

	entpackt, gzfehler := gzip.NewReader(datei)
	if gzfehler != nil {
		_ = datei.Close() //nolint:errcheck // der Fehler darunter ist der Grund

		return nil, fmt.Errorf("%w: die Sammeldatei ist kein gzip: %s",
			domain.ErrValidation, gzfehler.Error())
	}

	leser := NeuerBulkleser(entpackt)
	leser.offene = []io.Closer{entpackt, datei}

	return leser, nil
}

// Schliessen gibt die Datei frei.
func (b *Bulkleser) Schliessen() {
	for _, offenes := range b.offene {
		_ = offenes.Close() //nolint:errcheck // gelesen ist gelesen
	}
}

// Naechste gibt die nächste Karte — `nil` ohne Fehler heißt: das war alles.
//
// **Ein Befund ist kein Fehler.** Wer eine unleserliche Zeile trifft, bekommt
// die Karte `nil` **und** einen Befund und ruft weiter; ein Fehler kommt nur
// vom Strom selbst.
func (b *Bulkleser) Naechste() (*Karte, []domain.Befund, error) {
	for b.leser.Scan() {
		b.zeile++

		roh := b.leser.Bytes()
		if len(roh) == 0 {
			continue
		}

		karte, befunde, lesefehler := LiesKarte(roh)
		if lesefehler != nil {
			return nil, []domain.Befund{{
				Ort: fmt.Sprintf("Zeile %d", b.zeile),
				Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"die Zeile ließ sich nicht lesen und wird übergangen: %v",
					lesefehler),
			}}, nil
		}

		return karte, befunde, nil
	}

	if fehler := b.leser.Err(); fehler != nil {
		return nil, nil, fmt.Errorf("die Sammeldatei lesen: %w", fehler)
	}

	return nil, nil, nil
}
