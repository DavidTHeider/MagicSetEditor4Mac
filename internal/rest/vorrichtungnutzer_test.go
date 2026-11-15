// **Die willigen Antworten zu Nutzern, Konten und Faktoren** — Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, seit `vorrichtungbestand_test.go` mit dem zweiten
// Faktor aus Werkstück 7 über die Zeilengrenze aus §2 lief. Der Schnitt folgt
// der Frage: dort antwortet, was am **Bestand** hängt, hier, wer ihn benutzt.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) NutzerLoeschen(context.Context, string, string) error { return nil }
func (williger) ProfilLoeschen(context.Context, string, string) error { return nil }

// **Die vier Kontoarten** — Werkstück 6 (#1351).
func (williger) KontoEintragen(_ context.Context, kennung, _, name string) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{ID: "1", Name: name, Email: kennung, Profile: []rest.Profil{}}, nil
}

func (williger) KontoAnmelden(_ context.Context, kennung, _, _ string) (rest.Nutzer, error) {
	return rest.Nutzer{ID: "1", Name: "Ich", Email: kennung, Profile: []rest.Profil{}}, nil
}

// **Der zweite Faktor** — Werkstück 7 (#1351).
func (williger) Faktoren(context.Context, string) (rest.Faktorstand, error) {
	return rest.Faktorstand{Faktoren: []rest.Faktor{{Art: "totp"}}, Biometrie: true}, nil
}

func (williger) FaktorEinrichten(context.Context, string) (rest.Faktoreinrichtung, error) {
	return rest.Faktoreinrichtung{
		Adresse: "otpauth://totp/x", Geheimnis: "GEZDGNBVGY3TQOJQ", Codes: []string{"a"},
	}, nil
}

func (williger) FaktorBestaetigen(context.Context, string, string) error { return nil }

func (williger) BiometrieEinrichten(context.Context, string) error { return nil }

func (williger) FaktorEntfernen(context.Context, string, string) error { return nil }

func (williger) KontoAnonym(context.Context) (rest.Nutzer, error) {
	return rest.Nutzer{
		ID: "1", Name: "anonym-abc", Anonym: true, Profile: []rest.Profil{},
	}, nil
}
