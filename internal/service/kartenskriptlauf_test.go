package service_test

// **`new_card` — Karten aus einem Skript** — Werkstück B von Phase 4.11
// (#1337).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `src/script/functions/construction.cpp`, Zeilen 23–52 (Fassung f9d9356 vom
// 2021-07-06). Die Funktion baut eine Karte aus dem **Spiel** — also mit allen
// Feldern und ihren Vorgaben — und setzt darauf die übergebenen Werte:
//
//	IndexMap<FieldP,ValueP>::const_iterator value_it = new_card->data.find(name);
//	if (value_it == new_card->data.end()) {
//	  throw ScriptError(format_string(_("Card doesn't have a field named '%s'"),name));
//	}
//
// **Ein unbekannter Feldname ist ein Fehler**, kein stilles Übergehen. Und
// gesetzt wird nur, was sich setzen lässt:
//
//	if (TextValue*) … else if (ChoiceValue*) … else if (PackageChoiceValue*) …
//	else if (ColorValue*) … else
//	  throw ScriptError(_("Can not set value '%s', it is not of the right type"));
//
// # Warum sie hier steht und nicht in `internal/msescript`
//
// **§2**: „`internal/msescript` importiert nichts aus dem Projekt, auch nicht
// `domain`. Es ist ein Sprachwerkzeug; eine Kartenentität darin wäre eine
// Vermischung, die den Interpreter unprüfbar machte."
//
// `new_card` braucht die **Felder des Spiels**, um einen Namen abzuweisen —
// das ist Projektwissen. Sie gehört damit zu den Funktionen, die `service` in
// die Umgebung gibt, wie `symbol_variation` und `exists_in_package`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probefelder sind die Felder eines kleinen Spiels.
func probefelder() []domain.Field {
	return []domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "rarity", Kind: domain.KindChoice, Default: "common"},
		{Name: "kunst", Kind: domain.KindImage},
	}
}

// **`new_card` gibt eine Karte mit den gesetzten Werten.**
func TestNewCardGibtEineKarte(t *testing.T) {
	t.Parallel()

	karten, befunde, fehler := service.KartenAusSkript(t.Context(),
		`[ new_card([name: "Plains", rarity: "basic land"]) ]`, probefelder(), "")
	if fehler != nil {
		t.Fatalf("KartenAusSkript: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet war eine", len(karten))
	}

	werte := map[string]string{}
	for _, wert := range karten[0].Values {
		werte[wert.Field] = wert.Text
	}

	if werte["name"] != "Plains" || werte["rarity"] != "basic land" {
		t.Errorf("die Werte lauten %v", werte)
	}
}

// **Mehrere Karten in einer Liste.**
//
// So schreibt `magic.mse-game` seine Grundlande: fünf `new_card`-Aufrufe in
// einer Liste.
func TestNewCardGibtMehrereKarten(t *testing.T) {
	t.Parallel()

	karten, _, fehler := service.KartenAusSkript(t.Context(),
		`[ new_card([name: "Plains"]), new_card([name: "Island"]) ]`, probefelder(), "")
	if fehler != nil {
		t.Fatalf("KartenAusSkript: %v", fehler)
	}

	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet waren zwei", len(karten))
	}

	if karten[0].Values[0].Text != "Plains" || karten[1].Values[0].Text != "Island" {
		t.Errorf("die Karten lauten %+v und %+v", karten[0].Values, karten[1].Values)
	}
}

// **Ein unbekannter Feldname ist ein Fehler** — so wie im Altcode.
//
// **Nicht ein stilles Übergehen.** Ein Skript, das `nmae:` schreibt, legte
// sonst eine Karte ohne Namen an, und niemand sähe warum.
func TestNewCardWeistEinenUnbekanntenFeldnamenAb(t *testing.T) {
	t.Parallel()

	_, _, fehler := service.KartenAusSkript(t.Context(),
		`[ new_card([nmae: "Plains"]) ]`, probefelder(), "")
	if fehler == nil {
		t.Fatal("der unbekannte Feldname ging durch")
	}

	if !strings.Contains(fehler.Error(), "nmae") {
		t.Errorf("die Meldung nennt das Feld nicht: %v", fehler)
	}
}

// **Ein Feld, das sich nicht setzen lässt, ebenso.**
//
// `new_card` setzt Text, Wahl, Paketwahl und Farbe — mehr nicht. Ein Bildfeld
// trägt einen Dateinamen, den es beim Anlegen noch nicht gibt.
func TestNewCardWeistEinUnsetzbaresFeldAb(t *testing.T) {
	t.Parallel()

	_, _, fehler := service.KartenAusSkript(t.Context(),
		`[ new_card([kunst: "bild.png"]) ]`, probefelder(), "")
	if fehler == nil {
		t.Fatal("das Bildfeld ließ sich setzen")
	}

	if !strings.Contains(fehler.Error(), "kunst") {
		t.Errorf("die Meldung nennt das Feld nicht: %v", fehler)
	}
}

// **Ein Skript, das keine Liste gibt, ist ein Fehler.**
func TestEinSkriptOhneListeIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, _, fehler := service.KartenAusSkript(t.Context(),
		`"keine Liste"`, probefelder(), ""); fehler == nil {
		t.Error("die Zeichenkette ging als Kartenliste durch")
	}
}

// **Und der Interpreter kennt `new_card` von sich aus nicht.**
//
// Die Gegenprobe zu §2: die Funktion gehört `internal/service`, nicht der
// Standardbibliothek. Wer sie dorthin schöbe, hätte eine Kartenentität im
// Sprachwerkzeug.
func TestDerInterpreterKenntNewCardNichtVonSelbst(t *testing.T) {
	t.Parallel()

	_, fehler := msescript.AuswertenQuelle(t.Context(),
		`new_card([name: "Plains"])`, msescript.NeueUmgebung(map[string]msescript.Wert{}))
	if fehler == nil {
		t.Error("die Standardbibliothek kennt new_card")
	}
}

// **Die vier Wege, auf denen ein Kartenskript scheitert.**
//
// Sie stehen zusammen, weil sie dieselbe Frage beantworten: *was ist keine
// Karte?* — und weil §3 keinen Zweig zulässt, den kein Lauf erreicht.
func TestEinKartenskriptScheitertBenannt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		quelle string
		nennt  string
	}{
		// **Ein Ausdruck, der gar nicht aufgeht.**
		"unlesbar": {`[ new_card(`, "Kartenskript"},

		// **Eine Liste, deren Glied keine Abbildung ist.** `[ 1, 2 ]` sind
		// Zahlen; MSE2 nähme sie als Karten und stürbe beim ersten Zugriff.
		"eine Zahl in der Liste": {`[ 1 ]`, "keine Karte"},

		// **`new_card` mit etwas anderem als einer Abbildung.**
		"new_card mit Text": {`[ new_card("Plains") ]`, "Abbildung"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := service.KartenAusSkript(t.Context(), fall.quelle, probefelder(), "")
			if fehler == nil {
				t.Fatalf("%q ging durch", fall.quelle)
			}

			if !strings.Contains(fehler.Error(), fall.nennt) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.nennt, fehler)
			}
		})
	}
}

// **`new_card()` ohne Argument gibt eine leere Karte.**
//
// So steht es in `add_cards_scripts` — der auskommentierte Eintrag
// `#Many Empty Cards` ruft es genau so. Ein Fehler wäre hier falsch: eine
// leere Karte ist eine gültige Karte.
func TestNewCardOhneArgumentGibtEineLeereKarte(t *testing.T) {
	t.Parallel()

	karten, _, fehler := service.KartenAusSkript(t.Context(),
		`[ new_card() ]`, probefelder(), "")
	if fehler != nil {
		t.Fatalf("KartenAusSkript: %v", fehler)
	}

	if len(karten) != 1 || len(karten[0].Values) != 0 {
		t.Errorf("die Karten lauten %+v", karten)
	}
}

// **Ein Feldname in Skriptschreibweise landet in der Dateischreibweise.**
//
// Das Skript schreibt `casting_cost`, die `.mse-set` trägt `casting cost`.
// Gemessen am 2026-08-31: ohne diese Umsetzung liefen **null von zwölf**
// echten Skripten — sie scheiterten an `super_type` und `casting_cost`.
func TestEinFeldnameKommtInDerDateischreibweiseAn(t *testing.T) {
	t.Parallel()

	felder := append(probefelder(),
		domain.Field{Name: "casting cost", Kind: domain.KindText})

	karten, _, fehler := service.KartenAusSkript(t.Context(),
		`[ new_card([casting_cost: "W"]) ]`, felder, "")
	if fehler != nil {
		t.Fatalf("KartenAusSkript: %v", fehler)
	}

	if karten[0].Values[0].Field != "casting cost" {
		t.Errorf("das Feld heißt %q", karten[0].Values[0].Field)
	}
}
