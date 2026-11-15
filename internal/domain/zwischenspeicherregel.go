// **Die Grenze des Zwischenspeichers auf der Platte** — OPS-08.AK-03
// (#1725),
// Marke `new`.
//
// **In Megabyte, je Speicher**: die Antworten und die Bilder von Scryfall
// halten sie jeder für sich. Die Sammeldatei hat keine — sie ist **eine**
// Datei, die ein neuer Stand ersetzt.
//
// **Und kein einstellbarer Ablauf.** Scryfall bittet, seine Daten mindestens
// 24 Stunden zu halten (`internal/scryfall`, `Speicherfrist`); darunter wäre
// es unhöflich, darüber zeigte die Anwendung länger alte Preise, ohne dass es
// jemand bestellt hätte. `docs/ZWISCHENSPEICHER.md` sagt das bei jedem
// Speicher.

package domain

import "strconv"

// ZwischenspeicherGrenze ist der Schlüssel der Grenze in Megabyte.
const ZwischenspeicherGrenze = "zwischenspeicher.grenze"

// Die Grenzen der Grenze, in Megabyte.
const (
	zwischenspeichergrenzeVorgabe    = 512
	zwischenspeichergrenzeHoechstens = 8192
)

// zwischenspeichereinstellungen ist die Zeile für die Reihe.
func zwischenspeichereinstellungen() []einstellung {
	return []einstellung{{
		Schluessel: ZwischenspeicherGrenze, Wirkung: WirkungNeustart, Bereich: BereichAnwendung,
		Vorgabe:  strconv.Itoa(zwischenspeichergrenzeVorgabe),
		Pruefung: zahlZwischen(1, zwischenspeichergrenzeHoechstens),
	}}
}

// ZwischenspeichergrenzeAus gibt die Grenze in Bytes — ein unlesbarer Wert
// fällt auf die Vorgabe und nie auf „unbegrenzt".
func ZwischenspeichergrenzeAus(wert string) int64 {
	megabyte, fehler := strconv.Atoi(wert)
	if fehler != nil || megabyte < 1 || megabyte > zwischenspeichergrenzeHoechstens {
		megabyte = zwischenspeichergrenzeVorgabe
	}

	return int64(megabyte) << 20
}
