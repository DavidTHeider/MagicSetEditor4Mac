package msescript

// **Was ein gemerkter Baum belegt** — PERF03
// (#2353),
// Befund A-19 des Audits 0.9.0: bis dahin zählte der Baumspeicher Einträge,
// und unter 4.096 Bäumen ließ sich beliebig viel halten.
//
// **Eine obere Schätzung, keine Messung**: je Knoten seine feste Größe, sein
// Text und die Felder seiner Kinder und Namen, dazu der Schlüssel. Ein Text,
// den der Zerteiler aus dem Quelltext schneidet, teilt dessen Speicher und
// zählt hier trotzdem — gezählt wird lieber zu viel als zu wenig.

import "reflect"

// GrenzeBaumbytes ist, wie viel die gemerkten Bäume belegen dürfen — **64
// MiB**, PERF03 (#2353). An der Grenze räumt der Speicher ganz, wie an der
// Zahl ([GrenzeBaumspeicher]).
const GrenzeBaumbytes = 64 << 20

// baumgrenzeBytes ist die Grenze im Betrieb — eine Naht für den Lauf, der sie
// herabsetzt.
var baumgrenzeBytes int64 = GrenzeBaumbytes

var (
	knotengroesse     = int64(reflect.TypeFor[Knoten]().Size())
	zeigergroesse     = int64(reflect.TypeFor[*Knoten]().Size())
	zeichenkettenkopf = int64(reflect.TypeFor[string]().Size())
)

// baumBytes schätzt, was ein gemerkter Baum samt Schlüssel belegt.
func baumBytes(quelle string, baum *Knoten) int64 {
	bytes := int64(len(quelle))

	for stapel := []*Knoten{baum}; len(stapel) > 0; {
		knoten := stapel[len(stapel)-1]
		stapel = stapel[:len(stapel)-1]

		if knoten == nil {
			continue
		}

		bytes += knotengroesse + int64(len(knoten.Text)) +
			int64(cap(knoten.Kinder))*zeigergroesse + int64(cap(knoten.Namen))*zeichenkettenkopf

		for _, name := range knoten.Namen {
			bytes += int64(len(name))
		}

		stapel = append(stapel, knoten.Kinder...)
	}

	return bytes
}
