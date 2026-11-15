package austausch

// **Die Quarantäne misst, was ausgepackt wird** — QA-06b
// (#1967),
// beim Festlegen der Invariante von `FuzzUmschlag` gefunden.
//
// [pruefeManifest] hält jedes Objekt gegen [HoechstensObjekt] und die Summe
// gegen [HoechstensGesamt] — über `groesse` aus dem **inneren** Manifest.
// Gegen die Bytes gehalten wurde diese Zahl nirgends: `pruefeSummen` prüfte
// die Prüfsumme und nicht die Länge, und `sicherung` packt nach dem
// **äußeren** Manifest aus, dessen Grenzen 64 GB erlauben. Ein unterschriebener
// Umschlag, der `groesse: 1` sagte, brachte damit beliebig viel durch die
// Quarantäne — und der Kopf von `quarantaene.go` sagt ausdrücklich, dass eine
// Unterschrift hier nichts erspart.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

func TestEineFalscheGroesseImManifestWirdVorDemAuspackenAbgewiesen(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	inhalt := make([]byte, 4096)
	summe := sha256.Sum256(inhalt)

	for name, groesse := range map[string]int{"kleiner": 1, "groesser": 8192} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			roh := []byte(`{"fassung":1,"urheber":"` + string(ich.Kennung()) + `","objekte":[` +
				`{"art":"karte","name":"a.json","sha256":"` + hex.EncodeToString(summe[:]) +
				`","groesse":` + strconv.Itoa(groesse) + `}]}`)

			rahmen := t.TempDir()
			ort := filepath.Join(rahmen, "probe"+Endung)

			if packfehler := sicherung.PackenAls(ort, aeusseresManifest, []sicherung.Teil{
				{Name: "a.json", Inhalt: inhalt},
				{Name: Manifestname, Inhalt: roh},
				{Name: Unterschriftsname, Inhalt: ich.Signiere(roh)},
			}, sicherung.Angaben{}); packfehler != nil {
				t.Fatalf("packen: %v", packfehler)
			}

			_, lesefehler := Lies(ort, filepath.Join(rahmen, "ziel"))
			if !errors.Is(lesefehler, domain.ErrUmschlagInhalt) {
				t.Errorf("4096 Byte, das Manifest sagt %d: %v", groesse, lesefehler)
			}

			if _, da := os.Stat(filepath.Join(rahmen, "ziel", "inhalt", "a.json")); da == nil {
				t.Error("ausgepackt, bevor die Größe geprüft war")
			}
		})
	}
}
