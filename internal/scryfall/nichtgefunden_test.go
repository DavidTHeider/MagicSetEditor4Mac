package scryfall_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// abweisend gibt einen Dienst, der auf alles mit 404 antwortet.
func abweisend(t *testing.T) *scryfall.Klient {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)

			if _, err := schreiben.Write([]byte(`{"message":"Not Found"}`)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(dienst.Close)

	return klientOhneWarten(t, dienst.URL, netz.NeuerKlient(netz.UeblicheFristen()))
}

// **Ein 404 nennt, wonach gefragt wurde** —
// #1054.
//
// `holeMit` bediente jeden Weg mit demselben Satz: „Scryfall kennt diese Karte
// nicht". Es fragen aber auch der Satz, die Satzangaben, das Verzeichnis der
// Sammeldateien und die Symbologie — und wer den Satz las, suchte nach einer
// Karte, die niemand gesucht hat.
//
// Gemessen im Flächenlauf (#1028): `mse scryfall-set gibtsnicht magic` meldete
// „nicht gefunden: Scryfall kennt diese Karte nicht".
func TestEin404NenntDenGegenstand(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name  string
		ruf   func(*scryfall.Klient) error
		nennt string
	}{
		{
			"eine Karte nach Namen",
			func(k *scryfall.Klient) error {
				_, _, _, err := k.KarteNachName(t.Context(), "Gibt Es Nicht")

				return err
			},
			"Karte",
		},
		{
			"die Angaben eines Satzes",
			func(k *scryfall.Klient) error {
				_, _, _, err := k.Setangaben(t.Context(), "gibtsnicht")

				return err
			},
			"Satz",
		},
		{
			"die Regelhinweise einer Karte",
			func(k *scryfall.Klient) error {
				_, _, _, err := k.Rulings(t.Context(), "gibtsnicht")

				return err
			},
			"Regelhinweise",
		},
		{
			"das Verzeichnis der Sammeldateien",
			func(k *scryfall.Klient) error {
				_, _, err := k.Bulkverzeichnis(t.Context())

				return err
			},
			"Verzeichnis",
		},
		{
			"die Symbologie",
			func(k *scryfall.Klient) error {
				_, _, _, err := k.Symbologie(t.Context())

				return err
			},
			"Symbol",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			klient := abweisend(t)

			err := fall.ruf(klient)
			if err == nil {
				t.Fatal("ein 404 kam nicht als Fehler an")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.nennt, err)
			}

			// **Und niemand spricht mehr von „dieser Karte", wo keine gefragt
			// war.** Der Satz war der ganze Befund.
			if fall.nennt != "Karte" && strings.Contains(err.Error(), "diese Karte") {
				t.Errorf("die Meldung spricht von einer Karte: %v", err)
			}
		})
	}
}
