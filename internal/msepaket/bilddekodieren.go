package msepaket

import (
	"bytes"
	"fmt"
	"image"
	"io"
)

// DekodiereBild packt ein Bild aus — **erst vermessen, dann ausgepackt**
// (FMT-06.AK-01, #1695).
//
// Ein PNG von ein paar hundert Bytes darf 40.000 × 40.000 Bildpunkte
// behaupten, und `image.Decode` teilt dafür 6,4 GB zu, bevor es merkt, dass
// die Daten fehlen. `image.DecodeConfig` liest nur den Kopf; was dort über
// [GrenzeBildseite] oder [GrenzeBildpunkte] liegt, wird nicht ausgepackt.
//
// **Die eine Stelle, die auspackt** — `internal/werkzeuge/bilddekodieren_test.go`
// weist `image.Decode` und seine Verwandten überall sonst ab. Die Formate sind
// die, die das Programm registriert hat; ein neues läuft von selbst durch
// diese Prüfung.
func DekodiereBild(roh []byte) (image.Image, string, error) {
	mass, art, err := image.DecodeConfig(bildbytes(roh))
	if err != nil {
		return nil, "", err
	}

	if fehler := bildmassPruefen(mass); fehler != nil {
		return nil, art, fehler
	}

	if fehler := ausgepacktZuTeuer(mass, art); fehler != nil {
		return nil, art, fehler
	}

	return image.Decode(bildbytes(roh))
}

// bildleser gibt die Bytes mit `Peek` **und** `ReadAt` her.
//
// **Ohne `Peek` steckt `image` den Leser in einen `bufio.Reader`, und der
// kann kein `ReadAt`.** `golang.org/x/image/tiff` springt im Archiv umher
// und puffert einen solchen Strom bis zur gesuchten Stelle, je Schritt zehn
// Megabyte (`tiff/buffer.go`) — acht Byte mit einem Verweis ins Ferne
// kosteten so zehn Megabyte, gefunden von `FuzzDekodiereBild` (QA-06b, #1967).
type bildleser struct {
	*bytes.Reader

	roh []byte
}

func bildbytes(roh []byte) bildleser {
	return bildleser{Reader: bytes.NewReader(roh), roh: roh}
}

// Peek gibt die nächsten `n` Bytes, ohne weiterzulesen — wie `bufio.Reader`.
func (b bildleser) Peek(n int) ([]byte, error) {
	stelle := len(b.roh) - b.Len()
	if n > b.Len() {
		return b.roh[stelle:], io.EOF
	}

	return b.roh[stelle : stelle+n], nil
}

// DekodiereBildAus liest ein Bild aus einer Quelle — höchstens
// [GrenzePaketdatei] Byte — und packt es aus wie [DekodiereBild].
//
// **Für Dateien, die jemand wählt**: ein Bild vom Datenträger ist so fremd wie
// eines aus einem Paket, und eine Datei von zwei Gigabyte wäre sonst ganz im
// Speicher, bevor ihr Kopf gelesen ist.
func DekodiereBildAus(quelle io.Reader) (image.Image, string, error) {
	return dekodiereBildBisZu(quelle, GrenzePaketdatei)
}

// dekodiereBildBisZu ist [DekodiereBildAus] mit einer genannten Grenze —
// dieselbe Naht wie [leseBisZu].
func dekodiereBildBisZu(quelle io.Reader, grenze int64) (image.Image, string, error) {
	roh, err := io.ReadAll(io.LimitReader(quelle, grenze+1))
	if err != nil {
		return nil, "", err
	}

	if int64(len(roh)) > grenze {
		return nil, "", fmt.Errorf("%w: die Datei ist größer als %d Byte", ErrZuGross, grenze)
	}

	return DekodiereBild(roh)
}

// bildmassPruefen weist Maße über der Grenze ab — je Seite und zusammen.
func bildmassPruefen(mass image.Config) error {
	if mass.Width > GrenzeBildseite || mass.Height > GrenzeBildseite ||
		int64(mass.Width)*int64(mass.Height) > GrenzeBildpunkte {
		return bildZuGross("%d × %d Bildpunkte, höchstens %d je Seite und %d zusammen",
			mass.Width, mass.Height, GrenzeBildseite, GrenzeBildpunkte)
	}

	return nil
}
