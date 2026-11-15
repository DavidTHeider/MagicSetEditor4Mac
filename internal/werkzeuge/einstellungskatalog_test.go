package werkzeuge_test

// **Wie der Katalog der Einstellungen gelesen wird** — die Vorrichtung zu
// `einstellungstexte_test.go`.
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet** (§2): nebenan steht, **was** im
// Katalog stehen muss; hier steht, **wie** er zusammenkommt — aus mehreren
// Teildateien, und ohne die Hinweisblöcke.
//
// Die Zeilengrenze hat den Schnitt erzwungen; richtig ist er unabhängig davon.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// katalogtexte liest die Schlüssel des Einstellungskatalogs einer Sprache.
func katalogtexte(t *testing.T, sprache string) map[string]bool {
	t.Helper()

	// **Jede Teildatei und keine Aufzählung.** Der Katalog kommt **als
	// einer** an — `seiteneinstellungen.ts` streut hinein, was in den
	// `einstellung*.ts` daneben steht, weil er sonst über die Zeilengrenze
	// aus §2 ginge.
	//
	// **Hier stand eine feste Liste von drei Dateien**, und daneben der Satz:
	// „ein Wächter, der nur die erste läse, … würde bei der nächsten
	// Auslagerung stillschweigend blind". Genau das ist am 2026-09-18
	// eingetreten — `einstellungenhandel.ts` und `einstellungshinweise.ts`
	// kamen dazu, und die Liste kannte sie nicht. **Er wurde nicht blind,
	// sondern rot**, weil die Schlüssel fehlten; blind wäre er beim nächsten
	// Mal geworden, wenn die Auslagerung nichts Neues gebracht hätte.
	//
	// **Deshalb jetzt ein Muster statt einer Liste** (§10: nicht den Fall
	// behandeln, sondern die Klasse). Wer eine Teildatei anlegt, bekommt sie
	// gelesen, ohne sie zu bestellen.
	ordner := filepath.Join("..", "..", "frontend", "src", "sprache", sprache)

	teile, err := filepath.Glob(filepath.Join(ordner, "einstellung*.ts"))
	if err != nil {
		t.Fatalf("die Teilkataloge von %s finden: %v", sprache, err)
	}

	teile = append(teile, filepath.Join(ordner, "seiteneinstellungen.ts"))

	if len(teile) < 2 {
		t.Fatalf("in %s steht nur %v — der Wächter läse fast nichts",
			sprache, teile)
	}

	var roh []byte

	for _, datei := range teile {
		teil, err := os.ReadFile(datei) //nolint:gosec // Pfad aus dem Baum selbst
		if err != nil {
			t.Fatalf("den Katalog %s lesen: %v", datei, err)
		}

		roh = append(roh, teil...)
	}

	gefunden := map[string]bool{}

	// Sowohl `'sprache.de':` als auch `dienstAus:`.
	muster := regexp.MustCompile(`(?m)^\s*'([^']+)':|^\s*([A-Za-z][A-Za-z0-9]*):`)
	for _, treffer := range muster.FindAllStringSubmatch(
		ohneHinweise(t, string(roh)), -1) {
		for _, teil := range treffer[1:] {
			if teil != "" {
				gefunden[teil] = true
			}
		}
	}

	return gefunden
}

// eigennamenAusDemFenster liest die Kennungen aus `dienstnamen.ts`.

// hinweisblock erkennt den Kopf eines Hinweisblocks.
//
// **Am Namen und nicht an einer Aufzählung.** Eine Liste stünde beim nächsten
// `…hilfe` still zu kurz da; dass die Regel überhaupt greift, prüft
// [ohneHinweise] selbst.
var hinweisblock = regexp.MustCompile(
	`(?m)^(\s*)(?:export const )?\w*(?i:hilfe|hinweise)\s*[:=]\s*\{\s*$`)

// ohneHinweise schneidet die Hinweisblöcke aus dem Katalogtext.
//
// # Der Befund, für den das hier steht
//
// **Ein Hinweis ist keine Beschriftung.** Der Wächter sammelte jeden
// Schlüssel, der irgendwo im Katalog vorkommt — und ein Schlüssel, der nur
// einen **Hinweis** hat, sah damit aus wie einer mit Namen. Im Fenster stünde
// der rohe Schlüssel, und der Lauf wäre grün.
//
// **Gemessen am 2026-09-18**: `richtlinie.handel.menge` aus der Namensliste
// gestrichen — der Lauf blieb grün, weil derselbe Schlüssel im Hinweisblock
// steht. Eine Gegenprobe, die nichts ändert, ist keine (§10); diese hat etwas
// geändert, nur nicht das Erwartete.
//
// **Was er nicht hält**: dass eine Beschriftung die richtige ist. Er sieht,
// dass eine dasteht.
func ohneHinweise(t *testing.T, roh string) string {
	t.Helper()

	zeilen := strings.Split(roh, "\n")
	uebrig := make([]string, 0, len(zeilen))
	geschnitten := 0

	for stelle := 0; stelle < len(zeilen); stelle++ {
		kopf := hinweisblock.FindStringSubmatch(zeilen[stelle])
		if kopf == nil {
			uebrig = append(uebrig, zeilen[stelle])

			continue
		}

		geschnitten++
		schluss := kopf[1] + "}"
		anfang := stelle

		// **Auch `} as const` schließt einen Block** (OPS-04, #1721): hier
		// stand nur `}`, und ein Hinweisblock mit `as const` schnitt still
		// **jeden** Katalog dahinter weg — rot wurde es nur, weil zufällig
		// eine Teildatei mit Namen folgte.
		for stelle++; stelle < len(zeilen); stelle++ {
			if strings.TrimSuffix(strings.TrimRight(zeilen[stelle], ",;"), " as const") == schluss {
				break
			}
		}

		if stelle == len(zeilen) {
			t.Fatalf("der Hinweisblock %q endet nie — der Wächter schnitte "+
				"alles danach weg", strings.TrimSpace(zeilen[anfang]))
		}
	}

	// **Und die Regel greift wirklich.** Ohne diese Zeile wäre ein Muster,
	// das nichts trifft, von einem, das alles trifft, nicht zu
	// unterscheiden — derselbe Satz wie bei den Verbotswörtern des
	// Handelsvertrags.
	if geschnitten == 0 {
		t.Fatal("kein einziger Hinweisblock erkannt — das Muster trifft nicht")
	}

	return strings.Join(uebrig, "\n")
}
