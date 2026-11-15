// **Bewegung mit Grenze** — Vermerk
// [0065](../../docs/entscheidungen/0065-wie-diese-anwendung-aussieht.md),
// Festlegung 10 (AP-1,
// #2337).
//
// # Warum es ihn gibt
//
// 0.9.0 soll lebendiger werden — Hover, Druck, Aufklappen, Überlagerungen,
// die erscheinen. Festlegung 10 erlaubt das **mit Grenze**, und eine Grenze,
// die nur im Vermerk steht, hält bis zum ersten `duration-300`. Der Befund
// B-30 aus dem Audit zu 0.9.0 hat genau diese Reihenfolge verlangt: erst die
// Grenze und ihr Wächter, dann die Effekte.
//
// # Was er hält
//
//   - **Die zwei Merkmale** stehen in `stil.css`: `--dauer-kurz` mit 120 ms
//     und `--dauer-mittel` mit 160 ms — und beide fallen unter
//     `prefers-reduced-motion` und `[data-bewegung='reduziert']` auf 0 ms.
//   - **Keine Dauer über 200 ms** in `frontend/src`: weder als
//     Tailwind-Klasse (`duration-300`, `duration-[250ms]`) noch in CSS
//     (`transition`, `animation`) noch an einem Mantine-Bauteil
//     (`transitionProps={{ duration: … }}`) — **und seit dem Zweitblick über
//     Phase 0** (Befund 9) auch nicht in einer Variablen (`--x: 300ms`), in
//     einem Stilobjekt (`transitionDuration`), über die Web-Animations-API
//     (`.animate(…, { duration })`, **gelesen bis zur passenden Klammer**)
//     und nicht als Tailwinds eigene Animation (`animate-spin` und die
//     übrigen: alle länger). **Eine eigene Animation** (`animate-[…]`) geht
//     durch, wenn ihre Dauer darin steht und unter der Grenze liegt — ohne
//     lesbare Dauer nicht (Befund 6 des Nachlaufs). Ein Leerzeichen vor dem
//     Doppelpunkt und `_` im Namen einer Variablen ändern daran nichts.
//   - **Nur Farbe, Deckkraft und Lage**: kein `transition-all` und keine
//     Übergänge über Breite, Höhe, Abstand oder Position im Fluss — eine
//     Bewegung, die das Layout schiebt, ist ein Layoutsprung mit Ansage.
//
// # Was er **nicht** hält
//
//   - **Die Kartenbewegung am Tisch.** PixiJS zeichnet sie
//     (`frontend/src/tisch/bewegung.ts`, 220 ms), sie folgt Festlegung 9 und
//     `bewegungErwuenscht`; ein Klassenname kommt dort nicht vor. Der
//     Vermerk nennt sie ausdrücklich, und `tischbewegung_test.go` hält das
//     Dokument dazu.
//   - **Ob eine Bewegung etwas sagt.** Das ist die Frage aus Festlegung 5,
//     und sie beantwortet ein Mensch am Fenster.
//   - **Eine Dauer, die erst gerechnet wird** — `calc()`, ein Wert aus einer
//     Bibliothek ohne Klasse oder Stilobjekt, ein Name statt einer Zahl
//     (`{ duration: lang }`). Was im Quelltext nicht als Zahl dasteht, sieht
//     er nicht.
//   - **Kommentare.** `//`-Zeilen und `/* … */` in CSS zählen nicht.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// hoechsteDauer ist die Grenze aus Festlegung 10, in Millisekunden.
const hoechsteDauer = 200.0

var (
	// tailwinddauer findet `duration-150` und `duration-[250ms]`.
	tailwinddauer = regexp.MustCompile(`\bduration-(\d+)\b|\bduration-\[(\d+(?:\.\d+)?)(ms|s)\]`)

	// cssdauer findet eine Zeitangabe in `transition` oder `animation`.
	cssdauer = regexp.MustCompile(`(?i)\b(?:transition|animation)(?:-duration)?\s*:[^;{}]*`)

	// zeitangabe findet `120ms`, `0.2s`, `.15s`.
	zeitangabe = regexp.MustCompile(`(\d*\.?\d+)(ms|s)\b`)

	// mantinedauer findet `transitionProps={{ duration: 300 }}`.
	mantinedauer = regexp.MustCompile(`transitionProps=\{\{[^}]*duration\s*:\s*(\d+)`)

	// layoutuebergang findet Übergänge, die das Layout schieben.
	layoutuebergang = regexp.MustCompile(
		`\btransition-all\b|\btransition-\[[^\]]*(?:width|height|margin|padding|top|left|right|bottom|inset)[^\]]*\]|` +
			`(?i)\btransition(?:-property)?\s*:[^;{}]*\b(?:all|width|height|max-height|margin|padding|top|left|right|bottom|inset)\b`)

	// csskommentar ist ein Blockkommentar in CSS.
	csskommentar = regexp.MustCompile(`(?s)/\*.*?\*/`)

	// variablendauer findet eine Dauer in einer Variablen, `--dauer-lang: 300ms`.
	variablendauer = regexp.MustCompile(`--[A-Za-z0-9_-]+\s*:\s*(\d*\.?\d+)(ms|s)\b`)

	// animationsklasse findet Tailwinds eigene Animationen, `animate-spin`.
	animationsklasse = regexp.MustCompile(`\banimate-(?:spin|ping|pulse|bounce)\b`)

	// eigeneanimation findet `animate-[auftauchen_120ms_ease-out]` — die
	// Dauer steht darin und wird gelesen.
	eigeneanimation = regexp.MustCompile(`\banimate-\[([^\]]*)\]`)

	// skriptdauer findet `transitionDuration: '250ms'` in einem Stilobjekt.
	skriptdauer = regexp.MustCompile(`\b(?:transition|animation)Duration\s*:\s*['"]?(\d*\.?\d+)(ms|s)?`)

	// aufrufdauer findet `duration: 400` im Text eines `.animate(…)`, den
	// [animateaufrufe] bis zur passenden Klammer liest.
	aufrufdauer = regexp.MustCompile(`\bduration\s*:\s*(\d*\.?\d+)`)
)

// bewegungsfehler nennt, was in einem Quelltext gegen Festlegung 10 steht.
func bewegungsfehler(inhalt string) []string {
	var funde []string

	melde := func(was string, ms float64) {
		if ms > hoechsteDauer {
			funde = append(funde, was+" — "+strconv.FormatFloat(ms, 'f', -1, 64)+" ms")
		}
	}

	for _, treffer := range tailwinddauer.FindAllStringSubmatch(inhalt, -1) {
		melde(treffer[0], millisekunden(treffer[1]+treffer[2], treffer[3]))
	}

	for _, treffer := range cssdauer.FindAllString(inhalt, -1) {
		for _, zeit := range zeitangabe.FindAllStringSubmatch(treffer, -1) {
			melde(strings.TrimSpace(treffer), millisekunden(zeit[1], zeit[2]))
		}
	}

	for _, treffer := range mantinedauer.FindAllStringSubmatch(inhalt, -1) {
		melde(treffer[0], millisekunden(treffer[1], "ms"))
	}

	for _, treffer := range variablendauer.FindAllStringSubmatch(inhalt, -1) {
		melde(treffer[0], millisekunden(treffer[1], treffer[2]))
	}

	for _, treffer := range skriptdauer.FindAllStringSubmatch(inhalt, -1) {
		melde(treffer[0], millisekunden(treffer[1], treffer[2]))
	}

	for _, aufruf := range animateaufrufe(inhalt) {
		for _, treffer := range aufrufdauer.FindAllStringSubmatch(aufruf, -1) {
			melde(aufruf, millisekunden(treffer[1], "ms"))
		}
	}

	for _, treffer := range animationsklasse.FindAllString(inhalt, -1) {
		funde = append(funde, treffer+" — eine Animation von Tailwind läuft länger als 200 ms")
	}

	for _, treffer := range eigeneanimation.FindAllStringSubmatch(inhalt, -1) {
		// **Tailwind schreibt Leerzeichen als `_`** — und `_` ist ein
		// Wortzeichen, an dem `\b` nicht greift.
		zeiten := zeitangabe.FindAllStringSubmatch(strings.ReplaceAll(treffer[1], "_", " "), -1)
		if len(zeiten) == 0 {
			funde = append(funde, treffer[0]+" — eine eigene Animation ohne Dauer, die sich lesen ließe")
		}

		for _, zeit := range zeiten {
			melde(treffer[0], millisekunden(zeit[1], zeit[2]))
		}
	}

	for _, treffer := range layoutuebergang.FindAllString(inhalt, -1) {
		funde = append(funde, treffer+" — nur Farbe, Deckkraft und Lage")
	}

	return funde
}

// animateaufrufe gibt den Text jedes `.animate(…)` bis zur passenden
// Klammer — ein Muster endete an der ersten, und `animate(bilder(), …)`
// ging durch (Befund 6 des Nachlaufs über Phase 0).
func animateaufrufe(inhalt string) []string {
	var aufrufe []string

	for rest := inhalt; ; {
		_, danach, gefunden := strings.Cut(rest, ".animate(")
		if !gefunden {
			return aufrufe
		}

		tiefe, ende := 1, len(danach)

		for stelle, zeichen := range danach {
			switch zeichen {
			case '(':
				tiefe++
			case ')':
				tiefe--
			}

			if tiefe == 0 {
				ende = stelle

				break
			}
		}

		aufrufe = append(aufrufe, danach[:ende])
		rest = danach[ende:]
	}
}

// millisekunden rechnet eine Zeitangabe um; ohne Einheit sind es
// Millisekunden, wie bei Tailwind.
func millisekunden(zahl, einheit string) float64 {
	wert, fehler := strconv.ParseFloat(zahl, 64)
	if fehler != nil {
		return 0
	}

	if einheit == "s" {
		return wert * 1000
	}

	return wert
}

// **Die Gegenprobe zuerst**: was der Prüfer findet, ist an Beispielen zu
// sehen — und was er durchlässt, ebenso.
func TestDerBewegungspruefer(t *testing.T) {
	t.Parallel()

	for inhalt, erwartet := range map[string]int{
		`className="transition-colors duration-150"`:          0,
		`className="duration-200"`:                            0,
		`className="duration-300"`:                            1,
		`className="duration-[201ms]"`:                        1,
		`transition: color var(--dauer-kurz), opacity 160ms;`: 0,
		`transition: color 0.25s;`:                            1,
		`animation: auftauchen 300ms ease-out;`:               1,
		`transition-duration: 0.01ms !important;`:             0,
		`transitionProps={{ duration: 0 }}`:                   0,
		`transitionProps={{ duration: 250 }}`:                 1,
		`className="transition-all duration-100"`:             1,
		`transition: height 100ms;`:                           1,
		`className="transition-[max-height] duration-100"`:    1,
		`className="transition-transform duration-[160ms]"`:   0,
		// **Befund 9 des Zweitblicks über Phase 0**: eine Dauer in einer
		// Variablen, implizite Animationsklassen und Bewegung aus Skript.
		`--dauer-lang: 300ms;`:                                1,
		`--dauer-kurz: 120ms;`:                                0,
		`className="animate-spin"`:                            1,
		`className="animate-none"`:                            0,
		`style={{ transitionDuration: '250ms' }}`:             1,
		`style={{ animationDuration: '0.1s' }}`:               0,
		`knoten.animate([{ opacity: 0 }], { duration: 400 })`: 1,
		`knoten.animate([{ opacity: 0 }], { duration: 150 })`: 0,
		// **Befund 6 des Nachlaufs**: eine Klammer im Aufruf, ein Leerzeichen
		// vor dem Doppelpunkt, ein Unterstrich im Namen — und eine eigene
		// Animation, deren Dauer dasteht und unter der Grenze liegt.
		`knoten.animate(bilder(), { duration: 400 })`:     1,
		`knoten.animate(bilder(), { duration : 150 })`:    0,
		`style={{ transitionDuration : '250ms' }}`:        1,
		`transitionProps={{ duration : 250 }}`:            1,
		`--dauer_lang: 300ms;`:                            1,
		`className="animate-[auftauchen_120ms_ease-out]"`: 0,
		`className="animate-[auftauchen_0.3s_ease-out]"`:  1,
		`className="animate-[auftauchen_ease-out]"`:       1,
		`knoten.animate(bilder(), { duration: lang })`:    0,
	} {
		if gefunden := bewegungsfehler(inhalt); len(gefunden) != erwartet {
			t.Errorf("%s: %d Funde erwartet, %d gefunden: %v", inhalt, erwartet, len(gefunden), gefunden)
		}
	}
}

// TestKeineBewegungUeberDerGrenze hält Festlegung 10 im ganzen Fenster.
func TestKeineBewegungUeberDerGrenze(t *testing.T) {
	t.Parallel()

	ordner := filepath.Join(wurzel, "frontend", "src")
	gesehen := 0

	fehler := filepath.WalkDir(ordner, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() || !stilträger(pfad) {
			return err
		}

		gesehen++

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Baum
		if lesefehler != nil {
			return lesefehler
		}

		inhalt := csskommentar.ReplaceAllString(ohneZeilenkommentare(string(roh)), "")

		for _, fund := range bewegungsfehler(inhalt) {
			kurz, _ := filepath.Rel(wurzel, pfad) //nolint:errcheck // nur zur Anzeige
			t.Errorf("%s: %s — Vermerk 0065, Festlegung 10", filepath.ToSlash(kurz), fund)
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", ordner, fehler)
	}

	if gesehen < 100 {
		t.Fatalf("nur %d Dateien gelesen — der Wächter greift daneben", gesehen)
	}
}

// TestDieZweiDauernStehenUndSchweigenOhneBewegung hält die Merkmale.
func TestDieZweiDauernStehenUndSchweigenOhneBewegung(t *testing.T) {
	t.Parallel()

	stil := lies(t, "frontend/src/stil.css")
	ohne := lies(t, "frontend/src/barrierefreiheit.css")

	for _, muss := range []string{"--dauer-kurz: 120ms;", "--dauer-mittel: 160ms;"} {
		if !strings.Contains(stil, muss) {
			t.Errorf("frontend/src/stil.css führt %q nicht — Vermerk 0065, Festlegung 10", muss)
		}
	}

	// **Beide Auslöser, und je beide Merkmale auf 0 ms.**
	for _, ausloeser := range []string{"@media (prefers-reduced-motion: reduce)", "[data-bewegung='reduziert']"} {
		_, danach, gefunden := strings.Cut(ohne, ausloeser)
		block, _, _ := strings.Cut(danach, "}\n\n")

		if !gefunden || !strings.Contains(block, "--dauer-kurz: 0ms") ||
			!strings.Contains(block, "--dauer-mittel: 0ms") {
			t.Errorf("unter %s fallen die Dauern nicht auf 0 ms", ausloeser)
		}
	}
}

// TestOhneBewegungSinktKeinKnopf hält die Lage still — Befund 6 des
// Zweitblicks über Phase 6 von 0.9.0: die Regeln unter „wenig Bewegung“
// verkürzten nur Dauern, und `active:translate-y-px` senkte einen gedrückten
// Knopf weiter um einen Punkt. Eine Dauer von null ist keine ausbleibende
// Bewegung.
//
// **Und nur der Knopf** — Rest R4 des Nachlaufs: hier stand `*:active`, und
// das nahm jedem gedrückten Element seine Lage, auch den Griffen des
// Zuschnitts, die mit `translate` auf ihrer Ecke sitzen. Die Regel trifft
// deshalb die Klasse aus `rueckmeldung`, gelesen aus `knopfklassen.ts`.
func TestOhneBewegungSinktKeinKnopf(t *testing.T) {
	t.Parallel()

	ohne := lies(t, "frontend/src/barrierefreiheit.css")

	druck := regexp.MustCompile(`\S*active:translate\S*`).FindString(lies(t, "frontend/src/bausteine/knopfklassen.ts"))
	if druck == "" {
		t.Fatal("knopfklassen.ts nennt keine Druckbewegung — der Wächter liest an der falschen Stelle")
	}

	regel := "[class~='" + strings.Trim(druck, "'") + "']:active"

	// **Ein Selektor über zwei Zeilen ist derselbe** — Prettier bricht den
	// langen um.
	einzeilig := strings.Join(strings.Fields(ohne), " ")

	// **Je Auslöser sein eigener Block** — die Regel der Anwendung stünde
	// sonst für die des Systems ein.
	for _, stelle := range []struct{ von, bis string }{
		{"@media (prefers-reduced-motion: reduce)", "[data-bewegung='reduziert'] {"},
		{"[data-bewegung='reduziert'] " + regel, "}"},
	} {
		_, danach, gefunden := strings.Cut(einzeilig, stelle.von)
		block, _, _ := strings.Cut(danach, stelle.bis)

		if !gefunden || !strings.Contains(stelle.von+block, regel) || !strings.Contains(block, "translate: none !important") {
			t.Errorf("unter %s bleibt ein gedrückter Knopf nicht stehen — erwartet wird %s", stelle.von, regel)
		}
	}

	if regexp.MustCompile(`\*:active\s*\{`).MatchString(ohne) {
		t.Error("barrierefreiheit.css nimmt jedem gedrückten Element die Lage (`*:active`) — gemeint ist nur der Knopf")
	}
}
