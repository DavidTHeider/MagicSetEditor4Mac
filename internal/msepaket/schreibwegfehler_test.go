// **Die Fehlerwege des Schreibwegs — jeder betreten** (§3, §10)
// (#1164).
//
// # Warum das eine eigene Datei ist
//
// `schreibweg_test.go` beantwortet „schreibt er richtig?", diese hier „was tut
// er, wenn es schiefgeht?". Der Schnitt folgt der Frage (§2).
//
// # Wie die Wege erreichbar werden
//
// Drei Griffe, und keiner davon ist ein Kunstgriff:
//
//   - **Ein enges Verzeichnis** (`0o500`) — die Nachbardatei lässt sich nicht
//     anlegen. Das ist der Alltag eines Pakets, das jemand schreibgeschützt
//     abgelegt hat.
//   - **Eine `.bak`, die schon ein Verzeichnis ist** — das Umbenennen des
//     vorigen Stands scheitert. Ein Rest aus einem früheren Lauf.
//   - **Ein Schreiber, der ablehnt** — für `archivBauen`, das deshalb einen
//     [io.Writer] nimmt und keine Datei.
package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Name, der aus dem Paket herausführt, wird abgewiesen.**
func TestEinAusbrechenderNameWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	fehler := msepaket.DateiSchreiben(t.Context(), probeverzeichnis(t), "../draussen", []byte("x"), "")
	if !errors.Is(fehler, msepaket.ErrAusbruch) {
		t.Errorf("der Fehler ist %v — will ErrAusbruch", fehler)
	}
}

// **Ein Paket, das es nicht gibt, ist „nicht gefunden".**
func TestEinFehlendesPaketWirdEingeordnet(t *testing.T) {
	t.Parallel()

	fehler := msepaket.DateiSchreiben(t.Context(),
		filepath.Join(t.TempDir(), "nichts.mse-style"), "style", []byte("x"), "")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", fehler)
	}
}

// **Und eine Datei, die es im Paket nicht gibt, ebenso — in beiden Formen.**
func TestEineFehlendeDateiWirdEingeordnet(t *testing.T) {
	t.Parallel()

	for name, anlegen := range map[string]func(*testing.T) string{
		"Verzeichnis": probeverzeichnis,
		"Archiv":      probearchiv,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := msepaket.DateiSchreiben(t.Context(), anlegen(t), "gibtesnicht", []byte("x"), "")
			if !errors.Is(fehler, domain.ErrNotFound) {
				t.Errorf("der Fehler ist %v — will ErrNotFound", fehler)
			}
		})
	}
}

// **Eine Datei, die kein Archiv ist, wird beim Öffnen abgewiesen.**
func TestEinKaputtesArchivWirdBeimSchreibenAbgewiesen(t *testing.T) {
	t.Parallel()

	ort := filepath.Join(t.TempDir(), "probe.mse-style")
	if err := os.WriteFile(ort, []byte("kein Archiv"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x"), ""); fehler == nil {
		t.Error("die Datei kam als Archiv durch")
	}
}

// **In einem engen Verzeichnis entsteht keine Nachbardatei — und nichts geht
// verloren.**
func TestEinEngesVerzeichnisLaesstNichtsSchreiben(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anlegen func(*testing.T) string
		engen   func(ort string) string
	}{
		"Verzeichnis": {probeverzeichnis, func(ort string) string { return ort }},
		"Archiv":      {probearchiv, filepath.Dir},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := fall.anlegen(t)
			vorher := dateiAus(t, ort, "style")
			eng := fall.engen(ort)

			if err := os.Chmod(eng, 0o500); err != nil { // #nosec G302 -- Verzeichnis
				t.Fatalf("engen: %v", err)
			}

			t.Cleanup(func() { _ = os.Chmod(eng, 0o700) }) //nolint:errcheck,gosec // Aufräumen

			fehler := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"), summe([]byte(vorher)))
			if fehler == nil {
				t.Fatal("das enge Verzeichnis nahm die Datei an")
			}

			// **Der Schritt steht in der Meldung.** Welcher es ist, hängt am
			// Fall: ein enges Paketverzeichnis lässt die Nachbardatei nicht
			// anlegen, ein enges **Elternverzeichnis** schon die Sperrdatei
			// nicht. Beides ist dieselbe Ursache und beides sagt, woran es
			// lag — was nicht kommen darf, ist eine Meldung ohne Schritt.
			if !strings.Contains(fehler.Error(), "Nachbardatei") &&
				!strings.Contains(fehler.Error(), "Schreibsperre") {
				t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
			}
		})
	}
}

// **Ein `.bak`, an dessen Stelle eine Datei liegt, lässt den Rückweg nicht
// anlegen — und der vorige Stand bleibt stehen.**
//
// Nur beim Verzeichnispaket: dort ist der Rückweg selbst ein Verzeichnis
// (`<paket>.mse-style.bak/`), beim Archiv eine Datei.
func TestEinVerbauterRueckwegLaesstDenStandStehen(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)
	vorher := dateiAus(t, ort, "style")

	if err := os.WriteFile(ort+msepaket.Sicherungsendung, []byte("keine Ablage"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	fehler := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"), summe([]byte(vorher)))
	if fehler == nil {
		t.Fatal("der verbaute Rückweg kam durch")
	}

	if !strings.Contains(fehler.Error(), "Rückweg") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}

	if nachher := dateiAus(t, ort, "style"); nachher != vorher {
		t.Errorf("aus %q wurde %q", vorher, nachher)
	}
}

// **Ein `.bak`, das schon ein Verzeichnis ist, lässt das Sichern scheitern —
// und der vorige Stand bleibt stehen.**
//
// Das ist der Zweig, auf dem Schreiben, Schließen und beide Umbenennungen
// zusammenlaufen. **Dass der alte Stand danach noch dasteht, ist die halbe
// Zusage**: ein Abbruch mitten im Tausch wäre schlimmer als kein Schutz.
func TestEinBelegterRueckwegLaesstDenStandStehen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		anlegen func(*testing.T) string
		wo      func(ort string) string
	}{
		"Verzeichnis": {probeverzeichnis, func(ort string) string {
			return filepath.Join(ort+msepaket.Sicherungsendung, "style")
		}},
		"Archiv": {probearchiv, func(ort string) string { return ort + msepaket.Sicherungsendung }},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := fall.anlegen(t)
			vorher := dateiAus(t, ort, "style")

			// **Nicht leer.** Ein leeres Verzeichnis ersetzt `os.Rename` unter
			// macOS klaglos; erst ein belegtes weist ab.
			imWeg := fall.wo(ort)
			if err := os.MkdirAll(filepath.Join(imWeg, "belegt"), 0o750); err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			fehler := msepaket.DateiSchreiben(t.Context(), ort, "style", []byte("x\n"), summe([]byte(vorher)))
			if fehler == nil {
				t.Fatal("der belegte Rückweg kam durch")
			}

			if nachher := dateiAus(t, ort, "style"); nachher != vorher {
				t.Errorf("aus %q wurde %q", vorher, nachher)
			}
		})
	}
}
