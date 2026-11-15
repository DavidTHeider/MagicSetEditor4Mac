package app

// **Der Dienst beim Kennungsumzug, nach dem Nachlauf über Phase 3** — 0.9.0
// (#2324),
// Befund 4 des Nachlaufs: jeder Fehler von `launchctl print` hieß „nicht
// geladen", und der Rückweg von `umladen` löschte auch eine neue plist, die
// er nicht selbst geschrieben hatte.

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

// programmausgang ist ein Fehler wie der eines beendeten Programms.
type programmausgang int

func (p programmausgang) Error() string { return "exit status " + strconv.Itoa(int(p)) }

func (p programmausgang) ExitCode() int { return int(p) }

// beideDienstplists legt die alte und die neue plist an.
func beideDienstplists(t *testing.T, umzug kennungsumzug) (alt, neu string) {
	t.Helper()

	agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
	alt = filepath.Join(agenten, altdienstlabel+".plist")
	neu = filepath.Join(agenten, dienstlabel+".plist")
	legeAn(t, alt, "alt")
	legeAn(t, neu, "neu")

	return alt, neu
}

// **Ein Fragen, das scheitert, ist kein „nicht geladen"** — dann geschieht
// nichts, und der Fehler wird gemeldet.
func TestEinScheiterndesFragenLaesstBeideDienstplistsStehen(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	alt, neu := beideDienstplists(t, umzug)
	buch.scheitert = "launchctl print gui/501/" + dienstlabel

	if err := umzug.ziehen(t.Context()); err == nil {
		t.Error("ein gescheitertes Fragen wurde nicht gemeldet")
	}

	if befehle := buch.nur("launchctl"); !slices.EqualFunc(befehle,
		[][]string{{"launchctl", "print", "gui/501/" + dienstlabel}}, slices.Equal) {
		t.Errorf("Befehle %v — nach einem gescheiterten Fragen wird nichts umgeladen", befehle)
	}

	if inhalt(t, alt) != "alt" || inhalt(t, neu) != "neu" {
		t.Error("eine der beiden plists wurde angefasst")
	}
}

// **Eine neue plist, die schon dastand, bleibt** — auch wenn das Umladen
// scheitert; weg geht nur, was der Umzug selbst schrieb.
func TestEineVorgefundeneNeueDienstplistBleibtImRueckweg(t *testing.T) {
	t.Parallel()

	for name, scheitert := range map[string]string{
		"das Entladen des alten": "launchctl bootout gui/501/" + altdienstlabel,
		"das Laden des neuen":    "launchctl bootstrap gui/501 ",
	} {
		umzug, buch := umzugIn(t)
		alt, neu := beideDienstplists(t, umzug)
		buch.ausgaenge = map[string]int{"launchctl print gui/501/" + dienstlabel: nichtGeladenCode}
		buch.scheitert = scheitert
		if scheitert == "launchctl bootstrap gui/501 " {
			buch.scheitert += neu
		}

		if err := umzug.ziehen(t.Context()); err == nil {
			t.Errorf("%s: kein Fehler", name)
		}

		if _, err := os.Stat(neu); err != nil {
			t.Errorf("%s: die vorgefundene neue plist ist weg (%v)", name, err)
		}

		if inhalt(t, alt) != "alt" {
			t.Errorf("%s: die alte plist ist nicht mehr dieselbe", name)
		}
	}
}
