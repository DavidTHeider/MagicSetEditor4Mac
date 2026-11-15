package rest_test

// **Die Doppel des Durchreichens** — ausgelagert aus `durchreichen_test.go`,
// weil die Datei mit ihnen über die Zeilengrenze aus §2 ginge. Der Schnitt
// folgt der Frage: dort, **was durchgereicht werden muss** — hier, **wer
// beim Prüfen nicht mitspielt**.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// leererabgleich gibt einen Vorschlag ohne Änderungen und ohne Befunde.
type leererabgleich struct{ williger }

func (leererabgleich) Vorschlag(
	context.Context, int64, int64, string,
) (rest.Vorschlag, error) {
	return rest.Vorschlag{}, nil
}

// mitbefunden gibt einen Vorschlag mit einem Befund.
type mitbefunden struct{ williger }

func (mitbefunden) Vorschlag(
	context.Context, int64, int64, string,
) (rest.Vorschlag, error) {
	return rest.Vorschlag{Befunde: []string{"unbekannter Wert"}}, nil
}

// **Und auch der Volltext und der Lebenszyklus scheitern**
// (#1107).
func (u unwilliger) Felder(context.Context, int64) ([]string, error) {
	return nil, u.fehler
}

func (u unwilliger) Suche(context.Context, domain.SearchQuery) (domain.SearchResult, error) {
	return domain.SearchResult{}, u.fehler
}

func (u unwilliger) KarteAnlegen(context.Context, int64) (int64, error) {
	return 0, u.fehler
}

func (u unwilliger) KarteKopieren(context.Context, int64, int64) (int64, error) {
	return 0, u.fehler
}

func (u unwilliger) KarteLoeschen(context.Context, int64, int64, bool) error {
	return u.fehler
}

func (u unwilliger) KartenTauschen(context.Context, int64, int64, int64) error { return u.fehler }

// **Und auch die Setangaben scheitern** (#1107).
func (u unwilliger) Angaben(context.Context, int64) (
	[]rest.Feldbeschreibung, []rest.Setwert, error,
) {
	return nil, nil, u.fehler
}

func (u unwilliger) Setzen(context.Context, int64, string, string) (
	vorher string, hatte bool, fehler error,
) {
	return "", false, u.fehler
}

func (u unwilliger) Vorbehalte(context.Context, int64) ([]rest.Vorbehalt, error) {
	return nil, u.fehler
}

func (u unwilliger) Beispielsets() ([]rest.Beispielset, error) {
	return nil, u.fehler
}
