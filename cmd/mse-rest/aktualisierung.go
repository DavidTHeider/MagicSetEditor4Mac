// **Die Aktualisierung am Dienst** —
// #1100,
// Werkstück 7.
//
// Der Composition Root reicht dieselbe Naht durch, die das Fenster benutzt.
// **Und ohne Rücksicht auf die Einstellung**, aus demselben Grund wie bei
// `mse aktualisierung`: sie verhindert ungefragtes Verhalten, und ein Aufruf
// ist eine Frage.

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// fassungen bedient die beiden Endpunkte der Aktualisierung.
type fassungen struct{}

// NeuesteFassung sagt, was freigegeben ist.
func (fassungen) NeuesteFassung(ctx context.Context) (rest.Fassungsstand, error) {
	auskunft, err := app.NeuesteFassung(ctx)
	if err != nil {
		return rest.Fassungsstand{}, err
	}

	return rest.Fassungsstand{
		Laufend:     auskunft.Laufend,
		Freigegeben: auskunft.Freigegeben,
		Seite:       auskunft.Seite,
		Neuer:       auskunft.Neuer,
	}, nil
}

// FassungLaden holt das Abbild und gibt seinen Ort.
//
// **Ohne Fortschrittsmeldung**, und das ist eine Entscheidung: eine
// HTTP-Antwort kommt am Ende. Wer den Fortschritt sehen will, nimmt die
// Kommandozeile oder das Fenster — beides sagt es unterwegs.
func (fassungen) FassungLaden(ctx context.Context) (string, error) {
	return app.FassungLaden(ctx, nil)
}
