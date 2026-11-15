//go:build fremdpakete

package msepaket_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Das Einfügemenü gegen die echten Symbolschriften**
// (#564).
//
// Der Leser ist gegen ein selbst geschriebenes Paket geprüft, und das belegt
// die Syntax — nicht die **Vielfalt**. `magic-mana-large` trägt 61 Einträge
// über vier Ebenen; ob ein Untermenü in einem Untermenü aufgeht, sagt kein
// Nachbau mit zwei Kindern (§10: die Reichweite einer Messung ist die
// Reichweite ihrer Aussage).
//
// **Läuft nicht in der Pipeline** und zählt nicht zur Abdeckung (§3).
func TestDasEinfuegemenueUeberFremdpakete(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := filepath.Glob(filepath.Join(wurzel, "*.mse-symbol-font"))
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if len(eintraege) == 0 {
		t.Fatalf("keine Symbolschrift unter %q", wurzel)
	}

	var mitMenue, codes, trenner, eigene, untermenues, ohneNamen int

	for _, pfad := range eintraege {
		paket, err := msepaket.Oeffne(pfad)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(pfad), err)

			continue
		}

		schrift, _ := paket.Symbolmodell()

		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("%s schließen: %v", filepath.Base(pfad), fehler)
		}

		if schrift.Einfuegemenue == nil {
			continue
		}

		mitMenue++

		zaehle(t, filepath.Base(pfad), schrift.Einfuegemenue,
			&codes, &trenner, &eigene, &untermenues, &ohneNamen)
	}

	t.Logf("%d von %d Symbolschriften haben ein Einfügemenü", mitMenue, len(eintraege))
	t.Logf("  %d Codes, %d Trenner, %d eigene, %d Untermenüs",
		codes, trenner, eigene, untermenues)

	// **Drei Codes ohne Namen, und das ist richtig so.**
	//
	// `space-psi-xerent` schreibt ein nacktes `item:` dorthin, wo ein
	// `type: line` hingehörte; `magic-embossedletters` und
	// `magic-mana-large-extra-udelude` schreiben `item:` mit einem
	// Leerzeichen dahinter. MSE liest beides genauso leer — `trim_left("  ")`
	// ist `""` (`src/util/io/reader.cpp`, Zeile 224, Fassung f9d9356 vom
	// 2021-07-06) — und baut daraus einen Menüeintrag ohne Bild, der nichts
	// einfügt.
	//
	// **Der Leser gibt wieder, was dasteht** (§8). Weggelassen wird ein
	// solcher Eintrag dort, wo die Palette entsteht: eine Schaltfläche ohne
	// Aufschrift und ohne Wirkung ist keine.
	t.Logf("  %d Codes ohne Namen — Paketeigenheit, siehe Kommentar", ohneNamen)

	// **Die Gegenprobe zur Messung**: fände der Leser gar nichts, stünden hier
	// vier Nullen, und der Lauf wäre grün.
	if codes == 0 || trenner == 0 || eigene == 0 || untermenues == 0 {
		t.Errorf("eine der vier Arten kam über das ganze Pack nicht vor")
	}
}

// zaehle geht das Menü durch und prüft jeden Eintrag gegen seine Art.
func zaehle(
	t *testing.T, paket string, menue []msepaket.Menueeintrag,
	codes, trenner, eigene, untermenues, ohneNamen *int,
) {
	t.Helper()

	for i := range menue {
		eintrag := &menue[i]

		switch eintrag.Art {
		case msepaket.MenueCode:
			*codes++

			if strings.TrimSpace(eintrag.Name) == "" {
				*ohneNamen++
			}

		case msepaket.MenueTrenner:
			*trenner++

			if len(eintrag.Kinder) > 0 {
				t.Errorf("%s: ein Trenner mit Kindern bei %+v", paket, eintrag.Stelle)
			}

		case msepaket.MenueEigen:
			*eigene++

		case msepaket.MenueUntermenue:
			*untermenues++

			if len(eintrag.Kinder) == 0 {
				t.Errorf("%s: ein Untermenü ohne Kinder bei %+v", paket, eintrag.Stelle)
			}

			zaehle(t, paket, eintrag.Kinder, codes, trenner, eigene, untermenues, ohneNamen)

		default:
			t.Errorf("%s: unbekannte Art %q bei %+v", paket, eintrag.Art, eintrag.Stelle)
		}
	}
}
