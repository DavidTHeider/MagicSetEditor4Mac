package rest_test

// **Das Profilbild am willigen Doppel** — Werkstück 8 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `vorrichtungbestand_test.go` sonst über die
// Zeilengrenze aus §2 ginge.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// NutzerbildSetzen gibt den Nutzer mit dem Namen, den die Ablage vergäbe.
//
// **Der Name ist der, den `service.Bildablage` bildet** — ein Doppel, das
// einen anderen zurückgäbe, wäre großzügiger als der Betrieb (§3).
func (williger) NutzerbildSetzen(_ context.Context, nutzerID, _, _ string) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{
		ID: nutzerID, Name: "Ich", Bild: "nutzer-" + nutzerID + ".png",
		Fassung: "2", Profile: []rest.Profil{},
	}, nil
}

func (williger) NutzerbildEntfernen(_ context.Context, nutzerID, _ string) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{
		ID: nutzerID, Name: "Ich", Bild: "", Fassung: "2",
		Profile: []rest.Profil{},
	}, nil
}

// **Das Abzeichen eines Profils** — Werkstück 9 (#1351), Punkt 13.
//
// **Der Name ist der, den `service.Bildablage` bildet**: `profil-<id>.png`
// neben `nutzer-<id>.png`. Ein Doppel, das ihn anders bildete, wäre
// großzügiger als der Betrieb (§3).
func (williger) ProfilbadgeSetzen(_ context.Context, profilID, _, _ string) (
	rest.Profil, error,
) {
	return rest.Profil{
		ID: profilID, NutzerID: "1", Name: "Mein Profil",
		Badge: "profil-" + profilID + ".png", Fassung: "2",
	}, nil
}

func (williger) ProfilbadgeEntfernen(_ context.Context, profilID, _ string) (
	rest.Profil, error,
) {
	return rest.Profil{
		ID: profilID, NutzerID: "1", Name: "Mein Profil", Badge: "", Fassung: "2",
	}, nil
}
