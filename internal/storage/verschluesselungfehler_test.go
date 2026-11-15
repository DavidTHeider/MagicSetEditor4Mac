// **Was geschieht, wenn es nicht glückt** — die Fehlerwege der
// Verschlüsselung, Werkstück 14 von Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil es eine eigene Frage ist (§2): nebenan steht,
// dass die Chiffre taugt; hier steht, was ein Mensch erfährt, wenn sie es
// nicht tut. Die Zeilengrenze war der Anlass, der Schnitt folgt der Frage.
package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Fehlerwege der Verschlüsselung** — je einer, und alle betreten.
//
// **Ohne sie stünden vier Zweige da, die niemand je gegangen ist** (§3), und
// jeder von ihnen ist ein Weg, auf dem eine Auskunft an einen Menschen geht.
func TestDieFehlerwegeDerVerschluesselung(t *testing.T) {
	t.Run("ohne Schlüssel wird abgewiesen", func(t *testing.T) {
		klar := filepath.Join(t.TempDir(), "klar.db")
		bestandMit(t, klar, "x")

		db, err := storage.Oeffnen(probekontext(t), klar)
		if err != nil {
			t.Fatalf("öffnen: %v", err)
		}

		t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

		fehler := storage.Verschluesseln(
			probekontext(t), db, filepath.Join(t.TempDir(), "z.db"), "")

		if !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet war eine Abweisung", fehler)
		}
	})

	// **Ein belegtes Ziel ist ein Fehler und kein stilles Überschreiben.**
	// `VACUUM INTO` weigert sich, und diese Weigerung ist die Zusage.
	t.Run("ein belegtes Ziel scheitert", func(t *testing.T) {
		ort := t.TempDir()
		klar := filepath.Join(ort, "klar.db")
		belegt := filepath.Join(ort, "belegt.db")

		bestandMit(t, klar, "x")

		if err := os.WriteFile(belegt, []byte("da liegt schon etwas"), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		db, err := storage.Oeffnen(probekontext(t), klar)
		if err != nil {
			t.Fatalf("öffnen: %v", err)
		}

		t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

		if fehler := storage.Entschluesseln(
			probekontext(t), db, belegt); fehler == nil {
			t.Error("das Ziel wurde überschrieben")
		}
	})
}

// **Und was `IstVerschluesselt` nicht lesen kann, meldet sie.**
//
// Zwei Wege, und sie sind verschieden: eine Datei, die sich **nicht öffnen**
// lässt, und eine, die sich öffnen und **nicht lesen** lässt. Ein Verzeichnis
// ist der zweite Fall — `os.Open` glückt, `io.ReadFull` nicht.
func TestEineUnlesbareDateiMeldetSichStattAlsKlartextZuGelten(t *testing.T) {
	ort := t.TempDir()

	verschlossen := filepath.Join(ort, "zu.db")
	if err := os.WriteFile(verschlossen, []byte("egal"), 0o000); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	for name, pfad := range map[string]string{
		"eine Datei ohne Leserecht": verschlossen,
		"ein Verzeichnis":           ort,
	} {
		t.Run(name, func(t *testing.T) {
			ist, fehler := storage.IstVerschluesselt(pfad)

			if fehler == nil {
				t.Fatalf("kein Fehler, gemeldet wurde verschlüsselt=%v", ist)
			}

			// **Und die Antwort ist nicht „unverschlüsselt".** Ein `false`
			// neben einem Fehler verführt einen unaufmerksamen Aufrufer dazu,
			// den Bestand ohne Schlüssel zu öffnen.
			if ist {
				t.Error("bei einem Fehler soll nichts behauptet werden")
			}
		})
	}
}

// **Ein Schlüssel, der keiner ist, wird beim Öffnen abgewiesen.**
//
// # Warum das ein eigener Weg ist
//
// **Weil er nicht am Inhalt scheitert, sondern an der Form.** Ein *falscher*
// Schlüssel öffnet die Verbindung und liest dann Rauschen — das steht nebenan.
// Ein Schlüssel, der keine 64 Hexziffern ist, kommt gar nicht so weit: die
// Hülle weist das `PRAGMA` selbst ab, und dieser Zweig ist der einzige Weg,
// auf dem der Verbindungshaken einen Fehler meldet.
//
// **Von außen kann das nur eine Naht erzeugen**, denn `NeuerSchluessel`
// würfelt immer 64 Hexziffern — und der Schlüsselbund gibt zurück, was
// jemand hineingelegt hat. Genau dort ist der Fall echt: wer den Eintrag von
// Hand ändert, bekommt eine Auskunft und keinen stillen Fehlschlag.
func TestEinSchluesselDerKeinerIstWirdBeimOeffnenAbgewiesen(t *testing.T) {
	klar := filepath.Join(t.TempDir(), "klar.db")
	bestandMit(t, klar, "x")

	for name, schluessel := range map[string]string{
		"keine Hexziffern": "keinschluessel",
		"zu kurz":          "abcd",
	} {
		t.Run(name, func(t *testing.T) {
			db, fehler := storage.OeffnenMitSchluessel(probekontext(t), klar, schluessel)
			if fehler == nil {
				_ = db.Close() //nolint:errcheck // Test

				t.Fatal("der Bestand ging mit einem unbrauchbaren Schlüssel auf")
			}
		})
	}
}
