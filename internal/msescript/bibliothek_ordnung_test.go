// `sort_text` und seine Ordnungszeichenkette.
//
// # Wie die Regeln erschlossen wurden
//
// #98 stellte
// drei Fragen und hielt fest, warum sie an **einem** Vorkommen nicht zu
// beantworten sind. Die bezogenen Pakete tragen sechsundzwanzig verschiedene
// Ordnungszeichenketten; zwei der drei Fragen entscheiden sie, und zwar nicht
// über Plausibilität, sondern über **Namen und Kommentare der Vorlage selbst**:
//
//   - `<…>` — `color_filter := sort_text@(order: "<WUBRG>")`, darüber der
//     Kommentar `#### color based on mana cost, input == a mana cost`. Dazu
//     `tap_filter := sort_text@(order: "<TQ>")` unter `#### only allow tickets`
//     und `color_filterH := sort_text@(order: "</>")`. Entscheidend ist
//     `mana_is_wedge`: es zählt das Ergebnis von `sort_text(order:"<WUBRG>")`
//     und vergleicht mit **drei** — bei `WWU` dürfen das nicht drei sein.
//   - `[…]` — `mana_unsort := sort_text@(order:"[/\?XYZI0123456789VHSCAIEKPLOMWUBRG]")`.
//     Der **Name** entscheidet: eine Ordnung, in der alles in einer Klammer
//     steht, sortiert nicht. Also ist eine Klammer ein gemeinsamer Rang.
//
// Die dritte Frage — `(…)` — bleibt offen, und das ist der Punkt: `mana_sort`
// nimmt `(EKPLO)(WUBRG)`, `mana_sort_wedge` dieselbe Ordnung mit `(EPOKL)(WBGUR)`.
// Das passt auf eine **zyklische** Gruppe wie auf eine gewöhnliche Folge, die
// zwei Vorlagen verschieden anlegen. Beides erklärt die Daten; also wird nichts
// entschieden, sondern gemeldet.
//
// Quelle für jede Zeile: `magic.mse-game/script` 509–546 und 639–640 des
// bezogenen Pakets `M15-Magic-Pack`.
package msescript_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// ordne wertet `sort_text` mit einer Ordnung über einer Eingabe aus.
func ordne(t *testing.T, ordnung, eingabe string) (string, error) {
	t.Helper()

	quelle := `sort_text(order: "` + ordnung + `", "` + eingabe + `")`

	wert, err := msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
	if err != nil {
		return "", err
	}

	return wert.AlsText(), nil
}

func TestSortTextOrdnetNachDerAngegebenenReihe(t *testing.T) {
	faelle := []struct {
		name, ordnung, eingabe, erwartet string
	}{
		{
			// Die schlichte Form: jedes Zeichen ein eigener Rang.
			name: "einfache Zeichen tragen ihren Rang", ordnung: "WUBRG",
			eingabe: "GRBUW", erwartet: "WUBRG",
		},
		{
			// **Der Fall, um den es geht.** Ohne die Klammer würde aus `10`
			// ein `01`, und eine Karte kostete plötzlich eins statt zehn.
			name:    "eine Klammergruppe hält ihre Eingabereihenfolge",
			ordnung: "XYZI[0123456789]HSCA", eingabe: "10", erwartet: "10",
		},
		{
			name:    "die Klammergruppe steht als Ganzes an ihrer Stelle",
			ordnung: "XYZI[0123456789]HSCA", eingabe: "H10X", erwartet: "X10H",
		},
		{
			// `mana_unsort`: alles in einer Klammer sortiert nicht.
			name:    "alles in einer Klammer lässt die Eingabe stehen",
			ordnung: "[WUBRG0123456789]", eingabe: "G2WR", erwartet: "G2WR",
		},
		{
			// `\?` steht im Paket als `"\\?…"` — das Zeichen `?`, nicht ein
			// Sonderzeichen. Ein Gegenschrägstrich nimmt dem Nächsten seine
			// Bedeutung; anders wäre `\?` nicht zu lesen.
			name:    "ein Gegenschrägstrich macht das Nächste zum Zeichen",
			ordnung: `\?WU`, eingabe: "UW?", erwartet: "?WU",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gelesen, err := ordne(t, fall.ordnung, fall.eingabe)
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if gelesen != fall.erwartet {
				t.Errorf("%q, erwartet %q", gelesen, fall.erwartet)
			}
		})
	}
}

// **Die spitze Klammer filtert und entdoppelt.**
//
// `color_filter` macht aus einer Manakosten die Farben, die darin vorkommen —
// jede einmal. `mana_is_wedge` zählt das Ergebnis und vergleicht mit drei; wer
// nicht entdoppelte, hielte `WWU` für dreifarbig.
func TestSortTextMitSpitzerKlammerFiltertUndEntdoppelt(t *testing.T) {
	faelle := []struct {
		name, ordnung, eingabe, erwartet string
	}{
		{"die Farben einer Manakosten", "<WUBRG>", "2WWU", "WU"},
		{"in der Reihe der Ordnung", "<WUBRG>", "GW", "WG"},
		{"was nicht darin steht, fällt weg", "<WUBRG>", "2TX", ""},
		{"ein Zeichen genügt", "</>", "2W/U", "/"},
		{"und fehlt es, kommt nichts", "</>", "2WU", ""},
		{"drei Farben bleiben drei", "<WUBRG>", "RWB", "WBR"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gelesen, err := ordne(t, fall.ordnung, fall.eingabe)
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if gelesen != fall.erwartet {
				t.Errorf("%q, erwartet %q", gelesen, fall.erwartet)
			}
		})
	}
}

// **Was MSE kennt und dieses Programm nicht baut, wird gemeldet.**
//
// Hier stand bis #240 die runde Klammer selbst: sie sei „nicht erschlossen",
// weil die Pakete ihre Regel nicht entscheiden. Das war richtig, solange der
// Quelltext von MSE2 tabu war — seit dem Lizenzwechsel ist sie nachgelesen und
// gebaut (`bibliothek_zyklus.go`).
//
// **Übrig bleiben sechs Schlüsselwörter**, die `spec_sort.cpp` kennt und die
// in keinem vorliegenden Paket vorkommen. Sie zu bauen wäre Vorrat ohne
// Fundstelle; sie als Buchstabenfolge zu lesen wäre schlimmer als ein Fehler —
// es ergäbe eine Ordnung, die plausibel aussieht und falsch ist.
//
// **`in_place(` stand hier bis zum 2026-08-18 mit dabei**, und die Begründung
// lautete „kein Paket von M15-Magic-Pack führt es". Das **volle** Pack führt es
// (`magic.mse-game/script:521`), und es trifft den Regeltext — gebaut in
// #415. Die
// Regel war richtig, ihre Grundlage zu schmal.
func TestEinUngebautesSchluesselwortWirdGemeldet(t *testing.T) {
	for name, muster := range map[string]string{
		"mixed":    `mixed(WUBRG)`,
		"once":     `once(WUBRG)`,
		"ordered":  `ordered(WUBRG)`,
		"compound": `compound(WUBRG)`,
		"pattern":  `pattern(.. WUBRG)`,
		"any":      `any()`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ordne(t, muster, "WU")
			if err == nil {
				t.Fatal("kein Fehler — die Ordnung wurde geraten")
			}

			if !strings.Contains(err.Error(), "nicht gebaut") {
				t.Errorf("Meldung %q sagt nicht, dass es nicht gebaut ist", err)
			}
		})
	}
}

// **Und die runde Klammer geht jetzt** — die Gegenprobe dazu, dass die Liste
// oben `cycle(` **nicht** enthält.
func TestDieRundeKlammerWirdNichtMehrAbgewiesen(t *testing.T) {
	gebaut, err := ordne(t, `\?XYZI[0123456789]VHSFCAI(EKPLO)(WUBRG)`, "WU")
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if gebaut != "WU" {
		t.Errorf("%q", gebaut)
	}
}

// **Was in keiner Ordnung steht, bleibt stehen — hinten.**
//
// **Hier stand bis zum 2026-08-18 eine Vermutung**, und sie war falsch: „ein
// Zeichen, das die Ordnung nicht kennt, kommt hinten an". Nachgelesen hängt
// MSEs `spec_sort` an sein Ergebnis nur an, was die Spezifikation nennt, und
// wirft den Rest der Eingabe weg (`src/util/spec_sort.cpp`, Zeilen 266–352 und
// 355, Fassung f9d9356 vom 2021-07-06).
//
// Der Preis der Vermutung stand auf jeder gewöhnlichen Karte der kanonischen
// Sets: sie speichern Manakosten als `[3][W]`, und die Klammern wurden als
// leere Kästen gezeichnet
// (#411).
func TestUnbekannteZeichenFallenWeg(t *testing.T) {
	gelesen, err := ordne(t, "WUBRG", "R?W")
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if gelesen != "WR" {
		t.Errorf("%q, erwartet %q", gelesen, "WR")
	}
}

// **Eine Ordnung ohne schließende Klammer wird gemeldet.**
func TestEineOffeneKlammerWirdGemeldet(t *testing.T) {
	for name, ordnung := range map[string]string{
		"eckige": "XY[012",
		"spitze": "<WUB",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ordne(t, ordnung, "W"); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ohne `order:` wird gemeldet, nicht geraten.**
//
// Eine Ordnung, die fehlt, ist keine leere Ordnung: `sort_text` ohne sie wüsste
// nicht, wonach es sortieren soll, und eine erfundene Voreinstellung stellte
// Manakosten in eine Reihenfolge, die niemand angeordnet hat.
func TestSortTextOhneOrdnungWirdGemeldet(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		`sort_text("WU")`, msescript.Bibliothek())
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "order") {
		t.Errorf("Meldung %q nennt das fehlende Argument nicht", err)
	}
}

// **Und ohne Eingabe ebenso.**
func TestSortTextOhneEingabeWirdGemeldet(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`sort_text(order: "WU")`, msescript.Bibliothek()); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Gegenschrägstrich in der Ordnung wird geflohen geschrieben.**
//
// Hier stand: „ein Gegenschrägstrich am Ende meint sich selbst", begründet
// damit, dass kein gesichtetes Paket ihn schreibt. Die Exportvorlagen von
// `M15-Magic-Pack` schreiben ihn doch — und mit Fluchtfolgen
// (#51).
// Seitdem beendet ein Rückstrich vor dem Anführungszeichen die Zeichenkette
// nicht mehr, sondern flieht es.
//
// **Ein Beleg aus einem Fehlen ist so viel wert wie die Suche, die nichts
// fand** — der Satz stand schon in `lexer_test.go`, und er hat sich zum
// zweiten Mal bewahrheitet.
func TestEinGegenschraegstrichInDerOrdnungWirdGeflohen(t *testing.T) {
	gelesen, err := ordne(t, `WU\\`, `\\UW`)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if gelesen != `WU\` {
		t.Errorf("%q, erwartet %q", gelesen, `WU\`)
	}
}

// **Ein Zeichen, das zweimal in der Ordnung steht, gilt an seiner ersten
// Stelle.**
//
// Mehrere Ordnungen des bezogenen Pakets führen die `0` zweimal —
// `[XYZ∞01234567890HSCAPKWUBRG/|]` etwa. Wer das zweite gelten ließe, hinge ein
// Zeichen ans Ende, das vorn gemeint war.
func TestEinZeichenGiltAnSeinerErstenStelle(t *testing.T) {
	gelesen, err := ordne(t, "W0U0", "0U")
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if gelesen != "0U" {
		t.Errorf("%q, erwartet %q", gelesen, "0U")
	}
}
