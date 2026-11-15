package rest_test

// **Ein Sitz lässt sich an ein Profil binden** — UXR-F21
// (#2041).
//
// **Bis dahin wies der Dienst `profilID` als unbekanntes Feld ab** — ehrlich,
// aber ohne Weg zur Bindung. Hier steht, dass die Angabe die Naht erreicht und
// dass ihr Fehlen den Normalfall bleibt; dass dieselbe Partie an allen Flächen
// denselben eigenen Sitz hat, hält `internal/flaechen/sitzprofil_test.go`.

import (
	"net/http"
	"testing"
)

// **Die Profilkennung kommt bis zur Naht — je Sitz.**
func TestDerDienstReichtDasProfilDesSitzesDurch(t *testing.T) {
	t.Parallel()

	gezaehlt := &zaehlender{}

	antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt{
		methode: http.MethodPost, weg: "/partien",
		rumpf: `{"spiel":"magic","sitze":[` +
			`{"kennung":"a","name":"A","deckID":"1"},` +
			`{"kennung":"b","name":"B","deckID":"2","profilID":"3"}]}`,
	}, gueltig)

	if antwort.Code != http.StatusCreated {
		t.Fatalf("endete mit %d: %s", antwort.Code, antwort.Body.String())
	}

	if len(gezaehlt.sitze) != 2 ||
		gezaehlt.sitze[0].ProfilID != "" || gezaehlt.sitze[1].ProfilID != "3" {
		t.Errorf("angekommen: %+v", gezaehlt.sitze)
	}
}

// **Und die Abwahl der Runde kommt ebenso an** — P1a
// (#2335).
// Ob es den Bereich gibt, entscheidet der Kern; hier steht nur, dass er ihn
// zu sehen bekommt — und ohne Angabe nichts.
func TestDerDienstReichtDieAbwahlDurch(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		rumpf    string
		erwartet int
	}{
		"mit Abwahl":  {`"abgewaehlt":["zugstruktur"],`, 1},
		"ohne Abwahl": {``, 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gezaehlt := &zaehlender{}

			antwort := ruf(t, dienstMit(t, gezaehlt), endpunkt{
				methode: http.MethodPost, weg: "/partien",
				rumpf: `{"spiel":"magic",` + fall.rumpf + `"sitze":[` +
					`{"kennung":"a","name":"A","deckID":"1"},` +
					`{"kennung":"b","name":"B","deckID":"2"}]}`,
			}, gueltig)

			if antwort.Code != http.StatusCreated {
				t.Fatalf("endete mit %d: %s", antwort.Code, antwort.Body.String())
			}

			if len(gezaehlt.abgewaehlt) != fall.erwartet {
				t.Errorf("angekommen: %v", gezaehlt.abgewaehlt)
			}
		})
	}
}
