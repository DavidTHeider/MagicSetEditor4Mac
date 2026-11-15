package rest_test

// **Die Probleme eines Sets über REST** — PERF02
// (#2352):
// die Antwort trägt eine Liste und nie `null`, und eine unlesbare Kennung ist
// ein Prüffehler am Feld.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func TestDieProblemeEinesSetsKommenAlsStand(t *testing.T) {
	for _, fall := range []struct {
		methode, urteil string
		karten          int
	}{
		{http.MethodGet, "ungeprueft", 0},
		{http.MethodPost, "aktuell", 1},
	} {
		antwort := ruf(t, neuerDienst(t), endpunkt{methode: fall.methode, weg: "/sets/1/probleme"}, gueltig)
		if antwort.Code != http.StatusOK {
			t.Fatalf("%s: Status %d — %s", fall.methode, antwort.Code, antwort.Body.String())
		}

		var stand rest.Problemstand
		if err := json.Unmarshal(antwort.Body.Bytes(), &stand); err != nil {
			t.Fatal(err)
		}

		if stand.Urteil != fall.urteil || stand.Karten == nil || len(stand.Karten) != fall.karten {
			t.Errorf("%s: %+v", fall.methode, stand)
		}
	}
}

func TestEineUnlesbareSetkennungIstEinPruefFehler(t *testing.T) {
	for _, methode := range []string{http.MethodGet, http.MethodPost} {
		antwort := ruf(t, neuerDienst(t), endpunkt{methode: methode, weg: "/sets/keins/probleme"}, gueltig)
		if antwort.Code != http.StatusBadRequest || !strings.Contains(antwort.Body.String(), "setID") {
			t.Errorf("%s: %d %s", methode, antwort.Code, antwort.Body.String())
		}
	}
}
