package werkzeuge_test

// Die kleinste Ablage, an der sich ein Token bauen lässt — für
// `tokengestalt_test.go`.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func tokenprobe() domain.Zugriffstoken {
	return domain.Zugriffstoken{
		ProfilID:    1,
		Bezeichnung: "Wächterlauf",
		Scope:       domain.ScopeLesen,
	}
}

// tokenablage nimmt genau eines auf.
type tokenablage struct{}

func (*tokenablage) TokenAnlegen(
	_ context.Context, neu domain.Zugriffstoken, _ string,
) (domain.Zugriffstoken, error) {
	neu.ID = 1

	return neu, nil
}

func (*tokenablage) Token(context.Context, int64) ([]domain.Zugriffstoken, error) {
	return nil, nil
}

func (*tokenablage) TokenMitKennung(context.Context, int64) (
	domain.Zugriffstoken, domain.Gehuetetes, error,
) {
	return domain.Zugriffstoken{}, domain.Gehuetetes{}, nil
}

// **Werkstück O2 von Phase 5.2** (#1426): Rotation, Nutzung, Sammelwiderruf.
func (*tokenablage) TokenRotieren(_ context.Context, tokenID int64, _, vorgaengerBis string) (
	domain.Zugriffstoken, error,
) {
	return domain.Zugriffstoken{ID: tokenID, VorgaengerBis: vorgaengerBis}, nil
}

func (*tokenablage) TokenNutzungVermerken(context.Context, int64, string) error { return nil }

func (*tokenablage) TokenAlleLoeschen(context.Context, int64) (int64, error) { return 0, nil }

func (*tokenablage) TokenAendern(_ context.Context, geaendert domain.Zugriffstoken) (
	domain.Zugriffstoken, error,
) {
	return geaendert, nil
}

func (*tokenablage) TokenLoeschen(context.Context, int64) error { return nil }
