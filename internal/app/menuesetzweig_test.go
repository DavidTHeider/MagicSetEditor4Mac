package app_test

// **Was die Menüleiste seit Werkstück 4 von Phase 3.7 führt** —
// (#1168,
// Befund Z4-01 im Audit 0.3.0).
//
// **Eine eigene Datei neben `menue_test.go`**, weil jene mit ihr über 400
// Zeilen ginge (§2). Der Schnitt folgt einer Frage: dort steht die Ablage,
// hier das Setmenü und was in der Hilfe dazugekommen ist.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// zweig gibt ein Untermenü an seiner Beschriftung — oder bricht ab.
func zweig(t *testing.T, leiste *app.Menue, name string) *app.Menue {
	t.Helper()

	eintrag := untermenueMitNamen(leiste, name)
	if eintrag == nil {
		t.Fatalf("es gibt kein Menü %q", name)
	}

	return eintrag.Untermenue
}

// fuerBeschriftungen prüft, dass jede genannte Beschriftung dasteht — **in
// dieser Reihenfolge**.
//
// **Die Reihenfolge ist die Aussage.** Ein Menü, das dieselben Einträge in
// anderer Folge führt, ist ein anderes Menü: MSE2 setzt Rückgängig vor
// Wiederherstellen und Suchen vor Ersetzen, und wer daran dreht, dreht an
// einer Gewohnheit.
func fuerBeschriftungen(t *testing.T, zweig *app.Menue, erwartet []string) {
	t.Helper()

	var gesehen []string

	for _, eintrag := range zweig.Eintraege {
		if eintrag.Beschriftung != "" {
			gesehen = append(gesehen, eintrag.Beschriftung)
		}
	}

	stelle := 0

	for _, label := range gesehen {
		if stelle < len(erwartet) && label == erwartet[stelle] {
			stelle++
		}
	}

	if stelle != len(erwartet) {
		t.Errorf("erwartet %v in dieser Reihenfolge, gesehen %v", erwartet, gesehen)
	}
}

// klickeAlle löst jeden Eintrag eines Menüs aus.
func klickeAlle(zweig *app.Menue) {
	for _, eintrag := range zweig.Eintraege {
		if eintrag.Handlung != nil {
			eintrag.Handlung()
		}
	}
}

// **Das Setmenü führt die Vorgänge, die MSE2 im Bearbeitenmenü hat.**
//
//	add_menu_item(menuEdit, ID_EDIT_UNDO, "undo", …);
//	add_menu_item(menuEdit, ID_EDIT_REDO, "redo", …);
//	…
//	add_menu_item_tr(menuEdit, ID_EDIT_FIND, "find", "find");
//	add_menu_item_tr(menuEdit, ID_EDIT_FIND_NEXT, nullptr, "find_next");
//	add_menu_item_tr(menuEdit, ID_EDIT_REPLACE, nullptr, "replace");
//
// `src/gui/set/window.cpp`, Zeilen 81–96, Fassung f9d9356 vom 2021-07-06.
func TestDasSetmenueFuehrtDieVorgaengeAusMSE2(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

	// **Rückgängig und Wiederherstellen zuerst**, wie bei MSE2.
	fuerBeschriftungen(t, zweig(t, leiste, texte.Set), []string{
		texte.Rueckgaengig, texte.Wiederherstellen,
		texte.Suchen, texte.Ersetzen,
	})
}

// **Kein Kürzel im Setmenü — und das ist der Punkt.**
//
// # Warum
//
// ⌘Z liegt bei _Bearbeiten → Widerrufen_, an der Responder-Kette von macOS,
// und das muss dableiben: ohne es gäbe es kein ⌘C und ⌘V. Ein zweites ⌘Z hier nähme
// AppKit die Wahl ab — und **immer** die fachliche Rücknahme gemeint, auch
// mitten in einem Textfeld. Genau das verbietet `rahmen/tastatur.ts`: „Im
// Textfeld gehört ⌘Z dem Textfeld."
//
// **Die Taste geht deshalb weiter durchs Fenster**, wo sie den Fokus kennt;
// der Eintrag hier ist das Verzeichnis, das die Leiste sein soll (Z4-01).
func TestDasSetmenueTraegtKeineKuerzel(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

	for _, eintrag := range zweig(t, leiste, texte.Set).Eintraege {
		if eintrag.Kuerzel != nil {
			t.Errorf("%q trägt ein Kürzel", eintrag.Beschriftung)
		}
	}
}

// **Die Ablage führt ein Untermenü „Ausgeben".**
//
//	add_menu_item_tr(menuFile, wxID_ANY, "export", "export", wxITEM_NORMAL, makeExportMenu());
//
// dieselbe Datei, Zeile 63.
func TestDieAblageFuehrtEinAusgabemenue(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

	ablage := zweig(t, leiste, texte.Ablage)

	var ausgeben *app.Menue

	for _, eintrag := range ablage.Eintraege {
		if eintrag.Beschriftung == texte.Ausgeben && eintrag.Untermenue != nil {
			ausgeben = eintrag.Untermenue
		}
	}

	if ausgeben == nil {
		t.Fatalf("es gibt kein Untermenü %q", texte.Ausgeben)
	}

	fuerBeschriftungen(t, ausgeben, []string{
		texte.Exportvorlage, texte.Bilder, texte.Druckbogen,
	})
}

// **Und „Über" gibt es — im Anwendungsmenü.**
//
//	add_menu_item_tr(menuHelp, ID_HELP_ABOUT, nullptr, "about");
//
// dieselbe Datei, Zeile 106: MSE2 hängt es an die Hilfe, aber unter der
// Kennung `wxID_ABOUT`, und die stellt wxWidgets unter macOS ins
// Anwendungsmenü — dort steht es seit #255 auch hier. AGENTS.md §11 sagt seit
// jeher, die Rechtelage stehe „in der Anwendung selbst unter »Über«" — **sie
// stand dort bis zum 2026-08-27 nicht**.
func TestDieHilfeFuehrtUeber(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

	fuerBeschriftungen(t, zweig(t, leiste, texte.Hilfe), []string{texte.HilfeOeffnen})

	if zweig(t, leiste, texte.Anwendung).Eintraege[0].Beschriftung != texte.Ueber {
		t.Errorf("das Anwendungsmenü beginnt nicht mit %q", texte.Ueber)
	}
}

// **Jeder neue Eintrag ruft seine Handlung.**
func TestJederNeueMenueeintragRuftSeineHandlung(t *testing.T) {
	t.Parallel()

	var (
		zurueck, vor, ueber bool
		seiten              []string
	)

	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{
		Rueckgaengig:     func() { zurueck = true },
		Wiederherstellen: func() { vor = true },
		UeberOeffnen:     func() { ueber = true },
		SeiteOeffnen:     func(seite string) { seiten = append(seiten, seite) },
	}, texte, nil)

	klickeAlle(zweig(t, leiste, texte.Set))
	klickeAlle(zweig(t, leiste, texte.Hilfe))
	klickeAlle(zweig(t, leiste, texte.Anwendung))

	for _, eintrag := range zweig(t, leiste, texte.Ablage).Eintraege {
		if eintrag.Untermenue != nil && eintrag.Beschriftung == texte.Ausgeben {
			klickeAlle(eintrag.Untermenue)
		}
	}

	if !zurueck || !vor || !ueber {
		t.Errorf("zurueck=%v vor=%v ueber=%v", zurueck, vor, ueber)
	}

	// **Die Seiten sind Seitenschlüssel und keine erfundenen Namen.** Was
	// hier steht, muss `navigation.ts` kennen — der Wächter daneben hält es.
	erwartet := []string{"suche", "ersetzen", "ausgeben", "bilder", "druckbogen"}
	if len(seiten) != len(erwartet) {
		t.Fatalf("%v", seiten)
	}

	for i, seite := range erwartet {
		if seiten[i] != seite {
			t.Errorf("Stelle %d ist %q, erwartet %q", i, seiten[i], seite)
		}
	}
}

// **Und ohne Handlung tut ein Eintrag nichts, statt abzustürzen.**
func TestEinNeuerEintragOhneHandlungTutNichts(t *testing.T) {
	t.Parallel()

	texte := app.MenuetexteFuer("de")
	leiste := app.Menueleiste(app.Menuhandlungen{}, texte, nil)

	klickeAlle(zweig(t, leiste, texte.Set))
	klickeAlle(zweig(t, leiste, texte.Hilfe))
}
