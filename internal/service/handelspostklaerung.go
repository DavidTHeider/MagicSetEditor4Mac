package service

// **Ein Konflikt in der Übergabe führt in die Klärung** — TRADE-03.AK-04
// (#1717).
//
// # Warum nur dort
//
// **Weil die Matrix es nur dort kennt** — `klaeren` geht aus der offenen
// Übergabe, dem lokalen Verbuchen und dem Abgleich (`handelsmatrix.go`).
// Dort heißt dieselbe Operation mit anderem Inhalt: die beiden Seiten meinen
// nicht mehr dasselbe, und das entscheidet ein Mensch. In der Verhandlung ist
// ein abweichender Vorschlag ein Gegenangebot und kein Streit — dort bleibt
// es bei der Antwort von vorhin.
//
// **Geklärt wird von dieser Seite**, nicht vom Absender: er hat die
// Nachricht geschickt, wir haben den Widerspruch gesehen.

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Klaerungsstelle geht einen Schritt — die Leitung.
type Klaerungsstelle interface {
	Schreiten(
		ctx context.Context, vorgang string, schritt domain.Handelsschritt,
		wer string,
	) (domain.Handelsangebot, error)
}

// MitKlaerung lässt einen Konflikt in der Übergabe in die Klärung führen.
func (p *Handelspost) MitKlaerung(stelle Klaerungsstelle) *Handelspost {
	p.klaerung = stelle

	return p
}

// klaeren führt den Vorgang in die Klärung, wo die Matrix es kennt.
//
// **Ein Fehler geht zurück**: der Absender versucht es noch einmal, und die
// Klärung geschieht beim nächsten Anlauf — still vorbei ginge sie nie.
func (p *Handelspost) klaeren(
	ctx context.Context, sender, vorgang string,
) error {
	if p.klaerung == nil {
		return nil
	}

	stand, fehler := p.stand.Stand(ctx, vorgang)
	if fehler != nil {
		return fehler
	}

	for _, ich := range stand.Parteien {
		if ich == sender || !slices.Contains(stand.Moegliche(ich), domain.SchrittKlaeren) {
			continue
		}

		_, fehler := p.klaerung.Schreiten(ctx, vorgang, domain.SchrittKlaeren, ich)

		return fehler
	}

	return nil
}
