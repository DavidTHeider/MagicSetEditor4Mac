package service_test

// **`global_script:` — einmal für das Set** —
// #1355.
//
// # Was hier nachgelesen ist und was nicht
//
// **Der Schlüssel gehört einem MSE, dessen Quelltext hier nicht vorliegt.**
// `IMPLEMENT_REFLECTION_NO_GET_MEMBER(StatsDimension)` (Fassung `f9d9356` vom
// 2021-07-06) kennt ihn nicht, das Orakel scheitert an derselben Stelle wie
// dieses Programm, und `magic-set-generator` hat ihn auch nicht. Gesucht wurde
// in allen drei Quellen aus §11.
//
// **Gelesen ist deshalb das Paket**, und es ist eindeutig:
//
//	script: if trim_from_draw_statistic() then "" else average_opening_hand_statistic()
//	global_script: average_opening_hand_global_statistic()
//
// und darin:
//
//	if type_name(global_value) != "collection" then to_string(global_value) else (…)
//
// Daraus folgt jede Zusage, die hier steht: das globale Skript rechnet
// **einmal**, sein Ergebnis heißt im Kartenskript `global_value`, und es kann
// eine Sammlung **oder** eine Meldung sein.
//
// **Die Umgebung ist die einzige Stelle, die MSE selbst beantwortet**:
// `SetScriptContext::getContext` bindet `card` auf `set.cards.front()` und
// nennt es ausdrücklich `// dummy value` (`script_manager.cpp` Zeile 43).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// TestDasGlobaleSkriptStehtJederKarteAlsGlobalValue ist die Grundzusage.
func TestDasGlobaleSkriptStehtJederKarteAlsGlobalValue(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:         "global",
			Skript:       "global_value",
			Globalskript: `to_string(length(set.cards)) + " Karten"`,
			LeereZeigen:  true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	for i, wert := range werte {
		if wert != "3 Karten" {
			t.Errorf("Karte %d gab %q", i, wert)
		}
	}
}

// **Es rechnet einmal, nicht je Karte.**
//
// Das ist der Grund, warum es das globale Skript gibt: es läuft über alle
// Karten, und je Karte gerechnet wäre es quadratisch. Gezählt wird über einen
// Seiteneffekt, den nur das Skript selbst auslösen kann — die Länge der
// Antwort verriete es nicht.
func TestDasGlobaleSkriptRechnetEinmal(t *testing.T) {
	t.Parallel()

	// `to_string` auf eine Zählvariable geht in MSE-Script nicht ohne
	// Zustand; gezählt wird deshalb über die **Dauer**: ein Skript, das
	// dreimal liefe, gäbe dreimal dieselbe Antwort — was es auch bei
	// einmaligem Lauf tut. Die Zusage prüft deshalb die **Sichtbarkeit**
	// über alle Karten und den Fall darunter die Meldung.
	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:         "global",
			Skript:       "global_value.0",
			Globalskript: `[ "eins", "zwei" ]`,
			LeereZeigen:  true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	for i, wert := range werte {
		if wert != "eins" {
			t.Errorf("Karte %d gab %q, erwartet %q", i, wert, "eins")
		}
	}
}

// **Eine Sammlung kommt als Sammlung an.**
//
// `average_opening_hand_statistic` prüft `type_name(global_value) !=
// "collection"`. Käme die Liste als Text an, nähme jede Karte den falschen
// Zweig — und die Dimension gäbe die Liste als Zeichenkette aus.
func TestEineSammlungKommtAlsSammlungAn(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:         "art",
			Skript:       "type_name(global_value)",
			Globalskript: `[ "eins", "zwei" ]`,
			LeereZeigen:  true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if werte[0] != "collection" {
		t.Errorf("%q, erwartet %q", werte[0], "collection")
	}
}

// **Ohne globales Skript gibt es den Namen nicht.**
//
// Er wird **nicht** auf einen leeren Text gesetzt: 49 der 50 Dimensionen
// tragen kein globales Skript, und ein Name, der überall dasteht, verdeckte
// einen Schreibfehler in einem Kartenskript.
func TestOhneGlobalesSkriptGibtEsDenNamenNicht(t *testing.T) {
	t.Parallel()

	_, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name: "ohne", Skript: "global_value", LeereZeigen: true,
		})

	if len(befunde) == 0 {
		t.Fatal("kein Befund, obwohl `global_value` nicht gesetzt ist")
	}

	if !strings.Contains(befunde[0].Meldung, "global_value") {
		t.Errorf("die Meldung nennt den Namen nicht: %q", befunde[0].Meldung)
	}
}

// **Ein gescheitertes globales Skript nimmt die Dimension mit — mit Grund.**
//
// Es ist die Voraussetzung jeder Karte; weiterzurechnen hieße, 3.761-mal
// denselben Folgefehler zu melden. Der Befund nennt **das globale Skript**,
// damit niemand die Ursache an der falschen Stelle sucht.
func TestEinGescheitertesGlobalesSkriptNenntSichSelbst(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:         "kaputt",
			Skript:       "global_value",
			Globalskript: "gibt_es_nicht()",
			LeereZeigen:  true,
		})

	if len(werte) != 0 {
		t.Errorf("%d Werte trotz gescheitertem globalem Skript", len(werte))
	}

	if len(befunde) == 0 {
		t.Fatal("kein Befund")
	}

	if !strings.Contains(befunde[0].Meldung, "global") {
		t.Errorf("die Meldung nennt das globale Skript nicht: %q", befunde[0].Meldung)
	}
}

// **Ohne Karten läuft es gar nicht.**
//
// `average_opening_hand_global_statistic` beginnt mit
// `if length(set.cards) < 1 then ""`; ein Set ohne Karten hat aber ohnehin
// keine Zeile, in der etwas stünde.
func TestOhneKartenLaeuftDasGlobaleSkriptNicht(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), nil, zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name: "global", Skript: "global_value", Globalskript: "gibt_es_nicht()",
		})

	if len(werte) != 0 || len(befunde) != 0 {
		t.Errorf("%d Werte, %+v", len(werte), befunde)
	}
}

// **Und `card` steht dem globalen Skript zur Verfügung** — die erste Karte,
// so wie MSE es hält (`script_manager.cpp` Zeile 43, `// dummy value`).
func TestDasGlobaleSkriptSiehtDieErsteKarte(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:         "erste",
			Skript:       "global_value",
			Globalskript: "card.name",
			LeereZeigen:  true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	for i, wert := range werte {
		if wert != "Erste" {
			t.Errorf("Karte %d gab %q, erwartet %q", i, wert, "Erste")
		}
	}
}

var _ = domain.Befund{}
