package main

// **Die vier Fragen an eine geordnete Sammlung** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk [0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).

import (
	"strings"
	"testing"
)

// geordnet legt Besitz und Wunsch desselben Drucks an.
//
// **Derselbe Druck zweimal** — der Fall, für den die Eindeutigkeit `wunsch`
// aufnimmt: wer eine gespielte Fassung hat und eine schöne will, hat beides.
func geordnet(t *testing.T) string {
	t.Helper()

	ort := eigenesZuhause(t)

	if code, _, fehler := fahreAn(t, ort, "exemplar", "probe-druck", "4",
		"nearMint", "en", "nonfoil", "", "", "", "", "", "", "",
		"Kiste 4", "besitz"); code != CodeGut {
		t.Fatalf("Besitz anlegen: %d (%s)", code, fehler)
	}

	if code, _, fehler := fahreAn(t, ort, "exemplar", "probe-druck", "2",
		"played", "en", "nonfoil", "", "", "", "", "", "", "",
		"", "wunsch"); code != CodeGut {
		t.Fatalf("Wunsch anlegen: %d (%s)", code, fehler)
	}

	return ort
}

// **`besitz` trennt Besitz und Wunsch.**
func TestBesitzTrenntBesitzUndWunsch(t *testing.T) {
	ort := geordnet(t)

	code, aus, fehler := fahreAn(t, ort, "besitz", "probe-druck")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "4 besessen, 2 gewünscht") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}

	// **Beide Zeilen stehen da, jede mit ihrer Absicht.**
	if !strings.Contains(aus, "besitz") || !strings.Contains(aus, "wunsch") {
		t.Errorf("die Absichten fehlen: %q", aus)
	}
}

// **Ohne Kennung ist es ein Bedienfehler.**
func TestBesitzOhneKennung(t *testing.T) {
	if code, _, _ := fahreAn(t, eigenesZuhause(t), "besitz"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **`lagerorte` zählt auf, was wo liegt — und Wünsche liegen nirgends.**
func TestLagerorteZaehlenNurBesitz(t *testing.T) {
	ort := geordnet(t)

	code, aus, fehler := fahreAn(t, ort, "lagerorte")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Kiste 4") {
		t.Errorf("»Kiste 4« fehlt: %q", aus)
	}

	// **Vier Stück in einer Zeile** — der Wunsch ist nicht dabei.
	if !strings.Contains(aus, "4\t1\t") {
		t.Errorf("die Zahlen stimmen nicht: %q", aus)
	}
}

// **`dubletten` hängt an der Zahl, die mitgegeben wird.**
func TestDublettenHaengenAnDerZahl(t *testing.T) {
	ort := geordnet(t)

	for _, fall := range []struct {
		behalten string
		erwartet string
	}{
		{behalten: "1", erwartet: "3\t4\t"},
		{behalten: "3", erwartet: "1\t4\t"},
	} {
		code, aus, fehler := fahreAn(t, ort, "dubletten", fall.behalten)
		if code != CodeGut {
			t.Fatalf("Rückgabewert %d (%s)", code, fehler)
		}

		if !strings.Contains(aus, fall.erwartet) {
			t.Errorf("behalten %s: %q enthält %q nicht",
				fall.behalten, aus, fall.erwartet)
		}
	}

	// **Ohne Zahl gilt 1.**
	if _, aus, _ := fahreAn(t, ort, "dubletten"); !strings.Contains(aus, "3\t4\t") {
		t.Errorf("die Vorgabe greift nicht: %q", aus)
	}
}

// **Eine Zahl, die keine ist, ist ein Bedienfehler.**
func TestDublettenMitUnbrauchbarerZahl(t *testing.T) {
	if code, _, _ := fahreAn(t, eigenesZuhause(t),
		"dubletten", "viele"); code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
	}
}

// **Eine Zahl unter null ist ein fachliches Nein.**
func TestDublettenUnterNull(t *testing.T) {
	if code, _, _ := fahreAn(t, eigenesZuhause(t),
		"dubletten", "-1"); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}

// **`wuensche` gibt nur, was kein Besitz ist.**
func TestWuenscheGebenNurWuensche(t *testing.T) {
	ort := geordnet(t)

	code, aus, fehler := fahreAn(t, ort, "wuensche")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "played") {
		t.Errorf("der Wunsch fehlt: %q", aus)
	}

	// **Und der Besitz steht nicht dabei.**
	if strings.Contains(aus, "nearMint") {
		t.Errorf("der Besitz steht in der Wunschliste: %q", aus)
	}
}

// **Alle vier gehen auch als JSON.**
func TestDieVierFragenAlsJSON(t *testing.T) {
	ort := geordnet(t)

	for _, fall := range []struct {
		befehl   []string
		erwartet string
	}{
		{befehl: []string{"besitz", "probe-druck"}, erwartet: `"gewuenscht": 2`},
		{befehl: []string{"lagerorte"}, erwartet: `"lagerort": "Kiste 4"`},
		{befehl: []string{"dubletten", "1"}, erwartet: `"ueberzaehlig": 3`},
		{befehl: []string{"wuensche"}, erwartet: `"wunsch": true`},
	} {
		code, aus, fehler := fahreAn(t, ort, append(fall.befehl, "--json")...)
		if code != CodeGut {
			t.Fatalf("%v: Rückgabewert %d (%s)", fall.befehl, code, fehler)
		}

		if !strings.Contains(aus, fall.erwartet) {
			t.Errorf("%v: %q enthält %q nicht", fall.befehl, aus, fall.erwartet)
		}
	}
}
