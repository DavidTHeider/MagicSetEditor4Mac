package aktualisierung_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Antwort, die GitHub am 2026-08-25 wirklich schickt** — gekürzt auf
// die Felder, die dieser Klient liest, und um zwei, die er nicht kennt.
//
// **Eingecheckt und selbst geschrieben** (§16): geprüft wird gegen ein
// Beispiel im Baum, nicht gegen den Dienst. Ein Lauf gegen GitHub steht hinter
// `netzlauf`.
const echteAntwort = `{
  "tag_name": "v0.9.0",
  "name": "Fassung 0.9.0",
  "html_url": "https://github.com/DavidTHeider/MagicSetEditor4Mac/releases/tag/v0.9.0",
  "published_at": "2026-08-22T10:00:00Z",
  "draft": false,
  "prerelease": false,
  "assets": [
    {"name": "MagicSetEditor4Mac-0.9.0.dmg", "size": 187762389,
     "url": "https://api.github.com/repos/x/y/releases/assets/1"},
    {"name": "pruefsummen.txt", "size": 412,
     "url": "https://api.github.com/repos/x/y/releases/assets/2"}
  ],
  "ein_feld_das_wir_nicht_kennen": 42
}`

// gegenstelle gibt einen Dienst, der diese Antwort schickt — und sagt, was er
// gefragt wurde.
func gegenstelle(t *testing.T, rumpf string, status int) (*httptest.Server, *http.Header) {
	t.Helper()

	var gesehen http.Header

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gesehen = anfrage.Header.Clone()

			schreiben.Header().Set("Content-Type", "application/json")
			schreiben.WriteHeader(status)

			if _, err := schreiben.Write([]byte(rumpf)); err != nil {
				t.Errorf("schreiben: %v", err)
			}
		}))
	t.Cleanup(dienst.Close)

	return dienst, &gesehen
}

// probeklient baut einen Klienten gegen eine Gegenstelle.
func probeklient(t *testing.T, adresse string) *aktualisierung.Klient {
	t.Helper()

	// **Ohne Marke**, und das ist der Normalfall dieser Tests: die Gegenstelle
	// ist ein `httptest`-Dienst, und eine Marke prüft `marke_test.go` eigens.
	return aktualisierung.NeuerKlient(adresse,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "")
}

// **Die neueste Freigabe wird gelesen, und ihre Gestalt kommt an.**
func TestDieNeuesteFreigabeWirdGelesen(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t, echteAntwort, http.StatusOK)

	freigabe, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err != nil {
		t.Fatalf("holen: %v", err)
	}

	if freigabe.Fassung != "0.9.0" {
		t.Errorf("Fassung = %q", freigabe.Fassung)
	}

	if !strings.HasSuffix(freigabe.Abbild.Name, ".dmg") {
		t.Errorf("Abbild = %q", freigabe.Abbild.Name)
	}

	if freigabe.Pruefsummen.Name != "pruefsummen.txt" {
		t.Errorf("Prüfsummen = %q", freigabe.Pruefsummen.Name)
	}

	if freigabe.Seite == "" {
		t.Error("keine Seite — die Meldung soll dorthin führen können")
	}
}

// **Der `User-Agent` ist Pflicht, und er benennt diese Anwendung.**
//
// Nachgelesen am 2026-08-25: „Requests with no `User-Agent` header will be
// rejected." Ein Klient, der ihn vergisst, bekommt gar keine Antwort — und
// merkt es erst im Betrieb.
func TestDerKlientNenntSich(t *testing.T) {
	t.Parallel()

	dienst, gesehen := gegenstelle(t, echteAntwort, http.StatusOK)

	if _, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context()); err != nil {
		t.Fatalf("holen: %v", err)
	}

	if kenner := gesehen.Get("User-Agent"); !strings.Contains(kenner, "MagicSetEditor4Mac") {
		t.Errorf("User-Agent = %q", kenner)
	}

	if annahme := gesehen.Get("Accept"); !strings.Contains(annahme, "github") {
		t.Errorf("Accept = %q", annahme)
	}
}

// **Ein fehlendes Pflichtfeld bricht laut** (§16, „Drift wird erkannt").
func TestEineFreigabeOhneMarkeBrichtLaut(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t, `{"name":"ohne Marke","assets":[]}`, http.StatusOK)

	if _, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context()); err == nil {
		t.Error("eine Freigabe ohne tag_name wurde angenommen")
	}
}

// **Ein unbekanntes Zusatzfeld bricht nicht.** Dieselbe Regel, andere
// Richtung: Pflichtfelder streng, Zusatzfelder tolerant. `echteAntwort` trägt
// eines, und [TestDieNeuesteFreigabeWirdGelesen] geht durch.

// **Eine Freigabe ohne Abbild ist keine, aus der man laden kann.**
func TestEineFreigabeOhneAbbildSagtEs(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t,
		`{"tag_name":"v9.9.9","assets":[{"name":"nur.txt","size":1,"url":"x"}]}`,
		http.StatusOK)

	freigabe, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err != nil {
		t.Fatalf("holen: %v", err)
	}

	if freigabe.Abbild.Name != "" {
		t.Errorf("es wurde ein Abbild gefunden: %q", freigabe.Abbild.Name)
	}

	// **Die Fassung steht trotzdem da.** Melden geht ohne Abbild; laden nicht.
	if freigabe.Fassung != "9.9.9" {
		t.Errorf("Fassung = %q", freigabe.Fassung)
	}
}

// **Ein fachliches Nein ist kein Ausfall** (§16, Satz 4): 404 heißt „keine
// Freigabe" und wird nicht wiederholt.
func TestOhneFreigabeKommtEinFachlichesNein(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t, `{"message":"Not Found"}`, http.StatusNotFound)

	_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err == nil {
		t.Fatal("404 wurde als Freigabe gelesen")
	}

	if !strings.Contains(err.Error(), "keine Freigabe") {
		t.Errorf("die Meldung sagt nicht, was los ist: %v", err)
	}
}

// **Ohne Gegenstelle sagt der Klient, dass er nicht durchkam** — und nicht,
// dass es keine neue Fassung gebe. Das ist der Unterschied zwischen „nichts
// Neues" und „nicht gefragt" (§8).
func TestOhneGegenstelleKommtEinFehler(t *testing.T) {
	t.Parallel()

	// **Eine Adresse, an der nichts lauscht.** Der Dienst wird angelegt und
	// sofort geschlossen; sein Port ist damit sicher frei und gehört uns.
	dienst, _ := gegenstelle(t, "", http.StatusOK)
	adresse := dienst.URL

	dienst.Close()

	if _, err := probeklient(t, adresse).NeuesteFreigabe(t.Context()); err == nil {
		t.Error("ohne Gegenstelle kam keine Meldung")
	}
}

// **Ein Fehler auf der anderen Seite ist keiner dieser Anwendung**, und die
// Meldung nennt die Zahl — sonst sucht jemand bei sich.
func TestEinServerfehlerNenntSeineZahl(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t, `{"message":"Server Error"}`,
		http.StatusInternalServerError)

	_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())
	if err == nil {
		t.Fatal("ein 500 wurde als Freigabe gelesen")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("die Meldung nennt die Zahl nicht: %v", err)
	}
}

// **Was kein JSON ist, wird abgewiesen und nicht geraten.**
func TestEineUnlesbareAntwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst, _ := gegenstelle(t, "kein JSON", http.StatusOK)

	if _, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context()); err == nil {
		t.Error("eine unlesbare Antwort wurde angenommen")
	}
}
