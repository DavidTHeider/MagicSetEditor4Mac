// **Die Kürzeltabelle im Handbuch hält gegen `navigation.ts`** — ergänzt am
// 2026-09-07, im Zuge der Freigabe von 0.4.0.
//
// # Warum es diesen Wächter gibt
//
// **`docs/HANDBUCH.md` sagte „`⌘1` … `⌘7` — die sieben Seiten der Leiste".**
// Das war dreifach falsch: die Grenze ist 8, `⌘3` führt zu einer **Setlasche**
// und nicht in die Leiste, und `⌘'` für „Nutzer und Profile" fehlte ganz — die
// einzige Taste für die größte Seite aus Phase 4.12
// (#1351).
//
// **Niemand hat es gesehen, weil niemand nachzählte.** `navigation.test.ts`
// hält die Kürzel im Quelltext — jede Leistenseite hat ihre Taste, `⌘?` und
// `⌘,` sind namentlich geprüft. Die **Dokumentation** stand daneben und unter
// keiner Aufsicht; sie ist über drei Phasen mitgelaufen, ohne dass etwas rot
// wurde.
//
// # Was er hält und was nicht
//
//   - **Er hält**: jedes Kürzel aus `navigation.ts` kommt in der Tabelle vor,
//     und die Tabelle erfindet keines, das es nicht gibt.
//   - **Er hält nicht**, ob die Beschreibung daneben stimmt. Dass `⌘3` als
//     Setlasche und nicht als Leistenseite beschrieben ist, sieht nur ein
//     Mensch — der Wächter sähe eine Zeile mit `⌘3` und wäre zufrieden
//     (`AGENTS.md` §10: neben einen Wächter gehört, was er nicht prüft).
package werkzeuge_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

// kuerzelImQuelltext findet `kuerzel: '…'` in `navigation.ts`.
//
// **Nur die erste Reihe.** `zweitkuerzel` trägt ⌥⌘ und steht in der Tabelle
// als Bereich; einzeln aufzuführen hieße, zehn Laschen zu nennen, die niemand
// einzeln sucht.
//
// **Das erste Zeichen wird mitgenommen und verworfen.** RE2 kennt keinen
// Blick zurück, und `zweitkuerzel` endet auf dasselbe Wort — ohne diesen
// Kniff fände das Muster beide. Die Einträge stehen teils einzeilig
// (`{ schluessel: 'start', …, kuerzel: '1', … }`), teils über mehrere Zeilen;
// `[\s,{]` deckt beide.
//
// **Der erste Anlauf verlangte einen Zeilenumbruch davor** und fand genau
// **einen** von elf. Er wurde rot statt still — die Untergrenze unten ist
// genau dafür da (§10).
//
// **Beide Anführungsarten**, und das ist kein Schönheitsfehler: der
// Apostroph für „Nutzer und Profile" steht als `kuerzel: "'"` — in einfachen
// Anführungszeichen ginge er nicht. Der zweite Anlauf des Wächters fand ihn
// deshalb nicht und meldete ihn als erfunden.
var kuerzelImQuelltext = regexp.MustCompile(`[\s,{]kuerzel: (?:'([^']*)'|"([^"]*)")`)

// kuerzelImHandbuch findet die Tasten in der Tabelle — `⌘X` in Backticks.
var kuerzelImHandbuch = regexp.MustCompile("`⌘([^`]+)`")

func TestDieKuerzeltabelleNenntJedeTaste(t *testing.T) {
	t.Parallel()

	quelltext := lies(t, "frontend/src/navigation.ts")

	gebaut := map[string]bool{}

	for _, treffer := range kuerzelImQuelltext.FindAllStringSubmatch(quelltext, -1) {
		// **Zwei Gruppen, eine gefüllt** — je nach Anführungsart.
		taste := treffer[1] + treffer[2]
		if taste != "" {
			gebaut[taste] = true
		}
	}

	// **Neun Tasten sind der heutige Stand** — sechs Ziffern, `?`, `,` und
	// `'`. Die Zahl steht hier, damit ein leerer Fund auffällt und nicht als
	// „nichts fehlt" durchgeht (§10: eine Messung ohne Reichweite sagt
	// nichts).
	//
	// **Bis UX-02 waren es elf**
	// (#1674):
	// die erste Reihe hatte zehn Ziffern, und jede Seite der Leiste trug ihre
	// eigene. Seit die Leiste **Bereiche** führt, trägt der Bereich sie — und
	// ⌘7 bis ⌘0 sind frei. Die Untergrenze folgt der Zahl; sie stehen zu
	// lassen hieße, sie zugleich frei und vergeben zu nennen.
	if len(gebaut) < 9 {
		t.Fatalf("navigation.ts nennt nur %d Kürzel — heißt das Feld noch "+
			"`kuerzel`?", len(gebaut))
	}

	tabelle := kuerzeltabelleAus(t)

	genannt := map[string]bool{}
	for _, treffer := range kuerzelImHandbuch.FindAllStringSubmatch(tabelle, -1) {
		genannt[treffer[1]] = true
	}

	// **Ein Bereich zählt für jede Ziffer darin.** `⌘4` … `⌘8` in einer Zeile
	// ist für einen Menschen fünf Tasten; ohne diesen Schritt hielte der
	// Wächter drei davon für fehlend und zwänge eine Tabelle mit elf Zeilen,
	// die niemand besser läse.
	fuelleBereiche(tabelle, genannt)

	var fehlt []string

	for taste := range gebaut {
		if !genannt[taste] {
			fehlt = append(fehlt, "⌘"+taste)
		}
	}

	if len(fehlt) > 0 {
		sort.Strings(fehlt)
		t.Errorf("die Kürzeltabelle in docs/HANDBUCH.md nennt diese Tasten "+
			"nicht: %s — wer sie sucht, findet sie nicht",
			strings.Join(fehlt, ", "))
	}

	// **Und andersherum**: eine Taste in der Tabelle, die es nicht gibt,
	// schickt jemanden auf einen Weg ins Leere.
	//
	// **`K`, `F`, `N`, `W` und `Q` sind ausgenommen** — sie stehen nicht in
	// `navigation.ts`, weil sie keine Seite führen: die Befehlspalette, die
	// Volltextsuche, ein neues Set und die zwei Fensterbefehle von macOS.
	//
	// **`[` und `]` ebenso** (UX-03.AK-02, #1675): sie gehen den Gang der
	// Seiten zurück und vor und führen auf keine bestimmte. Dass es sie gibt,
	// hält `kuerzel.test.ts` gegen `deuten`.
	ausserhalbDerLeiste := map[string]bool{
		"K": true, "F": true, "N": true, "W": true, "Q": true,
		"[": true, "]": true,

		// ⌘↩ gehört einem Feld, nicht der Leiste: es wertet in der Skriptkonsole
		// aus (Werkstück U13 von Phase 5.1, #1424).
		"↩": true,
	}

	for taste := range genannt {
		if !gebaut[taste] && !ausserhalbDerLeiste[taste] {
			t.Errorf("die Kürzeltabelle nennt ⌘%s — navigation.ts kennt es "+
				"nicht", taste)
		}
	}
}

// fuelleBereiche trägt die Ziffern zwischen `⌘a` … `⌘b` nach.
//
// **Nur Ziffern**: ein Bereich über Zeichen hätte keine Ordnung, die jemand
// erriete, und die Tabelle nennt Zeichen ohnehin einzeln.
func fuelleBereiche(tabelle string, genannt map[string]bool) {
	for _, treffer := range bereichImHandbuch.FindAllStringSubmatch(tabelle, -1) {
		for _, taste := range tastenVon(treffer[1][0], treffer[2][0]) {
			genannt[taste] = true
		}
	}
}

// tastenVon zählt die Ziffern eines Bereichs auf — **die Null am Ende**.
//
// **`⌘4` … `⌘0` ist keine leere Menge, sondern sieben Tasten** (T9, #1474).
// Auf der Tastatur liegt die Null rechts neben der Neun, und ein Bereich
// bis dorthin meint 4, 5, 6, 7, 8, 9 und 0. Ein Zähler über die Zeichen
// liefe von `'4'` bis `'0'` und damit gar nicht — und der Wächter hielte
// **fünf** Tasten für ungenannt, die dastehen.
//
// **Gefunden hat es dieser Wächter selbst**, in dem Augenblick, in dem es
// die erste ⌘0 gab: bis dahin endete jeder Bereich auf einer Ziffer über
// seinem Anfang, und die Lücke war nicht zu sehen.
func tastenVon(von, bis byte) []string {
	if bis == '0' {
		bis = '9' + 1
	}

	tasten := make([]string, 0, 10)

	for zeichen := von; zeichen <= bis; zeichen++ {
		if zeichen > '9' {
			tasten = append(tasten, "0")

			continue
		}

		tasten = append(tasten, string(zeichen))
	}

	return tasten
}

// bereichImHandbuch findet `⌘4` … `⌘8` — mit dem Auslassungszeichen dazwischen.
var bereichImHandbuch = regexp.MustCompile("`⌘(\\d)` … `⌘(\\d)`")

// kuerzeltabelleAus schneidet die Tabelle aus dem Handbuch.
//
// **Und nicht die ganze Datei**: `⌘Z` steht an vielen Stellen im Fließtext,
// und wer die ganze Datei läse, hielte jede davon für eine Tabellenzeile.
func kuerzeltabelleAus(t *testing.T) string {
	t.Helper()

	handbuch := lies(t, "docs/HANDBUCH.md")

	anfang := strings.Index(handbuch, "| Kürzel")
	if anfang < 0 {
		t.Fatal("docs/HANDBUCH.md hat keine Tabelle mit `| Kürzel` — heißt " +
			"die Spalte noch so?")
	}

	rest := handbuch[anfang:]

	// **Die Tabelle endet an der ersten Leerzeile nach ihrem Kopf.**
	ende := strings.Index(rest, "\n\n")
	if ende < 0 {
		return rest
	}

	return rest[:ende]
}
