package service_test

// **Die drei Wege der Verwaltung am Doppel** — PLAY-03.AK-05
// (#1711).
//
// # Warum in einer eigenen Datei
//
// **Die Zeilengrenze** (§2): `partieprobe_test.go` stand bei 311 Zeilen. Der
// Typ bleibt einer — Methoden dürfen daneben liegen, und die Frage, die
// diese Datei beantwortet, ist eine eigene.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (p *partieprobe) Loeschen(_ context.Context, kennung string) error {
	if fehler := p.stolpernAuf("loeschen"); fehler != nil {
		return fehler
	}

	if _, gibtEs := p.partien[kennung]; !gibtEs {
		return fmt.Errorf("%w: die Partie %q", domain.ErrNotFound, kennung)
	}

	delete(p.partien, kennung)

	p.geloescht = append(p.geloescht, kennung)

	return nil
}

func (p *partieprobe) Kopieren(
	ctx context.Context, kennung, neueKennung string,
) (domain.Partie, error) {
	if fehler := p.stolpernAuf("kopieren"); fehler != nil {
		return domain.Partie{}, fehler
	}

	vorlage, fehler := p.Partie(ctx, kennung)
	if fehler != nil {
		return domain.Partie{}, fehler
	}

	return p.Anlegen(ctx, domain.Kopiere(vorlage, neueKennung))
}

func (p *partieprobe) Offenlegungszahl(
	_ context.Context, partieID string,
) (int, error) {
	if fehler := p.stolpernAuf("offenlegungszahl"); fehler != nil {
		return 0, fehler
	}

	return len(p.freigaben[partieID]), nil
}
