// **Ein Set wegnehmen, mit seinen Bildern** —
// #1218.
//
// # Warum das nicht der Bestand allein kann
//
// Die Kaskaden des Schemas räumen jede Zeile, die an `sets` hängt. Was sie
// **nicht** erreichen, liegt auf der Platte: die Bilder, die jemand in dieser
// Anwendung eingesetzt hat, stehen unter `<wurzel>/<setID>/` und nicht in der
// Datenbank. Ein Löschen, das nur die Zeilen nimmt, hinterlässt ein
// Verzeichnis, das niemandem mehr gehört und das nichts wieder aufräumt.
//
// # Warum es kein Papierkorb ist
//
// Siehe [0054](../../docs/entscheidungen/0054-ein-set-laesst-sich-loeschen.md).

package service

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setentferner nimmt ein Set aus dem Bestand.
//
// **Der Name kommt mit** und ist die Bestätigung: der Bestand weist ab, wenn
// er nicht stimmt.
type Setentferner interface {
	SetLoeschen(ctx context.Context, setID int64, name string) error
}

// Setloeschung nimmt ein Set und räumt hinterher.
type Setloeschung struct {
	speicher Setentferner

	// bilder ist die Ablage der eingesetzten Bilder — `nil` ohne Ablage.
	//
	// **Nachgereicht und nicht im Konstruktor**, aus demselben Grund wie bei
	// [Setschreiber.MitBildablage]: nicht jede Naht hat eine.
	bilder *Bildablage
}

// NeueSetloeschung baut sie über einem Speicher.
func NeueSetloeschung(speicher Setentferner) *Setloeschung {
	return &Setloeschung{speicher: speicher}
}

// MitBildablage sagt, wo die eingesetzten Bilder liegen.
func (s *Setloeschung) MitBildablage(ablage *Bildablage) *Setloeschung {
	s.bilder = ablage

	return s
}

// Loeschen nimmt das Set und gibt zurück, was dabei liegen blieb.
//
// **Erst der Bestand, dann die Platte** — und der Reihenfolge nach ist das
// §7 Regel 3 („geschrieben ist geschrieben"): scheitert das Löschen, bleiben
// die Bilder stehen, denn sie gehören einem Set, das es noch gibt. Glückt es
// und das Aufräumen scheitert, ist das ein **Befund** und keine Rücknahme —
// ein Fehler an dieser Stelle ließe die Oberfläche ein Set weiter anzeigen,
// das aus dem Bestand verschwunden ist.
func (s *Setloeschung) Loeschen(ctx context.Context, setID int64, name string) (
	[]domain.Befund, error,
) {
	if fehler := s.speicher.SetLoeschen(ctx, setID, name); fehler != nil {
		return nil, fehler
	}

	return s.bilderRaeumen(setID), nil
}

// bilderRaeumen nimmt das Bildverzeichnis des Sets weg.
func (s *Setloeschung) bilderRaeumen(setID int64) []domain.Befund {
	if s.bilders() == "" {
		return nil
	}

	if fehler := os.RemoveAll(s.bilder.Verzeichnis(setID)); fehler != nil {
		// **ArtUmgebung**: es hängt an diesem Rechner — an Rechten, an einer
		// Datei, die jemand offen hält. Auf einem anderen steht es anders.
		return []domain.Befund{{
			Datei: s.bilder.Verzeichnis(setID),
			Art:   domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Das Set ist weg; seine eingesetzten Bilder ließen sich nicht wegräumen: %v",
				fehler),
		}}
	}

	return nil
}

// bilders gibt den Ort der Ablage — leer, wenn es keine gibt.
func (s *Setloeschung) bilders() string {
	if s.bilder == nil {
		return ""
	}

	return s.bilder.Wurzel()
}
