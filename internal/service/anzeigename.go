package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Anzeigename ist, was über einem Paket steht — UX-05, S07.NAT-02
// (#1677).
//
// **ported (§15)**: der Kurzname, großgeschrieben (`PackageList::drawItem`,
// `src/gui/control/package_list.cpp`, Zeilen 39–43, Fassung f9d9356 vom
// 2021-07-06); ohne ihn der Name der Datei ohne Endung, wie
// `Packaged::validate` ihn einsetzt (`src/util/io/package.cpp`, Zeilen
// 622–626). **Den vollen Namen darunter zeigt MSE2 auch** — der Bestand führt
// ihn nicht (Migration 0012 nahm nur den Kurznamen auf).
func Anzeigename(paket domain.Paket) string {
	if paket.Kurzname != "" {
		return msescript.Grossgeschrieben(paket.Kurzname)
	}

	return msescript.Grossgeschrieben(paket.Name)
}
