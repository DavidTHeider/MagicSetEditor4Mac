package werkzeuge_test

// **Der Wächter über `docs/TISCHLEISTUNG.md`** — PLAY-06.AK-05
// (#1714).
//
// # Wofür
//
// **Weil das Dokument sich selbst widersprochen hat.** Die Tabelle oben führte
// die Rasterung als **offen**, während weiter unten ein Abschnitt
// „Die Rasterung — erhoben am 2026-09-16" mit Zahlen stand. Wer nur die
// Tabelle las, hielt eine erhobene Zahl für fehlend; wer nur den Abschnitt
// las, hielt eine offene Frage für beantwortet.
//
// **Ein Dokument, das sich selbst widerspricht, wird an der falschen Stelle
// geglaubt** — das ist der Satz, für den es dieses Issue gibt.
//
// # Was er hält
//
// Die Tabelle oben nennt je Zeile einen Zustand **und den Abschnitt, der ihn
// trägt**. Steht dort **gemessen am JJJJ-MM-TT**, muss es unterhalb genau
// diesen Abschnitt geben: „## <Abschnitt> — erhoben am <dasselbe Datum>".
// Steht dort **offen**, nennt die Zeile keinen Abschnitt, und es darf zu ihr
// **keinen** „— erhoben am" geben.
//
// **Bis zum 2026-09-25 genügte, dass das Datum irgendwo unterhalb stand** —
// Befund 5 des Zweitblicks über Phase 6.9. Fünf Zeilen tragen dasselbe Datum;
// fehlte der Abschnitt einer davon, blieb der Lauf grün, solange ein anderer
// es nannte. Die Gegenprobe steht deshalb als eigener Lauf daneben.
//
// # Was er **nicht** hält
//
// **Ob die Zahlen stimmen.** Er vergleicht zwei Stellen desselben Dokuments;
// dass eine Messung richtig erhoben wurde, sagt nur ihr Skript und die
// Maschine, auf der es lief (§10).
//
// **Und nicht, ob eine Zeile fehlt.** Ein Punkt, den niemand in die Tabelle
// schreibt, fällt keinem Wächter auf — dafür steht das Issue.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// zustandszeile liest „| **Die Rasterung** — … | **gemessen am …** … | Die
// Rasterung |".
var zustandszeile = regexp.MustCompile(
	`^\| \*\*([^*]+)\*\*[^|]*\| ([^|]+)\| ([^|]+)\|`)

// erhebungskopf liest „## Die Rasterung — erhoben am 2026-09-16".
var erhebungskopf = regexp.MustCompile(
	`(?m)^## (.+?) — erhoben am (\d{4}-\d{2}-\d{2})`)

// messdatum liest „gemessen am 2026-09-16" aus einem Zustand.
var messdatum = regexp.MustCompile(`gemessen am (\d{4}-\d{2}-\d{2})`)

func TestDieTischleistungWidersprichtSichNicht(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join("..", "..", "docs", "TISCHLEISTUNG.md")

	rohes, err := os.ReadFile(pfad) // #nosec G304 -- ein fester Pfad im Baum
	if err != nil {
		t.Fatalf("%s ist nicht zu lesen: %v", pfad, err)
	}

	befunde, gesehen := tischleistungsbefunde(string(rohes))

	for _, befund := range befunde {
		t.Error(befund)
	}

	// **Sonst wäre dieser Lauf grün, weil er nichts sieht** (§10): eine
	// Tabelle, die er nicht liest, erfüllt jede Zusage.
	if gesehen < 4 {
		t.Errorf("gelesen wurden %d Zustandszeilen — das sind zu wenige, "+
			"der Ausdruck trifft die Tabelle nicht mehr", gesehen)
	}
}

// **Die Gegenprobe, als Lauf** — ein Dokument, dem der Abschnitt einer
// gemessenen Zeile fehlt, obwohl ein anderer dasselbe Datum trägt.
func TestEinFehlenderAbschnittFaelltAuf(t *testing.T) {
	t.Parallel()

	const tabelle = "## Was gemessen ist und was nicht\n\n" +
		"| | Zustand | Abschnitt |\n| --- | --- | --- |\n" +
		"| **Ziehen** | **gemessen am 2026-09-25**, in WebKit | Ziehen |\n" +
		"| **Kontextverlust** | **gemessen am 2026-09-25** | Kontextverlust |\n" +
		"| **Fenster** | **offen** | — |\n\n"

	faelle := []struct {
		name  string
		rest  string
		nennt []string
	}{
		{
			"beide Abschnitte stehen da",
			"## Ziehen — erhoben am 2026-09-25\n\n## Kontextverlust — erhoben am 2026-09-25\n",
			nil,
		},
		{
			"einer fehlt, der andere nennt dasselbe Datum",
			"## Kontextverlust — erhoben am 2026-09-25\n",
			[]string{"Ziehen"},
		},
		{
			"einer trägt ein anderes Datum",
			"## Ziehen — erhoben am 2026-09-24\n\n## Kontextverlust — erhoben am 2026-09-25\n",
			[]string{"Ziehen"},
		},
		{
			"eine offene Zeile hat doch einen Abschnitt",
			"## Ziehen — erhoben am 2026-09-25\n\n## Kontextverlust — erhoben am 2026-09-25\n\n" +
				"## Fenster — erhoben am 2026-09-25\n",
			[]string{"Fenster"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befunde, _ := tischleistungsbefunde(tabelle + fall.rest)

			if len(befunde) != len(fall.nennt) {
				t.Fatalf("Befunde %q, erwartet einer je %v", befunde, fall.nennt)
			}

			for stelle, name := range fall.nennt {
				if !strings.Contains(befunde[stelle], name) {
					t.Errorf("der Befund %q nennt %q nicht", befunde[stelle], name)
				}
			}
		})
	}
}

// tischleistungsbefunde hält jede Zeile der Zustandstabelle gegen die
// Abschnitte darunter und gibt, was nicht passt — samt der Zahl gelesener
// Zeilen.
func tischleistungsbefunde(text string) (befunde []string, gesehen int) {
	tabelle, rest, gefunden := tabelleUndRest(text)
	if !gefunden {
		return []string{"der Abschnitt „Was gemessen ist und was nicht" +
			"\" fehlt, oder hinter ihm steht kein weiterer"}, 0
	}

	koepfe := map[string][]string{}
	for _, treffer := range erhebungskopf.FindAllStringSubmatch(rest, -1) {
		name := strings.TrimSpace(treffer[1])
		koepfe[name] = append(koepfe[name], treffer[2])
	}

	for _, zeile := range strings.Split(tabelle, "\n") {
		treffer := zustandszeile.FindStringSubmatch(zeile)
		if treffer == nil {
			continue
		}

		gesehen++

		if befund := zeilenbefund(strings.TrimSpace(treffer[1]),
			strings.TrimSpace(treffer[2]), strings.TrimSpace(treffer[3]),
			koepfe); befund != "" {
			befunde = append(befunde, befund)
		}
	}

	return befunde, gesehen
}

// zeilenbefund hält eine Zeile gegen die Abschnitte — oder gibt "".
func zeilenbefund(
	name, zustand, abschnitt string, koepfe map[string][]string,
) string {
	if datum := messdatum.FindStringSubmatch(zustand); datum != nil {
		if !slices.Contains(koepfe[abschnitt], datum[1]) {
			return "die Tabelle nennt für „" + name + "\" eine Messung am " +
				datum[1] + " im Abschnitt „" + abschnitt + "\" — unterhalb " +
				"steht „## " + abschnitt + " — erhoben am " + datum[1] +
				"\" nicht"
		}

		return ""
	}

	if !strings.Contains(zustand, "offen") {
		return "der Zustand von „" + name + "\" ist weder „offen\" noch " +
			"„gemessen am JJJJ-MM-TT\": " + zustand
	}

	if erhoben, gibtEs := koepfe[name]; gibtEs {
		return "die Tabelle führt „" + name + "\" als offen — unterhalb steht " +
			"„" + name + " — erhoben am " + erhoben[0] + "\" mit Zahlen"
	}

	return ""
}

// tabelleUndRest schneidet die Zustandstabelle vom Rest des Dokuments.
func tabelleUndRest(text string) (tabelle, rest string, gefunden bool) {
	const marke = "## Was gemessen ist und was nicht"

	anfang := strings.Index(text, marke)
	if anfang == -1 {
		return "", "", false
	}

	ende := strings.Index(text[anfang+len(marke):], "\n## ")
	if ende == -1 {
		return "", "", false
	}

	schnitt := anfang + len(marke) + ende

	return text[anfang:schnitt], text[schnitt:], true
}

// **Und der Tisch läuft nicht im Takt** — PLAY-06.AK-04
// (#1714):
// „Fehlender oder verlorener WebGL-Kontext darf weder dauerhafte CPU-Schleifen
// noch Unbedienbarkeit erzeugen."
//
// # Was er hält
//
// **`autoStart: false`** steht in `pixioptionen.ts`, und `pixinaht.ts` reicht
// genau diese Optionen an `init` — seit L02 (#2190) liegen sie dort, wo ein Lauf
// sie ohne Renderer messen kann. Ohne die Zeile fährt PixiJS
// seinen Ticker vom ersten Bild an durch — sechzig Bilder je Sekunde für
// einen stillstehenden Tisch, und nach einem verlorenen Kontext sechzig
// Fehlschläge je Sekunde. Am 2026-09-17 waren es **99 % CPU** im
// Vorschaurenderer und 4.700 Konsolenmeldungen (`docs/TISCHLEISTUNG.md`).
//
// # Was er **nicht** hält
//
// **Dass der Ticker danach wieder anhält.** Er wird für die Dauer eines
// Fluges angeworfen (`fluege.treibe`); dass er aufhört, steht in
// `pixibewegung.ts` und ist von hier aus nicht zu sehen. Die Zeile hier ist
// die **Voraussetzung**, nicht der ganze Beleg.
func TestDerTischLaeuftNichtVonSelbst(t *testing.T) {
	t.Parallel()

	lies := func(name string) string {
		pfad := filepath.Join("..", "..", "frontend", "src", "tisch", name)

		rohes, err := os.ReadFile(pfad) // #nosec G304 -- ein fester Pfad im Baum
		if err != nil {
			t.Fatalf("%s ist nicht zu lesen: %v", pfad, err)
		}

		return string(rohes)
	}

	if !strings.Contains(lies("pixioptionen.ts"), "autoStart: false") {
		t.Error("in `pixioptionen.ts` fehlt `autoStart: false` — PixiJS fährt " +
			"dann seinen Ticker durch, und ein verlorener Kontext wird zu " +
			"sechzig Fehlschlägen je Sekunde (PLAY-06.AK-04)")
	}

	if !strings.Contains(lies("pixinaht.ts"), "init(tischoptionen(") {
		t.Error("`pixinaht.ts` reicht nicht `tischoptionen` an `init` — die " +
			"gemessenen Optionen wären dann nicht die, mit denen gezeichnet wird")
	}
}
