//go:build fremdpakete

// **Was am Rastern teuer ist, je Ebenenart** —
// #1029,
// Werkstück Ω.4.
//
// # Warum diese Messung vor dem Schnitt steht
//
// Das Profil des ganzen Weges sagt: der Planbau trägt 55 Prozent, das Rastern
// 45. Eine Güte darf den Plan nicht anfassen — Layout und Umbruch bleiben
// gleich —, also muss sie aus diesen 45 Prozent schöpfen.
//
// **Aus welchem Teil davon, sagt keine Meinung.** Eine „schnelle" Güte, die
// den teuersten Schritt behält, ist keine; und eine, die den billigsten
// weglässt, spart nichts und verliert etwas.
//
// # Wie gemessen wird
//
// Der Plan wird in seine Ebenen zerlegt und je **Art** ein eigener Plan
// gezeichnet. Die Summe der Teile ist nicht das Ganze — jeder Lauf legt ein
// eigenes Ziel an —, aber das Verhältnis der Teile ist die Frage.
//
// # Was sie nicht sagt
//
// **Nicht, was eine Güte weglassen darf.** Eine teure Ebenenart kann die sein,
// die die Karte ausmacht; die Zahl sagt, was es kostet, nicht was es wert ist.
package app

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// TestWasAmRasternTeuerIst misst die Ebenenarten einzeln.
func TestWasAmRasternTeuerIst(t *testing.T) {
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

	jeArt := map[render.Ebenenart][]time.Duration{}
	ebenen := map[render.Ebenenart]int{}
	schritt := len(zeilen.Zeilen) / ProfilKarten

	for i := range ProfilKarten {
		plan := planZurKarte(t, anwendung, zeilen.Zeilen[i*schritt].ID)

		for art, teilplan := range plaeneJeArt(plan) {
			ebenen[art] += len(teilplan.Ebenen)

			beginn := time.Now()

			if _, _, err := rasterer.Zeichne(context.Background(), teilplan, 1, render.GueteMittel); err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			jeArt[art] = append(jeArt[art], time.Since(beginn))
		}
	}

	for _, art := range sortierteArten(jeArt) {
		meldeZeiten(t, string(art)+" ", jeArt[art])
		t.Logf("      %d Ebenen dieser Art über %d Karten",
			ebenen[art], len(jeArt[art]))
	}
}

// plaeneJeArt zerlegt einen Plan in je einen Plan pro Ebenenart.
//
// **Der Rahmen bleibt derselbe.** Ein Teilplan über einen kleineren Kasten
// zeichnete weniger und sagte etwas über den Kasten statt über die Art.
func plaeneJeArt(plan *render.Plan) map[render.Ebenenart]*render.Plan {
	aus := map[render.Ebenenart]*render.Plan{}

	for _, ebene := range plan.Ebenen {
		teil, da := aus[ebene.Art]
		if !da {
			kopie := *plan
			kopie.Ebenen = nil
			teil = &kopie
			aus[ebene.Art] = teil
		}

		teil.Ebenen = append(teil.Ebenen, ebene)
	}

	return aus
}

// sortierteArten gibt die Arten nach ihrer mittleren Dauer, teuerste zuerst.
func sortierteArten(jeArt map[render.Ebenenart][]time.Duration) []render.Ebenenart {
	arten := make([]render.Ebenenart, 0, len(jeArt))
	for art := range jeArt {
		arten = append(arten, art)
	}

	sort.Slice(arten, func(i, j int) bool {
		return mitte(jeArt[arten[i]]) > mitte(jeArt[arten[j]])
	})

	return arten
}

// mitte gibt den Median einer Messreihe.
func mitte(zeiten []time.Duration) time.Duration {
	if len(zeiten) == 0 {
		return 0
	}

	sortiert := append([]time.Duration(nil), zeiten...)
	sort.Slice(sortiert, func(i, j int) bool { return sortiert[i] < sortiert[j] })

	return sortiert[len(sortiert)/2]
}
