package app

import (
	"regexp"
	"strings"
)

// **Was von einer Absturzspur ins Diagnosepaket geht** — Befund 7 des
// Zweitblicks über Phase 1 von 0.9.0 (F04, #2343), und Befund 5 des
// Nachlaufs darüber.
//
// # Warum nicht die ganze Spur
//
// Das Paket hält Geheimnisse heraus, indem es ihre Quellen **gar nicht
// liest** (`redigieren.go`). Eine Spur ist eine solche Quelle: ihr Kopf trägt
// den Wert, den der Code mit `panic` warf, und das ist freier Text — ein
// Kennwort darin, eine Passphrase nur aus Buchstaben erkennt der Schwärzer an
// keiner Gestalt.
//
// # Eine Liste dessen, was bleibt — nicht dessen, was geht
//
// Bis zum Nachlauf erkannte der Filter den freien Text und ließ den Rest
// stehen. **Er ließ zu viel stehen**: ein Wert, der mit `runtime error:`
// begann, eine Leerzeile im Wert, eine Fortsetzung, die wie ein Kopf aussah.
// Jetzt bleibt eine Zeile **nur, wenn das Laufzeitsystem sie in dieser Gestalt
// schreibt**; jede andere wird geschwärzt, und **auf diesem Rechner steht die
// Spur vollständig** im Ordner der Sitzungen.
//
// **Die Abgrenzung kommt von Go selbst**: jede Folgezeile eines Panikwerts —
// auch eine leere — beginnt mit einem Tab (`printindented`, auch für die
// Meldung von `throw`), und verschachtelte Paniken stehen mit `\tpanic: `
// (`printpanics`; `runtime/error.go` und `runtime/panic.go`, Go 1.27.2). Was
// nach einem Wert mit Tab beginnt, gehört zu ihm. **Die echten Gestalten
// liegen unter `testdaten/absturzspuren/`**, von Go 1.27.2 geschrieben.
//
// # Was bleibt
//
//   - ein Panikwert, der **genau** einem Satz des Laufzeitsystems gleicht —
//     Stellen und Längen als Zahlen, Typen als Namen,
//   - `fatal error:` — die Sätze von `throw` sind fest, **bis auf einen**: wirft
//     das Schreiben des Werts selbst, folgt dessen Text, und er geht,
//   - Köpfe der Goroutinen **ohne ihre Marken** (`tracebacklabels`: dort
//     stünden Werte aus dem Code),
//   - Funktionszeilen (Argumente sind Maschinenwörter), Dateizeilen, Signale,
//     Register und die festen Zwischenzeilen.

var (
	// laufzeitwerte sind die Panikwerte des Laufzeitsystems, ganz.
	laufzeitwerte = regexp.MustCompile(`^(?:` + strings.Join([]string{
		`runtime error: index out of range \[-?\d+\](?: with length \d+)?`,
		`runtime error: slice bounds out of range \[[\d:-]*\](?: with (?:length|capacity) \d+)?`,
		`runtime error: (?:invalid memory address or nil pointer dereference|integer divide by zero|` +
			`negative shift amount|integer overflow|floating point error|unsafe pointer conversion|` +
			`makeslice: len out of range|makeslice: cap out of range|makechan: size out of range|` +
			`growslice: len out of range|cgo argument has Go pointer to unpinned Go pointer)`,
		`runtime error: (?:hash of unhashable type|comparing uncomparable type) ` + typname,
		`interface conversion: ` + typname + ` is (?:nil|` + typname + `), not ` + typname,
		`interface conversion: ` + typname + ` is not ` + typname + `: missing method \w+`,
		// Sätze ohne die Vorsilbe (`runtime/map.go`, `runtime/chan.go`).
		`assignment to entry in nil map`,
		`send on closed channel`,
		`close of closed channel`,
		`close of nil channel`,
	}, "|") + `)$`)

	// gopfzeile ist der Kopf einer Goroutine bis zu ihrem Zustand — was danach
	// bis zum Doppelpunkt steht, sind Marken.
	gopfzeile = regexp.MustCompile(`^goroutine \d+(?: gp=0x[0-9a-f]+ m=(?:\d+|nil)(?: mp=0x[0-9a-f]+)?)? \[[^\]]*\]`)

	// laufzeitzeilen stehen ganz so da, wie das Laufzeitsystem sie schreibt.
	laufzeitzeilen = regexp.MustCompile(`^(?:` + strings.Join([]string{
		`[^\s(]+\(.*\)`, // eine Funktion
		`\t[^\t]+:\d+(?: \+0x[0-9a-f]+)?(?: fp=0x[0-9a-f]+ sp=0x[0-9a-f]+ pc=0x[0-9a-f]+)?`, // ihre Datei
		`created by \S+(?: in goroutine \d+)?`,
		`\[signal [^\]]*\]`,
		`\.\.\.\d+ frames elided\.\.\.`,
		`\.\.\.additional frames elided\.\.\.`,
		`runtime stack:`,
		`SIG[A-Z0-9]+: [a-z ]+`,
		`PC=0x[0-9a-f]+ m=\S+ sigcode=\S+(?: addr=0x[0-9a-f]+)?`,
		`[a-z0-9]+ +0x[0-9a-f]+`, // ein Register
		`unexpected fault address 0x[0-9a-f]+`,
		`signal arrived during (?:cgo|external code) execution`,
		``,
	}, "|") + `)$`)
)

// typname ist ein Typ, wie Go ihn schreibt.
const typname = `[\w./*\[\]{}(), ]+?`

// panikVorsilbe und fatalVorsilbe beginnen den Kopf einer Spur.
const (
	panikVorsilbe = "panic: "
	fatalVorsilbe = "fatal error: "
	fatalImWert   = "panic while printing panic value"
)

// spurFuerDasPaket gibt die Spur ohne freien Text — Zeile für Zeile, damit die
// Gestalt lesbar bleibt.
func spurFuerDasPaket(roh string) string {
	zeilen := strings.Split(roh, "\n")
	imWert := false

	for i, zeile := range zeilen {
		if imWert && strings.HasPrefix(zeile, "\t") {
			zeilen[i] = folgezeile(zeile)

			continue
		}

		zeilen[i], imWert = kopfzeile(zeile)
	}

	return strings.Join(zeilen, "\n")
}

// folgezeile ist eine Zeile mit Tab unter einem Wert: eine verschachtelte
// Panik oder die Fortsetzung eines Werts. **Beide lassen sich nicht sicher
// trennen** — eine Fortsetzung kann `panic: ` tragen —, und deshalb gilt für
// beide dasselbe: es bleibt nur ein Satz des Laufzeitsystems.
func folgezeile(zeile string) string {
	rest := strings.TrimPrefix(zeile, "\t")
	if wert, da := strings.CutPrefix(rest, panikVorsilbe); da {
		return "\t" + panikVorsilbe + wertFuerDasPaket(wert)
	}

	return "\t" + Geschwaerzt + markenVon(rest)
}

// kopfzeile ist eine Zeile ohne Wert davor; `wert` sagt, ob ihr einer folgt.
func kopfzeile(zeile string) (gezeigt string, wert bool) {
	if rest, da := strings.CutPrefix(zeile, panikVorsilbe); da {
		return panikVorsilbe + wertFuerDasPaket(rest), true
	}

	if rest, da := strings.CutPrefix(zeile, fatalVorsilbe); da {
		if strings.HasPrefix(rest, fatalImWert) {
			return fatalVorsilbe + fatalImWert + ": " + Geschwaerzt, true
		}

		return zeile, true
	}

	if kopf := gopfzeile.FindString(zeile); kopf != "" {
		return kopf + ":", false
	}

	if laufzeitzeilen.MatchString(zeile) {
		return zeile, false
	}

	return Geschwaerzt, false
}

// wertFuerDasPaket gibt einen Panikwert, wenn das Laufzeitsystem ihn so
// schreibt — sonst die Schwärzung, und die Marke dahinter bleibt.
func wertFuerDasPaket(wert string) string {
	marke := markenVon(wert)
	if laufzeitwerte.MatchString(strings.TrimSuffix(wert, marke)) {
		return wert
	}

	return Geschwaerzt + marke
}

// markenVon gibt die Marken, die Go an einen Panikwert hängt — ` [recovered]`
// und ` [recovered, repanicked]`: sie sagen, wie es weiterging, und tragen
// nichts vom Wert.
func markenVon(wert string) string {
	for _, marke := range []string{" [recovered, repanicked]", " [recovered]"} {
		if strings.HasSuffix(wert, marke) {
			return marke
		}
	}

	return ""
}
