package service

// **Was mit einer schon vermerkten Nachricht geschieht** — aus `handelspost.go`
// gelöst, als die Handelsrichtlinie (#2435) die Datei über die Zeilengrenze
// aus §2 hob; die Frage, die sie beantwortet, ist eine eigene.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schonVermerkt antwortet auf einen schon vermerkten Schlüssel — **auch nach
// einem verlorenen Wettlauf**, wo „bekannt" stand, gleich welcher Inhalt (Z2).
func (p *Handelspost) schonVermerkt(
	ctx context.Context, sender string, nachricht Handelsnachricht,
	vorher *domain.Inboxeintrag, abdruck string,
) (Zustellungsantwort, error) {
	art := domain.Zustellung(vorher, abdruck)

	if art == domain.ZustellungKonflikt {
		if fehler := p.klaeren(ctx, sender, nachricht.Angebot.Vorgang); fehler != nil {
			return Zustellungsantwort{}, fehler
		}
	}

	// **Eine Urkunde wird bei jedem Anlauf gegengezeichnet** (TRADE-03,
	// #1717): kam sie, bevor hier verbucht war, scheiterte die
	// Gegenzeichnung nach dem Vermerk — und „bekannt" hieß danach, dass sie
	// nie mehr versucht wurde. Die Gegenzeichnung ist wiederholbar; die
	// Parteien stehen im Angebot, dessen Abdruck schon angenommen ist.
	if art == domain.ZustellungBekannt {
		if fehler := p.gegenbeleg(ctx, sender, nachricht, nachricht.Angebot); fehler != nil {
			return Zustellungsantwort{}, fehler
		}
	}

	return Zustellungsantwort{Art: art, Ergebnis: vorher.Ergebnis}, nil
}
