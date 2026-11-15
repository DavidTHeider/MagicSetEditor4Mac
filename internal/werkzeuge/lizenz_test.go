// Die Lizenz steht an sechs Stellen — dieser Wächter hält sie gegeneinander.
//
// # Warum das einen Wächter braucht
//
// Eine Lizenzangabe ist keine Dokumentation, sondern eine **Zusage an Dritte**.
// Sie steht in `LICENSE`, im „Über"-Dialog, in den Metadaten des Bündels, in
// `README.md`, in `DRITTANBIETER.md` und in beiden Leitlinien. Läuft eine davon
// aus dem Tritt, sagt die Anwendung etwas anderes als die Datei daneben — und
// die im Bündel ist die, die jemand zu sehen bekommt.
//
// Der Wechsel von MIT auf GPL-2.0 am 2026-08-16
// ([0019](../../docs/entscheidungen/0019-gpl-2-0-statt-mit.md)) hat genau das
// gezeigt: sechs Stellen, und beim ersten Anlauf trugen drei noch die alte.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dieLizenz ist die Kennung, auf die sich alle sechs Stellen einigen müssen.
//
// **Die genaue Variante, nicht die Familie.** Hier stand `"GPL-2.0"`, und
// damit ging `GPL-2.0-only` durch — eine **andere** Lizenz: sie verbietet die
// Weitergabe unter einer späteren Fassung, und `LICENSE` erlaubt sie
// ausdrücklich. Gefunden vom Zweitblick, der die Kennung testweise umschrieb
// und den Wächter grün bleiben sah.
const dieLizenz = "GPL-2.0-or-later"

// mitBehauptungen sind die Wendungen, mit denen dieses Projekt sich früher als
// MIT-lizenziert bezeichnet hat.
//
// **Nicht einfach „MIT".** `DRITTANBIETER.md` führt Wails, React, Vite und
// Tailwind auf, und die stehen weiterhin unter MIT — zu Recht. Ein Wächter, der
// jedes Vorkommen der drei Buchstaben abwiese, zwänge dazu, eine wahre Angabe
// über eine fremde Abhängigkeit zu löschen. Gesucht wird deshalb die
// **Behauptung über dieses Projekt**, nicht das Wort.
//
// **Was er dafür nicht kann: Zeitformen.** „steht unter MIT" und „stand bis zum
// 2026-08-16 unter MIT" sehen für ihn gleich aus, und der zweite Satz ist wahr.
// Eine Wendung von hier lässt sich deshalb nicht historisch verwenden — wer
// über die alte Lizenz schreibt, formuliert ohne sie. `DRITTANBIETER.md` tut
// das seit dem Wechsel, und `docs/entscheidungen/` steht gar nicht erst in
// [lizenzstellen]: die Vermerke sind Aufzeichnungen ihres Tages, und 0001 muss
// weiter sagen dürfen, was es damals entschieden hat.
var mitBehauptungen = []string{
	"MIT-Lizenz",
	"unter MIT",
	"unter **MIT**",
	"MIT — mit",
	"— MIT,",

	// **Und die falsche GPL-Variante.** `-only` ist keine Schreibweise von
	// `-or-later`, sondern ihr Gegenteil.
	"GPL-2.0-only",
}

// lizenzstellen sind die Dateien, die die Lizenz dieses Projekts nennen.
//
// **Die siebte ist am 2026-08-23 dazugekommen, und sie war die wichtigste.**
// `scripts/releasezettel.sh` schreibt den Zettel, der ins `.dmg` geht — die einzige
// Lizenzauskunft, die jemand bekommt, der die Anwendung nicht aus dem Quelltext
// baut. Er trug „Lizenz: MIT", und zwar in dem `.dmg`, das als `v0.1.0`
// veröffentlicht wurde.
//
// Der Wächter kannte die Wendung längst: `"MIT — mit"` steht seit dem Wechsel
// in [mitBehauptungen]. Er sah nur an dieser Stelle nicht hin. Das ist die
// Klasse aus §10 — ein Wächter, der die halbe Zusage hält, macht die andere
// unsichtbar —, und gefunden hat sie der Zweitblick über 0.1.0, nicht dieser
// Wächter (#894).
var lizenzstellen = []string{
	"README.md",
	"DRITTANBIETER.md",
	"AGENTS.md",
	"CLAUDE.md",
	// **Das Über-Fenster und das Bündel wanderten mit Wails 3** (ENG-02,
	// #1727): den Satz des Über-Fensters setzt jetzt die Hülle, und die
	// Angaben des Bündels stehen fest in `Info.plist` statt in `wails.json`.
	// Dieselbe Regel wie beim Zettel darunter — die Stelle wandert mit dem
	// Satz.
	"internal/fenster/wails.go",
	"build/darwin/Info.plist",
	// **Der Zettel steht seit O6 in einer eigenen Datei** (#1430): mit dem
	// Kompatibilitätsmanifest ging `release.sh` über die Zeilengrenze aus §2.
	// Die Stelle wandert mit ihm — ein Wächter auf der Datei, in der die
	// Angabe **nicht mehr** steht, ist grün und misst nichts (§10).
	"scripts/releasezettel.sh",
}

func TestJedeLizenzstelleNenntGPL(t *testing.T) {
	for _, stelle := range lizenzstellen {
		t.Run(stelle, func(t *testing.T) {
			inhalt := lies(t, stelle)

			if !strings.Contains(inhalt, dieLizenz) {
				t.Errorf("%s nennt %s nicht", stelle, dieLizenz)
			}

			if wendung, da := behauptetMIT(inhalt); da {
				t.Errorf("%s sagt weiterhin %q — die Lizenz ist %s",
					stelle, wendung, dieLizenz)
			}
		})
	}
}

// **`LICENSE` trägt den Lizenztext selbst, nicht nur seinen Namen.**
//
// Eine Kennung `GPL-2.0` in einer README ist eine Angabe; die GPL verlangt, dass
// dem Werk eine **Kopie der Lizenz** beiliegt (§1). Geprüft wird deshalb der
// Text an drei Stellen, die keine Zusammenfassung enthielte: die Überschrift,
// die Weitergabeklausel und der Gewährleistungsausschluss.
func TestLicenseTraegtDenGanzenText(t *testing.T) {
	inhalt := lies(t, "LICENSE")

	for _, gefordert := range []string{
		"GNU GENERAL PUBLIC LICENSE",
		"Version 2, June 1991",
		"You may modify your copy or copies of the Program",
		"NO WARRANTY",

		// **Die „or later"-Klausel steht im Anhang, nicht im Lizenztext.**
		// Ohne sie wäre das Werk `GPL-2.0-only`, und jede Kennung daneben
		// behauptete etwas, das die Datei nicht hergibt.
		"any later version",

		// **Der Geltungsbereich überlebt den Lizenzwechsel.** Er hielt schon
		// unter MIT fest, dass Kartenrahmen, Symbole und Marken nicht gedeckt
		// sind ([0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md)).
		// Wer den Lizenztext austauscht, ersetzt leicht die ganze Datei und
		// verliert ihn dabei.
		"SCOPE OF THIS LICENSE",
		"Wizards of the Coast",
	} {
		if !strings.Contains(inhalt, gefordert) {
			t.Errorf("LICENSE führt %q nicht", gefordert)
		}
	}

	// Der MIT-Text hat einen unverwechselbaren ersten Satz.
	if strings.Contains(inhalt, "Permission is hereby granted, free of charge") {
		t.Error("LICENSE trägt noch den MIT-Text")
	}
}

// **Die Gegenprobe** — ohne sie ist nicht belegt, dass der Wächter greift.
//
// [behauptetMIT] ist die Stelle, an der er zufassen soll. Geprüft wird beides:
// dass eine alte Angabe auffällt **und** dass eine wahre Angabe über eine
// fremde Abhängigkeit durchgeht. Nur die erste Hälfte zu prüfen ergäbe einen
// Wächter, der sich mit `grep MIT` zufriedengibt und dafür die Tabelle in
// `DRITTANBIETER.md` unschreibbar macht.
func TestDerLizenzwaechterGreift(t *testing.T) {
	faelle := []struct {
		name    string
		text    string
		erwartt bool
	}{
		{
			"die alte Angabe aus README.md",
			"| **Lizenz** | MIT — mit ausdrücklicher Ausnahme fremder Inhalte |", true,
		},
		{
			"die alte Angabe aus main.go",
			`Message: "MIT-Lizenz, ausgenommen Inhalte Dritter."`, true,
		},
		{
			"die alte Angabe aus wails.json",
			`"copyright": "© 2026 David T. Heider — MIT, ausgenommen Inhalte Dritter"`, true,
		},
		{
			"die alte Angabe aus AGENTS.md",
			"Dieses Projekt steht unter **MIT** (`LICENSE`).", true,
		},

		{
			"eine fremde Abhängigkeit darf MIT sein",
			"| [Wails v2](https://github.com/wailsapp/wails) | MIT | Fenster, Bindings |", false,
		},
		{
			"die neue Angabe",
			"Der Code steht unter der **GPL-2.0-or-later** (`LICENSE`).", false,
		},
		{
			"die falsche GPL-Variante",
			"| **Lizenz** | GPL-2.0-only |", true,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if _, da := behauptetMIT(fall.text); da != fall.erwartt {
				t.Errorf("%v, erwartet war %v", da, fall.erwartt)
			}
		})
	}
}

// behauptetMIT sagt, ob ein Text dieses Projekt als MIT-lizenziert bezeichnet.
// fremdeMITWerke sind die Werke Dritter, die unter MIT stehen und über die
// deshalb **erlaubterweise** „MIT" geschrieben werden darf.
//
// # Warum es diese Liste seit dem 2026-08-16 gibt
//
// Der Wächter las bis dahin die **ganze Datei** und konnte nicht unterscheiden,
// wessen Lizenz gemeint ist. Das war richtig, solange dieses Projekt keine
// fremde MIT-Quelle hatte. Mit `magic-set-generator` (§11) hat es eine, und
// `DRITTANBIETER.md` **muss** deren Lizenz nennen — die MIT-Lizenz verlangt
// das ausdrücklich. Ein Wächter, der die Pflichterfüllung verbietet, ist
// falsch.
//
// **Die Liste ist kurz und soll es bleiben.** Sie wächst nur, wenn jemand
// bewusst eine weitere fremde Quelle aufnimmt — und dann liest er diesen
// Absatz.
var fremdeMITWerke = []string{
	"magic-set-generator",
	"Fenhl",

	// **MTGJSON kam am 2026-09-02 dazu** — Phase 4.8
	// (#1321).
	// Seine Daten stehen unter MIT („Copyright © 2018 – Present, Zach
	// Halpern"), und `DRITTANBIETER.md` sagt es, weil die Lizenz verlangt,
	// dass der Vermerk bei jeder Weitergabe mitgeht.
	"MTGJSON",
	"Zach Halpern",
}

// behauptetMIT sucht **zeilenweise**, nicht über die ganze Datei.
//
// Eine Zeile darf „MIT" sagen, wenn sie ein Werk aus [fremdeMITWerke] nennt.
// Eine Zeile, die es nicht nennt, redet über dieses Projekt — und dieses
// Projekt steht unter GPL-2.0-or-later.
//
// **Was er damit nicht mehr kann**, und das gehört danebengeschrieben: eine
// Zeile, die einen fremden Namen nennt und trotzdem über uns spricht
// („anders als magic-set-generator stehen wir unter MIT"), kommt durch. Das ist
// der Preis dafür, dass die Pflichtangabe möglich bleibt; er ist kleiner als
// der umgekehrte, denn diesen Satz müsste jemand absichtlich schreiben.
func behauptetMIT(inhalt string) (string, bool) {
	for _, zeile := range strings.Split(inhalt, "\n") {
		if nenntFremdesWerk(zeile) {
			continue
		}

		for _, wendung := range mitBehauptungen {
			if strings.Contains(zeile, wendung) {
				return wendung, true
			}
		}
	}

	return "", false
}

// nenntFremdesWerk sagt, ob eine Zeile über ein fremdes MIT-Werk spricht.
func nenntFremdesWerk(zeile string) bool {
	for _, werk := range fremdeMITWerke {
		if strings.Contains(zeile, werk) {
			return true
		}
	}

	return false
}

func lies(t *testing.T, name string) string {
	t.Helper()

	inhalt, err := os.ReadFile(filepath.Join("..", "..", name)) //nolint:gosec // fester Name aus der Liste oben
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	return string(inhalt)
}
