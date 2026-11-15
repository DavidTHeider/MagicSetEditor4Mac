// **Plan, Ersetzen und Zurück für den Dienst** — DOM-01.AK-02 und AK-03
// (#1696).
//
// Gerechnet wird in `service`, übersetzt in `app` ([app.Paketweg.Plan]);
// hier wird nur die Gestalt des Plans in die von `internal/rest` gebracht.

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// PaketPlan sagt vor einer Änderung, was sie bewirkt.
func (p pakete) PaketPlan(ctx context.Context, vorgang, art, name, quelle string) (rest.Paketplan, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Paketplan{}, err
	}

	plan, fehler := weg.Plan(ctx, vorgang, art, name, quelle)
	if fehler != nil {
		return rest.Paketplan{}, fehler
	}

	return rest.Paketplan{
		Vorgang: plan.Vorgang, Paket: plan.Paket, Quelle: plan.Quelle, Von: plan.Von, Zu: plan.Zu,
		Geaendert: plan.Geaendert, Moeglich: plan.Moeglich, Sets: plan.Sets,
		Konflikte: alsRestkonflikte(plan.Konflikte),
	}, nil
}

// PaketErsetzen legt eine neue Fassung über die vorhandene.
func (p pakete) PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) (rest.Paket, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Paket{}, err
	}

	paket, fehler := weg.Ersetzen(ctx, quelle, bestaetigt)
	if fehler != nil {
		return rest.Paket{}, fehler
	}

	return alsRestpaket(paket), nil
}

// PaketZurueck tauscht die aktuelle Fassung gegen die vorige.
func (p pakete) PaketZurueck(ctx context.Context, art, name string) (rest.Paket, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Paket{}, err
	}

	paket, fehler := weg.Zurueck(ctx, art, name)
	if fehler != nil {
		return rest.Paket{}, fehler
	}

	return alsRestpaket(paket), nil
}

// PaketBedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04.
func (p pakete) PaketBedarf(ctx context.Context, setID int64) ([]rest.Paketkonflikt, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	befunde, fehler := weg.Setbedarf(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	return alsRestkonflikte(befunde), nil
}

// alsRestkonflikte übersetzt die Befunde der Anwendung — **eine Stelle** für
// den Plan und den Bedarf.
func alsRestkonflikte(befunde []app.AbhaengigkeitDTO) []rest.Paketkonflikt {
	konflikte := make([]rest.Paketkonflikt, 0, len(befunde))
	for _, befund := range befunde {
		konflikte = append(konflikte, rest.Paketkonflikt{
			Paket: befund.Paket, Braucht: befund.Braucht, Art: befund.Art,
			Verlangt: befund.Verlangt, Da: befund.Da,
		})
	}

	return konflikte
}

// SetVorschau sagt, was das Einfahren einer Datei täte — FMT-03.AK-01
// (#1692), über denselben Bestand, in den `POST /sets` schriebe.
func (p pakete) SetVorschau(ctx context.Context, pfad string) (any, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	return weg.Setvorschau(ctx, p.db, pfad)
}
