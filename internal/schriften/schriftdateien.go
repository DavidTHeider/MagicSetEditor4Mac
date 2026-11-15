package schriften

// **Die Schriftdateien eines Orts, auch in Unterordnern** —
// #2411.
//
// macOS aktiviert eine Schrift in einem Unterordner von `~/Library/Fonts`
// (gemessen am 2026-10-15 in einem Gast mit macOS 27.0.1,
// `docs/belege/huellenprobe/schriftordner-27.0.1.md`). Wer den Ordner
// `Magic - Fonts` des Packs als Ordner dorthin zieht, sieht die Schriften in
// jedem Programm; diese Suche las die Orte flach und zeichnete die Karte ohne
// Text.
//
// **Mit Grenze**: [schrifttiefe] Ebenen unter dem Ort, versteckte Ordner
// nicht, und einem Verweis auf einen Ordner wird nicht gefolgt — ein Ort wird
// bei jedem ersten Fehlschlag gelesen, und ein Ring aus Verweisen ginge nie
// zu Ende.
//
// **Was es kostet**, gemessen am 2026-10-15 auf Mac17,3 über die vier Orte
// dieses Rechners (673 Dateien): flach 0,28 ms je Durchgang, so 0,32 ms.
// `Supplemental` liegt dabei zweimal im Weg — als eigener Ort und als
// Unterordner von `/System/Library/Fonts` —, und das kostet weniger, als
// eine Ausnahme zu führen. Die Vorrangigkeit ändert es nicht: Ebene für
// Ebene kommt `Supplemental` gleich nach den Dateien des Systemorts, wie
// vorher als eigener Ort.
//
// **Was es nicht sieht**: eine Schrift, die jemand während der Sitzung in
// einen **schon vorhandenen** Unterordner legt. Gemerkte Fehlschläge hängen
// am Stand der Orte selbst ([Suche.ortestand]), und der ändert sich dann
// nicht — dieselbe Grenze wie bei den Paketen, bis zum nächsten Start.

import (
	"os"
	"path/filepath"
	"strings"
)

// schrifttiefe ist, wie viele Ebenen unter einem Ort gesucht wird.
const schrifttiefe = 4

// schriftdateien gibt die Schriftdateien unter `ort`, relativ zu ihm — **erst
// die des Orts selbst, dann Ebene für Ebene**: was oben liegt, hat jemand
// dorthin gelegt und geht vor, gleich wie die Ordner daneben heißen. Ein Ort,
// den es nicht gibt, gibt nichts: die Liste gilt für die Plattform, nicht für
// diesen Rechner.
func schriftdateien(ort string) []string {
	var dateien []string

	ebene := []string{""}

	for tiefe := 0; tiefe <= schrifttiefe && len(ebene) > 0; tiefe++ {
		var naechste []string

		for _, ordner := range ebene {
			eintraege, err := os.ReadDir(filepath.Join(ort, ordner))
			if err != nil {
				continue
			}

			for _, eintrag := range eintraege {
				name := filepath.Join(ordner, eintrag.Name())

				switch {
				case eintrag.IsDir():
					if !strings.HasPrefix(eintrag.Name(), ".") {
						naechste = append(naechste, name)
					}
				// **Ein Verweis auf eine Schriftdatei zählt** — die flache Suche
				// davor nahm alles, was kein Ordner ist, und dabei bleibt es.
				// Einem Verweis auf einen Ordner wird nicht gefolgt: `IsDir`
				// ist für ihn falsch, und eine Endung trägt er nicht.
				case passendeEndung(strings.ToLower(filepath.Ext(eintrag.Name()))):
					dateien = append(dateien, name)
				}
			}
		}

		ebene = naechste
	}

	return dateien
}
