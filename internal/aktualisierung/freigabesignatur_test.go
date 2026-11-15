// **Was der Freigabeweg schreibt, kann der Ladeweg prüfen**
// (#892).
//
// # Die Naht, die sonst niemand prüft
//
// `scripts/release.sh` schreibt `pruefsummen.txt` und `pruefsummen.txt.asc`;
// `internal/aktualisierung` liest beide. Das sind **zwei Programme in zwei
// Sprachen**, und zwischen ihnen liegt genau die Sorte Naht, über die §10
// spricht: jede Seite ist für sich geprüft, und ob sie zusammenpassen, prüft
// niemand.
//
// **Ein synthetischer Fall belegt das nicht.** Die übrigen Tests signieren
// selbst — mit derselben Bibliothek, die danach prüft. Hier liegt eine
// Signatur von **`gpg`**, aus einem echten Freigabelauf, und sie wird gegen
// den **eingebetteten** Schlüssel gehalten.
//
// # Warum er überspringt, statt zu scheitern
//
// Die Dateien entstehen unter `build/freigabe/` und sind nicht eingecheckt
// (sie gehören zu einem Abbild von 224 MB). Wer sie nicht gebaut hat, hat
// nichts zu prüfen — und ein Test, der das als Fehler meldete, wäre auf jedem
// frischen Klon rot.
//
// **Und er sagt, dass er übersprungen hat.** Ein Lauf, der still grün ist,
// behauptet etwas über eine Prüfung, die nicht stattgefunden hat.
package aktualisierung_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
)

func TestDieEchteFreigabesignaturPasstZumEingebettetenSchluessel(t *testing.T) {
	wurzel := filepath.Join("..", "..", "build", "freigabe")

	summen, fehler := os.ReadFile(filepath.Join(wurzel, "pruefsummen.txt")) // #nosec G304 -- fester Pfad im Baum
	if fehler != nil {
		t.Skip("keine gebaute Freigabe unter build/freigabe — `./scripts/release.sh <fassung>`")
	}

	signatur, fehler := os.ReadFile(filepath.Join(wurzel, "pruefsummen.txt.asc")) // #nosec G304 -- fester Pfad im Baum
	if fehler != nil {
		t.Fatalf("die Freigabe hat Prüfsummen, aber keine Signatur: %v", fehler)
	}

	if err := aktualisierung.SignaturPruefen(summen, signatur); err != nil {
		t.Errorf("die echte Freigabesignatur prüft nicht: %v", err)
	}

	// **Die Gegenprobe**: eine geänderte Zeile fällt durch. Ohne sie stünde
	// nur fest, dass die Prüfung etwas zurückgibt.
	verfaelscht := append([]byte("jemand hat etwas dazugeschrieben\n"), summen...)

	if err := aktualisierung.SignaturPruefen(verfaelscht, signatur); err == nil {
		t.Error("eine verfälschte Prüfsummenliste kam durch")
	}
}
