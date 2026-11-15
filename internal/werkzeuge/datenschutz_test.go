package werkzeuge_test

// **Die Datenschutzangaben stimmen mit dem überein, was die Anwendung tut** —
// Werkstück 13 von Phase 4.12
// (#1351).
//
// # Der Anlass
//
// **In der Hilfe stand „Sonst nichts. Keine Telemetrie, kein Konto, keine
// Anmeldung."** Der Satz war wahr, als er geschrieben wurde. Seit Phase 4.12
// gibt es Konten, eine Anmeldung über vier fremde Anbieter, einen zweiten
// Faktor und ein Formular mit Anrede, Namensteilen und Biografien — und der
// Satz stand unverändert da.
//
// **Eine Angabe über Datenschutz, die niemand nachzählt, veraltet still** —
// dieselbe Falle wie bei den Kontrasten, den Flächenzahlen und den
// Fähigkeiten, und dieselbe Antwort: ein Wächter, der sie gegen den Quelltext
// hält.
//
// # Was er hält
//
//  1. **Jedes personenbezogene Feld** aus `domain.Nutzer` wird in der Hilfe
//     benannt. Wer eines hinzufügt, muss sagen, dass es erhoben wird.
//  2. **Jede Adresse**, die diese Anwendung von sich aus anspricht, steht
//     dort. Wer einen Dienst aufnimmt, sagt es.
//
// # Was er nicht hält
//
// **Ob der Satz daneben wahr ist.** Er sieht, dass „Telefonnummer" vorkommt,
// nicht ob dort steht, wofür sie da ist. Und er sieht den **Katalog**, nicht
// das Fenster: dass die Seite ihn auch zeigt, hält
// `frontend/src/seiten/hilfe.test.tsx`.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// personenfelder sind die Felder von [domain.Nutzer], die eine Person
// beschreiben — mit dem Wort, unter dem sie in der Hilfe stehen.
//
// # Warum eine Tabelle und keine Reflexion über den Verbund
//
// **Weil ein Feldname kein deutsches Wort ist.** `Kurzebio` heißt in der Hilfe
// „kurze Biografie"; eine Ableitung aus dem Bezeichner gäbe „Kurzebio" und
// hielte damit nur sich selbst.
//
// **Und weil die Zuordnung die Entscheidung ist**: wer ein Feld hinzufügt,
// trägt es hier ein und sieht dabei, dass er über eine Person schreibt.
var personenfelder = map[string]string{
	"Name":         "Anzeigename",
	"Email":        "Adresse",
	"Zweitemail":   "Zweitadresse",
	"Telefon":      "Rufnummer",
	"Kennwort":     "Kennwort",
	"Anrede":       "Anrede",
	"Namenszusatz": "Namenszusatz",
	"Vorname":      "Vorname",
	"Weiterenamen": "weitere Vornamen",
	"Nachname":     "Nachname",
	"Slogan":       "Slogan",
	"Kurzebio":     "Biografie",
	"Langebio":     "Biografie",
}

// **Jedes Feld von `domain.Nutzer`, das eine Person beschreibt, steht in der
// Tabelle oben.**
//
// **Die Gegenprobe zur Tabelle** (§10): ohne sie bliebe ein neues Feld
// unbemerkt — die Tabelle wäre grün, weil sie sich selbst hält.
func TestJedesPersonenfeldStehtInDerTabelle(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile(filepath.Join("..", "domain", "nutzer.go"))
	if err != nil {
		t.Fatalf("nutzer.go lesen: %v", err)
	}

	// **Nur der Verbund `Nutzer`** — darunter stehen `Profil` und anderes.
	inhalt := string(roh)
	anfang := strings.Index(inhalt, "type Nutzer struct {")
	if anfang < 0 {
		t.Fatal("den Verbund Nutzer gibt es nicht mehr — das Muster greift daneben")
	}

	ende := strings.Index(inhalt[anfang:], "\n}")
	if ende < 0 {
		t.Fatal("das Ende des Verbunds ist nicht zu finden")
	}

	// **Die Felder, die keine Person beschreiben** — sie stehen hier
	// ausdrücklich und nicht als Regel über Namen: dass eine Kennung keine
	// Person beschreibt, ist eine **Entscheidung**.
	technisch := map[string]bool{
		"ID": true, "Fassung": true, "Anonym": true,
		"Bild": true, "Profile": true,
	}

	muster := regexp.MustCompile(`(?m)^\t([A-Z][A-Za-z]*)\s`)

	gesehen := 0

	for _, treffer := range muster.FindAllStringSubmatch(
		inhalt[anfang:anfang+ende], -1) {
		feld := treffer[1]
		if technisch[feld] {
			continue
		}

		gesehen++

		if _, steht := personenfelder[feld]; !steht {
			t.Errorf("das Feld %q beschreibt eine Person und steht in "+
				"`personenfelder` nicht — wird es in der Hilfe genannt?", feld)
		}
	}

	// **Und der Wächter selbst muss etwas gesehen haben.** Ein Muster, das
	// danebengreift, wäre grün und prüfte nichts.
	if gesehen < len(personenfelder)-2 {
		t.Errorf("nur %d Felder gesehen — das Muster greift daneben", gesehen)
	}
}

// **Und jedes Wort steht in der Hilfe — in beiden Sprachen.**
func TestJedesPersonenfeldStehtInDerHilfe(t *testing.T) {
	t.Parallel()

	for _, sprache := range []string{"de", "en"} {
		text := hilfetext(t, sprache)

		for feld, wort := range personenfelder {
			if sprache == "en" {
				continue
			}

			if !strings.Contains(text, wort) {
				t.Errorf("%s: %q (aus %s) steht nicht in der Hilfe",
					sprache, wort, feld)
			}
		}
	}
}

// **Und jede Adresse, die diese Anwendung anspricht, steht dort.**
//
// **Aus dem Quelltext gelesen und nicht abgeschrieben**: wer einen Dienst
// aufnimmt, ändert die Ausgabe dieses Wächters, und eine Konstante daneben
// liefe auseinander.
func TestJedeAngesprocheneAdresseStehtInDerHilfe(t *testing.T) {
	t.Parallel()

	text := hilfetext(t, "de")

	for _, wirt := range wirteAusDemQuelltext(t) {
		if !strings.Contains(text, wirt) {
			t.Errorf("%q wird angesprochen und steht nicht in der Hilfe", wirt)
		}
	}
}

// hilfetext gibt den ganzen Hilfekatalog einer Sprache als Text.
func hilfetext(t *testing.T, sprache string) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join( //nolint:gosec // fester Pfad im Repositorium
		"..", "..", "frontend", "src", "sprache", sprache, "hilfe.ts"))
	if err != nil {
		t.Fatalf("den Hilfekatalog %s lesen: %v", sprache, err)
	}

	return string(roh)
}

// wirteAusDemQuelltext sammelt die Adressen, die `internal/` anspricht.
//
// **Ohne Testdateien und ohne Beispieladressen**: `example.invalid` und
// `beispiel.test` sind reserviert und werden nie angesprochen (RFC 2606,
// RFC 6761).
//
// # Und ohne Kommentare
//
// **Eine zitierte Adresse ist keine angesprochene.** `docs.github.com` steht
// als Beleg neben einer Regel, `www.sqlite.org` als Verweis auf eine
// Fehlerliste, `media.wizards.com` als Herkunft eines Bildes, das jemand
// **von Hand** öffnet. Wer sie mitzählte, verlangte eine Datenschutzangabe
// über eine Fußnote — und der Wächter würde bei der nächsten Quellenangabe
// wieder rot, ohne dass sich etwas geändert hat.
//
// **Was gezählt wird, ist eine Adresse in einer Zeichenkette.** Das ist die
// Gestalt, in der ein Ruf entsteht.
func wirteAusDemQuelltext(t *testing.T) []string {
	t.Helper()

	erfunden := regexp.MustCompile(
		`(example|beispiel|invalid|test|angreifer|localhost|127\.0\.0\.1)`)
	muster := regexp.MustCompile(`https://([a-z0-9.-]+)`)

	gefunden := map[string]bool{}

	dateien, err := filepath.Glob(filepath.Join("..", "*", "*.go"))
	if err != nil {
		t.Fatalf("die Dateien finden: %v", err)
	}

	for _, pfad := range dateien {
		if strings.HasSuffix(pfad, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // aus Glob
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		for _, treffer := range muster.FindAllStringSubmatch(
			ohneKommentare(string(roh)), -1) {
			wirt := treffer[1]
			if erfunden.MatchString(wirt) || strings.Contains(wirt, "DavidTHeider") {
				continue
			}

			gefunden[wirt] = true
		}
	}

	// **Und der Wächter selbst muss etwas gefunden haben.**
	if len(gefunden) < 5 {
		t.Fatalf("nur %d Adressen gefunden — das Muster greift daneben",
			len(gefunden))
	}

	wirte := make([]string, 0, len(gefunden))
	for wirt := range gefunden {
		wirte = append(wirte, wirt)
	}

	return wirte
}

// ohneKommentare streicht Zeilenkommentare aus Go-Quelltext.
//
// **Zeilenweise und nicht über einen Zerteiler**: gesucht wird eine Adresse in
// einer Zeichenkette, und die steht nie hinter einem `//` derselben Zeile. Ein
// echter Zerteiler wäre genauer und hier ohne Gewinn.
//
// **Blockkommentare bleiben stehen** — dieses Repositorium schreibt keine.
func ohneKommentare(quelltext string) string {
	zeilen := strings.Split(quelltext, "\n")
	geputzt := make([]string, 0, len(zeilen))

	for _, zeile := range zeilen {
		geputzt = append(geputzt, zeile)

		gestutzt := strings.TrimSpace(zeile)
		if strings.HasPrefix(gestutzt, "//") {
			geputzt[len(geputzt)-1] = ""
		}
	}

	return strings.Join(geputzt, "\n")
}
