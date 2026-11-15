package main

// **Die Naht zur Einschätzung** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// **In einem Zug**, weil diese Fläche die Anwendung je Aufruf baut — und für
// eine Einschätzung ist das der ganze Weg: sie wird nicht übernommen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// KIEinschaetzung lässt ein Modell eine Karte beurteilen.
//
// **Die Antwort reist ungefiltert weiter** — die Gestalt steht an
// [app.KIEinschaetzungAntwort], und ein zweiter Verbund daneben liefe beim
// nächsten Feld auseinander.
func (p *preise) KIEinschaetzung(
	ctx context.Context, bezug rest.KIEinschaetzungsbezug,
) (any, error) {
	antwort := p.kianwendung().KIEinschaetzen(ctx,
		app.KIEinschaetzungsbezug{
			Art: bezug.Art, KarteID: bezug.KarteID, Anbieter: bezug.Anbieter,
			Alias: bezug.Alias, Frage: bezug.Frage, Format: bezug.Format,
			Freigegeben: bezug.Freigegeben,
		})
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
