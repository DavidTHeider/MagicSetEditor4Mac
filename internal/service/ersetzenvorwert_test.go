package service_test

// **Jeder Wechsel trägt den Text, aus dem er gerechnet ist** — #2534. Der
// Speicher prüft ihn in der Transaktion, die schreibt
// (`storage.Kartenschreiber.WerteSetzen`).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJederWechselTraegtSeinenVorwert(t *testing.T) {
	doppel := dreiKarten()

	if _, fehler := ersetzer(doppel).Ausfuehren(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"}); fehler != nil {
		t.Fatalf("Ausführen: %v", fehler)
	}

	vorher := map[int64]map[string]string{}
	for _, karte := range doppel.karten {
		vorher[karte.ID] = map[string]string{}
		for _, wert := range karte.Values {
			vorher[karte.ID][wert.Field] = wert.Text
		}
	}

	for _, wechsel := range doppel.geschrieben {
		if want := vorher[wechsel.KarteID][wechsel.Feld]; wechsel.Vorher != want {
			t.Errorf("Karte %d, %q: Vorwert %q, erwartet %q", wechsel.KarteID, wechsel.Feld, wechsel.Vorher, want)
		}
	}
}
