package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein verketteter Block sieht die Namen, die bei seiner Entstehung galten.**
//
// # Der Befund
//
// `magic.mse-game` verkettet eine eingebaute Funktion mit einem **Block**:
//
//	mana_filter := to_upper + {
//		if chosen(set.mana_cost_sorting, choice: "unsorted") then mana_unsort()
//		…
//	}
//
// Bei MSE ist das eine `ScriptCompose`, und ihr rechter Teil wird im
// **laufenden Kontext** ausgewertet (`src/script/context.cpp`, Zeilen 383–400,
// Fassung f9d9356 vom 2021-07-06) — `set` steht dort.
//
// Dieses Programm rechnete den Block bis zum 2026-08-17 in einer **frischen**
// Umgebung aus, die nur die Standardbibliothek kannte. Jeder Name der Vorlage
// war darin unbekannt, und `mana_filter` scheiterte an `set` — 1.330-mal über
// die 190 Stile, dazu 5.076-mal an `inserts_count` aus derselben Ursache
// (#345).
func TestEinVerketteterBlockSiehtSeineUmgebung(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("zusatz", msescript.Text("!"))

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`(to_upper + { input + zusatz })("ab")`, umgebung)
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "AB!" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "AB!")
	}
}

// **Auch der linke Teil einer Verkettung sieht sie.**
func TestAuchDerLinkeTeilSiehtSeineUmgebung(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("zusatz", msescript.Text("!"))

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`({ input + zusatz } + to_upper)("ab")`, umgebung)
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "AB!" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "AB!")
	}
}

// **Und eine Kette aus drei Gliedern trägt sie durch.**
//
// `text_filter` reiht rund dreißig; die Umgebung muss jedes davon erreichen.
func TestEineLangeKetteTraegtDieUmgebungDurch(t *testing.T) {
	umgebung := msescript.Bibliothek().Darueber()
	umgebung.Setze("a", msescript.Text("1"))
	umgebung.Setze("b", msescript.Text("2"))

	wert, err := msescript.AuswertenQuelle(t.Context(),
		`({ input + a } + { input + b } + to_upper)("x")`, umgebung)
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "X12" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "X12")
	}
}
