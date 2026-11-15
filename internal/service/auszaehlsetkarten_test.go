package service_test

// **Eine Dimension darf das ganze Set sehen** —
// #1356.
//
// # Der Befund
//
// Zwei Dimensionen von `magic.mse-game` — `land drop hit % (On The Play)` und
// `… (On The Draw)` — rechnen über **alle** Karten und geben ihr Ergebnis nur
// an der ersten aus:
//
//	if card != set.cards.0 then "" else (
//	  set_count := length(set.cards)-1
//	  cards := for x from 0 to set_count do (…)
//
// # Warum MSE das kann
//
// `SetScriptContext::getContext` bindet **das ganze Set-Objekt**
// (`src/script/script_manager.cpp` Zeile 39, Fassung `f9d9356` vom
// 2021-07-06), und `cards` löst über `mark_dependency_member`
// (`src/data/set.cpp` Zeilen 256–259) auf die Kartenliste auf. Das gilt in
// **jedem** Kontext, nicht nur in der Ausfuhr.
//
// # Und warum `card != set.cards.0` hier überhaupt trägt
//
// MSE vergleicht `ScriptObject<CardP>` über den **Zeiger**
// (`to_value.hpp`, Zeilen 184–187). Hier gibt es keine Zeiger; verglichen
// werden Abbildungen über Namen und Werte (`gleich`, §271). Damit die
// Antwort dieselbe ist, muss `set.cards.<n>` **dieselbe** Abbildung sein, die
// die n-te Karte als `card` bekommt — nicht eine zweite, aus rohen Feldern
// gebaute. Genau das prüft der zweite Fall hier.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dreiZaehlkarten gibt drei unterscheidbare Karten.
func dreiZaehlkarten() []*domain.CardRow {
	return []*domain.CardRow{
		{ID: 1, Number: 0, Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		{ID: 2, Number: 1, Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
		{ID: 3, Number: 2, Values: []domain.Value{{Field: "name", Text: "Dritte"}}},
	}
}

// zaehlvorlage ist eine Vorlage mit einem Namensfeld und ohne Skripte.
func zaehlvorlage() service.Vorlage {
	return service.Vorlage{
		Kartenfelder: []domain.Field{{Name: "name"}},
	}
}

// TestEineDimensionSiehtDasGanzeSet ist die Grundzusage.
func TestEineDimensionSiehtDasGanzeSet(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{Name: "zahl", Skript: "length(set.cards)"})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	for i, wert := range werte {
		if wert != "3" {
			t.Errorf("Karte %d gab %q, erwartet %q", i, wert, "3")
		}
	}
}

// **Und `set.cards.<n>` ist dieselbe Karte, die `card` an Stelle n ist.**
//
// Das ist die Zusage, an der die beiden echten Dimensionen hängen: ohne sie
// wäre `card != set.cards.0` für **jede** Karte wahr, jede gäbe `""`, und die
// Auszählung bliebe still leer — bei grünem Lauf.
func TestDieErsteKarteErkenntSichSelbst(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:        "erste",
			Skript:      `if card != set.cards.0 then "" else "ich bin die erste"`,
			LeereZeigen: true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != 3 {
		t.Fatalf("%d Werte, erwartet 3", len(werte))
	}

	if werte[0] != "ich bin die erste" {
		t.Errorf("die erste Karte gab %q", werte[0])
	}

	for i := 1; i < 3; i++ {
		if werte[i] != "" {
			t.Errorf("Karte %d gab %q, erwartet leer", i, werte[i])
		}
	}
}

// **Über alle Karten rechnen und einmal ausgeben** — die Gestalt, die die
// beiden echten Dimensionen haben.
func TestEineDimensionRechnetUeberAlleKarten(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name: "namen",
			Skript: `if card != set.cards.0 then "" else (
				for x from 0 to length(set.cards)-1 do set.cards[x].name
			)`,
			LeereZeigen: true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if werte[0] != "ErsteZweiteDritte" {
		t.Errorf("%q, erwartet %q", werte[0], "ErsteZweiteDritte")
	}
}

// **Zwei strukturgleiche Karten sind trotzdem zwei Karten.**
//
// # Der Befund
//
// Angestrichen als Befund 1 des Zweitblicks über Phase 4.11. Die Kartenabbildung
// trug keine Identität, und `gleich` vergleicht Namen und Werte — zwei Karten
// mit denselben Feldern galten damit als **dieselbe**. Der Ausdruck
//
//	if card != set.cards.0 then "" else (…)
//
// lief an jeder von ihnen, und eine Auszählung zählte still doppelt.
//
// **Der Fall stand in keinem Lauf**: `dreiZaehlkarten` gibt drei
// unterscheidbare Namen. Genau das ist die Gestalt aus §10 — eine Gegenprobe,
// die ihren Fall nicht enthält, prüft ihn nicht.
//
// MSE vergleicht `ScriptObject<CardP>` über den **Zeiger**
// (`src/script/to_value.hpp` Zeilen 184–187, Fassung `f9d9356` vom
// 2021-07-06); hier trägt die Abbildung eine Kennung.
func TestZweiStrukturgleicheKartenSindZweiKarten(t *testing.T) {
	t.Parallel()

	gleiche := []*domain.CardRow{
		{ID: 1, Number: 0, Values: []domain.Value{{Field: "name", Text: "Wald"}}},
		{ID: 2, Number: 1, Values: []domain.Value{{Field: "name", Text: "Wald"}}},
		{ID: 3, Number: 2, Values: []domain.Value{{Field: "name", Text: "Wald"}}},
	}

	werte, befunde := service.Auszaehlwerte(t.Context(), gleiche, zaehlvorlage(),
		nil, service.Auszaehlangabe{
			Name:        "einmal",
			Skript:      `if card != set.cards.0 then "" else "einmal"`,
			LeereZeigen: true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	gezaehlt := 0

	for _, wert := range werte {
		if wert != "" {
			gezaehlt++
		}
	}

	if gezaehlt != 1 {
		t.Errorf("%d von 3 strukturgleichen Karten gaben etwas — erwartet genau eine (%q)",
			gezaehlt, werte)
	}
}

// **Ohne Karten gibt es keine Liste, aber auch keinen Fehler.**
func TestOhneKartenIstDieListeLeer(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), nil, zaehlvorlage(),
		nil, service.Auszaehlangabe{Name: "zahl", Skript: "length(set.cards)"})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != 0 {
		t.Errorf("%d Werte ohne Karten", len(werte))
	}
}

// **Und die Setfelder stehen weiter da** — `set.cards` tritt neben sie, nicht
// an ihre Stelle.
func TestDieSetfelderBleibenNebenDenKarten(t *testing.T) {
	t.Parallel()

	vorlage := zaehlvorlage()
	vorlage.Setfelder = []domain.Field{{Name: "title"}}

	werte, befunde := service.Auszaehlwerte(t.Context(), dreiZaehlkarten(), vorlage,
		nil, service.Auszaehlangabe{
			Name:        "beides",
			Skript:      `to_string(length(set.cards)) + "/" + to_string(set.title)`,
			LeereZeigen: true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if !strings.HasPrefix(werte[0], "3/") {
		t.Errorf("%q — erwartet wurde „3/…"+`"`, werte[0])
	}
}

// **Eine Umgebung ohne Karte bekommt keine Kennung.**
//
// Der Rasterer baut die Kartenabbildung auch für eine **Attrappe** — dort gibt
// es keine Karte, und dann entscheidet wieder die Gestalt. Ohne diesen Zweig
// stünde in jeder Attrappenabbildung `karte:0`, und zwei Attrappen wären
// dieselbe.
func TestOhneKarteGibtEsKeineKennung(t *testing.T) {
	t.Parallel()

	if gefunden := service.Skriptkartenkennung(nil); gefunden != "" {
		t.Errorf("%q, erwartet leer", gefunden)
	}

	if gefunden := service.Skriptkartenkennung(
		&domain.CardRow{ID: 7}); gefunden != "karte:7" {
		t.Errorf("%q, erwartet %q", gefunden, "karte:7")
	}
}
