// **Die Sammlung pflegen** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Was diese Schicht dazutut
//
// Der Speicher schreibt Zeilen; hier steht, **was gilt**: eine Eingabe wird
// erst geglättet und dann geprüft, und was nicht durchkommt, kommt gar nicht
// erst in den Bestand. Ein `zustand: "prima"` wäre dort sonst eine Zeile, die
// jede spätere Auswertung stolpern lässt.

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungsablage ist, was die Pflege zum Lesen und Schreiben braucht.
//
// **Die Schnittstelle steht hier und wird in `storage` erfüllt** (§2).
type Sammlungsablage interface {
	Anlegen(ctx context.Context, exemplar *domain.Exemplar) (int64, error)
	Aendern(ctx context.Context, exemplar *domain.Exemplar) error
	Loeschen(ctx context.Context, id, fassung int64) error
	Liste(ctx context.Context) ([]*domain.Exemplar, error)
	Fuer(ctx context.Context, printing string) ([]*domain.Exemplar, error)

	// NachDruck gibt die Exemplare einer Druckzeile — auch ohne Alias (A2, #1441).
	NachDruck(ctx context.Context, druckID int64) ([]*domain.Exemplar, error)
}

// Sammlungspflege ist der Anwendungsfall „meine Sammlung führen".
type Sammlungspflege struct {
	ablage Sammlungsablage
}

// NeueSammlungspflege baut sie über einer Ablage.
func NeueSammlungspflege(ablage Sammlungsablage) *Sammlungspflege {
	return &Sammlungspflege{ablage: ablage}
}

// Anlegen nimmt ein Exemplar auf und gibt seine Kennung.
//
// **Führt die Sammlung es schon**, kommt eine [domain.Dublette] zurück — mit
// der Zeile, die es gibt (DOM-03, S04.NAT-01 und 03). Wer sie erhöhen will,
// braucht Kennung und Fassung, und wer zu ihr springen will, die Zeile.
func (s *Sammlungspflege) Anlegen(
	ctx context.Context, exemplar domain.Exemplar,
) (int64, error) {
	if err := vorbereiten(&exemplar); err != nil {
		return 0, err
	}

	id, fehler := s.ablage.Anlegen(ctx, &exemplar)

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) || befund.Key != domain.SchluesselDoppeltesExemplar {
		return id, fehler
	}

	return 0, s.dublette(ctx, exemplar, fehler)
}

// dublette sucht die Zeile, an der die Aufnahme scheiterte.
//
// **Bestenfalls**: lässt sie sich nicht lesen oder ist sie inzwischen weg,
// bleibt die Absage des Bestands — sie sagt, was zu tun ist, und ein
// Lesefehler an ihrer Stelle sagte es nicht.
func (s *Sammlungspflege) dublette(ctx context.Context, neu domain.Exemplar, absage error) error {
	var (
		kandidaten []*domain.Exemplar
		fehler     error
	)

	if neu.Printing != "" {
		kandidaten, fehler = s.ablage.Fuer(ctx, neu.Printing)
	} else {
		kandidaten, fehler = s.ablage.NachDruck(ctx, neu.DruckID)
	}

	if fehler != nil {
		return absage
	}

	for _, vorhanden := range kandidaten {
		if vorhanden.GleicheZeile(neu) {
			return domain.NeueDublette(*vorhanden, absage)
		}
	}

	return absage
}

// Aendern schreibt ein Exemplar zurück.
func (s *Sammlungspflege) Aendern(ctx context.Context, exemplar domain.Exemplar) error {
	if exemplar.ID == 0 {
		return fmt.Errorf("%w: ohne Kennung ist nicht zu sagen, was geändert werden soll",
			domain.ErrValidation)
	}

	if err := vorbereiten(&exemplar); err != nil {
		return err
	}

	return s.ablage.Aendern(ctx, &exemplar)
}

// Loeschen nimmt ein Exemplar aus der Sammlung.
// **Die Fassung gehört dazu.** Sie sagt, welchen Stand der Löschende gelesen
// hat; ein Löschen gegen einen überholten entfernte fremde Arbeit (Befund 10
// des Nachlaufs über Phase 3.4).
func (s *Sammlungspflege) Loeschen(ctx context.Context, id, fassung int64) error {
	return s.ablage.Loeschen(ctx, id, fassung)
}

// Liste gibt die ganze Sammlung.
func (s *Sammlungspflege) Liste(ctx context.Context) ([]*domain.Exemplar, error) {
	return s.ablage.Liste(ctx)
}

// Fuer gibt die Exemplare eines Drucks.
func (s *Sammlungspflege) Fuer(ctx context.Context, printing string) (
	[]*domain.Exemplar, error,
) {
	return s.ablage.Fuer(ctx, printing)
}

// vorbereiten glättet ein Exemplar und prüft es.
//
// **In dieser Reihenfolge.** Wer erst prüft, weist ein `"  en"` ab, das nach
// dem Glätten in Ordnung wäre — und wer nur glättet, lässt einen erfundenen
// Zustand durch.
func vorbereiten(exemplar *domain.Exemplar) error {
	exemplar.Normalisieren()

	befunde := exemplar.Validate()
	if len(befunde) == 0 {
		return nil
	}

	// **Der erste Befund geht hinaus, nicht alle.**
	//
	// # Warum, und was das kostet
	//
	// `app.Fehler` trägt **ein** Feld und **einen** Satz — die Naht zu allen
	// vier Flächen ist so gebaut, seit es sie gibt. Ein `errors.Join` über
	// drei Befunde änderte daran nichts: `Uebersetze` findet mit `errors.As`
	// den ersten und nimmt dessen Satz. Was ankäme, wäre derselbe eine Befund
	// — nur hätte der Code hier so getan, als reisten drei.
	//
	// **Der Preis steht ausgesprochen** (§10): wer drei Felder falsch
	// ausgefüllt hat, korrigiert sie **nacheinander**. Das zu ändern hieße,
	// die Fehlergestalt an allen vier Flächen zu ändern — eine Entscheidung
	// über die Naht (§16) und nicht eine dieses Werkstücks. Sie steht als
	// Frage in Phase 3.5, wo die Formulare gebaut werden.
	return befunde[0]
}
