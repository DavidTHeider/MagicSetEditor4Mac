// Die Prüfungen über **Sprachformen, die echte Pakete schreiben**.
//
// Von `zerteiler_test.go` abgetrennt, weil die Datei über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, welchen
// Baum ein Ausdruck ergibt, hier, welche Formen ein bezogenes Pack benutzt und
// woran der Zerteiler an ihnen gescheitert ist
// (#186).
package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Der `then`-Zweig darf selbst eine Bedingung sein.**
//
// # Woher der Fall kommt
//
// Aus `magic-DayNight.mse-style`, Zeile 21 seines `init script:`:
//
//	base :=
//	if      color_count == 0 then if land then default_images.l else default_images.c
//	else if color_count == 1 then default_images[colors]
//
// Der `else`-Zweig konnte seit jeher ein `if` sein — daher `else if`. Der
// `then`-Zweig nicht, und daran scheiterten **53 der 190** Initskripte eines
// bezogenen Packs. Ein Initskript, das nicht läuft, nimmt jedem Feldskript
// seine Namen mit: `face_coordinates_map` allein fehlte danach 1705-mal
// (#186).
//
// # Welches `if` das `else` bekommt
//
// Das **innere**, und das ist nicht beliebig: sonst läse sich die Zeile oben
// als „wenn `color_count == 0`, dann (wenn `land` dann `l`), sonst `c`" — und
// die Kette darunter hinge am falschen Zweig. Der Zerteiler bindet `else` an
// das nächstgelegene `if`, wie jede Sprache mit dieser Form.
func TestDerThenZweigDarfSelbstEineBedingungSein(t *testing.T) {
	umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{
		"aussen": msescript.Wahrheit(true),
		"innen":  msescript.Wahrheit(false),
	})

	faelle := map[string]string{
		"das innere else greift":  "if aussen then if innen then 1 else 2",
		"und die Kette darunter":  "if aussen then if innen then 1 else 2\nelse 3",
		"auch über Zeilen hinweg": "if aussen then\n\tif innen then 1\n\telse 2",
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), quelle, umgebung)
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.Zahl != 2 {
				t.Errorf("%v, erwartet 2", wert)
			}
		})
	}

	// **Und das äußere `else` greift, wenn das innere schon vergeben ist.**
	wert, err := msescript.AuswertenQuelle(t.Context(),
		"if innen then if aussen then 1 else 2\nelse 3", umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Zahl != 3 {
		t.Errorf("%v, erwartet 3", wert)
	}
}

// **Ein Semikolon trennt Anweisungen wie ein Umbruch.**
//
// # Woher der Fall kommt
//
// Aus rund sechzig Stilen eines bezogenen Packs, die ihr `init script:` so
// schreiben:
//
//	f := { a := 1; b := 2; a + b }
//
// Der Zerteiler wies das Zeichen ab, und jedes dieser Initskripte lief nicht —
// mit ihm keiner der Namen, die es anlegt
// (#186).
func TestEinSemikolonTrenntWieEinUmbruch(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		"a := 1; b := 2; a + b", msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Zahl != 3 {
		t.Errorf("%v, erwartet 3", wert)
	}
}

// **Nach dem Punkt darf eine Zahl stehen.**
//
// `default_images[colors.0]` und `template(pure.0, type: "pt")` stehen so in
// `magic-m15.mse-style` — ein Zugriff auf das erste Glied einer Liste. Der
// Zerteiler verlangte einen Namen und wies 53 Initskripte allein an dieser
// Stelle ab.
func TestNachDemPunktDarfEineZahlStehen(t *testing.T) {
	// **Eine Liste ist eine Abbildung mit Zahlen als Namen** — so baut der
	// Auswerter `[a, b]`, und genau deshalb ist `farben.0` ein Zugriff wie
	// jeder andere und keine eigene Form.
	umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{
		"farben": {Art: msescript.ArtAbbildung, Eintraege: []msescript.Eintrag{
			{Name: "0", Wert: msescript.Text("weiss")},
			{Name: "1", Wert: msescript.Text("blau")},
		}},
	})

	for quelle, erwartet := range map[string]string{
		"farben.0": "weiss",
		"farben.1": "blau",
	} {
		wert, err := msescript.AuswertenQuelle(t.Context(), quelle, umgebung)
		if err != nil {
			t.Fatalf("%s: %v", quelle, err)
		}

		if wert.AlsText() != erwartet {
			t.Errorf("%s ergab %q, erwartet %q", quelle, wert.AlsText(), erwartet)
		}
	}
}

// **In einem `[…]` dürfen die Namen fehlen — dann sind es Zahlen.**
//
// # Woher der Fall kommt
//
// Aus 33 Stilen eines bezogenen Packs. `magic-genevensis-90-counter.mse-style`
// schreibt eine Liste von Abbildungen:
//
//	faces_coordinates := {
//	[
//	  [left: 0, top: 0],
//	  [left: 100, top: 0],
//	]
//	}
//
// Die äußere Klammer trägt keine Namen, die inneren schon. Der Zerteiler
// verlangte sie überall und wies damit **jede** Liste ab, deren Glieder ohne
// `name:` dastehen (#186).
//
// **Die Vereinheitlichung ist keine Vereinfachung, sondern die Sache selbst**:
// eine Liste **ist** in MSE-Script eine Abbildung mit Zahlen als Namen — genau
// deshalb greift `colors.0` auf ihr erstes Glied zu.
func TestEineListeIstEineAbbildungMitZahlen(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()

	faelle := map[string]string{
		"eine schlichte Liste":       `["a", "b"].1`,
		"gemischt mit Namen":         `[erstes: "a", "b"].1`,
		"eine Liste von Abbildungen": `[[links: 1], [links: 2]].1.links`,
	}

	erwartet := map[string]string{
		"eine schlichte Liste": "b", "gemischt mit Namen": "b",
		"eine Liste von Abbildungen": "2",
	}

	for name, quelle := range faelle {
		t.Run(name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), quelle, umgebung)
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != erwartet[name] {
				t.Errorf("%q, erwartet %q", wert.AlsText(), erwartet[name])
			}
		})
	}
}

// **Ein Zweig einer Bedingung darf eine Zuweisung sein.**
//
// `magic-m15-compleated-planeswalkers.mse-style` schreibt:
//
//	output := []
//	if card.loyalty != "" then output := output + ["foil_mask_loyalty.png"]
//	output
//
// 31 Initskripte eines bezogenen Packs scheiterten daran
// (#186).
func TestEinZweigDarfEineZuweisungSein(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		"a := 1\nif true then a := 2\na", msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Zahl != 2 {
		t.Errorf("%v, erwartet 2", wert)
	}
}

// **`for each x in liste do ausdruck` — die Schleife der Exportvorlagen.**
//
// # Woher der Fall kommt
//
// Aus allen 19 Exportvorlagen von `M15-Magic-Pack`. `magic-text` schreibt:
//
//	write_cards := to_text(for each card in sort_list(cards, …) do write_card())
//
// Ohne sie erzeugt **keine einzige** Vorlage eine Ausgabe
// (#51).
//
// # Was sie ergibt
//
// Eine **Liste** — also eine Abbildung mit Zahlen als Namen, wie jede andere
// Liste auch (`docs/FORMATE.md`). `to_text` darüber setzt die Glieder
// aneinander, und genau so benutzen die Vorlagen sie.
//
// # Warum der Name **in** der Schleife lebt
//
// `card` gilt für den Durchgang und danach nicht mehr. Wer ihn in der Umgebung
// des Aufrufers stehen ließe, machte aus zwei Schleifen hintereinander eine,
// deren zweite den letzten Wert der ersten sieht — ein Fehler, den man einmal
// baut und dreimal sucht.
func TestForEachGehtEineListeDurch(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()

	if err := msescript.Vorbelegen(t.Context(),
		`namen := ["Feuerball", "Eisball"]`, umgebung); err != nil {
		t.Fatalf("vorbelegen: %v", err)
	}

	// **Mit einem Semikolon statt `\n`.** Ob ein Rückstrich in einer
	// Zeichenkette eine Fluchtfolge ist, führt `docs/FORMATE.md` seit Phase 3
	// als **offen** — dieser Test soll die Schleife prüfen und nicht eine
	// Frage mitentscheiden, die woanders steht.
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_text(for each name in namen do name + ";")`, umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "Feuerball;Eisball;" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Der Schleifenname überlebt den Durchgang — und den ganzen Aufruf.**
//
// Hier stand bis zum 2026-08-17 das Gegenteil, und es war eine **Annahme**:
// „der Schleifenname gilt nur im Durchgang". MSE führt eine einzige
// Variablentabelle und wechselt die Ebene allein beim Aufruf
// (`case I_CALL: new_scope.reset(new LocalScope(*this))`,
// `src/script/context.cpp`, Zeilen 168–170, Fassung f9d9356 vom 2021-07-06);
// eine Schleife öffnet keine.
//
// **Was die Annahme gekostet hat:** `sub_type_filter` in `magic.mse-game`
// schält in einer Zählschleife Wort für Wort von `input` ab, und mit einer
// eigenen Ebene je Durchgang sah jeder wieder den ganzen Text — aus
// `Goblin Mage` wurde `Goblin Goblin Goblin Goblin`
// (#345).
func TestDerSchleifennameUeberlebtDenDurchgang(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()

	if err := msescript.Vorbelegen(t.Context(), `zahlen := [1, 2]`, umgebung); err != nil {
		t.Fatalf("vorbelegen: %v", err)
	}

	wert, err := msescript.AuswertenQuelle(t.Context(),
		"for each eines in zahlen do eines\neines", umgebung)
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "2" {
		t.Errorf("nach der Schleife steht %q, erwartet den letzten Durchgang", wert.AlsText())
	}
}

// **Was keine Liste ist, wird benannt.**
func TestForEachUeberEtwasAnderemMeldetSich(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(t.Context(),
		`for each x in 3 do x`, msescript.Bibliothek().Darueber()); err == nil {
		t.Error("kein Fehler")
	}
}

// **`for x from a to b do ausdruck` — die zählende Schleife.**
//
// Aus `magic-field-test.mse-export-template` Zeile 252:
//
//	for x from 0 to length(cmds)-1 do (…)
//
// und aus `magic.mse-game/statistics_script`:
//
//	if count < 0 then [] else for x from 0 to count do [to_text(…)]
//
// **Die obere Grenze zählt mit.** `to length(cmds)-1` sagt es: wer bis
// `length` zählte und die letzte Stelle ausließe, schriebe `-1` nicht dazu.
func TestForFromZaehltEinschliesslich(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_text(for x from 1 to 3 do to_text(x))`, msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "123" {
		t.Errorf("%q, erwartet %q", wert.AlsText(), "123")
	}
}

// **Eine leere Spanne ergibt eine leere Liste**, keinen Fehler.
//
// `for x from 0 to count do …` mit `count` gleich −1 steht so in
// `statistics_script` — die Vorlage prüft davor selbst auf `count < 0`, aber
// verlassen kann man sich darauf nicht.
func TestEineLeereSpanneErgibtEineLeereListe(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`to_text(for x from 0 to 0 - 1 do to_text(x))`, msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "" {
		t.Errorf("%q, erwartet leer", wert.AlsText())
	}
}

// **Was an einer Schleife schiefgehen kann, meldet sich mit Stelle.**
func TestDieSchleifenMeldenIhreFehler(t *testing.T) {
	for name, quelle := range map[string]string{
		"for ohne Namen": `for each in x do 1`,

		// **Ein `for`, dem weder `each` noch ein Name folgt.** Er landet in der
		// Zählschleife, und die verlangt einen Namen.
		"for mit einer Zeichenkette": `for "a" from 1 to 2 do 1`,
		"for each ohne in":           `for each x y do 1`,
		"for each ohne do":           `for each x in [1] then 1`,
		"eine kaputte Liste":         `for each x in ( do 1`,
		"ein kaputter Rumpf":         `for each x in [1] do (`,
		"for ohne from":              `for x to 3 do 1`,
		"for ohne to":                `for x from 1 do 1`,
		"for ohne do":                `for x from 1 to 3 then 1`,
		"ein kaputter Anfang":        `for x from ( to 3 do 1`,
		"ein kaputtes Ende":          `for x from 1 to ( do 1`,
		"ein kaputter Zaehlrumpf":    `for x from 1 to 3 do (`,
		"ein Anfang ohne Zahl":       `for x from "a" to 3 do 1`,
		"ein Ende ohne Zahl":         `for x from 1 to "a" do 1`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Ein Fehler im Rumpf reist nach oben** — aus beiden Schleifenformen.
func TestEinFehlerImSchleifenrumpfReistNachOben(t *testing.T) {
	for name, quelle := range map[string]string{
		"aus der Liste":   `for each x in [1] do gibtsnicht()`,
		"aus der Zählung": `for x from 1 to 2 do gibtsnicht()`,

		// **Und aus dem, worüber gezählt wird.** Die Liste beziehungsweise die
		// Grenze wird ausgewertet, bevor der erste Durchgang läuft.
		"aus der Liste selbst": `for each x in gibtsnicht() do 1`,
		"aus dem Anfang":       `for x from gibtsnicht() to 3 do 1`,
		"aus dem Ende":         `for x from 1 to gibtsnicht() do 1`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber()); err == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
