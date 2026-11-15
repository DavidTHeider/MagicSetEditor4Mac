// **Die Ränder des Archivs** — die Fehlerwege, die erst der Zweitblick über
// #1162
// nötig gemacht hat.
//
// Sie stehen für sich, weil sie eine eigene Art von Vorrichtung brauchen: ein
// Archiv, das an genau einer Stelle nicht dem entspricht, was diese Anwendung
// schreibt.
package sicherung_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Ein Ziel, das sich nicht verknüpfen lässt, wird benannt gemeldet.**
//
// Nicht „liegt schon da": das wäre die falsche Auskunft und schickte jemanden
// auf die Suche nach einer Datei, die es nicht gibt.
func TestEinUnmoeglichesZielWirdGemeldet(t *testing.T) {
	// **Ein Name über NAME_MAX.** Die Nachbardatei entsteht noch — ihr Name
	// ist kurz —, das Verknüpfen scheitert.
	ziel := filepath.Join(t.TempDir(), strings.Repeat("z", 300))

	fehler := sicherung.Packen(ziel, probeteile(t), sicherung.Angaben{
		Fassung: "0.3.0",
	})
	if fehler == nil {
		t.Fatal("der unmögliche Name kam durch")
	}

	if strings.Contains(fehler.Error(), "liegt schon da") {
		t.Errorf("die Meldung schickt auf die falsche Fährte: %v", fehler)
	}
}

// **Ein Manifest, dessen Strom verdorben ist, wird gemeldet.**
func TestEinVerdorbenesManifestWirdGemeldet(t *testing.T) {
	pfad := gepacktesArchiv(t)

	roh, err := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	leser, err := zip.NewReader(bytes.NewReader(roh), int64(len(roh)))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	for _, eintrag := range leser.File {
		if eintrag.Name != sicherung.Manifestname {
			continue
		}

		beginn, offsetfehler := eintrag.DataOffset()
		if offsetfehler != nil {
			t.Fatalf("Datenversatz: %v", offsetfehler)
		}

		// **Genau im Manifest und sonst nirgends.** Ein Byte irgendwo im
		// Archiv verdürbe womöglich einen Teil, und dann prüfte der Lauf den
		// falschen Zweig (§10).
		roh[beginn+2] ^= 0xFF
	}

	// #nosec G703 -- der Pfad kommt aus t.TempDir(), nicht von außen.
	if fehler := os.WriteFile(pfad, roh, 0o600); fehler != nil {
		t.Fatalf("zurückschreiben: %v", fehler)
	}

	if _, fehler := sicherung.Entpacken(pfad, t.TempDir()); fehler == nil {
		t.Error("das verdorbene Manifest kam durch")
	}
}

// **Ein Manifest mit zu vielen Teilen wird abgewiesen.**
//
// Es ist der Fall, den die beiden Größengrenzen durchlassen: sehr viele sehr
// kleine Einträge.
func TestEinManifestMitZuVielenTeilenWirdAbgewiesen(t *testing.T) {
	pfad := gepacktesArchiv(t)

	umbauen(t, pfad, func(name string, roh []byte) ([]byte, bool) {
		if name != sicherung.Manifestname {
			return roh, true
		}

		var gebaut bytes.Buffer

		gebaut.WriteString(`{"fassung":"0.3.0","erstellt":"0001-01-01T00:00:00Z","teile":[`)

		for i := 0; i <= sicherung.HoechstensEintraege; i++ {
			if i > 0 {
				gebaut.WriteByte(',')
			}

			gebaut.WriteString("{}")
		}

		gebaut.WriteString(`]}`)

		return gebaut.Bytes(), true
	})

	_, fehler := sicherung.Entpacken(pfad, t.TempDir())
	if fehler == nil {
		t.Fatal("das Manifest mit zu vielen Teilen kam durch")
	}

	if !strings.Contains(fehler.Error(), "Teile") {
		t.Errorf("die Meldung sagt nicht, was zu viel ist: %v", fehler)
	}
}

// **Steht am Ort eines Teils ein Verzeichnis, wird es gemeldet.**
//
// Der Fall entsteht, wenn jemand in das Zielverzeichnis hineingeräumt hat —
// oder wenn zweimal in dasselbe ausgepackt wird.
func TestEinVerzeichnisAmOrtEinesTeilsWirdGemeldet(t *testing.T) {
	pfad := gepacktesArchiv(t)
	ziel := t.TempDir()

	imWeg := filepath.Join(ziel, "beiwerk", "4", "bild.png", "darin")
	if err := os.MkdirAll(imWeg, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, fehler := sicherung.Entpacken(pfad, ziel); fehler == nil {
		t.Error("das Verzeichnis im Weg wurde übergangen")
	}
}

// **Ein Eintrag in einem unbekannten Packverfahren wird gemeldet.**
//
// Ein ZIP darf Verfahren tragen, die dieses Programm nicht kennt — AES etwa.
// Es öffnet dann gar nicht erst, und das ist etwas anderes als eine falsche
// Prüfsumme.
func TestEinUnbekanntesPackverfahrenWirdGemeldet(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "fremd"+sicherung.Endung)

	datei, err := os.Create(pfad) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	schreiber := zip.NewWriter(datei)

	// **Der Schreiber kennt das Verfahren erst, wenn man es ihm sagt** — der
	// Leser in `Entpacken` nicht. Genau dieses Gefälle stellt den Fall her.
	schreiber.RegisterCompressor(99, func(ziel io.Writer) (io.WriteCloser, error) {
		return durchreichen{ziel}, nil
	})

	manifest, err := schreiber.Create(sicherung.Manifestname)
	if err != nil {
		t.Fatalf("eintragen: %v", err)
	}

	if fehler := json.NewEncoder(manifest).Encode(sicherung.Beschreibung{
		Teile: []sicherung.Teilangabe{{Name: "bestand.db", Groesse: 4, Summe: "x"}},
	}); fehler != nil {
		t.Fatalf("Manifest schreiben: %v", fehler)
	}

	// **Verfahren 99 ist AES-verschlüsselt** und in `archive/zip` nicht
	// eingetragen; `Open` gibt `ErrAlgorithm`.
	teil, err := schreiber.CreateHeader(&zip.FileHeader{
		Name: "bestand.db", Method: 99,
	})
	if err != nil {
		t.Fatalf("eintragen: %v", err)
	}

	if _, fehler := teil.Write([]byte("egal")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := datei.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := sicherung.Entpacken(pfad, t.TempDir()); fehler == nil {
		t.Error("das unbekannte Packverfahren kam durch")
	}
}

// durchreichen schreibt, was hereinkommt — ein Packverfahren, das nicht packt.
type durchreichen struct{ io.Writer }

func (durchreichen) Close() error { return nil }
