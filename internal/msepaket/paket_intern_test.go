// Die Fehlerpfade des Paketlesers, die von außen nicht erreichbar sind.
package msepaket

import (
	"errors"
	"io/fs"
	"testing"
)

var errNichtSchliessbar = errors.New("dieses Archiv lässt sich planmäßig nicht schließen")

type sperrschliesser struct{}

func (sperrschliesser) Close() error { return errNichtSchliessbar }

// Ein Archiv, das sich nicht schließen lässt, wird gemeldet.
//
// Auf einer gesunden Platte kommt der Fall nicht zustande; im Betrieb schon —
// ein entferntes Dateisystem, eine ausgeworfene Platte. Ein Fehlerpfad, den kein
// Test betritt, ist eine Behauptung darüber, was im Ernstfall geschieht.
func TestSchliessenMeldetEinSperrigesArchiv(t *testing.T) {
	paket := &Paket{Pfad: "probe.mse-set", schliesser: sperrschliesser{}}

	if err := paket.Schliessen(); !errors.Is(err, errNichtSchliessbar) {
		t.Errorf("Fehler %v, erwartet war der Schließfehler", err)
	}
}

// **Ein unlesbares Paket meldet sich, statt eine kurze Liste zu geben.**
//
// Paketintern und mit einer eigenen Dateiquelle: ein Verzeichnis ohne
// Leserechte wäre der naheliegende Weg, aber `chmod 0700` auf ein Verzeichnis
// stößt gosec an (G302 verlangt 0600 oder weniger — für ein Verzeichnis ist das
// unbrauchbar). Eine Quelle, die beim Gehen scheitert, ist ohnehin deutlicher:
// sie sagt, worum es geht, statt es über Rechte nachzustellen.
func TestDateienMeldetEineUnlesbareQuelle(t *testing.T) {
	paket := &Paket{KurzerName: "kaputt", dateien: stolperquelle{}}

	namen, fehler := paket.Dateien()
	if fehler == nil {
		t.Errorf("eine unlesbare Quelle wurde übergangen: %v", namen)
	}
}

// stolperquelle ist eine Dateiquelle, die sich nicht durchgehen lässt.
type stolperquelle struct{}

func (stolperquelle) Open(string) (fs.File, error) {
	return nil, errQuelleKaputt
}

var errQuelleKaputt = errors.New("diese Quelle lässt sich nicht lesen")
