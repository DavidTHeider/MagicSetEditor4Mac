// **Die Fehlerwege von Stufe 2**
// (#1100).
//
// **Eine eigene Datei**, seit `laden_test.go` über die Zeilengrenze aus §2
// ging — und der Schnitt folgt der Frage: dort steht, was ein geglückter
// Ladevorgang tut, hier, was jeder Weg daneben hinterlässt. In beiden Fällen
// ist die Antwort dieselbe und der Grund ein anderer: **nichts**.
package aktualisierung_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// launischeGegenstelle gibt einen Dienst, der einen Weg mit einem Status abtut.
func launischeGegenstelle(t *testing.T, weg string, status int) *httptest.Server {
	t.Helper()

	signiere := probeschluessel(t)
	summen := summe(abbild()) + "  ./MagicSetEditor4Mac-0.2.0.dmg\n"
	signatur := signiere(summen)

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if strings.HasSuffix(anfrage.URL.Path, weg) {
				schreiben.WriteHeader(status)

				return
			}

			// **Die Signatur muss stehen, auch wenn dieser Test sie nicht
			// meint.** Sonst scheitert der Weg an ihr, und der Zweig, den der
			// Test wirklich prüft, bleibt unbetreten — grün, ohne etwas zu
			// belegen (§10).
			if strings.HasSuffix(anfrage.URL.Path, "/signatur") {
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			}

			if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst
}

// **Eine Prüfsummenliste, die mit einem Fehlschlag kommt, wird nicht gedeutet.**
//
// Eine 503-Seite ist keine Prüfsummenliste; sie enthält bloß keine Zeile für
// das Abbild und liefe sonst als „nicht genannt" durch — mit einer Meldung, die
// in die Irre führt.
func TestPruefsummenMitFehlschlagWerdenGemeldet(t *testing.T) {
	gegenstelle := launischeGegenstelle(t, "/pruefsummen", http.StatusServiceUnavailable)

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, t.TempDir(), nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Und ein Abbild, das mit einem Fehlschlag kommt, ebenso.**
func TestEinAbbildMitFehlschlagWirdGemeldet(t *testing.T) {
	gegenstelle := launischeGegenstelle(t, "/abbild", http.StatusInternalServerError)

	ziel := t.TempDir()

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, ziel, nil)

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", err)
	}

	eintraege, lesefehler := os.ReadDir(ziel)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(eintraege) != 0 {
		t.Errorf("es blieb etwas liegen: %+v", eintraege)
	}
}

// **Ein Verzeichnis, das es nicht gibt, wird gemeldet, statt zu laden.**
func TestEinUnbrauchbaresZielverzeichnisWirdGemeldet(t *testing.T) {
	inhalt := abbild()
	gegenstelle := gegenstelleMit(t, inhalt, summe(inhalt)+"  ./MagicSetEditor4Mac-0.2.0.dmg\n")

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, filepath.Join(t.TempDir(), "gibtsnicht"), nil)

	if err == nil {
		t.Error("ein Ziel, das es nicht gibt, kam durch")
	}
}

// **Ein Zielname, an dem schon ein Verzeichnis steht, wird gemeldet.**
//
// Der Weg ist kurz vor dem Ziel zu Ende — die Datei ist geladen und geprüft,
// und das Umbenennen scheitert. Auch dann bleibt kein Nachbar liegen.
func TestEinBelegterZielnameWirdGemeldet(t *testing.T) {
	inhalt := abbild()
	gegenstelle := gegenstelleMit(t, inhalt, summe(inhalt)+"  ./MagicSetEditor4Mac-0.2.0.dmg\n")

	ziel := t.TempDir()
	if err := os.Mkdir(filepath.Join(ziel, "MagicSetEditor4Mac-0.2.0.dmg"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	_, err := probeklient(t, gegenstelle.URL).Laden(t.Context(),
		aktualisierung.Freigabe{
			Fassung:     "0.2.0",
			Abbild:      beigabe("MagicSetEditor4Mac-0.2.0.dmg", gegenstelle.URL+"/abbild"),
			Pruefsummen: beigabe("pruefsummen.txt", gegenstelle.URL+"/pruefsummen"),
			Signatur:    beigabe("pruefsummen.txt.asc", gegenstelle.URL+"/signatur"),
		}, ziel, nil)

	if err == nil {
		t.Fatal("ein belegter Zielname kam durch")
	}

	eintraege, lesefehler := os.ReadDir(ziel)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(eintraege) != 1 {
		t.Errorf("es blieb etwas liegen: %+v", eintraege)
	}
}
