// **Ein Dokument, das sich selbst widerspricht, wird an der falschen Stelle
// geglaubt** — PLAY-06.AK-05
// (#1714).
//
// # Der Befund
//
// `docs/TISCHLEISTUNG.md` führt in der Tabelle der nichtfunktionalen Punkte
// eine Zeile **Reduzierte Bewegung** und schreibt dazu:
//
//	Es gibt **keine** Animation: der Tisch fährt PixiJS ohne Ticker und
//	zeichnet auf Zuruf. Damit gibt es nichts zu reduzieren — und wer eine
//	Animation nachrüstet, muss `prefers-reduced-motion` mitbringen.
//
// **Es hat jemand nachgerüstet.** `frontend/src/tisch/bewegung.ts` führt seit
// Werkstück G4 eine `Bewegungsdauer` von 220 ms, und `pixibewegung.ts` treibt
// dafür den Ticker an.
//
// **Und er hat mitgebracht, was der Satz verlangt**: `bewegungErwuenscht`
// fragt `prefers-reduced-motion` und gibt der Systemeinstellung die stärkere
// Stimme. Der **Code ist also richtig** — falsch ist das Dokument, und es ist
// die gefährlichere Hälfte: wer die Zeile liest, hält den Punkt für „trivial
// erfüllt" und prüft ihn nicht.
//
// # Was er hält
//
// Solange `bewegung.ts` eine `Bewegungsdauer` führt, behauptet
// `TISCHLEISTUNG.md` nicht, es gebe keine Animation.
//
// # Was er **nicht** hält
//
//   - **Nicht, dass die Einstellung wirkt.** Das hält
//     `frontend/src/tisch/bewegung.test.ts` an `bewegungErwuenscht`, und der
//     Weg von dort bis zur Leinwand hält `tischflaeche`.
//   - **Nicht den umgekehrten Fall.** Nähme jemand die Bewegung wieder heraus,
//     dürfte der alte Satz zurück — dieser Wächter verlangt ihn dann nicht.
//     Ein Wächter, der beide Richtungen erzwänge, schriebe den Text vor statt
//     den Widerspruch fernzuhalten.
package werkzeuge

import (
	"path/filepath"
	"strings"
	"testing"
)

// keineAnimation sind die Wendungen, mit denen das Dokument die Abwesenheit
// behauptet hat.
//
// **Zwei und nicht eine**: der Satz steht einmal mit Auszeichnungssternen und
// einmal ohne, und wer nur die eine Gestalt suchte, ließe die andere durch.
var keineAnimation = []string{"keine** Animation", "keine Animation"}

// TestDasLeistungsdokumentLeugnetDieBewegungNicht hält PLAY-06.AK-05.
func TestDasLeistungsdokumentLeugnetDieBewegungNicht(t *testing.T) {
	t.Parallel()

	bewegung := lies(t, filepath.Join("frontend", "src", "tisch", "bewegung.ts"))
	if !strings.Contains(bewegung, "export const Bewegungsdauer") {
		t.Skip("es gibt keine Bewegung mehr — dann darf das Dokument sie leugnen")
	}

	dokument := lies(t, filepath.Join("docs", "TISCHLEISTUNG.md"))

	for _, wendung := range keineAnimation {
		if strings.Contains(dokument, wendung) {
			t.Errorf("docs/TISCHLEISTUNG.md schreibt %q, während "+
				"`bewegung.ts` eine Bewegungsdauer führt — ein Dokument, das "+
				"sich selbst widerspricht, wird an der falschen Stelle "+
				"geglaubt (PLAY-06.AK-05)", wendung)
		}
	}
}
