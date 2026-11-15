package anmeldung_test

// **Jeder Anbieter läuft gegen eine Nachbildung** — Nachlauf über ID-05,
// Befund 3 (#1704).
//
// # Wofür dieser Lauf steht
//
// **Er ist der Beleg hinter `domain.PruefstandSimuliert`.** Die Oberfläche
// sagt je Anbieter „gegen eine Nachbildung geprüft“ — und für Microsoft gab
// es bis dahin keinen durchlaufenen Weg, nur einen Test, der ein Geheimnis
// abweist. Hier geht **jeder** aus `domain.Anbieter()` den ganzen Weg:
// beginnen, tauschen, Identität lesen — mit den Einstellungen, mit denen er
// startet, und in der Gestalt, in der sein Anbieter antwortet. Ein neuer
// Anbieter ohne Nachbildung macht diesen Lauf rot, bevor die Oberfläche
// „simuliert“ über ihn sagt.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anmeldung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// nachbildung ist, wie ein Anbieter gebaut wird und was er antwortet.
type nachbildung struct {
	bau       func(kennung, geheimnis string) anmeldung.Anbieter
	geheimnis string

	// nutzer ist die Antwort des zweiten Rufs — leer bei OIDC.
	nutzer string
}

var nachbildungen = map[string]nachbildung{
	"github":    {bau: anmeldung.GitHub, geheimnis: "geheim", nutzer: `{"id":4711,"name":"Wer"}`},
	"google":    {bau: anmeldung.Google},
	"microsoft": {bau: anmeldung.Microsoft},
	"x":         {bau: anmeldung.X, nutzer: `{"data":{"id":"4711","name":"Wer"}}`},
}

func TestJederAnbieterLaeuftGegenEineNachbildung(t *testing.T) {
	t.Parallel()

	for _, name := range domain.Anbieter() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bild, gibt := nachbildungen[name]
			if !gibt {
				t.Fatalf("für %s gibt es keine Nachbildung — „simuliert“ wäre eine Behauptung", name)
			}

			if pruefstand := domain.AnbieterPruefstand(name); pruefstand != domain.PruefstandSimuliert {
				t.Fatalf("Prüfstand %q — dieser Lauf belegt nur die Nachbildung", pruefstand)
			}

			anbieter := bild.bau("kennung", bild.geheimnis)

			var jeton string

			gegenstelle := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, anfrage *http.Request) {
					switch {
					case strings.HasSuffix(anfrage.URL.Path, "/token") && anbieter.Gestalt == anmeldung.OIDC:
						antworte(t, schreiben, `{"access_token":"abc","id_token":"`+jeton+`"}`)
					case strings.HasSuffix(anfrage.URL.Path, "/token"):
						antworte(t, schreiben, `{"access_token":"abc"}`)
					default:
						antworte(t, schreiben, bild.nutzer)
					}
				}))
			t.Cleanup(gegenstelle.Close)

			anbieter.Tauschadresse = gegenstelle.URL + "/token"
			anbieter.Nutzeradresse = gegenstelle.URL + "/nutzer"

			lauf := mussBeginnen(t, anbieter)

			if anbieter.Gestalt == anmeldung.OIDC {
				jeton = jetonMit(t, map[string]any{
					"sub": "4711", "name": "Wer", "iss": anbieter.Aussteller, "aud": "kennung",
					"exp": time.Now().Add(time.Hour).Unix(), "nonce": nonceAus(t, lauf),
				})
			}

			kennung, fehler := lauf.Einloesen(t.Context(), anmeldung.RuferUeber(gegenstelle.Client()), "der-code")
			if fehler != nil {
				t.Fatalf("einlösen: %v", fehler)
			}

			if kennung.Subjekt != "4711" || kennung.Anbieter != name || kennung.Anzeigename != "Wer" {
				t.Errorf("Identität %+v", kennung)
			}
		})
	}
}

// **Startbereit sagt, ob ein Weg beginnen kann** — Nachlauf über ID-05,
// Befund 4. Dieselbe Prüfung wie beim Beginnen, damit „eingerichtet“ an der
// Oberfläche nicht mehr verspricht, als der Weg hält.
func TestStartbereitIstDiePruefungDesBeginnens(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anbieter anmeldung.Anbieter
		bereit   bool
	}{
		"GitHub ohne Geheimnis":    {anmeldung.GitHub("kennung", ""), false},
		"GitHub mit Geheimnis":     {anmeldung.GitHub("kennung", "geheim"), true},
		"Microsoft mit Geheimnis":  {anmeldung.Microsoft("kennung", "geheim"), false},
		"Microsoft ohne Geheimnis": {anmeldung.Microsoft("kennung", ""), true},
		"Google ohne Kennung":      {anmeldung.Google("", ""), false},
	} {
		fehler := fall.anbieter.Startbereit()
		if (fehler == nil) != fall.bereit {
			t.Errorf("%s: %v", name, fehler)
		}

		if _, beginn := anmeldung.Beginne(fall.anbieter, "http://127.0.0.1:1/r"); (beginn == nil) != fall.bereit {
			t.Errorf("%s: Beginne sagt %v, Startbereit %v", name, beginn, fehler)
		}
	}
}
