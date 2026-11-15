package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Arten stehen in einer Reihenfolge, und sie ist nicht alphabetisch.**
//
//	enum PackageType {
//	  TYPE_PROG,
//	  TYPE_LOCALE,
//	  TYPE_GAME,
//	  TYPE_STYLESHEET,
//	  TYPE_EXPORT_TEMPLATE,
//	  TYPE_SYMBOL_FONT,
//	  TYPE_INCLUDE,
//	  TYPE_FONT,
//	  TYPE_OTHER,
//	}
//
// `src/gui/package_update_list.hpp`, Zeilen 57–67, Fassung f9d9356 vom
// 2021-07-06. `package_type` liest sie aus der **Endung** des Paketnamens
// (`package_update_list.cpp`, Zeilen 83–95), und `compare_pos_hint` ordnet
// zuerst danach (Zeilen 61–67).
//
// **Sie ist eine Aussage darüber, was zuerst gebraucht wird**: ohne Spiel kein
// Stil, ohne Stil keine Symbolschrift. Alphabetisch stünde `mse-export-
// template` vor `mse-game`, und wer eine Vorlage sucht, fände zuerst die
// Ausgabe.
//
// **Zwei der neun kommen hier nicht vor.** `TYPE_PROG` ist MSE selbst — das
// Programm verwaltet sich in MSE2 mit; hier tut es das nicht (§12, kein
// Selbstaktualisieren). `TYPE_FONT` ist eine lose `.ttf`; dieses Projekt führt
// Schriften nicht als Paket, sondern findet sie über `internal/schriften`.
func TestDieArtenStehenInMSEsReihenfolge(t *testing.T) {
	t.Parallel()

	// **Als Paare und nicht als Zahlen.** Eine Erwartung `Artrang() == 2`
	// prüfte die Zahl; geprüft werden soll die *Ordnung*, und die überlebt es,
	// wenn jemand eine achte Art dazwischenschiebt.
	reihe := []string{
		"mse-locale",
		"mse-game",
		"mse-style",
		"mse-export-template",
		"mse-symbol-font",
		"mse-include",
	}

	for i := 1; i < len(reihe); i++ {
		vorher := domain.Paket{Art: reihe[i-1]}
		nachher := domain.Paket{Art: reihe[i]}

		if vorher.Artrang() >= nachher.Artrang() {
			t.Errorf("%s (%d) steht nicht vor %s (%d)",
				reihe[i-1], vorher.Artrang(), reihe[i], nachher.Artrang())
		}
	}
}

// **Was keine der Arten ist, steht hinten** — `TYPE_OTHER` ist der letzte Wert
// des Enums, und `package_type` gibt ihn für jeden unbekannten Namen zurück.
func TestEineUnbekannteArtStehtHinten(t *testing.T) {
	t.Parallel()

	bekannt := domain.Paket{Art: "mse-include"}
	fremd := domain.Paket{Art: "mse-erfunden"}

	if fremd.Artrang() <= bekannt.Artrang() {
		t.Errorf("fremd = %d, bekannt = %d", fremd.Artrang(), bekannt.Artrang())
	}
}

// **Ein Set ist kein Katalogeintrag** — und trotzdem beantwortet der Rang die
// Frage, statt zu entscheiden, ob sie gestellt werden durfte. `paketlesen`
// weist Sets vorher ab; hier wäre eine zweite Prüfung eine zweite Wahrheit.
func TestEinSetBekommtDenRangDerUebrigen(t *testing.T) {
	t.Parallel()

	set := domain.Paket{Art: "mse-set"}

	if set.Artrang() != domain.ArtrangUebrige {
		t.Errorf("Artrang = %d", set.Artrang())
	}
}

// **Eine Paketliste steht in MSE2s Ordnung** — der kleinere `position hint:`
// zuerst, bei Gleichstand der Name (`PackageList::ComparePackagePosHint`,
// `src/gui/control/package_list.cpp`, Zeilen 52–60, Fassung f9d9356 vom
// 2021-07-06). Nach ihr stehen Spiele und Stile beim Anlegen eines Sets, und
// der erste ist der vorgewählte (UX-05, #1677).
func TestPaketeStehenNachRangUndName(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name string
		a, b domain.Paket
		ist  int
	}{
		{
			"der kleinere Rang zuerst, auch gegen den Namen",
			domain.Paket{Name: "magic-m15-altered", Rang: 1},
			domain.Paket{Name: "magic-DayNight", Rang: 34},
			-1,
		},
		{
			"ohne Hinweis nach hinten",
			domain.Paket{Name: "aaa", Rang: 1000000},
			domain.Paket{Name: "zzz", Rang: 2},
			1,
		},
		{
			"gleicher Rang: der Name entscheidet",
			domain.Paket{Name: "magic-m15-adventure", Rang: 11},
			domain.Paket{Name: "magic-m15-aftermath", Rang: 11},
			-1,
		},
		// **Nach Codeeinheiten, wie `wxString` vergleicht**: das große D steht
		// vor dem kleinen c — ohne Rücksicht auf Groß und klein stünde es
		// dahinter.
		{
			"groß vor klein",
			domain.Paket{Name: "magic-DayNight", Rang: 10},
			domain.Paket{Name: "magic-custom", Rang: 10},
			-1,
		},
		{"dasselbe", domain.Paket{Name: "probe", Rang: 5}, domain.Paket{Name: "probe", Rang: 5}, 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ist := domain.VergleicheNachRang(fall.a, fall.b); ist != fall.ist {
				t.Errorf("VergleicheNachRang(%s, %s) = %d, erwartet %d", fall.a.Name, fall.b.Name, ist, fall.ist)
			}
		})
	}
}
