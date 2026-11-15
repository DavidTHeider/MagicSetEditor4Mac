package msepaket_test

// **Ein Bild wird mit wahlfreiem Zugriff gelesen** — gefunden von
// `FuzzDekodiereBild` in QA-06b
// (#1967).
//
// Acht Byte — `II*\0` und ein Verzeichnis bei 808 Millionen — ließen
// [msepaket.DekodiereBild] zehn Megabyte zuteilen, und zwar schon beim
// Vermessen. `golang.org/x/image/tiff` springt im Archiv umher; bekommt es
// einen Strom ohne `ReadAt`, puffert es bis zur gesuchten Stelle, in
// Schritten von zehn Megabyte (`tiff/buffer.go`, `fillChunkSize`). Einen
// solchen Strom bekam es, weil `image.DecodeConfig` einen Leser ohne `Peek`
// in einen `bufio.Reader` steckt — und der kann kein `ReadAt`.
//
// **Nicht parallel**, wie [TestEinBildWirdVermessenBevorEsAusgepacktWird]:
// die Zählung des Speichers gilt für den ganzen Prozess.

import (
	"runtime"
	"testing"

	_ "golang.org/x/image/tiff" // dieselbe Registrierung wie das Programm

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinTIFFMitFernemVerzeichnisTeiltNichtZu(t *testing.T) {
	roh := []byte("II*\x000000")

	var vorher, nachher runtime.MemStats

	runtime.ReadMemStats(&vorher)

	_, _, fehler := msepaket.DekodiereBild(roh)

	runtime.ReadMemStats(&nachher)

	if fehler == nil {
		t.Fatal("acht Byte als Bild angenommen")
	}

	if zugeteilt := nachher.TotalAlloc - vorher.TotalAlloc; zugeteilt > 1<<20 {
		t.Errorf("%d Byte zugeteilt für eine Eingabe von %d", zugeteilt, len(roh))
	}
}
