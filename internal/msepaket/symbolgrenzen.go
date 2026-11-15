package msepaket

// **Wie groß ein Symbol sein darf** — DOM-02.AK-04
// (#1697):
// „Schadhafte oder überkomplexe Vektor-/Schriftdateien sind begrenzt und
// erhalten verständliche Fehler; FMT-06 gilt auch hier."
//
// # Woher die Zahlen kommen
//
// **Gemessen am 2026-09-27** über die 21 `.mse-symbol`-Dateien des
// installierten `Full-Magic-Pack`: höchstens **73 Punkte**, **16 Teile**,
// **drei Ebenen**, **11.045 Byte** (`magic-comic.mse-style/b.mse-symbol`,
// `magic-etherium.mse-style/etherium page.mse-symbol`). Die Grenzen liegen
// jeweils mehr als das Fünfzigfache darüber — ein nachgezeichnetes Bild darf
// viele Punkte haben, und keine Vorlage soll an ihnen scheitern. **Was sie
// abfangen**, ist das Gegenteil: eine Datei, deren Rasterung den Rechner
// minutenlang beschäftigte, oder eine Verschachtelung, die den Leser in die
// Tiefe schickt.
//
// Portierungsmarke **new**: MSE2 begrenzt nichts davon — gesucht in
// `src/data/symbol.cpp` und `src/gui/symbol/` (Fassung f9d9356 vom
// 2021-07-06). Die Dateigrenze folgt dem Muster von FMT-06
// ([GrenzePaketdatei]): erst gegen die Angabe, dann gegen das Gelesene.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const (
	// GrenzeSymboldatei ist, wie groß eine `.mse-symbol` sein darf — 1 MiB.
	GrenzeSymboldatei int64 = 1 << 20

	// GrenzeSymbolteile ist, wie viele Teile ein Symbol insgesamt hat,
	// Verschachtelte mitgezählt.
	GrenzeSymbolteile = 4096

	// GrenzeSymbolpunkte ist, wie viele Punkte alle Formen zusammen haben.
	GrenzeSymbolpunkte = 65536

	// GrenzeSymboltiefe ist, wie tief Gruppen und Symmetrien geschachtelt
	// sein dürfen.
	GrenzeSymboltiefe = 32
)

// PruefeSymbolumfang weist ein Symbol ab, das eine der Grenzen überschreitet
// — mit einem Satz, der sagt, welche und wie weit.
func PruefeSymbolumfang(zeichnung Symbolzeichnung) error {
	var zaehlung symbolzaehlung

	zaehlung.gehe(zeichnung.Teile, 1)

	switch {
	case zaehlung.tiefe > GrenzeSymboltiefe:
		return zuviel(domain.SchluesselSymbolZuTief,
			"Das Symbol ist %d Ebenen tief geschachtelt; hier sind höchstens %d erlaubt.",
			zaehlung.tiefe, GrenzeSymboltiefe)

	case zaehlung.teile > GrenzeSymbolteile:
		return zuviel(domain.SchluesselSymbolZuVieleTeile,
			"Das Symbol hat %d Teile; hier sind höchstens %d erlaubt.",
			zaehlung.teile, GrenzeSymbolteile)

	case zaehlung.punkte > GrenzeSymbolpunkte:
		return zuviel(domain.SchluesselSymbolZuVielePunkte,
			"Das Symbol hat %d Punkte; hier sind höchstens %d erlaubt.",
			zaehlung.punkte, GrenzeSymbolpunkte)
	}

	// **Und geschrieben passt es in eine Datei, die sich wieder lesen lässt.**
	// Ein Punkt kostet geschrieben mindestens achtzig Byte, und unter der
	// Punktegrenze lagen damit fünf Megabyte — abgelegt und danach nicht mehr
	// zu öffnen. Gefunden beim Festlegen von `FuzzSymbolzeichnung` (QA-06b, #1967).
	if geschrieben := int64(len(SchreibeSymbolzeichnung(zeichnung))); geschrieben > GrenzeSymboldatei {
		return domain.NewFieldError("symbol", domain.SchluesselSymbolGespeichertZuGross,
			"Das Symbol wäre geschrieben %d Byte groß; eine Symboldatei hat höchstens %d.",
			geschrieben, GrenzeSymboldatei).MitWerten(map[string]string{"grenze": "1 MiB"})
	}

	return nil
}

// zuviel baut den Befund samt den Werten für den Satz im Katalog.
func zuviel(schluessel domain.Schluessel, vorlage string, anzahl, grenze int) error {
	return domain.NewFieldError("symbol", schluessel, vorlage, anzahl, grenze).
		MitWerten(map[string]string{"anzahl": strconv.Itoa(anzahl), "grenze": strconv.Itoa(grenze)})
}

// symbolzaehlung zählt Teile, Punkte und die tiefste Ebene.
type symbolzaehlung struct{ teile, punkte, tiefe int }

// gehe zählt `teile` auf Ebene `ebene` und steigt in ihre Kinder.
//
// **Es bricht nicht ab, wenn eine Grenze fällt** — der Satz soll die ganze
// Zahl nennen und nicht die erste, die zu viel war. Tiefer als die Datei
// selbst geht es nicht, und die ist begrenzt.
func (z *symbolzaehlung) gehe(teile []Symbolteil, ebene int) {
	if len(teile) > 0 {
		z.tiefe = max(z.tiefe, ebene)
	}

	for _, teil := range teile {
		z.teile++
		z.punkte += len(teil.Punkte)
		z.gehe(teil.Teile, ebene+1)
	}
}

// SymbolzeichnungAus liest eine `.mse-symbol` aus Bytes — begrenzt wie von
// der Platte. `name` steht in den Befunden.
func SymbolzeichnungAus(name string, roh []byte) (Symbolzeichnung, []Befund, error) {
	if int64(len(roh)) > GrenzeSymboldatei {
		return Symbolzeichnung{}, nil, fmt.Errorf("%w: %s hat %d Byte, höchstens %d",
			ErrZuGross, name, len(roh), GrenzeSymboldatei)
	}

	baum, err := Zerteile(name, roh)
	if err != nil {
		return Symbolzeichnung{}, nil, err
	}

	zeichnung, befunde := Symbolzeichnungsmodell(baum)
	if err := PruefeSymbolumfang(zeichnung); err != nil {
		return Symbolzeichnung{}, befunde, err
	}

	return zeichnung, befunde, nil
}

// leseSymboldatei liest eine `.mse-symbol` mit ihrer eigenen Grenze.
func leseSymboldatei(pfad string) ([]byte, error) {
	return leseBisZu(os.DirFS(filepath.Dir(pfad)), filepath.Base(pfad), GrenzeSymboldatei)
}
