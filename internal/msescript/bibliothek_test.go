package msescript_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func mitBibliothek(t *testing.T, quelle string) (msescript.Wert, error) {
	t.Helper()

	return msescript.AuswertenQuelle(t.Context(), quelle, msescript.Bibliothek())
}

// Normalfall, Randfall, Fehlerfall — für jede Funktion, die wirklich gerufen wird.
func TestDieBibliothekRechnetWasDieVorlagenRufen(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   string
		erwartet string
	}{
		// `to_lower(…)` aus game_script Zeile 62.
		{"to_lower schreibt klein", `to_lower("M15-Altered")`, `"m15-altered"`},
		{"to_lower auf leer", `to_lower("")`, `""`},
		{"to_lower auf einer Zahl", `to_lower(17)`, `"17"`},

		// `to_real(styling.imageheight)` aus kamen-decade Zeile 105.
		{"to_real liest eine Zahl aus Text", `to_real("420")`, "420"},
		{"to_real mit Nachkomma", `to_real("45.5")`, "45.5"},
		// **Leerraum vorn, nicht hinten** — `wxString::ToDouble` über `strtod`
		// (am Orakel gefragt; der Fall dahinter steht in
		// `umwandlungfarbe_test.go`).
		{"to_real mit Leerraum vorn", `to_real(" 12")`, "12"},
		{"to_real auf einer Zahl", `to_real(17)`, "17"},

		// `length(remove_tags(card.type_1))` aus yugioh Zeile 732.
		{"length zählt Zeichen", `length("abc")`, "3"},
		{"length zählt keine Bytes", `length("没阴影")`, "3"},
		{"length auf leer", `length("")`, "0"},
		{"length einer Abbildung", `length([a: 1, b: 2])`, "2"},

		// `max(22, card_style.rarity.content_width)` aus magic-m15.
		{"max von zweien", `max(22, 30)`, "30"},
		{"max nimmt den ersten bei Gleichstand", `max(22, 22)`, "22"},
		{"max von einem", `max(7)`, "7"},

		// `rgb(156, 130, 90)` — in drei Spielen.
		{"rgb", `rgb(156, 130, 90)`, "rgba(156,130,90,255)"},
		{"rgba", `rgba(0,0,0,0)`, "rgba(0,0,0,0)"},
		{"rgb begrenzt nach oben", `rgb(300, 0, 0)`, "rgba(255,0,0,255)"},
		{"rgb begrenzt nach unten", `rgb(-5, 0, 0)`, "rgba(0,0,0,255)"},
		{"rgb schneidet Nachkommastellen ab", `rgb(17.9, 0, 0)`, "rgba(17,0,0,255)"},
		{"zwei gleiche Farben", `rgb(1,2,3) == rgb(1,2,3)`, "true"},

		// `map["stylesheet"] or else "m15-altered"` aus game_script Zeile 62.
		// **`nil` ist kein Fehler.** `or else` fängt allein den Fehlerwert
		// (`case I_OR_ELSE: if (at == SCRIPT_ERROR) a = b`,
		// `src/script/context.cpp`, Zeilen 535–537, Fassung f9d9356 vom
		// 2021-07-06) — hier stand bis zum 2026-08-17 das Gegenteil.
		{"or else lässt nichts stehen", `to_string(nil or else "vorgabe")`, `""`},
		{"or else nimmt die linke, wenn sie da ist", `"da" or else "vorgabe"`, `"da"`},
		{"or else hinter einem fehlenden Schlüssel", `[a: 1]["b"] or else 9`, "9"},
		{"or else rechnet rechts nicht, wenn links da ist", `1 or else gibtsnicht`, "1"},
		{"or else auch bei false", `false or else 9`, "false"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := mitBibliothek(t, fall.quelle)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if wert.String() != fall.erwartet {
				t.Errorf("%s, erwartet war %s", wert.String(), fall.erwartet)
			}
		})
	}
}

// Jeder Fehlerfall nennt seine Stelle.
func TestDieBibliothekMeldetFalscheAufrufeMitStelle(t *testing.T) {
	faelle := []struct {
		name   string
		quelle string
	}{
		{"to_lower ohne Argument", `to_lower()`},
		// `to_real("keine")` und `to_real("")` stehen hier nicht mehr: das
		// eine ist ein **verzögerter** Fehler, das andere null — wie bei MSE
		// (`umwandlungfarbe_test.go`, F02).
		{"to_real ohne Argument", `to_real()`},
		{"length ohne Argument", `length()`},
		{"max ohne Argument", `max()`},
		{"max auf einer Zeichenkette", `max("a", 1)`},
		{"max auf einer Zeichenkette rechts", `max(1, "a")`},
		{"rgb mit zwei Zahlen", `rgb(1, 2)`},
		{"rgb auf einer Zeichenkette", `rgb("a", 2, 3)`},
		{"rgba mit dreien", `rgba(1, 2, 3)`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := mitBibliothek(t, fall.quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}

			if befund.Zeile == 0 {
				t.Errorf("die Meldung %q nennt keine Stelle", befund.Meldung)
			}
		})
	}
}

// Ein Farbwert lässt sich als Farbe herausholen — dafür ist er da.
func TestEinFarbwertGibtSeineAnteileHer(t *testing.T) {
	wert, err := mitBibliothek(t, `rgb(156, 130, 90)`)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	farbe, ist := wert.AlsFarbe()
	if !ist {
		t.Fatal("kein Farbwert")
	}

	if farbe != (msescript.Farbe{R: 156, G: 130, B: 90, A: 255}) {
		t.Errorf("%+v", farbe)
	}

	if _, ist := msescript.Text("keine Farbe").AlsFarbe(); ist {
		t.Error("eine Zeichenkette gab sich als Farbe aus")
	}
}

// Eine eingebaute Funktion nimmt auch vorab gebundene Argumente.
//
// Belegt ist das nicht — `@` steht in den Vorlagen nur an `sort_text`, das die
// Bibliothek (noch) nicht führt. Gebaut ist es trotzdem, weil der Aufrufpfad
// sonst zwei Fälle hätte, von denen einer nie liefe.
func TestEineEingebauteFunktionNimmtGebundeneArgumente(t *testing.T) {
	wert, err := mitBibliothek(t, `to_lower@()("ABC")`)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if wert.String() != `"abc"` {
		t.Errorf("%s", wert.String())
	}
}

// Die Bibliothek bringt true und false mit.
func TestDieBibliothekKenntDieGrundnamen(t *testing.T) {
	wert, err := mitBibliothek(t, `true and not false`)
	if err != nil || wert.String() != "true" {
		t.Errorf("%v / %v", wert, err)
	}
}

// Ein Funktionswert ohne Rumpf wird gemeldet, nicht abgestürzt.
//
// Der Zerteiler erzeugt ihn nicht; dieser Test baut ihn. Er stand am Anfang als
// echter Absturz: `binden` gab die eingebaute Funktion nicht weiter, und der
// Aufruf danach fasste in einen leeren Rumpf. Über eine Wails-Bindung heißt das
// nicht Fehlermeldung, sondern beendete Anwendung.
//
// Zwei Formen davon, seit der Funktionsteil hinter einem Zeiger liegt: einer
// ohne Teil und einer mit leerem Teil. Der erste ist der wahrscheinlichere —
// `Wert{Art: ArtFunktion}` schreibt sich versehentlich.
func TestEinFunktionswertOhneRumpfWirdGemeldet(t *testing.T) {
	faelle := map[string]msescript.Wert{
		"ohne Funktionsteil":       {Art: msescript.ArtFunktion},
		"mit leerem Funktionsteil": {Art: msescript.ArtFunktion, Funktion: &msescript.Funktionsteil{}},
	}

	for name, wert := range faelle {
		t.Run(name, func(t *testing.T) {
			umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{"leer": wert})

			_, err := msescript.AuswertenQuelle(t.Context(), `leer()`, umgebung)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}
		})
	}

	// Und `@` an einem Wert ohne Funktionsteil ebenso.
	umgebung := msescript.NeueUmgebung(map[string]msescript.Wert{
		"leer": {Art: msescript.ArtFunktion},
	})
	if _, err := msescript.AuswertenQuelle(t.Context(), `leer@(a: 1)`, umgebung); err == nil {
		t.Error("`@` an einem Wert ohne Funktionsteil kam durch")
	}
}

// `or` und `else` gehören zusammen — mit Umbruch dazwischen sind es zwei.
//
// Der erste Anlauf ließ einen Umbruch zu und erfand damit eine Form, die in
// keinem Paket steht. Geschenkt bekam er eine Mehrdeutigkeit: `if a then b or`
// gefolgt von `else c` wäre nicht mehr von `b or else c` zu unterscheiden.
func TestOrUndElseGehoerenZusammen(t *testing.T) {
	if _, err := mitBibliothek(t, "nil or\nelse \"vorgabe\""); err == nil {
		t.Error("`or` und `else` über zwei Zeilen wurden als ein Operator gelesen")
	}

	if _, err := mitBibliothek(t, `gibtsnicht or else 1`); err == nil {
		t.Error("ein unbekannter Name links von `or else` kam durch")
	}
}

// Text und Auszeichnung — remove_tags und contains.
func TestDieTextfunktionen(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		// `<i-flavor></i-flavor>` aus magic-m15 Zeile 290.
		{
			"remove_tags nimmt Auszeichnung heraus",
			`remove_tags("<i-flavor>kursiv</i-flavor>")`, `"kursiv"`,
		},
		{"remove_tags auf reinem Text", `remove_tags("nichts")`, `"nichts"`},
		{"remove_tags auf leer", `remove_tags("")`, `""`},
		{
			"remove_tags bei mehreren Auszeichnungen",
			`remove_tags("<b>a</b> und <i>b</i>")`, `"a und b"`,
		},
		// **Ein einzelnes Kleinerzeichen bleibt stehen.** Ein Kartentext, der ab
		// dem ersten `<` verschwindet, wäre schlimmer als einer mit einem
		// Zeichen zu viel.
		{"eine offene Klammer ohne Ende bleibt", `remove_tags("2 < 3")`, `"2 < 3"`},
		{"remove_tags zählt für length", `length(remove_tags("<b>abc</b>"))`, "3"},

		// `contains(style, match:"walker")` aus game_script Zeile 63.
		{"contains findet", `contains("m15-walker", match:"walker")`, "true"},
		{"contains findet nicht", `contains("m15", match:"walker")`, "false"},
		{"contains auf leerem Muster", `contains("abc", match:"")`, "true"},
		// **`match:` ist eine feste Zeichenkette, kein Muster.** Aus auto_replace:
		// `match: (C)` mit `replace: ©` wäre als regulärer Ausdruck eine Gruppe.
		{"contains nimmt das Muster wörtlich", `contains("a(C)b", match:"(C)")`, "true"},
		{"ein Punkt ist ein Punkt", `contains("axb", match:"a.b")`, "false"},
		{
			"contains mit umgekehrter Reihenfolge",
			`contains(match:"walker", "m15-walker")`, "true",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := mitBibliothek(t, fall.quelle)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if wert.String() != fall.erwartet {
				t.Errorf("%s, erwartet war %s", wert.String(), fall.erwartet)
			}
		})
	}
}

func TestDieTextfunktionenMeldenFalscheAufrufe(t *testing.T) {
	for _, quelle := range []string{
		`remove_tags()`, `contains("a")`, `contains(match:"a")`,
	} {
		t.Run(quelle, func(t *testing.T) {
			_, err := mitBibliothek(t, quelle)

			var befund *msescript.Fehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v, erwartet war ein *msescript.Fehler", err)
			}
		})
	}
}
