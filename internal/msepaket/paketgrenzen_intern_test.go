package msepaket

import (
	"archive/zip"
	"errors"
	"hash/crc32"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// **Der Kopf eines Zip-Eintrags wird nicht geglaubt** — FMT-06.AK-01
// (#1695).
//
// `fs.ReadFile` teilt Speicher in der Größe zu, die `Stat` nennt, **bevor**
// ein Byte gelesen ist (`io/fs/readfile.go`, Go 1.27.1), und für einen
// Zip-Eintrag ist das die Zahl aus seinem Kopf. `archive/zip` prüft sie erst
// beim Lesen (`reader.go`, Zeilen 301–308). Ein Eintrag von neunzehn Bytes,
// der drei Gibibyte behauptet, holte sich drei Gibibyte.
//
// **Nicht parallel**, und das ist Absicht: gemessen wird die Zuteilung des
// ganzen Prozesses, und parallele Tests warten, bis die folgenden fertig sind.
func TestEinZipKopfWirdNichtGeglaubt(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "luege.mse-set")
	inhalt := []byte("mse version: 2.0.0\n")

	datei, err := os.Create(pfad) //nolint:gosec // im Temporärverzeichnis des Tests
	if err != nil {
		t.Fatal(err)
	}

	schreiber := zip.NewWriter(datei)

	roh, err := schreiber.CreateRaw(&zip.FileHeader{
		Name:               "set",
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE(inhalt),
		CompressedSize64:   uint64(len(inhalt)),
		UncompressedSize64: 3 << 30,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, fehler := roh.Write(inhalt); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := datei.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	var vorher, nachher runtime.MemStats

	runtime.ReadMemStats(&vorher)

	paket, err := Oeffne(pfad)

	runtime.ReadMemStats(&nachher)

	if err == nil {
		_ = paket.Schliessen() //nolint:errcheck // nur aufgeräumt
	}

	if zugeteilt := nachher.TotalAlloc - vorher.TotalAlloc; zugeteilt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — der Kopf wurde geglaubt", zugeteilt>>20)
	}

	if !errors.Is(err, ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", err)
	}
}

// **Auch ein Baum, der über die Größe lügt, wird begrenzt.** `archive/zip`
// und eine Wurzel nennen ehrliche Größen; ein Baum, der weniger nennt, als er
// liefert, wird trotzdem an der Grenze abgeschnitten — die zweite Schranke in
// [leseBisZu] hält, was die erste nicht sieht. Gemessen an einem Kilobyte,
// denn dieselbe Rechnung trägt jede Grenze.
func TestEinBaumDerUeberDieGroesseLuegtWirdBegrenzt(t *testing.T) {
	t.Parallel()

	_, err := leseBisZu(luegenderBaum{}, "set", 1<<10)
	if !errors.Is(err, ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", err)
	}
}

// luegenderBaum hat eine Datei, die ein Byte nennt und ohne Ende liefert.
type luegenderBaum struct{}

func (luegenderBaum) Open(string) (fs.File, error) { return luegendeDatei{}, nil }

type luegendeDatei struct{}

func (luegendeDatei) Stat() (fs.FileInfo, error) { return kleineAngabe{}, nil }
func (luegendeDatei) Close() error               { return nil }

func (luegendeDatei) Read(puffer []byte) (int, error) {
	clear(puffer)

	return len(puffer), nil
}

// kleineAngabe behauptet ein Byte.
type kleineAngabe struct{ fs.FileInfo }

func (kleineAngabe) Size() int64 { return 1 }
