// **Die vier Kontoarten über die Fassade** — Werkstück 6 von Phase 4.12
// (#1351).
package app_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein eingetragenes Konto kommt als Nutzer samt Profil an.**
func TestEinEingetragenesKontoKommtAlsNutzerAn(t *testing.T) {
	anwendung := kontoanwendung(&kontofuehrer{
		nutzer: domain.Nutzer{ID: 1, Name: "wer", Email: "wer@example.org"},
	})

	antwort := anwendung.KontoEintragen(t.Context(),
		"wer@example.org", "einlangesgeheimnis", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Nutzer == nil || antwort.Nutzer.ID != "1" {
		t.Fatalf("%+v", antwort.Nutzer)
	}

	if len(antwort.Nutzer.Profile) != 1 {
		t.Errorf("ohne Profil: %+v", antwort.Nutzer)
	}
}

// **Das Kennwort erreicht die Antwort nie — auch nicht als Feld, das leer
// bleibt.**
//
// # Warum das über JSON geprüft wird und nicht über das DTO
//
// **Weil zwischen beiden die Marken liegen** (§3): ein Feld ohne
// `json:"-"` steht in der Antwort, gleich wie es im Go-Typ heißt. Ein Test
// gegen `antwort.Nutzer.Kennwort` prüfte die Struktur und nicht die Zeile,
// die über die Naht geht.
func TestDasKennwortErreichtDieAntwortNie(t *testing.T) {
	anwendung := kontoanwendung(&kontofuehrer{nutzer: domain.Nutzer{
		ID: 1, Name: "wer", Email: "wer@example.org",
		Kennwort: "$argon2id$v=19$m=19456,t=2,p=1$c2FsemNoZW4$aGFzaA",
	}})

	antwort := anwendung.KontoAnmelden(t.Context(),
		"wer@example.org", "einlangesgeheimnis", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, fehler := json.Marshal(antwort)
	if fehler != nil {
		t.Fatalf("marshal: %v", fehler)
	}

	for _, verboten := range []string{"argon2id", "kennwort", "Kennwort"} {
		if strings.Contains(string(roh), verboten) {
			t.Errorf("%q steht in der Antwort: %s", verboten, roh)
		}
	}
}

// **Die Antwort sagt, welcher Art dieses Konto ist.**
//
// **Sonst wäre die Seite auf Raten angewiesen**: ein Nutzer ohne Adresse ist
// entweder anonym, oder der erste, oder einer über einen Anbieter — und die
// drei bekommen verschiedene Sätze zu lesen (Werkstück 13).
func TestDieAntwortSagtWelcherArtDiesesKontoIst(t *testing.T) {
	faelle := map[string]struct {
		nutzer   domain.Nutzer
		erwartet app.NutzerDTO
	}{
		"Adresse": {
			nutzer:   domain.Nutzer{ID: 1, Name: "wer", Email: "wer@example.org"},
			erwartet: app.NutzerDTO{Email: "wer@example.org"},
		},
		"Rufnummer": {
			nutzer:   domain.Nutzer{ID: 1, Name: "wer", Telefon: "+4915112345678"},
			erwartet: app.NutzerDTO{Telefon: "+4915112345678"},
		},
		"anonym": {
			nutzer:   domain.Nutzer{ID: 1, Name: "anonym-abc", Anonym: true},
			erwartet: app.NutzerDTO{Anonym: true},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			anwendung := kontoanwendung(&kontofuehrer{nutzer: fall.nutzer})

			antwort := anwendung.KontoAnonym(t.Context())
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			if antwort.Nutzer.Email != fall.erwartet.Email ||
				antwort.Nutzer.Telefon != fall.erwartet.Telefon ||
				antwort.Nutzer.Anonym != fall.erwartet.Anonym {
				t.Errorf("angekommen ist %+v", antwort.Nutzer)
			}
		})
	}
}

// **Ein Fehler der Anwendung kommt als Umschlag an** (§8).
func TestEinKontofehlerKommtAlsUmschlagAn(t *testing.T) {
	wege := map[string]func(*app.App) app.EinNutzerAntwort{
		"eintragen": func(a *app.App) app.EinNutzerAntwort {
			return a.KontoEintragen(t.Context(), "wer@example.org", "kurz", "")
		},
		"anmelden": func(a *app.App) app.EinNutzerAntwort {
			return a.KontoAnmelden(t.Context(), "wer@example.org", "falsch", "")
		},
		"anonym": func(a *app.App) app.EinNutzerAntwort {
			return a.KontoAnonym(t.Context())
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			anwendung := kontoanwendung(&kontofuehrer{
				fehler: domain.NewFieldError("kennung",
					domain.SchluesselValidierung, "krumm"),
			})

			antwort := weg(anwendung)
			if antwort.Fehler == nil {
				t.Fatal("der Fehler ging verloren")
			}

			if antwort.Fehler.Feld != "kennung" {
				t.Errorf("%+v", antwort.Fehler)
			}
		})
	}
}

// kontoanwendung baut eine Anwendung über einem Kontoführer.
func kontoanwendung(fuehrer *kontofuehrer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitKontoarten(fuehrer)
	anwendung.MitNutzern(&nutzerfuehrer{zeilen: []service.Nutzerzeile{{
		Nutzer:  fuehrer.nutzer,
		Profile: []domain.Profil{{ID: 10, NutzerID: 1, Name: "Mein Profil"}},
	}}})

	return anwendung
}

// kontofuehrer ist die Naht zum Dienst der Kontoarten.
type kontofuehrer struct {
	nutzer domain.Nutzer
	fehler error
}

func (k *kontofuehrer) Eintragen(_ context.Context, _ service.Kontowunsch) (
	domain.Nutzer, error,
) {
	return k.nutzer, k.fehler
}

func (k *kontofuehrer) Anmelden(_ context.Context, _, _, _ string) (domain.Nutzer, error) {
	return k.nutzer, k.fehler
}

func (k *kontofuehrer) Anonym(_ context.Context) (domain.Nutzer, error) {
	return k.nutzer, k.fehler
}
