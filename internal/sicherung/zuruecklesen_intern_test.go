// **Eine Sicherung wird zurückgelesen, bevor sie ihren Namen bekommt** —
// OPS01 (#2347).
//
// # Warum innen
//
// Weil eine Datei, die beim Zurücklesen anders ist als beim Schreiben, von
// außen nur über eine kaputte Platte herzustellen wäre. Eine Hülle, die ein
// Byte kippt, nachdem die Summen schon gerechnet sind, ist genau dieser Fall.
package sicherung

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kippenderSchreiber kippt das Byte an Stelle `bei` — **unter** dem Archiv,
// also nachdem `teilSchreiben` seine Summe gerechnet hat.
type kippenderSchreiber struct {
	nach io.Writer
	bei  int
	hat  int
}

func (k *kippenderSchreiber) Write(teil []byte) (int, error) {
	kopie := append([]byte(nil), teil...)
	if stelle := k.bei - k.hat; stelle >= 0 && stelle < len(kopie) {
		kopie[stelle] ^= 0xff
	}

	k.hat += len(kopie)

	return k.nach.Write(kopie)
}

// einTeil legt eine Quelle mit unkomprimierbarem Inhalt an.
func einTeil(t *testing.T) []Teil {
	t.Helper()

	quelle := filepath.Join(t.TempDir(), "teil")
	if fehler := os.WriteFile(quelle, zufall(t, 64<<10), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	return []Teil{{Name: "bestand.db", Quelle: quelle}}
}

// abdruckVon rechnet nach, was auf der Platte liegt.
func abdruckVon(t *testing.T, pfad string) Abdruck {
	t.Helper()

	roh, fehler := os.ReadFile(pfad) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	summe := sha256.Sum256(roh)

	return Abdruck{Groesse: int64(len(roh)), Summe: hex.EncodeToString(summe[:])}
}

func TestDerAbdruckIstDerDerGeschriebenenDatei(t *testing.T) {
	defer ArbeitsfaktorSetzen(1)()

	for _, kennwort := range []string{"", "ein langes Kennwort"} {
		ziel := filepath.Join(t.TempDir(), "bestand.msesicherung")

		abdruck, fehler := PackenGeprueft(ziel, einTeil(t), Angaben{Fassung: "0.9.0"}, kennwort)
		if fehler != nil {
			t.Fatalf("mit Kennwort %v: %v", kennwort != "", fehler)
		}

		if erwartet := abdruckVon(t, ziel); abdruck != erwartet {
			t.Errorf("mit Kennwort %v: Abdruck %+v, auf der Platte %+v", kennwort != "", abdruck, erwartet)
		}
	}
}

func TestEineVerdorbeneSicherungBekommtIhrenNamenNicht(t *testing.T) {
	ordner := t.TempDir()
	ziel := filepath.Join(ordner, "bestand.msesicherung")
	teile := einTeil(t)

	// **Mitten in den Inhalt des Teils** — nicht in den Kopf, den `zip`
	// ohnehin anders abwiese.
	kippen := func(nach io.Writer, schreiben func(io.Writer) error) error {
		return schreiben(&kippenderSchreiber{nach: nach, bei: 32 << 10})
	}

	_, fehler := packenIn(ziel, Manifestname, teile, Angaben{}, kippen, offenGelesen)
	if !errors.Is(fehler, domain.ErrSicherungVerworfen) {
		t.Fatalf("Fehler %v, erwartet ErrSicherungVerworfen", fehler)
	}

	if eintraege, _ := os.ReadDir(ordner); len(eintraege) != 0 { //nolint:errcheck // leer ist leer
		t.Errorf("im Ordner liegt %v — eine verworfene Sicherung hinterlässt nichts", eintraege)
	}
}

// **Jede Stelle, an der das Zurücklesen scheitert, scheitert laut** — und
// keine davon ließe sich über eine heile Platte herstellen.
func TestDasZuruecklesenWeistJedeAbweichungAb(t *testing.T) {
	ordner := t.TempDir()

	schreiben := func(name string, fuellen func(*zip.Writer)) *os.File {
		datei, fehler := os.Create(filepath.Join(ordner, name)) // #nosec G304 -- aus t.TempDir()
		if fehler != nil {
			t.Fatal(fehler)
		}

		t.Cleanup(func() { _ = datei.Close() }) //nolint:errcheck // nur gelesen

		if fuellen == nil {
			_, _ = datei.WriteString("kein Archiv") //nolint:errcheck // geprüft wird das Lesen

			return datei
		}

		archiv := zip.NewWriter(datei)
		fuellen(archiv)

		if fehler := archiv.Close(); fehler != nil {
			t.Fatal(fehler)
		}

		return datei
	}

	manifest := func(archiv *zip.Writer, inhalt string) {
		eintrag, _ := archiv.Create(Manifestname) //nolint:errcheck // ins Gedächtnis
		_, _ = io.WriteString(eintrag, inhalt)    //nolint:errcheck // ins Gedächtnis
	}

	scheitert := func(*os.File, int64) (io.ReaderAt, int64, error) {
		return nil, 0, errors.New("die Entschlüsselung scheitert")
	}

	faelle := []struct {
		name  string
		datei *os.File
		lesen lesart
	}{
		{"die Hülle öffnet sich nicht", schreiben("hülle", func(*zip.Writer) {}), scheitert},
		{"kein Archiv", schreiben("kein", nil), offenGelesen},
		{"kein Manifest", schreiben("ohne", func(*zip.Writer) {}), offenGelesen},
		{"ein Teil fehlt", schreiben("fehlt", func(archiv *zip.Writer) {
			manifest(archiv, `{"teile":[{"name":"bestand.db","groesse":1,"sha256":"00"}]}`)
		}), offenGelesen},
		{"ein Teil hat eine andere Summe", schreiben("summe", func(archiv *zip.Writer) {
			eintrag, _ := archiv.Create("bestand.db") //nolint:errcheck // ins Gedächtnis
			_, _ = io.WriteString(eintrag, "x")       //nolint:errcheck // ins Gedächtnis
			manifest(archiv, `{"teile":[{"name":"bestand.db","groesse":1,"sha256":"00"}]}`)
		}), offenGelesen},
	}

	for _, fall := range faelle {
		if _, fehler := zuruecklesen(fall.datei, Manifestname, fall.lesen); fehler == nil {
			t.Errorf("%s: angenommen", fall.name)
		}
	}
}
