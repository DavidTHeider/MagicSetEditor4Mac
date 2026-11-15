// **Die Wege, auf denen eine Sicherung nicht zustande kommt** —
// #1162.
//
// Sie stehen für sich, weil `archiv_test.go` sagt, was gelten soll, und hier
// steht, was geschieht, wenn es nicht geht. Jeder dieser Zweige ist einer, den
// jemand im Ernstfall liest — und ein Zweig, den kein Test betritt, ist eine
// Behauptung darüber, was dann passiert (§3).
package sicherung_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Eine Sicherung überschreibt nichts.**
func TestEinBelegtesZielWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "probe"+sicherung.Endung)
	if err := os.WriteFile(ziel, []byte("da liegt schon was"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	err := sicherung.Packen(ziel, probeteile(t), sicherung.Angaben{})

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Ein Ziel, das sich nicht anlegen lässt, wird gemeldet.**
func TestEinUnbrauchbaresZielWirdGemeldet(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "gibtsnicht", "probe"+sicherung.Endung)

	if err := sicherung.Packen(ziel, probeteile(t), sicherung.Angaben{}); err == nil {
		t.Error("ein Ziel ohne Verzeichnis kam durch")
	}
}

// **Ein Teil, den es nicht gibt, lässt nichts liegen.**
//
// **Der zweite Satz ist der wichtigere.** Eine halb geschriebene Sicherung
// sieht aus wie eine; sie fällt erst auf, wenn jemand sie braucht.
func TestEinFehlenderTeilLaesstNichtsLiegen(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "probe"+sicherung.Endung)

	err := sicherung.Packen(ziel, []sicherung.Teil{
		{Name: "weg.db", Quelle: filepath.Join(t.TempDir(), "gibtsnicht")},
	}, sicherung.Angaben{})

	if err == nil {
		t.Fatal("ein fehlender Teil kam durch")
	}

	if _, statfehler := os.Stat(ziel); !errors.Is(statfehler, os.ErrNotExist) {
		t.Error("es blieb eine halbe Sicherung liegen")
	}
}

// **Eine Datei, die kein Archiv ist, wird als solche gemeldet.**
func TestEineFremdeDateiWirdGemeldet(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "fremd"+sicherung.Endung)
	if err := os.WriteFile(pfad, []byte("kein ZIP"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := sicherung.Entpacken(pfad, t.TempDir()); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **Ein unlesbares Manifest ebenso.**
func TestEinUnlesbaresManifestWirdGemeldet(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "krumm"+sicherung.Endung)

	nurEintrag(t, pfad, sicherung.Manifestname, []byte("{ das ist kein JSON"))

	if _, err := sicherung.Entpacken(pfad, t.TempDir()); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **Ein Eintrag, der aus der Sicherung hinausführt, wird abgewiesen.**
//
// # Warum das hier steht und nicht nur beim Paketbezug
//
// Eine Sicherung kommt von derselben Person, die sie zurückspielt — **aber
// eine Datei aus dem Netz sieht genauso aus.** Wer eine fremde `.msesicherung`
// bekommt und einspielt, gäbe ihr sonst einen Schreibweg an jede Stelle, an
// die dieses Benutzerkonto darf.
func TestEinEintragDerHinausfuehrtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"hinauf":  "../hinaus.txt",
		"absolut": "/etc/hinaus.txt",
		"getarnt": "beiwerk/../../hinaus.txt",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pfad := filepath.Join(t.TempDir(), "boese"+sicherung.Endung)

			bauenMitManifest(t, pfad, weg, []byte("hallo"))

			if _, err := sicherung.Entpacken(pfad, t.TempDir()); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("der Fehler ist %v — will ErrValidation", err)
			}
		})
	}
}

// bauenMitManifest schreibt ein Archiv mit einem Teil und passendem Manifest.
func bauenMitManifest(t *testing.T, pfad, name string, inhalt []byte) {
	t.Helper()

	quelle := filepath.Join(t.TempDir(), "teil")
	if err := os.WriteFile(quelle, inhalt, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	echt := filepath.Join(t.TempDir(), "echt"+sicherung.Endung)
	if err := sicherung.Packen(echt, []sicherung.Teil{
		{Name: "teil", Quelle: quelle},
	}, sicherung.Angaben{}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	// **Den Namen im Manifest **und** im Archiv umbiegen.** Ein Manifest, das
	// einen Teil nennt, den es nicht gibt, scheitert schon an der
	// Vollständigkeit — und dann prüfte der Test den falschen Zweig.
	umbauen(t, echt, func(eintrag string, roh []byte) ([]byte, bool) {
		if eintrag != sicherung.Manifestname {
			return roh, true
		}

		var beschreibung sicherung.Beschreibung
		if fehler := json.Unmarshal(roh, &beschreibung); fehler != nil {
			t.Fatalf("Manifest lesen: %v", fehler)
		}

		beschreibung.Teile[0].Name = name

		neu, fehler := json.Marshal(beschreibung)
		if fehler != nil {
			t.Fatalf("Manifest schreiben: %v", fehler)
		}

		return neu, true
	})

	umbauen(t, echt, func(_ string, roh []byte) ([]byte, bool) {
		return roh, true
	})

	benennen(t, echt, "teil", name)

	if err := os.Rename(echt, pfad); err != nil {
		t.Fatalf("umbenennen: %v", err)
	}
}

// **Ein Zielverzeichnis, das eine Datei ist, wird gemeldet.**
func TestEinZielDasEineDateiIstWirdGemeldet(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "probe"+sicherung.Endung)
	if err := sicherung.Packen(pfad, probeteile(t), sicherung.Angaben{}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	ordner := t.TempDir()

	// **`beiwerk` ist eine Datei**, und darunter soll ein Teil liegen.
	if err := os.WriteFile(filepath.Join(ordner, "beiwerk"), []byte("!"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := sicherung.Entpacken(pfad, ordner); err == nil {
		t.Error("ein Ziel, das eine Datei ist, kam durch")
	}
}

// **Ein Zielverzeichnis, in das niemand schreiben darf, ebenso.**
func TestEinUnbeschreibbaresZielverzeichnisWirdGemeldet(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "probe"+sicherung.Endung)
	if err := sicherung.Packen(pfad, []sicherung.Teil{probeteile(t)[0]},
		sicherung.Angaben{}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	ordner := t.TempDir()

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	if err := os.Chmod(ordner, 0o500); err != nil {
		t.Fatalf("Rechte setzen: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(ordner, 0o700) }) //nolint:errcheck // nur zum Aufräumen

	if _, err := sicherung.Entpacken(pfad, ordner); err == nil {
		t.Error("ein unbeschreibbares Verzeichnis kam durch")
	}
}

// **Ein Teil, dessen Bytes im Archiv verdorben sind, wird gemeldet.**
//
// # Der Unterschied zu „verändert"
//
// [TestEineVeraenderteSicherungWirdAbgewiesen] schreibt das Archiv **richtig**
// neu und tauscht nur den Inhalt — die Prüfsumme des Manifests fängt es.
// Hier werden **rohe Bytes** im gepackten Strom umgedreht, und dann scheitert
// schon das Lesen: `archive/zip` prüft seine eigene CRC.
//
// **Zwei Wege, ein Schutz** — und beide gehören geprüft, denn eine kaputte
// Platte macht das eine und ein Eingriff das andere.
func TestEinVerdorbenerStromWirdGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	// **Unkomprimierbar und groß**, damit die Mitte der Datei sicher im
	// gepackten Strom liegt und nicht in einem Kopf.
	roh := make([]byte, 256<<10)
	for i := range roh {
		roh[i] = byte(i*7 + i/251)
	}

	quelle := filepath.Join(wurzel, "teil")
	if err := os.WriteFile(quelle, roh, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	pfad := filepath.Join(wurzel, "probe"+sicherung.Endung)
	if err := sicherung.Packen(pfad, []sicherung.Teil{
		{Name: "teil", Quelle: quelle},
	}, sicherung.Angaben{}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	archiv, err := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	mitte := len(archiv) / 2
	archiv[mitte] ^= 0xFF

	// #nosec G703 -- der Pfad kommt aus t.TempDir(), nicht von außen.
	if fehler := os.WriteFile(pfad, archiv, 0o600); fehler != nil {
		t.Fatalf("zurückschreiben: %v", fehler)
	}

	if _, entpackfehler := sicherung.Entpacken(pfad, t.TempDir()); entpackfehler == nil {
		t.Error("ein verdorbener Strom kam durch")
	}
}
