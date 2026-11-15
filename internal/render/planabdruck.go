package render

// **Der Abdruck eines Zeichenplans** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wofür es ihn gibt
//
// **Damit ein gerastertes Bild wiedererkannt wird.** Ein Spieltisch zeigt
// dieselbe Karte siebenmal; sie siebenmal zu zeichnen kostet siebenmal so
// viel. Was zwei Ansichten gleich macht, ist nicht die Kartenkennung — es ist
// **das, was gezeichnet wird**.
//
// # Warum die Kartenkennung nicht genügt
//
// **Weil sie weder portabel noch inhaltsadressiert ist** (#1459 wörtlich).
// Zwei Bestände geben derselben Karte verschiedene Zeilennummern, und
// **dieselbe** Zeilennummer trägt nach einer Änderung etwas anderes. Ein
// Cache über ihr zeigte das Bild von vorhin.
//
// # Was er umfasst und was ausdrücklich nicht
//
// **Nur, was gezeichnet wird**: Maße, Auflösung und die Ebenen in ihrer
// Reihenfolge. [Plan.Feldflaechen] und [Plan.Vorgaben] sind die Auskunft an
// die Oberfläche — der Rasterer liest sie nicht. Stünden sie darin, verwürfe
// eine geänderte Herleitung ein Bild, das Pixel für Pixel dasselbe wäre.
//
// # Und was er **nicht weiß**
//
// **Was in den Dateien steht, die er nennt.** Ein Bildpfad ist ein Pfad, ein
// Schriftname ein Name; ändert sich die Datei dahinter, ändert sich der
// Abdruck nicht. Deshalb gibt es [Bildpfade] und [Schriftnamen] daneben: wer
// einen Schlüssel baut, hängt ihre Inhalte an. Ein Abdruck, der so täte, als
// deckte er sie mit ab, wäre eine Zusage, die ihr Rumpf nicht hält (§10).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
)

// Fassung ist die Fassung dieses Rasterers.
//
// # Warum eine Zahl von Hand und keine gerechnete
//
// **Weil sich ein Rasterer ändert, ohne dass eine Datei es verrät.** Eine
// andere Glättung, eine andere Reihenfolge beim Mischen, ein behobener Fehler
// im Satz — jedes davon zeichnet dieselbe Vorlage anders, und jedes gehört in
// den Schlüssel. Wer sie vergäße, zeigte Bilder von gestern.
//
// **Gehalten wird sie von einem Wächter**, nicht von der Umsicht:
// `internal/werkzeuge/rasterfassung_test.go` hält den Quelltext dieses Pakets
// gegen einen festgehaltenen Abdruck — wer ihn ändert und die Fassung stehen
// lässt, macht die Pipeline rot.
//
// **2 seit QA-03** (#1656, #1805, #1806): Kerning mit Klassenpaaren, MSEs
// vierfacher Textweg, Ligaturen und Quartz' Schriftglättung — dieselbe
// Vorlage steht mit jedem davon anders da.
//
// **3 seit dem 2026-09-24**: ein ausgewichenes Zeichen wird mit seiner
// Ersatzschrift gemessen und steht in voller Breite da, und was über die
// Zeile ragt — die Punkte eines Ä bei Helvetica —, fällt nicht mehr weg
// (`zeichenrand_test.go`). Bis die Fassung stieg, zeigte die Ablage die
// alten Bilder: gesehen in der WebKit-Probe, „AOU" nach dem Beheben.
//
// **4 seit dem 2026-09-26**: der Ebenenspeicher zählt und lässt sich leeren
// (OPS-08, #1725). Kein Punkt fällt anders; die Fassung steigt, weil der
// Wächter jede Änderung am Rasterer verlangt — und das Fach der Spielbilder
// liegt im Arbeitsspeicher, der Preis ist also ein Neuzeichnen je Sitzung.
//
// **5 seit dem 2026-10-03**: jede Rasterung fragt eine Pforte nach einem
// Platz (#2020, `pforte.go`). Auch hier fällt kein Punkt anders, und auch hier
// ist der Preis ein Neuzeichnen je Sitzung.
//
// **6 seit dem 2026-10-11**: Text wirft seinen Schatten, und ein leeres
// Bildfeld zeigt das Schachbrett (F02, #2244). Ein Plan ohne beides zeichnet
// wie vorher; die Fassung steigt, damit kein Bild von gestern stehen bleibt.
const Fassung = "7"

// Planabdruck gibt den Abdruck dessen, was dieser Plan zeichnet.
func Planabdruck(plan Plan) string {
	// **Über JSON und nicht über einen eigenen Schreiber.** `encoding/json`
	// schreibt Struct-Felder in Deklarationsreihenfolge und Map-Schlüssel
	// **sortiert** — derselbe Grund wie bei `domain.Ereignis.Abdruck`, und
	// mehr braucht ein Abdruck nicht.
	//
	// **Der Fehlerweg fehlt, und das ist nachgelesen**: `json.Marshal`
	// scheitert an Kanälen, Funktionen und Zyklen. Ein [Plan] trägt Zahlen,
	// Zeichenketten, Verbünde und Zeiger auf Verbünde — nichts davon. Ein
	// Zweig dafür wäre einer, den kein Lauf betreten kann (§3); dieselbe
	// Überlegung steht an `storage.alsPartiejson`.
	roh, _ := json.Marshal(gezeichnetes{ //nolint:errcheck // siehe darüber
		Breite: plan.Breite, Hoehe: plan.Hoehe, DPI: plan.DPI,
		Ebenen: plan.Ebenen,
	})

	summe := sha256.Sum256(roh)

	return hex.EncodeToString(summe[:])
}

// gezeichnetes ist der Teil eines Plans, der auf dem Bild landet.
//
// **Ein eigener Verbund und kein Leeren der Felder.** Wer stattdessen eine
// Kopie mit geleerten Feldern hashte, nähme beim nächsten neuen Feld
// stillschweigend etwas dazu — hier steht ausdrücklich da, **was** zählt.
// Derselbe Schnitt wie bei `domain.abdruckbaresEreignis`.
type gezeichnetes struct {
	Breite float64 `json:"breite"`
	Hoehe  float64 `json:"hoehe"`
	DPI    int     `json:"dpi"`
	Ebenen []Ebene `json:"ebenen"`
}

// Bildpfade gibt die Paketpfade aller Bilder dieses Plans.
//
// **Geordnet und ohne Dubletten**: wer daraus einen Schlüssel baut, braucht
// bei gleichem Plan dieselbe Reihenfolge — und ein Rahmen, der zweimal
// vorkommt, ist eine Datei.
func Bildpfade(plan Plan) []string {
	gesehen := map[string]bool{}

	for _, ebene := range plan.Ebenen {
		sammleBildpfade(ebene.Bild, gesehen)
		sammleBildpfade(ebene.Maske, gesehen)
	}

	return geordnet(gesehen)
}

// sammleBildpfade läuft einen Bildbaum ab.
func sammleBildpfade(baum Bildbaum, gesehen map[string]bool) {
	if baum.Pfad != "" {
		gesehen[baum.Pfad] = true
	}

	for _, kind := range baum.Kinder {
		sammleBildpfade(kind.Baum, gesehen)
	}
}

// Schriftnamen gibt die Namen aller Schriften dieses Plans.
//
// **Aufrecht und kursiv**, denn beide werden gesucht und beide zeichnen. Was
// das Betriebssystem daraus macht, weiß [Schriftquelle] — und **welche Datei**
// es findet, gehört in den Schlüssel und nicht hierher.
func Schriftnamen(plan Plan) []string {
	gesehen := map[string]bool{}

	for _, ebene := range plan.Ebenen {
		for _, name := range []string{
			ebene.Satz.Schrift, ebene.Satz.Kursivschrift,
		} {
			if name != "" {
				gesehen[name] = true
			}
		}
	}

	return geordnet(gesehen)
}

// geordnet macht aus einer Menge eine sortierte Liste — leer und nicht `nil`.
func geordnet(gesehen map[string]bool) []string {
	reihe := make([]string, 0, len(gesehen))
	for eintrag := range gesehen {
		reihe = append(reihe, eintrag)
	}

	slices.Sort(reihe)

	return reihe
}
