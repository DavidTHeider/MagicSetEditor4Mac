// **Die unwilligen Antworten der Partien** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `unwilligeausfuhr_test.go` herausgelöst**, als sie über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort scheitert, was
// die Platte verlässt oder auf sie kommt, hier, was an einem Tisch geschieht.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Und die Partien** (G1, #1458).
func (u unwilliger) PartieAnlegen(
	context.Context, string, string, []rest.Sitzwunsch, []string,
) (rest.Partieangaben, error) {
	return rest.Partieangaben{}, u.fehler
}

// Partien klagt wie alles andere — die Liste ist ein eigener Weg (#1600).
func (u unwilliger) Partien(context.Context) (
	[]rest.Partiezeilenangaben, error,
) {
	return nil, u.fehler
}

func (u unwilliger) Partie(context.Context, string, string) (
	rest.Partieangaben, error,
) {
	return rest.Partieangaben{}, u.fehler
}

func (u unwilliger) PartieBefehl(context.Context, rest.Spielbefehl) (
	rest.Befehlsangaben, error,
) {
	return rest.Befehlsangaben{}, u.fehler
}

func (u unwilliger) PartieStand(context.Context, string, string, string) (
	rest.Partieangaben, error,
) {
	return rest.Partieangaben{}, u.fehler
}

func (u unwilliger) PartieVerlauf(context.Context, string, string) (
	[]rest.Partieereignis, error,
) {
	return nil, u.fehler
}
