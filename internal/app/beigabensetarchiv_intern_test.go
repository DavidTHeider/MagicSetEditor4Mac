package app

// **`write_set_file` legt ein Archiv, kein Verzeichnis** — Werkstück C von
// Phase 4.9 (#1324).
//
// # Was nachgelesen wurde
//
//	SCRIPT_FUNCTION(write_set_file) {
//	  …
//	  set->saveCopy(out_path); // TODO: use export_set instead?
//	}
//
// `src/script/functions/export.cpp`, Zeilen 429–439 — und `saveCopy` ist
//
//	void Package::saveCopy(const String& name) {
//	  saveToZipfile(name, true, true);
//	  clearKeepFlag();
//	}
//
// `src/util/io/package.cpp`, Zeilen 112–115. Beides Fassung f9d9356 vom
// 2021-07-06. **MSE2 schreibt ein ZIP.**
//
// # Warum hier ein Verzeichnis stand
//
// Der Kommentar an `setleger` berief sich auf
// [0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md):
// ein ZIP-Schreiber weiche die Byte-Zusage auf. **Der Vermerk deckt diesen
// Fall nicht** — er heißt „wird nicht **zurück**geschrieben", und eine Kopie
// neben die Ausgabe schreibt nichts zurück. Die Zusage aus #48 gilt dem
// Rundlauf über die **Herkunft**; die bleibt unberührt.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// setmitverzeichnis schreibt ein Set als Verzeichnis — wie der echte Schreiber.
type setmitverzeichnis struct{ dateien map[string]string }

func (s setmitverzeichnis) Schreibe(
	_ context.Context, _ int64, ziel string,
) (*service.Speicherung, error) {
	if err := os.MkdirAll(ziel, 0o750); err != nil {
		return nil, err
	}

	for name, inhalt := range s.dateien {
		pfad := filepath.Join(ziel, name)

		if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
			return nil, err
		}

		if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
			return nil, err
		}
	}

	return &service.Speicherung{}, nil
}

func TestWriteSetFileLegtEinArchiv(t *testing.T) {
	t.Parallel()

	ziel := t.TempDir()

	beigabe := beigaben{
		ziel: ziel,
		setschreibt: setmitverzeichnis{dateien: map[string]string{
			"set":             "mse version: 2.0.0\ngame: magic\n",
			"bilder/eins.png": "kein echtes Bild",
		}},
	}

	name, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set")
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if name != "kopie.mse-set" {
		t.Errorf("der Name ist %q, erwartet %q", name, "kopie.mse-set")
	}

	gelegt := filepath.Join(ziel, name)

	angaben, statfehler := os.Stat(gelegt)
	if statfehler != nil {
		t.Fatalf("nachsehen: %v", statfehler)
	}

	// **Eine Datei und kein Verzeichnis** — das ist der ganze Befund.
	if angaben.IsDir() {
		t.Fatal("es entstand ein Verzeichnis, erwartet ein Archiv")
	}

	// **Und es lässt sich öffnen** — mit allem, was das Beiwerk mitgebracht hat.
	paket, oeffnefehler := msepaket.Oeffne(gelegt)
	if oeffnefehler != nil {
		t.Fatalf("öffnen: %v", oeffnefehler)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	for _, name := range []string{"set", "bilder/eins.png"} {
		if _, lesefehler := paket.Datei(name); lesefehler != nil {
			t.Errorf("%s fehlt im Archiv: %v", name, lesefehler)
		}
	}

	// **Und das Temp-Verzeichnis bleibt nicht liegen.**
	eintraege, lesefehler := os.ReadDir(ziel)
	if lesefehler != nil {
		t.Fatalf("das Ziel lesen: %v", lesefehler)
	}

	if len(eintraege) != 1 {
		t.Errorf("im Ziel liegen %d Einträge, erwartet nur das Archiv",
			len(eintraege))
	}
}

// **Und die Fehlerwege** (§3).
func TestDasSetarchivMeldetWasSchiefgeht(t *testing.T) {
	t.Parallel()

	// **Ein Set ohne Hauptdatei kommt nicht ins Archiv.** `ArchivAnlegen`
	// weist es ab, und die Meldung sagt, welche Datei fehlt.
	t.Run("ohne Hauptdatei", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{
			ziel: t.TempDir(),
			setschreibt: setmitverzeichnis{dateien: map[string]string{
				"bilder/eins.png": "kein echtes Bild",
			}},
		}

		if _, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set"); err == nil {
			t.Error("ein Set ohne Hauptdatei kam durch")
		}
	})

	// **Ein Ziel, das es nicht gibt**, trägt kein Zwischenverzeichnis: es
	// entsteht **neben** dem Ziel, und dessen Elternverzeichnis muss da sein.
	t.Run("in ein Ziel, dessen Ort fehlt", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{
			ziel:        filepath.Join(t.TempDir(), "gibtsnicht", "auchnicht"),
			setschreibt: setmitverzeichnis{dateien: map[string]string{"set": "x"}},
		}

		if _, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set"); err == nil {
			t.Error("ein Ziel ohne Ort kam durch")
		}
	})

	// **Und eine Datei, die sich nicht lesen lässt.** Der Setschreiber legt
	// sie an, danach wird ihr das Leserecht genommen — das ist der Zweig, den
	// `ausVerzeichnis` für einen Lesefehler hält.
	t.Run("mit einer unlesbaren Datei", func(t *testing.T) {
		t.Parallel()

		beigabe := beigaben{
			ziel:        t.TempDir(),
			setschreibt: unlesbaresSet{},
		}

		if _, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set"); err == nil {
			t.Error("eine unlesbare Datei kam durch")
		}
	})
}

// **Ein Schreiber, der Erfolg meldet und nichts anlegt, wird gemeldet.**
//
// **Und nicht als leeres Archiv weggeschrieben.** Genau diese Vorrichtung gab
// es hier: `setmitschrift` gab „geschrieben" zurück, ohne ein Verzeichnis
// anzulegen — großzügiger als der Betrieb (§3). Seit `LegeSet` daraus ein
// Archiv packt, fällt es auf; der Fall bleibt als Test stehen, damit er es
// auch weiterhin tut.
func TestEinSchreiberOhneVerzeichnisWirdGemeldet(t *testing.T) {
	t.Parallel()

	beigabe := beigaben{ziel: t.TempDir(), setschreibt: schreiberOhneWirkung{}}

	if _, err := beigabe.LegeSet(probekontext(t), "kopie.mse-set"); err == nil {
		t.Error("ein Schreiber ohne Wirkung kam durch")
	}
}

// schreiberOhneWirkung meldet Erfolg und legt nichts an.
type schreiberOhneWirkung struct{}

func (schreiberOhneWirkung) Schreibe(
	context.Context, int64, string,
) (*service.Speicherung, error) {
	return &service.Speicherung{}, nil
}

// unlesbaresSet legt ein Set an, dessen Beiwerk sich nicht lesen lässt.
type unlesbaresSet struct{}

func (unlesbaresSet) Schreibe(
	_ context.Context, _ int64, ziel string,
) (*service.Speicherung, error) {
	if err := os.MkdirAll(ziel, 0o750); err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(ziel, "set"), []byte("x"), 0o600); err != nil {
		return nil, err
	}

	// **Ohne Leserecht** — und als Datei, nicht als Verzeichnis: ein
	// Verzeichnis ohne Recht meldete `WalkDir` schon beim Betreten.
	return &service.Speicherung{},
		os.WriteFile(filepath.Join(ziel, "geheim"), []byte("x"), 0o000)
}
