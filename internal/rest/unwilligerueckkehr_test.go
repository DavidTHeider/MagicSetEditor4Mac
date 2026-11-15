package rest_test

// **Der unwillige Rücksprung** — Werkstück O7 von Phase 5.2
// (#1431).
//
// Eine eigene Datei, wie `unwilligenutzer_test.go` daneben:
// `unwilligeausfuhr_test.go` ginge mit ihnen über die Zeilengrenze aus §2.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (u unwilliger) Ruecksprungarten(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Ruecksprunglage(context.Context, string, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Rueckkehren(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) RueckkehrQuittieren(context.Context, int64) error { return u.fehler }

func (u unwilliger) Pflegevorschau(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Pflegen(context.Context, string, []rest.Pflegewahl) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Kostenstand(context.Context, string, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Kostenposten(context.Context, string, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) KostenbudgetSetzen(
	context.Context, string, string, int64, string,
) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Ruecksetzungsarten(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Ruecksetzungsvorschau(context.Context, string, string, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Zuruecksetzen(
	context.Context, string, string, string, string, string, string,
) (any, error) {
	return nil, u.fehler
}
