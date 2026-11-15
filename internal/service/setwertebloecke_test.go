// **Die Angaben des Sets auf dem Schreibweg** —
// #890.
//
// **Eine eigene Datei**, weil `setschreiber_test.go` über die Zeilengrenze
// wuchs (§2). Der Schnitt folgt der Frage: dort **die Datei als Ganzes** —
// Bytezusage, Karten, Fehlerwege —, hier die **Angaben des Sets**.
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// geschriebeneQuelle schreibt ein Set und gibt seine Hauptdatei.
func geschriebeneQuelle(t *testing.T, quelle quellenDouble) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(quelle, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return string(gelesen)
}

// **Ein geänderter Setwert kommt in der Datei an** —
// #890.
//
// # Der Befund
//
// Der Schreiber glich **Karten** und **Schlüsselwörter** ab — die Angaben des
// Sets selbst nicht. `SetwertSetzen` schrieb in den Bestand, und beim Sichern
// blieb in der Datei stehen, was vorher dort stand: Titel, Urheberrecht,
// Setkürzel, **und das Symbol**.
//
// **Gefunden hat es die Abnahme von Phase X**, nicht ein Test: ein gezeichnetes
// Symbol wurde abgelegt, als `set.symbol` eingetragen, das Set gesichert — und
// die geschriebene Datei trug `symbol:` **leer**. Nachgemessen an
// `copyright`: im Bestand `PROBE-MARKE`, in der Datei unverändert
// `™ & © Wizards of the Coast`.
//
// **Wie lange es dastand, ist nicht gemessen** — nur, dass es dastand.
func TestEinGeaenderterSetwertKommtInDerDateiAn(t *testing.T) {
	t.Parallel()

	quelle := mitQuelle()
	quelle.werte = []domain.Value{{Field: "title", Text: "Ein anderer Titel"}}

	geschrieben := geschriebeneQuelle(t, quelle)

	if !strings.Contains(geschrieben, "\ttitle: Ein anderer Titel\n") {
		t.Errorf("der Titel steht nicht in\n%s", geschrieben)
	}

	if strings.Contains(geschrieben, "Mein Probeset") {
		t.Errorf("der alte Titel steht noch in\n%s", geschrieben)
	}
}

// **Und ein Setwert, den es in der Datei noch nicht gibt, bekommt seine Zeile.**
//
// Ein Symbol ist genau dieser Fall: `set.symbol` steht in vielen Sets gar
// nicht, und ein gezeichnetes muss trotzdem hinein.
func TestEinNeuerSetwertBekommtSeineZeile(t *testing.T) {
	t.Parallel()

	quelle := mitQuelle()
	quelle.werte = []domain.Value{
		{Field: "title", Text: "Mein Probeset"},
		{Field: "symbol", Text: "symbol1.mse-symbol"},
	}

	geschrieben := geschriebeneQuelle(t, quelle)

	if !strings.Contains(geschrieben, "\tsymbol: symbol1.mse-symbol\n") {
		t.Errorf("das Symbol steht nicht in\n%s", geschrieben)
	}
}

// **Und ein Set ohne `set info:`-Block bekommt keinen erfunden.**
//
// Ein Block, den niemand angeordnet hat, wäre eine Änderung an einer fremden
// Datei — und die Bytezusage gilt für alles, was niemand angefasst hat (0002).
func TestOhneSetinfoblockWirdKeinerErfunden(t *testing.T) {
	t.Parallel()

	quelle := mitQuelle()
	quelle.quelle = &domain.Quelle{
		Datei:  "set",
		Inhalt: []byte("mse version: 2.0.0\ngame: probe\ncard:\n\tname: Feuerball\n"),
	}
	quelle.werte = []domain.Value{{Field: "title", Text: "Ein Titel"}}

	geschrieben := geschriebeneQuelle(t, quelle)

	if strings.Contains(geschrieben, "set info") {
		t.Errorf("ein `set info:` ist entstanden:\n%s", geschrieben)
	}
}

// **Ein Kommentar im `set info:`-Block bleibt, was er ist.**
//
// Er ist keine Angabe, und ihn als eine zu behandeln hieße, ihn mit einem
// Wert zu überschreiben, den niemand ihm gegeben hat.
func TestEinKommentarImSetinfoBlockBleibt(t *testing.T) {
	t.Parallel()

	quelle := mitQuelle()
	quelle.quelle = &domain.Quelle{
		Datei: "set",
		Inhalt: []byte("mse version: 2.0.0\ngame: probe\nset info:\n" +
			"\t# ein Hinweis für Menschen\n\ttitle: Mein Probeset\n" +
			"card:\n\tname: Feuerball\n"),
	}
	quelle.werte = []domain.Value{{Field: "title", Text: "Ein anderer Titel"}}

	geschrieben := geschriebeneQuelle(t, quelle)

	if !strings.Contains(geschrieben, "\t# ein Hinweis für Menschen\n") {
		t.Errorf("der Kommentar fehlt in\n%s", geschrieben)
	}

	if !strings.Contains(geschrieben, "\ttitle: Ein anderer Titel\n") {
		t.Errorf("der Titel steht nicht in\n%s", geschrieben)
	}
}

// **Ein mehrzeiliger Setwert wird ersetzt und nicht verdoppelt.**
//
//	} else if (value.empty()) {
//	  // a multiline string
//	  readLine(true);
//
// `src/util/io/reader.cpp`, Zeilen 258–260, Fassung f9d9356 vom 2021-07-06:
// hinter dem Doppelpunkt steht nichts, der Wert liegt eine Ebene tiefer.
//
// **Am Kartenblock hat genau das einmal jede Karte getroffen** (#704): `notes`
// verdoppelte sich bei jedem Speichern, weil der Vergleich gegen das leere
// `kind.Wert` ging und die alten Kinder stehen blieben. Derselbe Zweig steht
// hier, und er wird hier geprüft.
func TestEinMehrzeiligerSetwertWirdErsetzt(t *testing.T) {
	t.Parallel()

	quelle := mitQuelle()
	quelle.quelle = &domain.Quelle{
		Datei: "set",
		Inhalt: []byte("mse version: 2.0.0\ngame: probe\nset info:\n" +
			"\tdescription:\n\t\tErste Zeile\n\t\tZweite Zeile\n" +
			"card:\n\tname: Feuerball\n"),
	}
	quelle.werte = []domain.Value{{Field: "description", Text: "Nur noch eine"}}

	geschrieben := geschriebeneQuelle(t, quelle)

	if strings.Contains(geschrieben, "Erste Zeile") {
		t.Errorf("der alte Wert steht noch in\n%s", geschrieben)
	}

	if !strings.Contains(geschrieben, "\tdescription: Nur noch eine\n") {
		t.Errorf("der neue Wert steht nicht in\n%s", geschrieben)
	}
}
