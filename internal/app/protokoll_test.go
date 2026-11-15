// **Das Protokoll** (#848).
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Satz landet in der Datei.**
func TestEinSatzStehtInDerProtokolldatei(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	// **`anzahl` und nicht `sets`** — ein Merkmal, das niemand angemeldet
	// hat, verliert seit #857 seinen Wert, und dieser Test prüft ja gerade,
	// dass ein Wert in der Datei ankommt.
	schreiber.Info("der Bestand ist offen", "anzahl", 3)

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	for _, teil := range []string{"der Bestand ist offen", "anzahl", "3"} {
		if !strings.Contains(string(gelesen), teil) {
			t.Errorf("die Zeile nennt %q nicht: %s", teil, gelesen)
		}
	}
}

// **Ein Pfad aus dem Benutzerverzeichnis wird entfernt, nicht durchgereicht.**
//
// # Warum das mehr ist als „wir schreiben ihn nicht hin"
//
// Der Abnahmekasten verlangt es **geprüft**. Ein Test, der nur zusieht, dass
// nichts Verbotenes zufällig dasteht, hält nichts: er wäre grün, solange
// niemand einen Pfad übergibt — und rot erst, wenn es zu spät ist.
//
// Hier wird einer **absichtlich** übergeben, und der Wächter muss ihn
// wegnehmen.
func TestEinBenutzerpfadWirdEntfernt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	zuhause, heimfehler := os.UserHomeDir()
	if heimfehler != nil {
		t.Fatalf("Benutzerverzeichnis: %v", heimfehler)
	}

	geheim := filepath.Join(zuhause, "Dokumente", "Mein Geheimset.mse-set")

	schreiber.Error("das Set ließ sich nicht öffnen", "datei", geheim)

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if strings.Contains(string(gelesen), zuhause) {
		t.Errorf("das Benutzerverzeichnis steht im Protokoll: %s", gelesen)
	}

	// **Und der Dateiname bleibt** — er ist die Auskunft, die den Fehler
	// verständlich macht. Verschwiegen würde die Meldung nutzlos.
	if !strings.Contains(string(gelesen), "Mein Geheimset.mse-set") {
		t.Errorf("der Dateiname fehlt ebenfalls: %s", gelesen)
	}
}

// **Der Ort der Protokolldatei liegt neben dem Bestand.**
func TestDerProtokollpfadLiegtNebenDemBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	protokoll, fehler := app.Protokollpfad()
	if fehler != nil {
		t.Fatalf("Protokollpfad: %v", fehler)
	}

	bestand, fehler := app.Datenbankpfad()
	if fehler != nil {
		t.Fatalf("Datenbankpfad: %v", fehler)
	}

	if filepath.Dir(protokoll) != filepath.Dir(bestand) {
		t.Errorf("%s liegt nicht neben %s", protokoll, bestand)
	}

	if filepath.Base(protokoll) != app.Protokollname {
		t.Errorf("Name %q, erwartet war %q", filepath.Base(protokoll), app.Protokollname)
	}
}

// **Ohne Benutzerverzeichnis gibt es keinen Ort** — und keinen geratenen.
func TestOhneBenutzerverzeichnisGibtEsKeinenProtokollpfad(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.Protokollpfad(); fehler == nil {
		t.Error("kein Fehler")
	}
}

// **Ein unbeschreibbarer Ort meldet sich, statt still zu schweigen.**
func TestEinUnbeschreibbarerOrtMeldetSich(t *testing.T) {
	sperre := filepath.Join(t.TempDir(), "zu")
	if fehler := os.MkdirAll(sperre, 0o500); fehler != nil {
		t.Fatalf("Sperre anlegen: %v", fehler)
	}

	if _, _, fehler := app.Protokoll(filepath.Join(sperre, "tief", "mse.log")); fehler == nil {
		t.Error("kein Fehler für ein unbeschreibbares Verzeichnis")
	}
}

// **Gruppen und Merkmale reichen durch den Schwärzer hindurch.**
//
// `slog` ruft `WithAttrs` und `WithGroup`, sobald jemand `With(...)` oder
// `WithGroup(...)` benutzt. Ohne sie fiele die Schwärzung dort weg — und
// gerade dort steht ein Pfad, den jemand einmal gebunden hat.
func TestGruppenUndMerkmaleBleibenGeschwaerzt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	zuhause, heimfehler := os.UserHomeDir()
	if heimfehler != nil {
		t.Fatalf("Benutzerverzeichnis: %v", heimfehler)
	}

	schreiber.With("ort", filepath.Join(zuhause, "geheim")).
		WithGroup("lauf").
		Info("etwas geschah", "datei", filepath.Join(zuhause, "auch geheim"))

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if strings.Contains(string(gelesen), zuhause) {
		t.Errorf("das Benutzerverzeichnis steht im Protokoll: %s", gelesen)
	}
}

// **Ein Ort, an dem die Datei nicht angelegt werden kann, meldet sich** — und
// zwar anders als ein Verzeichnis, das gar nicht entsteht.
//
// Zwei Fehlerwege, zwei Fälle: `MkdirAll` scheitert am gesperrten Elternteil,
// `OpenFile` an einem Namen, der ein **Verzeichnis** ist.
func TestEinNameDerEinVerzeichnisIstMeldetSich(t *testing.T) {
	ordner := filepath.Join(t.TempDir(), app.Protokollname)
	if fehler := os.MkdirAll(ordner, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if _, _, fehler := app.Protokoll(ordner); fehler == nil {
		t.Error("kein Fehler für einen Namen, der ein Verzeichnis ist")
	}

	// **Und ein anderer Name ist keines** — ein Protokoll heißt `mse.log`.
	if _, _, fehler := app.Protokoll(filepath.Join(t.TempDir(), "anders.log")); fehler == nil {
		t.Error("kein Fehler für einen fremden Namen")
	}
}

// **Ohne Benutzerverzeichnis wird nichts geschwärzt — und nichts verschluckt.**
//
// Der Zweig ist der Rückfall in `schwaerze`: gibt es kein Zuhause, bleibt der
// Text stehen. Das ist richtig — geschwärzt wird ein bekannter Pfad, nicht
// vorsichtshalber alles.
func TestOhneBenutzerverzeichnisBleibtDerTextStehen(t *testing.T) {
	t.Setenv("HOME", "")

	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll anlegen: %v", fehler)
	}

	schreiber.Info("ein Satz", "ort", "/Users/jemand/etwas")

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	gelesen, lesefehler := os.ReadFile(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !strings.Contains(string(gelesen), "/Users/jemand/etwas") {
		t.Errorf("der Text wurde verändert, obwohl es kein Zuhause gibt: %s", gelesen)
	}
}
