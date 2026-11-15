// **Was die Vorschau anfasst und was nicht** (#1162).
//
// Eine eigene Datei, weil sie eine eigene Frage beantwortet: nicht *ob* eine
// Sicherung zurückkommt, sondern *wie viel Arbeit* eine Rückfrage kostet — und
// worüber ein Auszug etwas aussagt.
package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Die Vorschau packt nicht die Bilder aus.**
//
// Sie will wissen, was im Bestand steht — und sonst nichts. Ein Archiv mit
// einem Gigabyte Bildern zweimal auszupacken (einmal für die Vorschau, einmal
// für das Zurückspielen) wäre Arbeit, die niemand angefordert hat, und die
// Wartezeit steht vor der Rückfrage, also vor der Entscheidung.
func TestDieVorschauPacktNichtDieBilderAus(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if err := os.MkdirAll(filepath.Join(bilder, "7"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "7", "kunst.png"),
		[]byte("ein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "mit"+sicherung.Endung))

	var danebenLag []string

	weg.MitSicherungspruefer(
		func(_ context.Context, pfad string) (Sicherungsbefund, error) {
			eintraege, fehler := os.ReadDir(filepath.Dir(pfad))
			if fehler != nil {
				t.Errorf("den Zwischenplatz lesen: %v", fehler)
			}

			for _, eintrag := range eintraege {
				danebenLag = append(danebenLag, eintrag.Name())
			}

			return Sicherungsbefund{}, nil
		})

	// Der Dialog liegt in der Fassade; hier wird der Pfad genannt: ziel

	if _, schaufehler := weg.Beschauen(t.Context(), ziel); schaufehler != nil {
		t.Fatalf("vorschau: %v", schaufehler)
	}

	for _, name := range danebenLag {
		if name == Beiwerkordner {
			t.Errorf("die Vorschau packte auch die Bilder aus: %v", danebenLag)
		}
	}

	if len(danebenLag) != 1 || danebenLag[0] != Bestandsname {
		t.Errorf("neben dem geprüften Bestand lag %v", danebenLag)
	}
}

// **Ein Auszug ersetzt keine Bilder — und ein voller Auszug schon.**
//
// # Warum dieser Lauf für sich steht
//
// Die Zusicherung hängt heute an **einem** Aufrufer: `Beschauen` liest
// `BilderErsetzen` gar nicht, also fiel bei einer Gegenprobe — das Feld immer
// `true` — **kein** Lauf um. Ein Wächter, den man entfernen kann, ohne dass
// etwas rot wird, ist keiner (§10).
//
// Geprüft wird deshalb die Zusicherung selbst, an der Stelle, an der sie
// entsteht: **ein Auszug beschreibt keinen Bilderstand.**
func TestEinAuszugErsetztKeineBilder(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if err := os.MkdirAll(filepath.Join(bilder, "7"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "7", "kunst.png"),
		[]byte("ein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "mit"+sicherung.Endung))

	auszug, fehler := weg.auspacken(t.Context(), ziel, weg.orte.Datenbank, Bestandsname)
	if fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	defer auszug.Wegraeumen()

	if auszug.BilderErsetzen {
		t.Error("ein Auszug gibt sich als Bilderstand aus")
	}

	// **Die Gegenprobe im selben Lauf**: dieselbe Sicherung, ganz ausgepackt,
	// beschreibt sehr wohl einen. Ohne sie sagte der Lauf oben auch dann ja,
	// wenn das Feld nie wahr würde.
	ganz, fehler := weg.auspacken(t.Context(), ziel, weg.orte.Datenbank)
	if fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	defer ganz.Wegraeumen()

	if !ganz.BilderErsetzen {
		t.Error("ein vollständiger Auszug beschreibt keinen Bilderstand")
	}

	if ganz.Beiwerk == "" {
		t.Error("die Bilder liegen nirgends")
	}
}

// **Die Vorschau sagt, was mit den Bildern geschieht.**
//
// Drei Fälle, drei Antworten — und sie hängen am **Format** der gewählten
// Datei. Ohne diese Angabe sagte die Rückfrage im Fenster für jede Sicherung
// dasselbe, und für eine nackte `.db` sagte sie damit das Gegenteil dessen,
// was geschieht (Befund 2 des Nachlaufs über #1162).
func TestDieVorschauSagtWasMitDenBildernGeschieht(t *testing.T) {
	mitBildern := func(t *testing.T, weg *Bestandsweg, ort, bilder string) string {
		t.Helper()

		if err := os.MkdirAll(filepath.Join(bilder, "7"), 0o700); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if err := os.WriteFile(filepath.Join(bilder, "7", "kunst.png"),
			[]byte("ein Bild"), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		return gesichert(t, weg, filepath.Join(ort, "mit"+sicherung.Endung))
	}

	ohneBilder := func(t *testing.T, weg *Bestandsweg, ort, _ string) string {
		t.Helper()

		return gesichert(t, weg, filepath.Join(ort, "ohne"+sicherung.Endung))
	}

	alsNackteDatenbank := func(t *testing.T, weg *Bestandsweg, ort, bilder string) string {
		t.Helper()

		ausgepackt := t.TempDir()
		if _, err := sicherung.Entpacken(
			mitBildern(t, weg, ort, bilder), ausgepackt); err != nil {
			t.Fatalf("die alte Gestalt herstellen: %v", err)
		}

		return filepath.Join(ausgepackt, Bestandsname)
	}

	for name, fall := range map[string]struct {
		bauen    func(*testing.T, *Bestandsweg, string, string) string
		erwartet string
	}{
		"ein Archiv mit Bildern": {mitBildern, BilderErsetzt},
		"ein Archiv ohne Bilder": {ohneBilder, BilderGeleert},
		"eine nackte Datenbank":  {alsNackteDatenbank, BilderUnberuehrt},
	} {
		t.Run(name, func(t *testing.T) {
			weg, ort, bilder := mitBilderort(t)
			weg.MitSicherungspruefer(gepruefteSicherung())

			ziel := fall.bauen(t, weg, ort, bilder)

			// Der Dialog liegt in der Fassade; hier wird der Pfad genannt: ziel

			schau, schaufehler := weg.Beschauen(t.Context(), ziel)
			if schaufehler != nil {
				t.Fatalf("vorschau: %v", schaufehler)
			}

			if schau.Bilder != fall.erwartet {
				t.Errorf("die Vorschau sagt %q, erwartet war %q",
					schau.Bilder, fall.erwartet)
			}
		})
	}
}
