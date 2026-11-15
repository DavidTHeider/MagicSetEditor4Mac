package msepaket_test

// **Ein Paket als Archiv neu anlegen** — Werkstück C von Phase 4.9
// (#1324).
//
// # Die Lücke, und woher sie bekannt ist
//
// `docs/MSE2BESTAND.md` führt sie beim Befehl `save set as directory`:
//
//	**Und mehr als das: es ist unsere einzige Form.** `Setschreiber.Schreibe`
//	legt ein **Verzeichnis** an (`os.MkdirAll`) und schreibt die Hauptdatei
//	hinein; MSE2 hat beides und macht das Archiv zur Vorgabe
//	(`Set::saveAs(…, true, true)`, `window.cpp` 605–614). **Was uns fehlt, ist
//	also das Gegenteil**: ein Set als `.mse-set`-**Archiv** zu schreiben.
//
// `msepaket` konnte ein **vorhandenes** Archiv umschreiben (`insArchiv`),
// aber keines anlegen. Damit war MSEs Vorgabeform hier gar nicht erreichbar.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein angelegtes Archiv lässt sich wieder öffnen.**
//
// **Das ist die Abnahme**, und sie geht durch den vorhandenen Leser: was
// `Oeffne` liest, liest auch MSE2 — derselbe Rundlauf, den
// `TestDerRundlaufGegenEineEchtePaketdatei` für die andere Richtung fährt.
func TestEinAngelegtesArchivLaesstSichOeffnen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	dateien := map[string][]byte{
		"set":             []byte("mse version: 2.0.0\ngame: magic\n"),
		"bilder/eins.png": []byte("nicht wirklich ein Bild"),
	}

	if err := msepaket.ArchivAnlegen(ziel, dateien); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	paket, err := msepaket.Oeffne(ziel)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	if paket.Art != msepaket.ArtSet {
		t.Errorf("die Art ist %q, erwartet %q", paket.Art, msepaket.ArtSet)
	}

	for name, erwartet := range dateien {
		gelesen, lesefehler := paket.Datei(name)
		if lesefehler != nil {
			t.Errorf("%s lesen: %v", name, lesefehler)

			continue
		}

		if !bytes.Equal(gelesen, erwartet) {
			t.Errorf("%s enthält %q, erwartet %q", name, gelesen, erwartet)
		}
	}
}

// **Es entsteht eine Datei und kein Verzeichnis.**
//
// Das ist der ganze Unterschied zur bisherigen Form, und er ist der Grund für
// dieses Werkstück: wer ein Set weitergibt, gibt **eine** Datei weiter.
func TestEinAngelegtesArchivIstEineDatei(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "Probe.mse-set")

	if err := msepaket.ArchivAnlegen(ziel,
		map[string][]byte{"set": []byte("mse version: 2.0.0\n")}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	angaben, err := os.Stat(ziel)
	if err != nil {
		t.Fatalf("nachsehen: %v", err)
	}

	if angaben.IsDir() {
		t.Error("es entstand ein Verzeichnis, erwartet eine Datei")
	}
}

// **Und die Fehlerwege** (§3).
func TestArchivAnlegenMeldetWasNichtGeht(t *testing.T) {
	t.Parallel()

	t.Run("ohne Hauptdatei", func(t *testing.T) {
		t.Parallel()

		err := msepaket.ArchivAnlegen(
			filepath.Join(t.TempDir(), "Probe.mse-set"),
			map[string][]byte{"bilder/eins.png": []byte("x")})

		if err == nil {
			t.Fatal("das Anlegen glückte ohne Hauptdatei")
		}
	})

	t.Run("in ein Verzeichnis, das es nicht gibt", func(t *testing.T) {
		t.Parallel()

		err := msepaket.ArchivAnlegen(
			filepath.Join(t.TempDir(), "gibtsnicht", "Probe.mse-set"),
			map[string][]byte{"set": []byte("x")})

		if err == nil {
			t.Fatal("das Anlegen glückte ohne Verzeichnis")
		}
	})

	// **Eine unbekannte Endung hat keine Hauptdatei** — und damit auch keine
	// Zusage darüber, was drin stehen muss.
	t.Run("ohne Paketendung", func(t *testing.T) {
		t.Parallel()

		err := msepaket.ArchivAnlegen(
			filepath.Join(t.TempDir(), "Probe.zip"),
			map[string][]byte{"set": []byte("x")})

		if err == nil {
			t.Fatal("das Anlegen glückte ohne Paketendung")
		}
	})

	// **Ein Name, der aus dem Paket hinausführt, wird abgewiesen** — dieselbe
	// Prüfung wie beim Lesen (`paketpfad`).
	t.Run("mit einem Namen, der hinausführt", func(t *testing.T) {
		t.Parallel()

		err := msepaket.ArchivAnlegen(
			filepath.Join(t.TempDir(), "Probe.mse-set"),
			map[string][]byte{
				"set":         []byte("x"),
				"../draussen": []byte("x"),
			})

		if err == nil {
			t.Fatal("ein Name, der hinausführt, kam durch")
		}
	})
}
