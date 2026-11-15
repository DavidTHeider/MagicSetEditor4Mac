// **Ein unvollendetes Zurückspielen bleibt sichtbar** — der zweite Zweitblick
// über die Freigabe 0.3.0, Befund 2
// (#1170).
//
// # Der Befund, in zwei Hälften
//
// **Die erste**: scheitert der nachgeholte Bildertausch beim nächsten Start,
// wird er nur **protokolliert**. `App.Starten` öffnet danach trotzdem die
// bereits ausgetauschte Datenbank — neue Datenbank, alte Bilder, und im
// Fenster steht nichts davon. Ein Protokolleintrag ist keine Meldung an
// jemanden, der die Anwendung benutzt.
//
// **Die zweite**: fehlt der Auszug, galt der Tausch **ohne Gegenbeweis** als
// erledigt. Der Auszug ist ein Nachbarverzeichnis `.rueckspiel-*`; dass er
// weg ist, kann heißen „umbenannt, also getauscht" — oder „von jemandem
// weggeräumt, also nie getauscht". Beides sah gleich aus, und das Programm
// entschied sich stumm für das Angenehmere.
//
// # Der Gegenbeweis
//
// Der **Elternordner** des Auszugs. `beiwerkErsetzen` benennt das `beiwerk/`
// darin um und lässt den Ordner selbst stehen; wegräumen tut ihn erst der
// Aufrufer. Steht er noch und das `beiwerk/` fehlt, **ist getauscht worden**.
// Sind beide weg, weiß niemand etwas — und dann sagt die Anwendung das.
package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// abgebrochenesRueckspielen stellt den Zustand nach einem Kill her: die
// Datenbank ist getauscht, der Auszug liegt da, die Bilder sind die alten.
func abgebrochenesRueckspielen(t *testing.T) (frisch *Bestandsweg, beiwerk, bilder string) {
	t.Helper()

	weg, ort, bilderort := mitBilderort(t)

	if err := os.MkdirAll(filepath.Join(bilderort, "7"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilderort, "7", "kunst.png"),
		[]byte("der alte Stand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

	ausgepackt, fehler := weg.auspacken(t.Context(), ziel, weg.orte.Datenbank)
	if fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	if fehler := weg.markeSchreiben(rueckspielstand{
		Quelle: ziel, Beiwerk: ausgepackt.Beiwerk, Bilder: BilderErsetzt,
	}); fehler != nil {
		t.Fatalf("Marke schreiben: %v", fehler)
	}

	frisch = wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilderort, nil }

	return frisch, ausgepackt.Beiwerk, bilderort
}

// **Ein gescheitertes Nachholen meldet sich.**
//
// Erzwungen über ein Bilderverzeichnis, das sich nicht finden lässt — derselbe
// Weg, den `bildertauschVollenden` schon kennt.
func TestEinGescheitertesNachholenMeldetSich(t *testing.T) {
	t.Parallel()

	frisch, _, _ := abgebrochenesRueckspielen(t)
	frisch.orte.Bilder = func() (string, error) { return "", os.ErrPermission }

	rest := frisch.RueckspielVollenden()
	if rest == "" {
		t.Fatal("kein Rest gemeldet — das Fenster öffnet die neue Datenbank " +
			"mit den alten Bildern und sagt nichts davon")
	}

	// **Und die Marke bleibt liegen.** Der nächste Start soll es wieder
	// versuchen; sie wegzuräumen hieße, den halben Stand für endgültig zu
	// erklären.
	if _, err := os.Stat(frisch.markenpfad()); err != nil {
		t.Errorf("die Marke ist weg: %v", err)
	}
}

// **Ein geglücktes Nachholen meldet nichts.**
func TestEinGegluecktesNachholenMeldetNichts(t *testing.T) {
	t.Parallel()

	frisch, _, _ := abgebrochenesRueckspielen(t)

	if rest := frisch.RueckspielVollenden(); rest != "" {
		t.Errorf("Rest %q, erwartet keinen", rest)
	}
}

// **Ein weggeräumter Auszug ohne Elternordner ist kein Beleg für „getauscht".**
//
// Das ist die zweite Hälfte des Befunds: hier galt der Tausch als erledigt,
// obwohl niemand ihn gesehen hat.
func TestEinSpurlosVerschwundenerAuszugGiltNichtAlsGetauscht(t *testing.T) {
	t.Parallel()

	frisch, beiwerk, bilder := abgebrochenesRueckspielen(t)

	// **Der ganze Zwischenplatz weg** — Auszug und Elternordner.
	if err := os.RemoveAll(filepath.Dir(beiwerk)); err != nil {
		t.Fatalf("räumen: %v", err)
	}

	rest := frisch.RueckspielVollenden()
	if rest == "" {
		t.Fatal("still als erledigt verbucht — niemand hat den Tausch gesehen")
	}

	if !strings.Contains(rest, "Bilder") {
		t.Errorf("die Meldung %q sagt nicht, worum es geht", rest)
	}

	// **Die alten Bilder stehen unverändert da** — und genau das ist der
	// Zustand, über den jemand Bescheid wissen muss.
	roh, err := os.ReadFile(filepath.Join(bilder, "7", "kunst.png")) // #nosec G304 -- t.TempDir()
	if err != nil || string(roh) != "der alte Stand" {
		t.Errorf("Bild %q (%v)", string(roh), err)
	}
}

// **Ein umbenannter Auszug bei stehendem Elternordner ist einer.**
//
// Das ist der Fall, den die alte Fassung richtig traf — er muss weiterhin
// stumm durchgehen, sonst meldete jeder saubere Wiederanlauf einen Rest.
func TestEinUmbenannterAuszugGiltAlsGetauscht(t *testing.T) {
	t.Parallel()

	frisch, beiwerk, _ := abgebrochenesRueckspielen(t)

	// **Nur das `beiwerk/`**, wie `os.Rename` es hinterlässt.
	if err := os.RemoveAll(beiwerk); err != nil {
		t.Fatalf("räumen: %v", err)
	}

	if rest := frisch.RueckspielVollenden(); rest != "" {
		t.Errorf("Rest %q — der Elternordner steht noch, der Tausch ist belegt", rest)
	}

	if _, err := os.Stat(frisch.markenpfad()); !os.IsNotExist(err) {
		t.Error("die Marke liegt noch da")
	}
}
