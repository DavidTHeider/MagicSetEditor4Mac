// **Zwei Klienten teilen die Drossel** —
// zweiter Teil von Befund 2 des Zweitblicks über Phase V, 2026-08-24.
//
// # Der Befund
//
// Das Fenster baute **vier** `scryfall.Klient`, jeden mit eigener Drossel;
// MCP und REST bauten ihre Naht **je Aufruf** neu. Weil der erste Aufruf einer
// Drossel absichtlich nicht wartet, sah eine solche Drossel praktisch nie
// einen zweiten — die Grenze war wirkungslos.
package scryfall_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

func TestZweiKlientenAusEinemWerkTeilenDieDrossel(t *testing.T) {
	t.Parallel()

	var zeiten []time.Time

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			zeiten = append(zeiten, time.Now())

			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(
				`{"object":"card","id":"a","name":"Prüfstein"}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	// **Ein Werk, zwei Klienten** — wie das Fenster sie baut.
	werk := netz.NeuesDrosselwerk(time.Now, nil)

	baue := func() *scryfall.Klient {
		return scryfall.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
			netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
			time.Now, "prüfung").MitDrosselwerk(werk)
	}

	erster, zweiter := baue(), baue()

	if _, _, _, fehler := erster.KarteNachName(t.Context(), "Eins"); fehler != nil {
		t.Fatalf("erster: %v", fehler)
	}

	if _, _, _, fehler := zweiter.KarteNachName(t.Context(), "Zwei"); fehler != nil {
		t.Fatalf("zweiter: %v", fehler)
	}

	if len(zeiten) != 2 {
		t.Fatalf("%d Anfragen", len(zeiten))
	}

	// **Der zweite Klient wartet**, obwohl er selbst noch nie gefragt hat.
	if abstand := zeiten[1].Sub(zeiten[0]); abstand < 500*time.Millisecond {
		t.Errorf("Abstand = %v, erwartet mindestens 500ms — die beiden Klienten "+
			"kannten einander nicht", abstand)
	}
}

// **Ohne Werk baut ein Klient seine eigenen Drosseln.**
//
// Ein Test oder ein Werkzeug, das nur einen Klienten braucht, soll kein Werk
// bauen müssen. Und die Grenze wird trotzdem eingehalten — nur eben nicht
// über mehrere Klienten hinweg.
func TestOhneWerkHatEinKlientEigeneDrosseln(t *testing.T) {
	t.Parallel()

	var zeiten []time.Time

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			zeiten = append(zeiten, time.Now())
			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(
				`{"object":"card","id":"a","name":"Prüfstein"}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	klient := scryfall.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "prüfung")

	for _, name := range []string{"Eins", "Zwei"} {
		if _, _, _, fehler := klient.KarteNachName(t.Context(), name); fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}
	}

	if len(zeiten) != 2 {
		t.Fatalf("%d Anfragen", len(zeiten))
	}

	if abstand := zeiten[1].Sub(zeiten[0]); abstand < 500*time.Millisecond {
		t.Errorf("Abstand = %v — auch ohne Werk gilt die Grenze", abstand)
	}
}

// **Ein leeres Werk lässt den Klienten, wie er ist.**
//
// # Warum das ein Fall ist und keine Vorsicht
//
// [scryfall.Klient.MitDrosselwerk] hängt sich an den Bau an. Ein Aufrufer, der
// das Werk aus einer Naht bekommt, kann `nil` in der Hand haben — und dann soll
// der Klient seine eigenen Drosseln behalten, nicht keine.
func TestEinLeeresWerkLaesstDenKlientenWieErIst(t *testing.T) {
	t.Parallel()

	var zeiten []time.Time

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			zeiten = append(zeiten, time.Now())
			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(
				`{"object":"card","id":"a","name":"Prüfstein"}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	klient := scryfall.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "prüfung").MitDrosselwerk(nil)

	for _, name := range []string{"Eins", "Zwei"} {
		if _, _, _, fehler := klient.KarteNachName(t.Context(), name); fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}
	}

	if len(zeiten) != 2 {
		t.Fatalf("%d Anfragen", len(zeiten))
	}

	// **Die eigene Drossel greift weiter** — sie wurde nicht ersetzt.
	if abstand := zeiten[1].Sub(zeiten[0]); abstand < 500*time.Millisecond {
		t.Errorf("Abstand = %v — das leere Werk hat die Drosseln genommen", abstand)
	}
}
