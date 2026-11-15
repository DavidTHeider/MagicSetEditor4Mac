package schriften

import (
	"errors"
	"fmt"
	"os"
)

// ErrSchriftZuGross meldet eine Schriftdatei über [GrenzeSchriftdatei].
var ErrSchriftZuGross = errors.New("die Schriftdatei ist größer, als eine Schrift sein darf")

// GrenzeSchriftdatei ist, wie groß eine Schriftdatei höchstens sein darf —
// FMT-06.AK-01 (#1695).
//
// **Gemessen** (2026-09-23): die größte Schrift dieses Rechners ist
// `Apple Color Emoji.ttc` mit 192 MB, die größte des vollen Packs 1,6 MB.
// 256 MiB lassen jede davon zu und halten eine Datei auf, die jemand als
// Schrift in ein Paket legt und die keine ist.
const GrenzeSchriftdatei int64 = 256 << 20

// schriftLesen liest eine Schriftdatei — **die eine Stelle**, an der dieses
// Paket eine liest, und erst nachdem ihre Größe unter der Grenze liegt.
//
// Der Pfad kommt aus der Verzeichnisliste eines Ortes dieser Suche oder aus
// dem Index eines Pakets, nicht aus einer Eingabe (#nosec G304 unten).
func schriftLesen(pfad string) ([]byte, error) {
	angaben, err := os.Stat(pfad)
	if err != nil {
		return nil, err
	}

	if angaben.Size() > GrenzeSchriftdatei {
		return nil, fmt.Errorf("%w: %s hat %d Byte, höchstens %d", ErrSchriftZuGross, pfad,
			angaben.Size(), GrenzeSchriftdatei)
	}

	return os.ReadFile(pfad) // #nosec G304 -- siehe darüber
}

// schriftOeffnen öffnet eine Schriftdatei zum **teilweisen** Lesen — unter
// derselben Grenze wie [schriftLesen], geprüft am geöffneten Griff.
//
// **Für ihre Namen braucht eine Schrift ihre Tabellen, nicht ihre Glyphen**
// (E03, #2240).
func schriftOeffnen(pfad string) (*os.File, error) {
	datei, err := os.Open(pfad) // #nosec G304 -- siehe [schriftLesen]
	if err != nil {
		return nil, err
	}

	angaben, err := datei.Stat()
	if err == nil && angaben.Size() > GrenzeSchriftdatei {
		err = fmt.Errorf("%w: %s hat %d Byte, höchstens %d", ErrSchriftZuGross, pfad,
			angaben.Size(), GrenzeSchriftdatei)
	}

	if err != nil {
		_ = datei.Close() //nolint:errcheck // der Fehler davor ist der, der zählt

		return nil, err
	}

	return datei, nil
}
