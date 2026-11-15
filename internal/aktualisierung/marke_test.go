// **Die Marke**
// (#1100,
// Werkstück 2).
//
// # Woher sie kommt, und warum nicht von woanders
//
// Aus der **Umgebung** — `GH_TOKEN`, sonst `GITHUB_TOKEN`. Das sind die
// beiden Namen, die `gh` selbst liest, und wer Lesezugriff auf ein privates
// Repositorium hat, hat `gh` eingerichtet.
//
// **Nicht aus den Einstellungen**: die liegen im Bestand, und der ist eine
// SQLite-Datei im Klartext. Eine Marke, die dort steht, steht in jeder
// Sicherung — auch in denen, die jemand weitergibt.
//
// **Nicht aus dem Schlüsselbund**: das verlangte ein Berechtigungsprofil und
// eine Eingabefläche für etwas, das ein Werkzeug daneben ohnehin schon hält.
// Der Preis steht dabei: eine Anwendung, die aus dem Finder startet, sieht die
// Umgebung der Anmeldesitzung und nicht die einer Shell — wer die Marke nur in
// `~/.zshrc` setzt, hat sie im Fenster nicht.
package aktualisierung_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Marke geht an jede der drei Anfragen.**
//
// Freigabe, Prüfsummen und Abbild liegen alle drei hinter demselben privaten
// Repositorium. Eine Marke, die nur an der ersten hängt, ergäbe eine
// Anwendung, die die neue Fassung **sieht** und nicht laden kann.
func TestDieMarkeGehtAnJedeAnfrage(t *testing.T) {
	inhalt := abbild()

	signiere := probeschluessel(t)
	summen := summe(inhalt) + "  ./MagicSetEditor4Mac-99.0.0.dmg\n"
	signatur := signiere(summen)

	var marken []string

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			marken = append(marken, anfrage.Header.Get("Authorization"))

			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/latest"):
				if _, fehler := schreiben.Write([]byte(`{"tag_name":"v99.0.0","assets":[]}`)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/pruefsummen"):
				if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			default:
				schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))
				if _, fehler := schreiben.Write(inhalt); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			}
		}))
	t.Cleanup(dienst.Close)

	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, func(time.Duration) {}), "ghp_geheim")

	if _, err := klient.NeuesteFreigabe(t.Context()); err != nil {
		t.Fatalf("NeuesteFreigabe: %v", err)
	}

	if _, err := klient.Laden(t.Context(), aktualisierung.Freigabe{
		Fassung:     "99.0.0",
		Abbild:      beigabe("MagicSetEditor4Mac-99.0.0.dmg", dienst.URL+"/abbild"),
		Pruefsummen: beigabe("pruefsummen.txt", dienst.URL+"/pruefsummen"),
		Signatur:    beigabe("pruefsummen.txt.asc", dienst.URL+"/signatur"),
	}, t.TempDir(), nil); err != nil {
		t.Fatalf("Laden: %v", err)
	}

	// **Vier**: Freigabe, Prüfsummen, Signatur, Abbild.
	if len(marken) != 4 {
		t.Fatalf("es gingen %d Anfragen hinaus", len(marken))
	}

	for _, marke := range marken {
		if marke != "Bearer ghp_geheim" {
			t.Errorf("eine Anfrage trug %q", marke)
		}
	}
}

// **Ohne Marke geht keine hinaus** — und die Anwendung fragt trotzdem.
//
// Ein öffentliches Repositorium antwortet auch unangemeldet; die Meldung
// „kein Zugriff" entsteht aus dem 404, nicht aus einer fehlenden Zeichenkette.
func TestOhneMarkeWirdTrotzdemGefragt(t *testing.T) {
	var marken []string

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			marken = append(marken, anfrage.Header.Get("Authorization"))

			if _, fehler := schreiben.Write([]byte(`{"tag_name":"v99.0.0","assets":[]}`)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(dienst.Close)

	if _, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context()); err != nil {
		t.Fatalf("NeuesteFreigabe: %v", err)
	}

	if len(marken) != 1 || marken[0] != "" {
		t.Errorf("die Anfragen trugen %+v", marken)
	}
}

// **`GH_TOKEN` schlägt `GITHUB_TOKEN`** — dieselbe Vorrangfolge wie bei `gh`.
func TestDieMarkeKommtAusDerUmgebung(t *testing.T) {
	for _, fall := range []struct {
		name, gh, github, will string
	}{
		{"beide", "aus-gh", "aus-github", "aus-gh"},
		{"nur GITHUB_TOKEN", "", "aus-github", "aus-github"},
		{"keine", "", "", ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Setenv("GH_TOKEN", fall.gh)
			t.Setenv("GITHUB_TOKEN", fall.github)

			if marke := aktualisierung.MarkeAusUmgebung(); marke != fall.will {
				t.Errorf("die Marke ist %q, erwartet war %q", marke, fall.will)
			}
		})
	}
}
