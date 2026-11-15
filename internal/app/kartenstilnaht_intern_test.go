//go:build fremdpakete

// Der Wächter über die Naht, die #395
// gefehlt hat: **eine Karte wird über ihren eigenen Stil gezeichnet.**
//
//	const StyleSheet& Set::stylesheetFor(const CardP& card) {
//	  if (card && card->stylesheet) return *card->stylesheet;
//	  else                          return *stylesheet;
//	}
//
// `src/data/set.cpp`, Zeilen 90–93, Fassung f9d9356 vom 2021-07-06 (§15).
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Die Karten eines kanonischen Sets tragen ihren eigenen Stil.**
//
// In `Archenemy.mse-set` steht `stylesheet: archenemy` **70-mal an den Karten**
// und `m15-altered` einmal am Set. Gezeichnet werden muss `magic-archenemy` —
// und dass es klappt, sagt hier die **Kartengröße**:
//
//	magic-archenemy.mse-style:28   card width: 435, card height: 621
//	magic-m15-altered.mse-style:32 card width: 375, card height: 523
//
// **Eine Größe und kein Bildvergleich**: sie unterscheidet die beiden Vorlagen
// eindeutig und hängt an nichts anderem (§10 — die Reichweite einer Messung).
// **Abgelesen, nicht geraten**: der erste Anlauf prüfte „quer statt hoch" und
// lag falsch — `magic-archenemy` ist hoch, nur größer.
func TestDerKartenstilDurchDieNaht(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	if _, err := os.Stat(filepath.Join(wurzel, "magic-archenemy.mse-style")); err != nil { //nolint:gosec // ein Pfad aus MSE_FREMDPAKETE oder $HOME — dieser Lauf misst gegen bezogene Pakete (§3)
		t.Skip("magic-archenemy liegt in diesem Paketverzeichnis nicht")
	}

	// **Über `kanonverzeichnis` und nicht über `HOME`** — hier stand der Pfad
	// aus `HOME` zusammengesetzt, und unter dem eigenen Zuhause des Testlaufs
	// lag dort nichts: der Lauf sprang in jedem `make fremdpakete` und hieß
	// `ok`. Sichtbar gemacht hat es erst der Beleg (QA-01, #1733) — dieselbe
	// Klasse wie #1656, an einer Stelle, die der Fix dort nicht erreichte.
	ort := filepath.Join(kanonverzeichnis(), "Archenemy.mse-set")
	if _, err := os.Stat(ort); err != nil { //nolint:gosec // ein Pfad aus MSE_KANON oder $HOME — dieser Lauf misst gegen bezogene Pakete (§3)
		t.Skip("die kanonischen Sets liegen nicht vor")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	geoeffnet := anwendung.SetOeffnen(t.Context(), ort)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	bild := bildAus(t, anwendung, vorschau.Vorschau.Bild)

	t.Logf("%dx%d, %d Befunde", bild.Bounds().Dx(), bild.Bounds().Dy(),
		len(vorschau.Vorschau.Befunde))

	for _, meldung := range vorschau.Vorschau.Befunde {
		if strings.Contains(meldung, "Stilpaket") || strings.Contains(meldung, "Paketverzeichnis") {
			t.Errorf("der Stil der Karte wurde nicht gefunden: %s", meldung)
		}
	}

	if bild.Bounds().Dx() != 435 || bild.Bounds().Dy() != 621 {
		t.Errorf("die Vorschau ist %dx%d, erwartet 435x621 aus `magic-archenemy` — "+
			"375x523 wäre der Stil des **Sets**",
			bild.Bounds().Dx(), bild.Bounds().Dy())
	}
}
