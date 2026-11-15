// **Das Doppel für den Einstellungsaustausch** — Werkstück O3 von Phase 5.2
// (#1427).
//
// **Ein eigener Typ und keine Methode an `williger`**, weil `Uebernehmen`
// dort schon die Massenänderung bedient — dieselbe Naht kann nicht zwei
// Gestalten desselben Namens tragen. Der Fehler wird durchgereicht wie bei
// `unwilliger`: ohne Fehler antwortet das Doppel knapp und mit der Gestalt,
// die der Dienst weitergibt.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

type austauschprobe struct{ fehler error }

func (a austauschprobe) Ausfuhr(context.Context, []string) (string, error) {
	if a.fehler != nil {
		return "", a.fehler
	}

	return `{"fassung":1,"bereiche":{}}`, nil
}

func (a austauschprobe) Vorschau(context.Context, []byte) (rest.Einstellungssicht, error) {
	if a.fehler != nil {
		return rest.Einstellungssicht{}, a.fehler
	}

	return rest.Einstellungssicht{
		Fassung: 1, Revision: "abc123",
		Zeilen: []rest.Einstellungsdiffzeile{{
			Ebene: "profil", Schluessel: "sprache",
			Vorher: "de", Nachher: "en", Herkunft: "profil",
		}},
		Befunde: []rest.Einstellungsbefund{},
	}, nil
}

func (a austauschprobe) Uebernehmen(
	context.Context, []byte, string, []string,
) (rest.Einstellungsbericht, error) {
	if a.fehler != nil {
		return rest.Einstellungsbericht{}, a.fehler
	}

	return rest.Einstellungsbericht{
		Uebernommen: 1, Revision: "def456",
		Zeilen:  []rest.Einstellungsdiffzeile{},
		Befunde: []rest.Einstellungsbefund{},
	}, nil
}

// austauschzaehler zählt die Rufe an der Austauschnaht in den gemeinsamen
// Zähler — damit „gültig" auch hier den Weg bis zum Anwendungsfall belegt.
type austauschzaehler struct{ z *zaehlender }

func (a austauschzaehler) Ausfuhr(ctx context.Context, abschnitte []string) (string, error) {
	a.z.rufe++

	return austauschprobe{}.Ausfuhr(ctx, abschnitte)
}

func (a austauschzaehler) Vorschau(ctx context.Context, roh []byte) (rest.Einstellungssicht, error) {
	a.z.rufe++

	return austauschprobe{}.Vorschau(ctx, roh)
}

func (a austauschzaehler) Uebernehmen(
	ctx context.Context, roh []byte, revision string, nur []string,
) (rest.Einstellungsbericht, error) {
	a.z.rufe++

	return austauschprobe{}.Uebernehmen(ctx, roh, revision, nur)
}
