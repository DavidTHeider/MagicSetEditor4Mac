package netz_test

// **Wohin ein Klient wählen darf** — SEC-P4
// (#2000),
// Befund P4 in `docs/PENTEST.md`.
//
// # Woran gemessen
//
// **An einer Gegenstelle, die mitzählt** — nicht an der Fehlermeldung. Eine
// Sperre, die erst nach der Anfrage griffe, gäbe dieselbe Meldung und hätte
// trotzdem gefragt. Und jede Sperre hat ihren grünen Zwilling: dieselbe
// Gegenstelle ist über einen erlaubten Weg erreichbar, sonst hieße die Null
// nur, dass niemand zuhört (§3).
//
// **Kein `t.Parallel`**: der benannte Schalter steht in der Umgebung, und die
// gehört dem ganzen Prozess.
//
// **Und nur Ziele, die diesen Rechner nicht verlassen** (`docs/PENTEST.md`
// §1): die Schleife in ihren Schreibweisen und `0.0.0.0`, das auf macOS
// ebenfalls hier landet. Ein Lauf gegen `169.254.169.254` oder `10.0.0.1`
// schickte im roten Zustand ein Paket ins Netz — die Bereiche selbst prüft
// deshalb die Tafel in `ziel_intern_test.go`, ohne zu wählen. Einzige
// Ausnahme ist der Proxyfall: dort geht jede Anfrage an den eigenen Lauscher.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// zaehlendeGegenstelle antwortet `200` und zählt jede Anfrage.
func zaehlendeGegenstelle(t *testing.T) (*httptest.Server, *atomic.Int64) {
	t.Helper()

	var gefragt atomic.Int64

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt.Add(1)
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(gegenstelle.Close)

	return gegenstelle, &gefragt
}

// portVon gibt den Port einer Adresse.
func portVon(t *testing.T, adresse string) string {
	t.Helper()

	zerlegt, fehler := url.Parse(adresse)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return zerlegt.Port()
}

// holeZiel holt eine Adresse und gibt nur den Fehler.
func holeZiel(t *testing.T, klient *netz.Klient, adresse string) error {
	t.Helper()

	_, fehler := netz.Hole(t.Context(), klient, nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, adresse, 1<<10)

	return fehler
}

func TestEineAdresseVonInnenWirdNichtGewaehlt(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "")

	gegenstelle, gefragt := zaehlendeGegenstelle(t)
	port := portVon(t, gegenstelle.URL)
	klient := netz.NeuerKlientNach(netz.UeblicheFristen(), nil, netz.ZielOeffentlich)

	for _, adresse := range []string{
		gegenstelle.URL,
		"http://localhost:" + port,
		"http://0.0.0.0:" + port,
		"http://[::ffff:127.0.0.1]:" + port,
		"http://[::1]:" + port,
		"ftp://example.org/",
	} {
		if fehler := holeZiel(t, klient, adresse); !errors.Is(fehler, netz.ErrZielGesperrt) {
			t.Errorf("%s: %v statt einer Sperre", adresse, fehler)
		}
	}

	if n := gefragt.Load(); n != 0 {
		t.Errorf("die Gegenstelle auf der Schleife wurde %d-mal gefragt", n)
	}

	// **Der benannte Schalter öffnet die Schleife — und nur sie.**
	t.Setenv(netz.SchleifeUmgebung, "an")

	if fehler := holeZiel(t, klient, gegenstelle.URL); fehler != nil {
		t.Errorf("mit dem Schalter: %v", fehler)
	}

	if gefragt.Load() != 1 {
		t.Errorf("mit dem Schalter kam %d-mal etwas an, erwartet einmal", gefragt.Load())
	}

	for _, adresse := range []string{"http://0.0.0.0:" + port, "http://[::ffff:0.0.0.0]:" + port} {
		if fehler := holeZiel(t, klient, adresse); !errors.Is(fehler, netz.ErrZielGesperrt) {
			t.Errorf("mit dem Schalter %s: %v statt einer Sperre", adresse, fehler)
		}
	}
}

func TestDieGrundsperreGiltJedemKlienten(t *testing.T) {
	gegenstelle, gefragt := zaehlendeGegenstelle(t)
	port := portVon(t, gegenstelle.URL)
	klient := netz.NeuerKlient(netz.UeblicheFristen())

	for _, adresse := range []string{
		"http://0.0.0.0:" + port,
		"http://[::]:" + port,
		"http://[::ffff:0.0.0.0]:" + port,
	} {
		if fehler := holeZiel(t, klient, adresse); !errors.Is(fehler, netz.ErrZielGesperrt) {
			t.Errorf("%s: %v statt einer Sperre", adresse, fehler)
		}
	}

	if n := gefragt.Load(); n != 0 {
		t.Errorf("die Gegenstelle wurde %d-mal über die Grundsperre hinweg gefragt", n)
	}

	// **Die Schleife bleibt jedem Dienst mit fester Adresse offen** — so
	// erreichen ihn die Läufe über `MSE_*_ADRESSE`.
	if fehler := holeZiel(t, klient, gegenstelle.URL); fehler != nil || gefragt.Load() != 1 {
		t.Errorf("über die Schleife: %v, %d Anfragen", fehler, gefragt.Load())
	}
}

func TestEineWeiterleitungFuehrtNichtHinterDieSperre(t *testing.T) {
	hinter, gefragtHinter := zaehlendeGegenstelle(t)
	hinterPort := portVon(t, hinter.URL)

	for name, ziel := range map[string]string{
		"unbestimmt":     "http://0.0.0.0:" + hinterPort + "/",
		"abgebildet":     "http://[::ffff:0.0.0.0]:" + hinterPort + "/",
		"fremdes Schema": "file:///etc/passwd",
	} {
		t.Run(name, func(t *testing.T) {
			var gefragtVorn atomic.Int64

			vorn := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, anfrage *http.Request) {
					gefragtVorn.Add(1)
					http.Redirect(schreiben, anfrage, ziel, http.StatusFound)
				}))
			t.Cleanup(vorn.Close)

			fehler := holeZiel(t, netz.NeuerKlient(netz.UeblicheFristen()), vorn.URL)
			if !errors.Is(fehler, netz.ErrZielGesperrt) {
				t.Errorf("die Weiterleitung nach %s: %v statt einer Sperre", ziel, fehler)
			}

			if gefragtVorn.Load() != 1 {
				t.Errorf("vorn %d Anfragen, erwartet eine", gefragtVorn.Load())
			}
		})
	}

	if n := gefragtHinter.Load(); n != 0 {
		t.Errorf("hinter der Weiterleitung kamen %d Anfragen an", n)
	}
}

func TestWeiterleitungenEndenNachZehn(t *testing.T) {
	var gefragt atomic.Int64

	kreis := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gefragt.Add(1)
			http.Redirect(schreiben, anfrage, "/noch-einmal", http.StatusFound)
		}))
	t.Cleanup(kreis.Close)

	fehler := holeZiel(t, netz.NeuerKlient(netz.UeblicheFristen()), kreis.URL)
	if fehler == nil || !strings.Contains(fehler.Error(), "10") {
		t.Errorf("ein Kreis endete mit %v", fehler)
	}

	// **Zehn Anfragen, wie unter Gos Vorgabe** — die zehnte Weiterleitung
	// wird nicht mehr gestellt. Ein eigenes `CheckRedirect` ersetzt die
	// Vorgabe ganz; ohne die Grenze darin liefe der Kreis bis zur Frist.
	if n := gefragt.Load(); n != 10 {
		t.Errorf("%d Anfragen im Kreis, erwartet zehn", n)
	}
}

func TestEinProxyDarfImEigenenNetzStehen(t *testing.T) {
	proxy, gefragt := zaehlendeGegenstelle(t)

	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv(netz.SchleifeUmgebung, "")

	klient := netz.NeuerKlientNach(netz.UeblicheFristen(), nil, netz.ZielOeffentlich)

	// **Der Proxy steht auf der Schleife, und er wird erreicht** — wie ein
	// Firmenproxy im eigenen Netz. Die Sperre gilt dem Ziel, nicht der
	// Leitung dorthin.
	if fehler := holeZiel(t, klient, "http://ziel.invalid/etwas"); fehler != nil {
		t.Fatalf("über den Proxy: %v", fehler)
	}

	if gefragt.Load() != 1 {
		t.Fatalf("der Proxy sah %d Anfragen, erwartet eine", gefragt.Load())
	}

	// **Und ein gesperrtes Ziel geht auch nicht über ihn**: was drüben
	// aufgelöst wird, sieht dieser Rechner nicht — eine Adresse in Ziffern
	// aber schon.
	for _, adresse := range []string{"http://10.1.2.3/", "http://169.254.169.254/"} {
		if fehler := holeZiel(t, klient, adresse); !errors.Is(fehler, netz.ErrZielGesperrt) {
			t.Errorf("%s über den Proxy: %v statt einer Sperre", adresse, fehler)
		}
	}

	if gefragt.Load() != 1 {
		t.Errorf("der Proxy sah %d Anfragen, erwartet eine", gefragt.Load())
	}
}
