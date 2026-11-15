package bezug

// **Ein Schließfehler ist ein Installationsfehler** — BK-06c
// (#2132).
//
// `schreiben` verwarf den Fehler von `Close` — „der Schreibfehler kommt schon
// aus io.Copy". Das stimmt nicht: ein Dateisystem darf einen Fehler erst beim
// Schließen melden (ein voller Datenträger, ein Netzlaufwerk), und dann stand
// auf der Platte eine Datei, die aussah wie eine ganze, und der Bezug meldete
// Erfolg. **Das ist die einzige schreibende Schließung des Pakets**; die
// übrigen drei schließen Leser (gzip, den Strom, den Rumpf einer
// Fehlerantwort) und haben nichts mehr zu verlieren.

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

var (
	errSchliessen = errors.New("planmäßig: das Dateisystem meldet beim Schließen")
	errSchreiben  = errors.New("planmäßig: das Dateisystem meldet beim Schreiben")
)

// stolperdatei schreibt in eine echte Datei und scheitert, wo verlangt.
//
// **Die Datei ist ein Feld und nicht eingebettet**: eingebettet reichte sie
// `ReadFrom` durch, und `io.Copy` ginge an `Write` vorbei.
type stolperdatei struct {
	datei         *os.File
	beimSchreiben bool
}

func (d stolperdatei) Write(daten []byte) (int, error) {
	if d.beimSchreiben {
		return 0, errSchreiben
	}

	return d.datei.Write(daten)
}

func (d stolperdatei) Close() error {
	return errors.Join(d.datei.Close(), errSchliessen)
}

func mitStolperdatei(t *testing.T, beimSchreiben bool) {
	t.Helper()

	vorher := zielOeffnen
	t.Cleanup(func() { zielOeffnen = vorher })

	zielOeffnen = func(name string) (io.WriteCloser, error) {
		datei, err := vorher(name)
		if err != nil {
			return nil, err
		}

		echt, ok := datei.(*os.File)
		if !ok {
			t.Fatalf("zielOeffnen gab %T statt einer Datei", datei)
		}

		return stolperdatei{datei: echt, beimSchreiben: beimSchreiben}, nil
	}
}

func TestEinSchliessfehlerIstEinInstallationsfehler(t *testing.T) {
	mitStolperdatei(t, false)

	_, err := auspacken(bytes.NewReader(archivOhneVerdichtung(t, "probe/a.txt", 10)), "", t.TempDir())
	if !errors.Is(err, errSchliessen) {
		t.Fatalf("Kopieren geglückt, Schließen gescheitert — und der Bezug meldet %v", err)
	}
}

func TestSchreibUndSchliessfehlerBleibenBeide(t *testing.T) {
	mitStolperdatei(t, true)

	_, err := auspacken(bytes.NewReader(archivOhneVerdichtung(t, "probe/a.txt", 10)), "", t.TempDir())
	if !errors.Is(err, errSchreiben) || !errors.Is(err, errSchliessen) {
		t.Fatalf("beide Gründe gehören in den Fehler: %v", err)
	}
}
