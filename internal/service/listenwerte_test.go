// **Eine berechnete Spalte bleibt nicht leer** —
// #1238.
//
// # Der Befund
//
// Die Spaltenwahl bietet **jedes** Feld des Spiels an; `zeilenAus` füllte die
// Zellen aber ausschließlich aus `CardRow.Values`. In `magic.mse-game` sind
// zahlreiche sichtbare Listenfelder **nicht gespeichert, sondern gerechnet**
// — `full name`, `cmc`, `type`, `pt`, `card number`; sie tragen
// `save value: false` und entstehen erst aus ihrem `script:`.
//
// Was Nutzende sahen: eine Spalte ließ sich wählen, und die Zellen blieben
// leer. Eine Sortierung danach ordnete nach lauter leeren Schlüsseln.
//
// # Warum nicht der ganze Planbau
//
// `umgebungAus` rechnet die Werte längst — mit Stil, Schriften und Bildern,
// und das kostet je Karte Millisekunden. Bei 3.788 Karten eines kanonischen
// Sets wäre die Liste damit unbenutzbar.
//
// Gerechnet wird deshalb **nur, was eine sichtbare Spalte braucht**, und in
// einer Umgebung ohne Stil: die Werte der Karte, ihre Abbildung als `card`,
// und das Wertskript des Feldes. Was darin nicht aufgeht, bleibt leer — wie
// vorher, aber eben nur noch das.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// listenaufbau baut die Liste über einem Doppel aus Feldern und einer Karte.
func listenaufbau(
	t *testing.T, felder []domain.Field, karten ...*domain.CardRow,
) domain.Kartenlistenaufbau {
	t.Helper()

	aufbau, fehler := kartenliste(&listendoppel{
		spiel:   "magic",
		ordnung: domain.StandardKartenordnung(),
		karten:  karten,
		felder:  felder,
	}).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	return aufbau
}

// listenaufbauSortiert baut die Liste mit einer Ordnung nach einem Feld.
func listenaufbauSortiert(
	t *testing.T, felder []domain.Field, nach string, karten ...*domain.CardRow,
) domain.Kartenlistenaufbau {
	t.Helper()

	aufbau, fehler := kartenliste(&listendoppel{
		spiel:   "magic",
		ordnung: domain.Kartenordnung{Feld: nach, Aufsteigend: true},
		karten:  karten,
		felder:  felder,
	}).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	return aufbau
}

func TestEineBerechneteSpalteWirdGefuellt(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{
		{Name: "name", Kind: domain.KindText, Listenangaben: domain.Listenangaben{
			Listenerlaubt: true, Listensichtbar: true, Listenstelle: 100,
		}},
		{
			// **`full name` aus `magic.mse-game`**, vereinfacht: es setzt zwei
			// Namen zusammen und wird nie gespeichert.
			Name:   "full name",
			Kind:   domain.KindText,
			Skript: `card.name + " // " + card.name_2`,
			Listenangaben: domain.Listenangaben{
				Listenerlaubt: true, Listensichtbar: true, Listenstelle: 200,
			},
		},
		{Name: "name 2", Kind: domain.KindText, Listenangaben: domain.Listenangaben{
			Listenerlaubt: true, Listenstelle: 300,
		}},
	}

	karte := &domain.CardRow{
		ID: 1, Number: 1,
		Values: []domain.Value{
			{Field: "name", Text: "Fire"},
			{Field: "name 2", Text: "Ice"},
		},
	}

	aufbau := listenaufbau(t, felder, karte)

	if got := aufbau.Zeilen[0].Werte["full name"]; got != "Fire // Ice" {
		t.Errorf("die Zelle ist %q, erwartet \"Fire // Ice\" — eine Spalte, "+
			"die sich wählen lässt und leer bleibt, ist eine Falle", got)
	}
}

// **Ein gespeicherter Wert schlägt das Skript.** Wer etwas eingetippt hat,
// soll es sehen — dieselbe Regel wie beim Vorgabewert.
func TestEinGespeicherterWertSchlaegtDasSkript(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "name", Kind: domain.KindText, Skript: `"gerechnet"`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	karte := &domain.CardRow{
		ID: 1, Number: 1,
		Values: []domain.Value{{Field: "name", Text: "getippt"}},
	}

	if got := listenaufbau(t, felder, karte).Zeilen[0].Werte["name"]; got != "getippt" {
		t.Errorf("die Zelle ist %q, erwartet \"getippt\"", got)
	}
}

// **Ein Skript, das nicht aufgeht, lässt die Zelle leer** — und hält die
// Liste nicht auf (§8).
func TestEinGescheitertesSkriptLaesstDieZelleLeer(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "kaputt", Kind: domain.KindText, Skript: `gibtsnicht(1)`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	karte := &domain.CardRow{ID: 1, Number: 1}

	if got := listenaufbau(t, felder, karte).Zeilen[0].Werte["kaputt"]; got != "" {
		t.Errorf("die Zelle ist %q, erwartet leer", got)
	}
}

// **Eine unsichtbare Spalte wird nicht gerechnet.** Das ist die
// Sparsamkeitszusage: `magic.mse-game` führt 142 Wertskripte, und die Liste
// zeigt ein halbes Dutzend Spalten.
func TestEineUnsichtbareSpalteWirdNichtGerechnet(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{
		{Name: "name", Kind: domain.KindText, Listenangaben: domain.Listenangaben{
			Listenerlaubt: true, Listensichtbar: true,
		}},
		{
			Name: "teuer", Kind: domain.KindText, Skript: `"gerechnet"`,
			Listenangaben: domain.Listenangaben{
				Listenerlaubt: true, Listensichtbar: false, Listenstelle: 200,
			},
		},
	}

	aufbau := listenaufbau(t, felder, &domain.CardRow{ID: 1, Number: 1})

	if got := aufbau.Zeilen[0].Werte["teuer"]; got != "" {
		t.Errorf("die unsichtbare Spalte wurde gerechnet: %q", got)
	}
}

// **Ein `sort script:` ordnet nach seinem Ergebnis** —
// #1232.
//
// # Der Befund
//
//	inline String getSortKey() const {
//	  return fieldP->sort_script ? sort_value : toString();
//	}
//
// `src/data/field.hpp`, Zeilen 234–237, Fassung f9d9356 vom 2021-07-06. Wo
// ein Feld ein Sortierskript trägt, ordnet MSE nach dessen **Ergebnis** und
// nicht nach dem Wert. In `magic.mse-game` sind das **15** Felder — vier
// Farbfelder, vier Indikatoren, drei Seltenheiten und die Kartennummer.
//
// Sichtbar war es an der Seltenheit: geordnet wurde
// `common, mythic rare, rare, uncommon` statt in der Reihenfolge des Spiels.
func TestEinSortierskriptOrdnetNachSeinemErgebnis(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "rarity", Kind: domain.KindChoice,
		// **Die Ordnung des Spiels**, nicht das Alphabet.
		Sortierskript: `if value == "common" then "1" else if value == "uncommon" ` +
			`then "2" else if value == "rare" then "3" else "4"`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	karten := []*domain.CardRow{
		{ID: 1, Number: 1, Values: []domain.Value{{Field: "rarity", Text: "mythic rare"}}},
		{ID: 2, Number: 2, Values: []domain.Value{{Field: "rarity", Text: "common"}}},
		{ID: 3, Number: 3, Values: []domain.Value{{Field: "rarity", Text: "rare"}}},
		{ID: 4, Number: 4, Values: []domain.Value{{Field: "rarity", Text: "uncommon"}}},
	}

	aufbau := listenaufbauSortiert(t, felder, "rarity", karten...)

	ordnung := make([]string, 0, len(aufbau.Zeilen))
	for _, zeile := range aufbau.Zeilen {
		ordnung = append(ordnung, zeile.Werte["rarity"])
	}

	erwartet := []string{"common", "uncommon", "rare", "mythic rare"}
	for i := range erwartet {
		if ordnung[i] != erwartet[i] {
			t.Fatalf("Reihenfolge %v, erwartet %v — alphabetisch geordnet "+
				"steht »common, mythic rare, rare, uncommon« da", ordnung, erwartet)
		}
	}
}

// **Und die Zelle zeigt weiterhin den Wert, nicht den Schlüssel.** Der
// Sortierschlüssel ist eine Ordnungsangabe und kein Text für Menschen.
func TestDerSortierschluesselStehtNichtInDerZelle(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "rarity", Kind: domain.KindChoice, Sortierskript: `"9"`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	karte := &domain.CardRow{
		ID: 1, Number: 1,
		Values: []domain.Value{{Field: "rarity", Text: "rare"}},
	}

	if got := listenaufbau(t, felder, karte).Zeilen[0].Werte["rarity"]; got != "rare" {
		t.Errorf("die Zelle ist %q, erwartet \"rare\"", got)
	}
}

// **Ein gescheitertes Sortierskript ordnet nach dem Wert.** Das ist MSEs
// Verhalten ohne Skript und die harmlose Richtung.
func TestEinGescheitertesSortierskriptOrdnetNachDemWert(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "rarity", Kind: domain.KindChoice, Sortierskript: `gibtsnicht(1)`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	karten := []*domain.CardRow{
		{ID: 1, Number: 1, Values: []domain.Value{{Field: "rarity", Text: "rare"}}},
		{ID: 2, Number: 2, Values: []domain.Value{{Field: "rarity", Text: "common"}}},
	}

	aufbau := listenaufbauSortiert(t, felder, "rarity", karten...)

	if erste := aufbau.Zeilen[0].Werte["rarity"]; erste != "common" {
		t.Errorf("die erste Zeile ist %q, erwartet \"common\" — ohne "+
			"brauchbaren Schlüssel gilt der Text", erste)
	}
}

// **Ein abgebrochener Kontext hält die Rechnung an — und die Liste ist dann
// keine.** Wer während des Ladens das Set wechselt, wartet sonst auf eine
// Liste, die niemand mehr sieht. Hier stand bis zum Zweitblick über Phase G,
// dass dabei eine Zeile mit leerer Zelle **als Erfolg** herauskommt; das ist
// ein halber Stand, der wie ein ganzer aussieht (Befund 3).
func TestEinAbbruchHaeltDieBerechnungAn(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "gerechnet", Kind: domain.KindText, Skript: `"da"`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	aufbau, fehler := kartenliste(&listendoppel{
		spiel:   "magic",
		ordnung: domain.StandardKartenordnung(),
		karten:  []*domain.CardRow{{ID: 1, Number: 1}},
		felder:  felder,
	}).Aufbau(ctx, 1)
	if !errors.Is(fehler, context.Canceled) {
		t.Errorf("Aufbau = %v, %v — erwartet den Abbruch", aufbau.Zeilen, fehler)
	}
}

// **Das Spielskript legt die Funktionen an, die die Feldskripte rufen.**
//
// Das ist der Kern von #1238: `full name` heißt `script: card_full_name()`,
// und diese Funktion kommt aus dem `init script:` des Spiels.
func TestDasSpielskriptStelltSeineFunktionenBereit(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{
		Name: "voller name", Kind: domain.KindText, Skript: `card_full_name()`,
		Listenangaben: domain.Listenangaben{Listenerlaubt: true, Listensichtbar: true},
	}}

	aufbau, fehler := kartenliste(&listendoppel{
		spiel:   "magic",
		ordnung: domain.StandardKartenordnung(),
		karten: []*domain.CardRow{{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "Fire"},
		}}},
		felder: append(felder, domain.Field{Name: "name", Kind: domain.KindText}),
		// **Und `set` führt seine Felder** — sonst bricht ein Spielskript, das
		// eines liest, an der Stelle ab und alles danach fehlt.
		setfelder:   []domain.Field{{Name: "custom super types"}},
		spielskript: `card_full_name := { set.custom_super_types; "voll: " + card.name }`,
	}).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if got := aufbau.Zeilen[0].Werte["voller name"]; got != "voll: Fire" {
		t.Errorf("die Zelle ist %q, erwartet \"voll: Fire\"", got)
	}
}
