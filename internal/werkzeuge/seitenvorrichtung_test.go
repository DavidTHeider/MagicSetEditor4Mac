package werkzeuge_test

// **Die Vorrichtung der Seitenprüfungen führt jeden Dienst und jede Frage** —
// Zusatzbefund 4 des Nachlaufs über Phase 4.8
// (#1321).
//
// # Warum es diesen Wächter gibt
//
// `frontend/src/test/seitenvorrichtung.tsx` baut den Einstellungsstand, mit
// dem `barrierefreiheit.test.tsx` und `auszeichnung.test.tsx` **jede** Seite
// prüfen. Die vier Dienstschalter und die Wahl je Frage stehen dort von Hand.
//
// **Eine Kopie, die niemand hält, läuft auseinander** — und beide Prüfungen
// blieben dabei grün: ein neuer Dienst erschiene im Fenster und fehlte in der
// Vorrichtung, also sähe ihn weder axe noch der Auszeichnungswächter. Genau
// diese Falle hat in derselben Phase schon einmal zugeschnappt, als der Stand
// dort **leer** war.
//
// # Was er nicht hält
//
// **Ob die Werte stimmen.** Er sieht, dass jeder Dienst und jede Frage
// vorkommt, nicht ob „an" dort „an" heißt. Und er sieht die Vorrichtung,
// nicht die Seite: dass die Seite sie benutzt, prüfen die beiden Läufe dort.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieSeitenvorrichtungFuehrtJedenDienst(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile(filepath.Join( //nolint:gosec // ein fester Pfad im Repositorium
		"..", "..", "frontend", "src", "test", "seitenvorrichtung.tsx"))
	if err != nil {
		t.Fatalf("seitenvorrichtung.tsx lesen: %v", err)
	}

	// **Die Liste und nicht die Datei.**
	//
	// Der erste Anlauf suchte `'scrydex'` irgendwo im Inhalt — und fand es:
	// die Zugangsvorrichtung in derselben Datei nennt denselben Namen. Die
	// Gegenprobe blieb dabei grün, und ein Wächter, den keine Verletzung rot
	// macht, ist keiner (§10). Hier steht deshalb der Block, der die
	// Dienstschalter aufspannt — **der Block und nicht die Zeile**:
	// prettier bricht eine Aufzählung um, sobald sie zu lang wird, und
	// ein Wächter, der eine Zeile liest, wird dann für jeden Eintrag rot.
	dienstliste := abschnittZwischen(t, string(roh),
		"const dienstzeilen", "] as const")
	fragenliste := string(roh)

	gesehen := 0

	for _, schluessel := range domain.Einstellungsreihe() {
		if domain.EinstellungsBereich(schluessel) != domain.BereichDienste {
			continue
		}

		// **Eine Adresse ist kein Dienst** — Werkstück 3 von A18 (#1533).
		// `dienst.openai.adresse` sagt, **wo** ein Dienst erreicht wird; die
		// Vorrichtung führt Dienste, und ein Präfixvergleich verlangte für
		// sie einen Eintrag, den es nicht geben soll.
		if istDienstadresse(schluessel) {
			continue
		}

		gesehen++

		wo, gesucht := dienstliste, strings.TrimPrefix(schluessel, "dienst.")
		if !strings.HasPrefix(schluessel, "dienst.") {
			wo, gesucht = fragenliste, schluessel
		}

		if !strings.Contains(wo, `'`+gesucht+`'`) {
			t.Errorf("die Seitenvorrichtung führt %q nicht — axe und der "+
				"Auszeichnungswächter sähen ihn dann auf keiner Seite",
				schluessel)
		}
	}

	// **Die Gegenprobe zur Schleife**: findet sie gar nichts, ist sie grün
	// und sagt nichts. Sieben sind es heute — sechs Dienste und eine Frage.
	if gesehen == 0 {
		t.Fatal("domain führt keine Einstellung im Bereich `dienste` — " +
			"entweder ist der Bereich umbenannt oder dieser Wächter sucht falsch")
	}
}

// abschnittZwischen gibt den Text zwischen zwei Marken.
//
// **Ein Instrument, das die ganze Datei liest, misst sich selbst** (§10):
// `'scrydex'` steht in `seitenvorrichtung.tsx` auch bei den erlaubten Werten
// der Frage und in der Zugangsvorrichtung, und die Gegenprobe blieb deshalb
// grün. Gelesen wird hier die **Aufzählung der Kennungen** und sonst nichts.
//
// Hier stand bis zum 2026-09-12 `zeileMit`, das **eine** Zeile suchte. Mit
// dem sechsten Dienst brach prettier die Aufzählung auf sieben Zeilen um, und
// der Wächter wurde für jeden Dienst rot — bei unveränderter Vorrichtung.
func abschnittZwischen(t *testing.T, inhalt, anfang, ende string) string {
	t.Helper()

	if zahl := strings.Count(inhalt, anfang); zahl != 1 {
		t.Fatalf("in seitenvorrichtung.tsx steht %q %dmal — erwartet genau "+
			"einmal; entweder ist die Liste umgebaut oder dieser Wächter "+
			"sucht falsch", anfang, zahl)
	}

	// **`Count` hat eben gesagt, dass es genau einen gibt** — `Cut` sagt es
	// noch einmal, und zwar so, dass kein Index dazwischen entsteht, der
	// auch −1 sein könnte.
	_, ab, gefunden := strings.Cut(inhalt, anfang)
	if !gefunden {
		t.Fatalf("in seitenvorrichtung.tsx steht %q nicht mehr", anfang)
	}

	ab = anfang + ab

	bis := strings.Index(ab, ende)
	if bis < 0 {
		t.Fatalf("in seitenvorrichtung.tsx folgt auf %q kein %q — entweder "+
			"ist die Liste umgebaut oder dieser Wächter sucht falsch",
			anfang, ende)
	}

	return ab[:bis+len(ende)]
}

// istDienstadresse sagt, ob ein Schlüssel die Adresse eines Dienstes ist.
//
// **Über `Dienste()` und nicht über eine Endung**: eine Prüfung auf
// `.adresse` nähme jeden künftigen Schlüssel mit, der so endet.
func istDienstadresse(schluessel string) bool {
	for _, dienst := range domain.Dienste() {
		if schluessel == domain.DienstAdresse(dienst) {
			return true
		}
	}

	return false
}
