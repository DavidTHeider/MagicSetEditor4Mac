// **Die Paketpflege des unwilligen Doppels** — jede Naht scheitert
// (#1291, #1350, FMT-02 #1691).
//
// **Eine eigene Datei**, weil `unwilligeausfuhr_test.go` mit der eigenen Kopie
// über die Zeilengrenze aus §2 ging.

package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (u unwilliger) PaketInstallieren(context.Context, string) (rest.Paket, error) {
	return rest.Paket{}, u.fehler
}

func (u unwilliger) PaketAnlegen(context.Context, string, string, string, string) (
	rest.Paket, error,
) {
	return rest.Paket{}, u.fehler
}

func (u unwilliger) PaketKopieren(context.Context, string, string, string) (rest.Paket, error) {
	return rest.Paket{}, u.fehler
}

func (u unwilliger) PaketEntfernen(context.Context, string, string, string) error {
	return u.fehler
}

func (u unwilliger) PaketUebernehmen(context.Context, string) (rest.Uebernahme, error) {
	return rest.Uebernahme{}, u.fehler
}

func (unwilliger) MSE2Orte() []string { return []string{"/Anwendungen/Magic Set Editor/data"} }
