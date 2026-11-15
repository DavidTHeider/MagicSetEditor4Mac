package app_test

// **Je Anmeldeanbieter: eingerichtet, und wogegen geprüft** — ID-05.AK-01
// (#1704).
//
// **Eingerichtet heißt: der Weg kann beginnen** (Nachlauf über ID-05, Befund
// 4). Vorher genügte eine `client_id`, und GitHub ohne Geheimnis hieß
// eingerichtet — und scheiterte beim ersten Klick.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eingerichtete gibt je Anbieter, ob er eingerichtet heißt.
func eingerichtete(t *testing.T, anwendung *app.App) map[string]bool {
	t.Helper()

	antwort := anwendung.Anbieter(t.Context())
	if antwort.Fehler != nil || len(antwort.Staende) != 4 {
		t.Fatalf("%+v", antwort)
	}

	stand := map[string]bool{}

	for _, eintrag := range antwort.Staende {
		if eintrag.Geprueft != string(domain.PruefstandSimuliert) {
			t.Errorf("%s: geprüft %q", eintrag.Anbieter, eintrag.Geprueft)
		}

		stand[eintrag.Anbieter] = eintrag.Eingerichtet
	}

	return stand
}

func TestJederAnbieterSagtObErEingerichtetUndWieErGeprueftIst(t *testing.T) {
	anwendung := faktoranwendung(t)

	for _, name := range []string{"github", "microsoft"} {
		if antwort := anwendung.EinstellungSetzen(t.Context(), domain.AnbieterKennung(name), "meine-kennung"); antwort.Fehler != nil {
			t.Fatal(antwort.Fehler)
		}
	}

	// **GitHub braucht noch sein Geheimnis**, Microsoft keines.
	if stand := eingerichtete(t, anwendung); stand["github"] || !stand["microsoft"] || stand["google"] || stand["x"] {
		t.Errorf("nur mit Kennung: %v", stand)
	}

	if antwort := anwendung.AnbietergeheimnisSetzen(t.Context(), "github", "geheim"); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if stand := eingerichtete(t, anwendung); !stand["github"] || !stand["microsoft"] {
		t.Errorf("mit Geheimnis: %v", stand)
	}
}
