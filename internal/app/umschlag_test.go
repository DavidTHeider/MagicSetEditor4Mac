package app_test

import (
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Die Umschläge, wie sie beim Frontend ankommen — vollständig.
//
// # Warum das nicht die übrigen Tests schon leisten
//
// Sie prüfen jeweils **ein herausgelöstes Listenfeld** (`json.Marshal(antwort.Treffer)`)
// und belegen damit `[]` statt `null`. Was sie nicht belegen: die Marken des
// Umschlags selbst. Wer `json:"treffer"` in `json:"hits"` ändert, macht keinen
// Go-Test rot — und die Frontend-Tests auch nicht, weil deren Vorrichtung die
// Antwort **zusammensetzt**, statt sie zu serialisieren.
//
// Genau diese Naht nennt AGENTS.md §3: „Zwischen den Prüfebenen". Hier steht
// sie als Zeichenkette, und eine geänderte Marke ist ein Textunterschied.
func TestJederUmschlagKommtSoAn(t *testing.T) {
	faelle := []struct {
		name     string
		antwort  any
		erwartet string
	}{
		{
			"Sets",
			app.SetsAntwort{Sets: []app.SetDTO{}},
			`{"sets":[],"fehler":null}`,
		},
		{
			"Karten",
			app.KartenAntwort{Karten: []app.KarteDTO{}},
			`{"karten":[],"fehler":null}`,
		},
		{
			"Vorbehalte",
			app.VorbehalteAntwort{Vorbehalte: []app.VorbehaltDTO{}},
			`{"vorbehalte":[],"fehler":null}`,
		},
		{
			"Schreiben",
			app.SchreibAntwort{},
			`{"fehler":null}`,
		},
		{
			"Vorschau",
			app.VorschauAntwort{},
			`{"vorschau":null,"fehler":null}`,
		},
		{
			"Öffnen",
			app.OeffnenAntwort{Befunde: []string{}, Angelegt: []string{}},
			`{"setID":"","karten":0,"befunde":[],"schonOffen":false,` +
				`"angelegt":[],"unvollstaendig":false,"abgebrochen":false,"fehler":null}`,
		},
		{
			"Suche",
			app.SuchAntwort{Treffer: []app.TrefferDTO{}},
			`{"treffer":[],"gekuerzt":false,"fehler":null}`,
		},
		{
			"Suchfelder",
			app.FelderAntwort{Felder: []string{}},
			`{"felder":[],"fehler":null}`,
		},
		{
			"Massenänderung",
			app.MassenAntwort{Betroffen: []string{}},
			`{"betroffen":[],"vorgang":"","fehler":null}`,
		},
		{
			"Statistik",
			app.StatistikAntwort{Statistik: app.StatistikDTO{Belegung: []app.ZahlDTO{}}},
			`{"statistik":{"karten":0,"felder":0,"leere":0,"belegung":[]},"fehler":null}`,
		},
		{
			"Verteilung",
			app.VerteilungAntwort{Verteilung: []app.WertzahlDTO{}},
			`{"verteilung":[],"fehler":null}`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gebaut, err := json.Marshal(fall.antwort)
			if err != nil {
				t.Fatalf("serialisieren: %v", err)
			}

			if string(gebaut) != fall.erwartet {
				t.Errorf("%s\nerwartet war\n%s", gebaut, fall.erwartet)
			}
		})
	}
}

// Und die Bestandteile, die in den Umschlägen stecken.
func TestJedesDTOKommtSoAn(t *testing.T) {
	faelle := []struct {
		name     string
		wert     any
		erwartet string
	}{
		{
			"Set",
			app.SetDTO{ID: "1", Name: "n", Spiel: "s", Stil: "m15", Herkunft: "h", Karten: 3},
			`{"id":"1","name":"n","spiel":"s","stil":"m15","herkunft":"h","karten":3,"herkunftFehlt":false}`,
		},
		{
			"Karte",
			app.KarteDTO{ID: "1", Nummer: 2, Werte: []app.WertDTO{{Feld: "name", Wert: "F"}}},
			`{"id":"1","nummer":2,"werte":[{"feld":"name","wert":"F"}]}`,
		},
		{
			"Treffer",
			app.TrefferDTO{
				KarteID: "1", SetID: "7", Nummer: 0,
				Kartenname: "F", Feld: "text", Wert: "w",
			},
			`{"karteID":"1","setID":"7","nummer":0,"kartenname":"F","feld":"text","wert":"w"}`,
		},
		{
			"Zahl",
			app.ZahlDTO{Name: "name", Karten: 12},
			`{"name":"name","karten":12}`,
		},
		{
			"Vorbehalt",
			app.VorbehaltDTO{Datei: "set", Zeile: 3, Inhalt: "x"},
			`{"datei":"set","zeile":3,"inhalt":"x"}`,
		},
		{
			// **Alle vier Listen leer und nicht `null`.** Der Grund steht
			// oben: eine Antwort, die `null` schickt, zwingt jede Seite zu
			// zwei Prüfungen statt einer.
			//
			// **Die vierte kam mit #1025 dazu**, und dieser Lauf hat sie
			// sofort angestrichen: `Nennungen` war `null`, weil die
			// Fehlerbehandlung eine nackte Rückgabe hatte. Genau dafür gibt es
			// ihn — die Go-Tests sehen die Struct-Marken nicht, die
			// Frontend-Tests gehen an der Serialisierung vorbei (§3).
			"Vorschau",
			app.VorschauDTO{
				Bild: "data:", Breite: 1, Hoehe: 2,
				Befunde:   []string{},
				Nennungen: []app.NennungDTO{},
				Flaechen:  []app.FeldflaecheDTO{},
				Vorgaben:  []app.VorgabeDTO{},
			},
			`{"bild":"data:","breite":1,"hoehe":2,"dpi":0,"befunde":[],"nennungen":[],` +
				`"flaechen":[],"vorgaben":[],"stil":"","gezeichnet":""}`,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gebaut, err := json.Marshal(fall.wert)
			if err != nil {
				t.Fatalf("serialisieren: %v", err)
			}

			if string(gebaut) != fall.erwartet {
				t.Errorf("%s\nerwartet war\n%s", gebaut, fall.erwartet)
			}
		})
	}
}
