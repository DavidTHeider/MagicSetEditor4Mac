// **Ein Paketpfad und wie er zerlegt wird** — die Adresse einer Datei in
// einem fremden Paket.
//
// **Aus `einbinden.go` herausgelöst**, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort steht, **wie** ein Verweis aufgelöst wird —
// Kreise, Befunde, Sprachfassungen —, hier, **was** ein Verweis überhaupt ist
// und wie aus `/a.mse-style/bild.png` ein Paket und eine Datei werden.

package msepaket

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

type Paketpfad struct {
	Paket string
	Datei string
}

// ZerlegePaketpfad liest `/magic-modules.mse-include/scripts`.
func ZerlegePaketpfad(verweis string) (Paketpfad, error) {
	sauber := strings.TrimSpace(verweis)

	if !strings.HasPrefix(sauber, "/") {
		return Paketpfad{}, fmt.Errorf("%w: %q beginnt nicht mit `/`", ErrPaketpfad, verweis)
	}

	if fehler := gueltigerText(verweis, sauber); fehler != nil {
		return Paketpfad{}, fehler
	}

	// **`.` und `..` werden am rohen Verweis abgewiesen, nicht wegkanonisiert.**
	//
	// Der erste Anlauf ließ `path.Clean` räumen und prüfte danach. Das fing den
	// Ausbruch aufs Dateisystem, ließ aber `/a.mse-include/../b.mse-include/x`
	// durch — ein gültiger Verweis auf ein **anderes** Paket als das genannte.
	// Kein Ausbruch, aber ein Verweis, der etwas anderes tut, als er sagt.
	// Gefunden vom Zweitblick über Phase 3.
	for teil := range strings.SplitSeq(sauber, "/") {
		if teil == "." || teil == ".." {
			return Paketpfad{}, fmt.Errorf("%w: %q enthält %q", ErrPaketpfad, verweis, teil)
		}

		if fehler := ohneRandleerraum(verweis, teil); fehler != nil {
			return Paketpfad{}, fehler
		}
	}

	geteilt := strings.SplitN(strings.TrimPrefix(path.Clean(sauber), "/"), "/", 2)

	if len(geteilt) != 2 || geteilt[0] == "" || geteilt[1] == "" {
		return Paketpfad{}, fmt.Errorf("%w: %q nennt kein Paket und keine Datei",
			ErrPaketpfad, verweis)
	}

	if !strings.Contains(geteilt[0], ".mse-") {
		return Paketpfad{}, fmt.Errorf("%w: %q sieht nicht wie ein Paketname aus",
			ErrPaketpfad, geteilt[0])
	}

	return Paketpfad{Paket: geteilt[0], Datei: geteilt[1]}, nil
}

// ZerlegeEigenpfad liest einen Verweis **ohne** führenden Schrägstrich.
//
// # Was das ist
//
// Ein `include file:` ohne Schrägstrich meint eine Datei im **eigenen** Paket:
// `magic.mse-game` schreibt `include file: script` und meint
// `/magic.mse-game/script`. Erschlossen durch Auszählen der 327 Pakete von
// `M15-Magic-Pack` — 134 Angaben absolut, 27 eigen, und in jedem nachgesehenen
// Fall liegt die genannte Datei im Paket selbst (`docs/FORMATE.md`).
//
// **Ohne diese Form gibt es kein Spielskript**: `magic.mse-game` bindet all
// seine Module so ein
// (#195).
//
// # Was sie nicht darf
//
// **Hinausgreifen.** Tiefer schon — `watermarks/watermarks` steht so in
// `magic-m15-extra.mse-style` —, aber `.` und `..` sind abgewiesen, aus
// demselben Grund wie in [ZerlegePaketpfad]: sonst öffnete der eigene Verweis
// genau die Tür, die der Zweitblick über Phase 3 dort geschlossen hat.
func ZerlegeEigenpfad(paket, verweis string) (Paketpfad, error) {
	sauber := strings.TrimSpace(verweis)

	if sauber == "" || strings.HasPrefix(sauber, "/") {
		return Paketpfad{}, fmt.Errorf("%w: %q ist kein Verweis ins eigene Paket",
			ErrPaketpfad, verweis)
	}

	if fehler := gueltigerText(verweis, sauber); fehler != nil {
		return Paketpfad{}, fehler
	}

	for teil := range strings.SplitSeq(sauber, "/") {
		if teil == "." || teil == ".." || teil == "" {
			return Paketpfad{}, fmt.Errorf("%w: %q enthält %q", ErrPaketpfad, verweis, teil)
		}

		if fehler := ohneRandleerraum(verweis, teil); fehler != nil {
			return Paketpfad{}, fehler
		}
	}

	return Paketpfad{Paket: paket, Datei: sauber}, nil
}

// ohneRandleerraum weist einen Teil ab, der mit Leerraum beginnt oder endet —
// FZ-01 (#2135).
//
// **Weil er mehrdeutig ist.** Außen um den ganzen Verweis wird Leerraum
// abgeschnitten (`include file:  /a.mse-include/x` steht so in Paketen); an
// einem inneren Teil hieße dasselbe Abschneiden, eine andere Datei zu meinen.
// `/magic.mse-game/ /` gab die Datei `" "`, und zusammengesetzt und wieder
// zerlegt war daraus `/magic.mse-game/` geworden — ein Verweis, der nicht
// zurückkommt, wie er ging. Gemessen am 2026-10-07: von 6.670 Verweisen in
// `full-magic-pack` trägt **keiner** Leerraum am Rand eines Teils; Leerraum
// **im** Namen (`Mein Set.mse-set`) bleibt erlaubt.
func ohneRandleerraum(verweis, teil string) error {
	if teil != strings.TrimSpace(teil) {
		return fmt.Errorf("%w: %q trägt Leerraum am Rand von %q", ErrPaketpfad, verweis, teil)
	}

	return nil
}

// gueltigerText weist einen Verweis ab, der kein UTF-8 ist — **ein `fs.FS`
// öffnet ihn nicht** (`fs.ValidPath`), und abgewiesen wird hier mit dem Satz
// über den Verweis statt später mit `invalid argument`. Gefunden von
// `FuzzPaketpfad` (QA-06b, #1967).
func gueltigerText(verweis, sauber string) error {
	if !utf8.ValidString(sauber) {
		return fmt.Errorf("%w: %q ist kein gültiger Text", ErrPaketpfad, verweis)
	}

	return nil
}

// Paketquelle liefert den Inhalt einer Datei aus einem anderen Paket.
//
// Eine Schnittstelle und keine feste Umsetzung, weil `msepaket` nicht weiß, wo
// Pakete liegen — das entscheidet der Paketverwalter aus Phase 7. Hier steht
// nur, was gebraucht wird.
