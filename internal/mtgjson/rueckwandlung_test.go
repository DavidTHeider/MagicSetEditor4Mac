package mtgjson_test

import (
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// Die Rückwandlungen, jede an ihren Rändern.
func TestDieRueckwandlungen(t *testing.T) {
	faelle := map[string]struct {
		wandlung mtgjson.Rueckwandlung
		ein      string
		aus      any
	}{
		"Manakosten aus Zahl und Buchstaben":  {mtgjson.Manasymbole, "3WW", "{3}{W}{W}"},
		"Manakosten, die auf eine Zahl enden": {mtgjson.Manasymbole, "W2", "{W}{2}"},
		"eine mehrstellige Zahl bleibt eine":  {mtgjson.Manasymbole, "12U", "{12}{U}"},
		"leere Manakosten":                    {mtgjson.Manasymbole, "", ""},

		// **Der Fall, den die MSE-Schreibweise nicht entscheidet**: `W/U` kann
		// ein hybrides Symbol sein oder zwei. Es geht als **eine** Klammer
		// hinaus — der Rundlauf hält, die Bedeutung ist geraten.
		"hybride Kosten gehen als eine Klammer": {mtgjson.Manasymbole, "W/U", "{W/U}"},

		// Kleinbuchstaben kommen in MSE-Kosten vor (`x` für das X-Symbol);
		// sie sind Symbole wie die großen.
		"ein Kleinbuchstabe ist ein Symbol": {mtgjson.Manasymbole, "x2", "{x}{2}"},

		"Seltenheit aus MSE":  {mtgjson.SeltenheitZurueck, "mythic rare", "mythic"},
		"Seltenheit wörtlich": {mtgjson.SeltenheitZurueck, "common", "common"},

		"leere Wortliste": {mtgjson.WoerterZurueck, "", []string{}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			gefunden := fall.wandlung(fall.ein)

			if erwartet, ist := fall.aus.([]string); ist {
				liste, passt := gefunden.([]string)
				if !passt || len(liste) != len(erwartet) {
					t.Errorf("%v, erwartet %v", gefunden, erwartet)
				}

				return
			}

			if gefunden != fall.aus {
				t.Errorf("%v, erwartet %v", gefunden, fall.aus)
			}
		})
	}
}

// Ein Set ohne Code bekommt seinen Namen als Code.
//
// MTG JSON verlangt einen; ein leerer machte die Ausgabe unbrauchbar. Ein Set,
// das in der Anwendung entstanden ist, hat keinen — es kam nie aus einer
// MTG-JSON-Datei.
func TestEinSetOhneCodeBekommtSeinenNamen(t *testing.T) {
	set := &domain.Set{
		Name:  "Ohne Code",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Eine"}}}},
	}

	inhalt, _ := mtgjson.ZuordnungMagic.Schreibe(set)

	var gebaut struct {
		Daten struct {
			Code string `json:"code"`
		} `json:"data"`
	}

	if err := json.Unmarshal(inhalt, &gebaut); err != nil {
		t.Fatalf("die Ausgabe ist kein JSON: %v", err)
	}

	if gebaut.Daten.Code != "Ohne Code" {
		t.Errorf("Code: %q, erwartet %q", gebaut.Daten.Code, "Ohne Code")
	}
}

// Ein Feld, das jemand in der Anwendung angelegt hat, geht als Zeichenkette
// hinaus.
//
// Mitgeführte JSON-Schlüssel liegen als rohes JSON an der Karte; ein Feld, das
// die Anwendung selbst angelegt hat, tut das nicht. Es als rohes JSON zu deuten
// hieße, aus `hallo` einen Fehler zu machen — und aus einem Feldwert, der
// zufällig wie JSON aussieht, etwas anderes als eine Zeichenkette.
func TestEinEigenesFeldGehtAlsZeichenketteHinaus(t *testing.T) {
	set := &domain.Set{
		Name: "Probeset",
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Eine"},
			{Field: "notiz", Text: "noch nachsehen"},
		}}},
	}

	inhalt, _ := mtgjson.ZuordnungMagic.Schreibe(set)

	var gebaut struct {
		Daten struct {
			Karten []map[string]json.RawMessage `json:"cards"`
		} `json:"data"`
	}

	if err := json.Unmarshal(inhalt, &gebaut); err != nil {
		t.Fatalf("die Ausgabe ist kein JSON: %v", err)
	}

	if gefunden := string(gebaut.Daten.Karten[0]["notiz"]); gefunden != `"noch nachsehen"` {
		t.Errorf("notiz: %s", gefunden)
	}
}

// Die Tabelle der Schlüsselarten ist eine Kopie, kein Griff nach innen.
//
// Wer sie ändern könnte, änderte die Prüfung von außen — und ein Prüfer, den
// sein Prüfling umschreiben kann, prüft nichts.
func TestDieSchluesseltabelleWirdAlsKopieGegeben(t *testing.T) {
	arten := mtgjson.Schluesselarten()
	if len(arten) == 0 {
		t.Fatal("die Tabelle ist leer")
	}

	arten["name"] = mtgjson.ArtZahl

	if mtgjson.Schluesselarten()["name"] != mtgjson.ArtZeichenkette {
		t.Error("die Tabelle ließ sich von außen ändern")
	}
}

// Die Artprüfung kennt jede Art, die die Tabelle führt.
func TestDieArtpruefungKenntJedeArt(t *testing.T) {
	faelle := map[mtgjson.Art]string{
		mtgjson.ArtZeichenkette: `"a"`,
		mtgjson.ArtWortliste:    `["a"]`,
		mtgjson.ArtZahl:         `3`,
		mtgjson.ArtWahrheit:     `true`,
		mtgjson.ArtObjekt:       `{"a":"b"}`,
	}

	for art, roh := range faelle {
		t.Run(string(art), func(t *testing.T) {
			if !mtgjson.PasstZurArt(json.RawMessage(roh), art) {
				t.Errorf("%s passt nicht zu %q", roh, art)
			}

			if mtgjson.PasstZurArt(json.RawMessage(`"passt nie"`), art) && art != mtgjson.ArtZeichenkette {
				t.Errorf("eine Zeichenkette gilt als %q", art)
			}
		})
	}

	// Eine Art, die die Tabelle nicht führt, wird durchgelassen statt geraten.
	if !mtgjson.PasstZurArt(json.RawMessage(`1`), mtgjson.Art("erfunden")) {
		t.Error("eine unbekannte Art wird abgewiesen")
	}
}
