package service_test

import (
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Ein Spiel- und ein Stilpaket, klein genug, dass jede Zeile zählbar bleibt.
const (
	spieldatei = "mse version: 2.0.0\n" +
		"card field:\n" + // 2
		"\ttype: text\n" + // 3
		"\tname: casting cost\n" + // 4
		"\tscript: to_upper(value)\n" + // 5
		"set field:\n" + // 6
		"\tname: title\n" + // 7
		"styling field:\n" + // 8
		"\ttype: boolean\n" + // 9
		"\tname: bold text\n" + // 10
		"init script:\n" + // 11
		"\tkosten := card.casting_cost\n" + // 12
		"\t# subtitle ist kein Titel\n" + // 13
		"\tzeigen := { kosten + card.title }\n" // 14

	stildatei = "mse version: 2.0.0\n" +
		"card style:\n" + // 2
		"\tcasting cost:\n" + // 3
		"\t\tleft: 10\n" + // 4
		"extra card field:\n" + // 5
		"\tname: rahmen\n" + // 6
		"extra card style:\n" + // 7
		"\trahmen:\n" + // 8
		"\t\tscript: kosten\n" + // 9
		"\t\tvisible: titles\n" // 10
)

func probepaket() []service.Paketdateitext {
	return []service.Paketdateitext{
		{Name: "game", Inhalt: spieldatei},
		{Name: "style", Inhalt: stildatei},
	}
}

// **Ein Feld, ein Setfeld, ein Stilfeld, ein Skriptname — jeder mit seiner
// Stelle** (U7, #1418). Der Index steht in Go, damit die Werkstatt keine
// zweite Lesart des Paketformats braucht.
func TestDerIndexNenntJedeDefinitionMitIhrerStelle(t *testing.T) {
	symbole := service.Paketsymbole(probepaket())

	erwartet := []service.Paketsymbol{
		{Name: "casting cost", Art: "card field", Datei: "game", Zeile: 4, Spalte: 2},
		{Name: "title", Art: "set field", Datei: "game", Zeile: 7, Spalte: 2},
		{Name: "bold text", Art: "styling field", Datei: "game", Zeile: 10, Spalte: 2},
		{Name: "kosten", Art: "script", Datei: "game", Zeile: 12, Spalte: 2},
		{Name: "zeigen", Art: "script", Datei: "game", Zeile: 14, Spalte: 2},
		{Name: "rahmen", Art: "extra card field", Datei: "style", Zeile: 6, Spalte: 2},
	}

	if !reflect.DeepEqual(symbole, erwartet) {
		t.Errorf("gefunden:\n%+v\nerwartet:\n%+v", symbole, erwartet)
	}
}

// **Verwendungen sind ganze Wörter, in beiden Schreibweisen** — ein Feld
// heißt in der Datei `casting cost` und im Skript `casting_cost`; MSE
// ersetzt das Leerzeichen selbst. Die Zeile der Definition zählt nicht als
// Verwendung, und `subtitle` ist kein `title`.
func TestVerwendungenSindGanzeWoerterInBeidenSchreibweisen(t *testing.T) {
	faelle := []struct {
		name     string
		erwartet []service.Symbolverwendung
	}{
		{"casting cost", []service.Symbolverwendung{
			{Datei: "game", Zeile: 12, Text: "kosten := card.casting_cost"},
			{Datei: "style", Zeile: 3, Text: "casting cost:"},
		}},
		{"kosten", []service.Symbolverwendung{
			{Datei: "game", Zeile: 14, Text: "zeigen := { kosten + card.title }"},
			{Datei: "style", Zeile: 9, Text: "script: kosten"},
		}},
		{"title", []service.Symbolverwendung{
			{Datei: "game", Zeile: 14, Text: "zeigen := { kosten + card.title }"},
		}},
		{"", nil},
		{"gibt es nicht", nil},
	}

	for _, fall := range faelle {
		t.Run("Verwendungen von "+fall.name, func(t *testing.T) {
			gefunden := service.Symbolverwendungen(probepaket(), fall.name)

			if !reflect.DeepEqual(gefunden, fall.erwartet) {
				t.Errorf("gefunden:\n%+v\nerwartet:\n%+v", gefunden, fall.erwartet)
			}
		})
	}
}

// **Ein Skript, das nicht zerteilbar ist, nimmt dem Index nicht die
// Felder** — dieselbe Regel wie in §8 für fremde Daten: gelesen wird, so
// weit es geht. Und ein Skript ohne Zuweisung definiert nichts.
func TestEinKaputtesSkriptLaesstDieFelderStehen(t *testing.T) {
	symbole := service.Paketsymbole([]service.Paketdateitext{{
		Name: "game",
		Inhalt: "card field:\n\tname: name\n\tscript: { 1 +\n" +
			"init script:\n\tto_upper(\"x\")\n",
	}})

	erwartet := []service.Paketsymbol{
		{Name: "name", Art: "card field", Datei: "game", Zeile: 2, Spalte: 2},
	}

	if !reflect.DeepEqual(symbole, erwartet) {
		t.Errorf("gefunden:\n%+v\nerwartet:\n%+v", symbole, erwartet)
	}
}

// **Ein Block ohne Namen definiert nichts** — und eine leere Datei auch nicht.
func TestEinBlockOhneNamenDefiniertNichts(t *testing.T) {
	symbole := service.Paketsymbole([]service.Paketdateitext{
		{Name: "game", Inhalt: "card field:\n\ttype: text\n"},
		{Name: "leer", Inhalt: ""},
	})

	if len(symbole) != 0 {
		t.Errorf("gefunden: %+v", symbole)
	}
}

// **Eine Datei, die sich nicht zerteilen lässt, trägt nichts bei** — und ein
// `name:` ohne Wert ist kein Name.
func TestEineUnzerteilbareDateiUndEinLeererNameTragenNichtsBei(t *testing.T) {
	symbole := service.Paketsymbole([]service.Paketdateitext{
		{Name: "kaputt", Inhalt: "mse version: 2.0.0\n: ohne schlüssel\n"},
		{Name: "game", Inhalt: "card field:\n\tname:\n\ttype: text\n"},
	})

	if len(symbole) != 0 {
		t.Errorf("gefunden: %+v", symbole)
	}
}
