package service

// **Ein Tischbild aus einem gelesenen Stand** — #2481 (Zweitblick über
// Phase 7 von 0.9.0, Befund 8).
//
// # Was offen stand
//
// **Sicht, Erlaubnis und Handlungen kamen aus drei Lesevorgängen.** Schrieb
// dazwischen ein anderer Zug, trug ein Bild die Sicht der einen Fassung und
// die Erlaubnis einer anderen — etwa eine laufende Partie, an der nichts
// mehr geht. Hier stand es schon für zwei der drei („aus demselben Stand wie
// die Erlaubnis"), nur las jede der drei Methoden für sich.

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Tischbild rechnet, was `fuer` sieht, stellen darf und an jedem Objekt tun
// kann — aus **einem** gelesenen Stand.
//
// **Über die Projektion und nicht über den Zustand** — der Grund steht bei
// [domain.Partie.Sicht]: sie ist die Regel, nicht ihre Anwendung.
func (l *Partieleitung) Tischbild(
	ctx context.Context, kennung, fuer string,
) (domain.Tischbild, error) {
	// **Ein an einen Sitz gebundenes Token sieht nur ihn** (#1819).
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, fuer); fehler != nil {
		return domain.Tischbild{}, fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return domain.Tischbild{}, fehler
	}

	// **Für den Bestand, der liest** (P9, #2377) — am Gastplatz der des
	// Gastes.
	bild := domain.Tischbild{
		Sicht: partie.SichtVon(fuer, BestandssichtAus(ctx)),
		// **Leere Listen und kein `nil`** — „nichts geht gerade" ist eine
		// Antwort, und sie sieht anders aus als keine.
		Erlaubt: []string{}, Handlungen: []domain.Objekthandlungen{},
	}

	if partie.Stand != domain.StandLaeuft {
		return bild, nil
	}

	erlaubt := l.ausleger.Erlaubt(partie, fuer)
	bild.Handlungen = Handlungsauskunft(partie, fuer, erlaubt)

	// **Die Rücknahme darf jeder stellen, der überhaupt etwas darf** — P8
	// (#2376). Ob sie wirkt, entscheidet der Kern am Strom; im Beginn gibt
	// es nichts, das auf ihrer Liste stünde.
	if len(erlaubt) > 0 && !beginnHaelt(partie) {
		erlaubt = append(slices.Clone(erlaubt), Leitungsbefehle()...)
	}

	if erlaubt != nil {
		bild.Erlaubt = erlaubt
	}

	return bild, nil
}
