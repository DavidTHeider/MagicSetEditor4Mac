//go:build fremdpakete

// **Wie oft ein Schlüsselwort wirklich vorkommt** — Werkstück M14 von
// Phase 5.5
// (#1616).
//
// # Wofür es diesen Lauf gibt
//
// **Weil das Issue den Schnitt *gezählt* verlangt und nicht gefühlt.**
// CR 702 trägt 771 Regeln — ein Viertel des Regelwerks —, und „nach
// Häufigkeit geschnitten" ist die richtige Idee und noch kein Schnitt. Was
// hier läuft, erhebt die Zahl; was daraus folgt, steht in
// `docs/SCHLUESSELWOERTER.md`.
//
// # Die Reichweite der Messung ist die Reichweite ihrer Aussage (§10)
//
// **Gemessen wird über die zehn kanonischen Sets** unter
// `~/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/` — die
// Sammlung aus [0026](../../docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md),
// Daten aus Scryfall mit Stand 2022-10-24. **Das ist nicht „Magic"**: es ist
// eine nach Farben geschnittene Auswahl, und was nach 2022 gedruckt wurde,
// fehlt darin ganz.
//
// **Hinter `fremdpakete`**, weil die Sets nicht im Baum liegen (§11) und
// weil ein Lauf, der sie braucht, in keiner Pipeline stehen darf (§3).
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// schluesselwoerter sind die Kandidaten, nach denen gezählt wird.
//
// **Aus CR 702 und nicht aus dem Gefühl.** Die Liste führt die
// Schlüsselwortfähigkeiten, die **ohne Parameter** auf einer Zeile stehen —
// genau die, die sich aus einem gedruckten Text erkennen lassen, ohne ihn zu
// deuten. Was einen Wert trägt („Trample over", „Ward {2}") zählt hier nicht
// mit; das ist `M8`.
var schluesselwoerter = []string{
	"flying", "first strike", "double strike", "deathtouch", "defender",
	"haste", "hexproof", "indestructible", "lifelink", "menace",
	"protection", "reach", "shadow", "trample", "vigilance", "banding",
	"flash", "fear", "intimidate", "shroud", "horsemanship", "islandwalk",
	"swampwalk", "forestwalk", "mountainwalk", "plainswalk", "landwalk",
	"changeling", "devoid", "ingest", "infect", "wither", "persist",
	"undying", "exalted", "battle cry", "soulbond", "unleash", "evolve",
	"extort", "myriad", "skulk", "melee", "afflict", "mentor", "riot",
	"convoke", "delve", "improvise", "cascade", "storm", "split second",
	"shield counter", "toxic", "backup", "bargain", "prowess",
}

// **Die Zählung über die kanonischen Sets.**
//
// **Kein Wächter, sondern eine Messung**: er behauptet nichts und scheitert
// an nichts — er schreibt auf, was dasteht. Die Zahlen daraus stehen in
// `docs/SCHLUESSELWOERTER.md`, und wer sie ändern will, fährt ihn wieder.
func TestWieOftEinSchluesselwortVorkommt(t *testing.T) {
	zuhause, fehler := os.UserHomeDir()
	if fehler != nil {
		t.Skipf("kein Zuhause: %v", fehler)
	}

	wurzel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "kanonische-sets")

	dateien, fehler := filepath.Glob(filepath.Join(wurzel, "*.mse-set"))
	if fehler != nil || len(dateien) == 0 {
		t.Skipf("keine kanonischen Sets unter %q", wurzel)
	}

	sort.Strings(dateien)

	gezaehlt := map[string]int{}
	karten, mitText := 0, 0

	for _, datei := range dateien {
		eine, andere, weitere := setzaehlen(t, datei, gezaehlt)
		karten += eine
		mitText += andere

		_ = weitere
	}

	t.Logf("%d Sets, %d Karten, %d mit Regeltext", len(dateien), karten, mitText)

	art := make([]string, 0, len(gezaehlt))
	for wort := range gezaehlt {
		art = append(art, wort)
	}

	sort.Slice(art, func(i, j int) bool {
		if gezaehlt[art[i]] != gezaehlt[art[j]] {
			return gezaehlt[art[i]] > gezaehlt[art[j]]
		}

		return art[i] < art[j]
	})

	for _, wort := range art {
		t.Logf("%6d  %s", gezaehlt[wort], wort)
	}
}

// setzaehlen zählt die Schlüsselwörter eines Sets.
func setzaehlen(
	t *testing.T, datei string, gezaehlt map[string]int,
) (karten, mitText, uebrig int) {
	t.Helper()

	paket, fehler := msepaket.Oeffne(datei)
	if fehler != nil {
		t.Logf("%s: %v", filepath.Base(datei), fehler)

		return 0, 0, 0
	}

	defer func() {
		if offen := paket.Schliessen(); offen != nil {
			t.Logf("%s schließen: %v", filepath.Base(datei), offen)
		}
	}()

	inhalt, fehler := paket.Datei("set")
	if fehler != nil {
		t.Logf("%s: %v", filepath.Base(datei), fehler)

		return 0, 0, 0
	}

	baum, fehler := msepaket.Zerteile("set", inhalt)
	if fehler != nil {
		t.Logf("%s: %v", filepath.Base(datei), fehler)

		return 0, 0, 0
	}

	for _, knoten := range baum.Wurzel {
		if knoten.Schluessel != "card" {
			continue
		}

		karten++

		text := ""

		for _, feld := range knoten.Kinder {
			if msepaket.Entkanonisch(feld.Schluessel) == "rule text" {
				text = msepaket.WertOderBlock(feld)
			}
		}

		if strings.TrimSpace(text) == "" {
			continue
		}

		mitText++

		zaehleZeile(text, gezaehlt)
	}

	return karten, mitText, 0
}

// zaehleZeile zählt, welche Schlüsselwörter eine Karte **trägt**.
//
// # Zwei Zahlen, und nur eine beantwortet die Frage
//
// **„Kommt das Wort vor" ist nicht „hat die Karte die Fähigkeit".** „Kann
// nicht von Kreaturen ohne Flugfähigkeit geblockt werden" nennt das Wort und
// gibt es niemandem. Der erste Anlauf zählte so und kam bei `flying` auf
// 3.349 von 24.422 Karten — eine Zahl über den **Text** und nicht über die
// Fähigkeit (§10).
//
// **Gezählt wird deshalb die Zeile.** Eine Schlüsselwortfähigkeit steht nach
// CR 702.1 allein auf einer Zeile, mehrere durch Kommas getrennt; dahinter
// darf Erinnerungstext in Klammern stehen (CR 207.2a). Eine Zeile, die
// **nur** aus bekannten Wörtern besteht, gibt sie — jede andere nicht.
//
// **Je Karte höchstens einmal je Wort**: gefragt ist, auf wie vielen Karten
// es steht.
func zaehleZeile(text string, gezaehlt map[string]int) {
	gesehen := map[string]bool{}

	for _, zeile := range strings.Split(text, "\n") {
		for _, wort := range woerterDerZeile(zeile) {
			if !gesehen[wort] {
				gesehen[wort] = true
				gezaehlt[wort]++
			}
		}
	}
}

// woerterDerZeile gibt die Schlüsselwörter einer Zeile — oder nichts.
//
// **Alles oder nichts**: steht in der Aufzählung etwas, das kein bekanntes
// Schlüsselwort ist, ist die Zeile ein Satz und keine Aufzählung.
func woerterDerZeile(zeile string) []string {
	ohne := ohneKlammern(strings.ToLower(zeile))

	// **Die Auszeichnungen des Pakets fallen weg** — MSE schreibt
	// `<i-flavor>` und `<soft-line>` in denselben Text.
	ohne = ohneKlammernMit(ohne, '<', '>')

	teile := strings.Split(ohne, ",")

	woerter := make([]string, 0, len(teile))

	for _, teil := range teile {
		teil = strings.TrimSpace(strings.Trim(strings.TrimSpace(teil), "."))
		if teil == "" {
			continue
		}

		if !istSchluesselwort(teil) {
			return nil
		}

		woerter = append(woerter, teil)
	}

	return woerter
}

// istSchluesselwort sagt, ob dieses Wort in der Liste steht.
//
// **Mit den Parametrischen**: „protection from red" und „ward {2}" sind
// Schlüsselwörter mit einem Wert (CR 702.16, CR 702.21); gezählt werden sie
// unter ihrem Namen.
func istSchluesselwort(teil string) bool {
	for _, wort := range schluesselwoerter {
		if teil == wort || strings.HasPrefix(teil, wort+" ") {
			return true
		}
	}

	return false
}

// ohneKlammern schneidet den Erinnerungstext weg — CR 207.2.
func ohneKlammern(zeile string) string {
	return ohneKlammernMit(zeile, '(', ')')
}

// ohneKlammernMit schneidet alles zwischen zwei Zeichen weg.
func ohneKlammernMit(zeile string, auf, zu rune) string {
	var gebaut strings.Builder

	tiefe := 0

	for _, zeichen := range zeile {
		switch {
		case zeichen == auf:
			tiefe++
		case zeichen == zu && tiefe > 0:
			tiefe--
		case tiefe == 0:
			gebaut.WriteRune(zeichen)
		}
	}

	return gebaut.String()
}
