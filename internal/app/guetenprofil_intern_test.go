//go:build fremdpakete

// **Wohin die Zeit beim Rastern geht** —
// #1029,
// Werkstücke Ω.4 und Ω.5.
//
// # Warum diese Messung vor den Güten steht
//
// Ω.4 verlangt vier Güten — Fast, Medium, High-Res, Ultra-High-Res —, und eine
// Güte ist eine Wahl darüber, **was weggelassen wird**. Was sich wegzulassen
// lohnt, sagt keine Meinung, sondern eine Messung: eine „schnelle" Güte, die
// den teuersten Schritt behält, ist keine.
//
// **Es gibt dafür keine Zahl.** `BenchmarkWegRastern` misst den ganzen Weg —
// 71 bis 78 ms je Karte —, aber nicht, welcher Teil davon wie viel trägt.
//
// # Was gemessen wird
//
// Der Weg zerfällt in drei Abschnitte, und jeder ist einzeln zu haben:
//
//  1. **Die Karte lesen** — Bestand, Feldwerte, Rohblöcke
//  2. **Den Plan bauen** — Skripte auswerten, Stil auflösen, Text setzen
//  3. **Rastern** — Ebenen malen, Glyphen zeichnen, Masken mischen
//
// Gemessen wird über **dieselbe** Karte, damit die drei Zahlen vergleichbar
// sind; und über mehrere Karten, weil eine einzelne nichts über die Klasse
// sagt (§10).
package app

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// ProfilKarten ist, wie viele Karten das Profil zieht.
//
// **Zwanzig**: genug, dass ein Ausreißer nicht die Aussage trägt, wenig genug,
// dass der Lauf in Sekunden bleibt.
const ProfilKarten = 20

// TestWohinDieZeitBeimRasternGeht misst die drei Abschnitte einzeln.
//
// **Ein Messlauf, kein Wächter** — wie `TestDieZeilenlageGegenMSE2`. Er sagt,
// wo die Zeit liegt; was daraus folgt, entscheidet Ω.4.
func TestWohinDieZeitBeimRasternGeht(t *testing.T) {
	anwendung := gestartetMit(t, func() string { return fremdpaketwurzelOderSkip(t) })

	geoeffnet := anwendung.SetOeffnen(context.Background(),
		kanonsatzpfad(t, "Black"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	zeilen := anwendung.Kartenzeilen(context.Background(), geoeffnet.SetID)
	if zeilen.Fehler != nil {
		t.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	if len(zeilen.Zeilen) < ProfilKarten {
		t.Skipf("nur %d Karten", len(zeilen.Zeilen))
	}

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		t.Skip("kein Rasterer")
	}

	var lesen, planen, rastern []time.Duration

	// **Jede zwanzigste Karte und nicht die ersten zwanzig.** Die ersten eines
	// Sets teilen oft Stil und Länge; über den Satz gezogen streut es.
	schritt := len(zeilen.Zeilen) / ProfilKarten

	for i := 0; i < ProfilKarten; i++ {
		karteID := zeilen.Zeilen[i*schritt].ID

		beginn := time.Now()

		einzeln := anwendung.Karte(context.Background(), karteID)
		if einzeln.Fehler != nil {
			t.Fatalf("Karte: %+v", einzeln.Fehler)
		}

		lesen = append(lesen, time.Since(beginn))

		beginn = time.Now()

		plan := planZurKarte(t, anwendung, karteID)

		planen = append(planen, time.Since(beginn))

		beginn = time.Now()

		if _, _, err := rasterer.Zeichne(context.Background(), plan, 1, render.GueteMittel); err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		rastern = append(rastern, time.Since(beginn))
	}

	meldeZeiten(t, "die Karte lesen", lesen)
	meldeZeiten(t, "den Plan bauen ", planen)
	meldeZeiten(t, "rastern        ", rastern)
}

// meldeZeiten schreibt Mitte, Mittel und Spanne.
//
// **Die Mitte zuerst.** Ein Mittelwert über zwanzig Messungen trägt jeden
// Ausreißer voll mit; die Mitte sagt, was der Normalfall kostet.
func meldeZeiten(t *testing.T, was string, zeiten []time.Duration) {
	t.Helper()

	if len(zeiten) == 0 {
		t.Logf("  %s: nichts gemessen", was)

		return
	}

	sortiert := append([]time.Duration(nil), zeiten...)
	sort.Slice(sortiert, func(i, j int) bool { return sortiert[i] < sortiert[j] })

	var summe time.Duration
	for _, eine := range sortiert {
		summe += eine
	}

	t.Logf("  %s: Mitte %8v   Mittel %8v   von %v bis %v   (%d Karten)",
		was, sortiert[len(sortiert)/2].Round(time.Microsecond),
		(summe / time.Duration(len(sortiert))).Round(time.Microsecond),
		sortiert[0].Round(time.Microsecond),
		sortiert[len(sortiert)-1].Round(time.Microsecond), len(sortiert))
}
