package mtgjson_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// Ein Set wird als MTG JSON ausgegeben.
//
// Die Gegenrichtung zu #141, und sie ist **nicht** deren Spiegelbild: MSE führt
// Felder, die MTG JSON nicht kennt, und umgekehrt. Was dabei nicht abbildbar
// ist, steht als Befund am Ergebnis (#142).
func TestEinSetWirdAlsMTGJSONGeschrieben(t *testing.T) {
	set := &domain.Set{
		Name:   "Probeset",
		Game:   "magic",
		Values: []domain.Value{{Field: "code", Text: "PRB"}},
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Flammenball"},
			{Field: "casting cost", Text: "2R"},
			{Field: "super type", Text: "Instant"},
			{Field: "rarity", Text: "common"},
			{Field: "illustrator", Text: "Selbst gezeichnet"},
			{Field: "uuid", Text: `"00000000-0000-0000-0000-000000000001"`},
		}}},
	}

	inhalt, _ := mtgjson.ZuordnungMagic.Schreibe(set)

	var gebaut struct {
		Daten struct {
			Code   string                       `json:"code"`
			Name   string                       `json:"name"`
			Karten []map[string]json.RawMessage `json:"cards"`
		} `json:"data"`
	}

	if err := json.Unmarshal(inhalt, &gebaut); err != nil {
		t.Fatalf("die Ausgabe ist kein JSON: %v\n%s", err, inhalt)
	}

	if gebaut.Daten.Code != "PRB" || gebaut.Daten.Name != "Probeset" {
		t.Errorf("Set: %+v", gebaut.Daten)
	}

	if len(gebaut.Daten.Karten) != 1 {
		t.Fatalf("%d Karten, erwartet eine", len(gebaut.Daten.Karten))
	}

	karte := gebaut.Daten.Karten[0]

	faelle := map[string]string{
		"name":     `"Flammenball"`,
		"manaCost": `"{2}{R}"`,
		"types":    `["Instant"]`,
		"rarity":   `"common"`,
		"artist":   `"Selbst gezeichnet"`,

		// Der mitgeführte Schlüssel geht **roh** zurück — als Zeichenkette und
		// nicht als in Anführungszeichen gesetzte Zeichenkette.
		"uuid": `"00000000-0000-0000-0000-000000000001"`,
	}

	for schluessel, erwartet := range faelle {
		// Verglichen wird der **gedeutete** Wert, nicht seine Schreibweise: die
		// Ausgabe ist eingerückt, und Einrückung ist keine Aussage über Daten.
		var eng bytes.Buffer
		if err := json.Compact(&eng, karte[schluessel]); err != nil {
			t.Errorf("%s fehlt in der Ausgabe", schluessel)

			continue
		}

		if eng.String() != erwartet {
			t.Errorf("%s: %s, erwartet %s", schluessel, eng.String(), erwartet)
		}
	}
}

// **Was nicht abbildbar ist, steht am Ergebnis** — nicht in einem Kommentar.
//
// MSE führt `super type` als **ein** Feld: `add_cards_scripts` legt
// Standardländer mit `super_type: "Basic Land"` an, wo „Basic" der Obertyp und
// „Land" der Kartentyp ist. MTG JSON trennt beides in `supertypes` und `types`,
// und die Trennung ist aus dem MSE-Wert nicht wiederherzustellen: sie stünde
// nur da, wenn wir eine Liste der Obertypen erfänden.
func TestWasNichtAbbildbarIstStehtAmErgebnis(t *testing.T) {
	set := &domain.Set{
		Name:   "Probeset",
		Values: []domain.Value{{Field: "code", Text: "PRB"}},
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Ebene"},
			{Field: "super type", Text: "Basic Land"},
		}}},
	}

	_, befunde := mtgjson.ZuordnungMagic.Schreibe(set)

	if len(befunde) == 0 {
		t.Fatal("kein Befund — der Verlust wird stillschweigend hingenommen")
	}

	gefunden := false

	for _, befund := range befunde {
		if befund.Ort == "PRB → super type" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("Befunde: %v — der Verlust an »super type« ist nicht benannt", befunde)
	}
}

// **Der Rundlauf: lesen → schreiben → lesen ergibt dieselben Kartenwerte.**
//
// Er ist **schwächer als der über `.mse-set`** (#48), und das gehört
// ausgesprochen: dort ist die Datei die Quelle und muss Byte für Byte
// wiederkommen. MTG JSON beschreibt Karten, nicht ihre Darstellung; was MSE
// führt — Stil, Styling, Rohblöcke — hat dort keinen Platz. Zugesagt sind
// deshalb die **Kartenwerte**, nicht die Bytes.
//
// **Und er sagt nichts über Gültigkeit.** Beide Richtungen benutzen dieselbe
// Tabelle; was die eine verbiegt, biegt die andere zurück. Nimmt man die
// Rückwandlungen heraus — `2R` statt `{2}{R}` —, bleibt dieser Test grün. Dafür
// steht `gueltigkeit_test.go` daneben, und dort ist die Gegenprobe gefahren.
func TestDerRundlaufUeberMTGJSON(t *testing.T) {
	erste, err := mtgjson.Lies(pfad("probe.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("erstes Lesen: %v", err)
	}

	inhalt, _ := mtgjson.ZuordnungMagic.Schreibe(alsDomaene(erste)[0])

	zweite, err := mtgjson.LiesInhalt(inhalt, "rundlauf.json", mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("zweites Lesen: %v\n%s", err, inhalt)
	}

	if len(zweite.Sets) != len(erste.Sets) {
		t.Fatalf("%d Sets, erwartet %d", len(zweite.Sets), len(erste.Sets))
	}

	for i, vorher := range erste.Sets {
		nachher := zweite.Sets[i]

		if len(nachher.Karten) != len(vorher.Karten) {
			t.Fatalf("Set %s: %d Karten, erwartet %d", vorher.Code, len(nachher.Karten), len(vorher.Karten))
		}

		for k, karteVorher := range vorher.Karten {
			vergleiche(t, vorher.Code, k, karteVorher, nachher.Karten[k])
		}
	}
}

// vergleiche hält die Werte zweier Karten gegeneinander.
func vergleiche(t *testing.T, code string, nummer int, vorher, nachher *mtgjson.Karte) {
	t.Helper()

	fuer := func(karte *mtgjson.Karte) map[string]string {
		werte := map[string]string{}
		for _, wert := range append(append([]domain.Value{}, karte.Werte...), karte.Rest...) {
			werte[wert.Field] = wert.Text
		}

		return werte
	}

	alt, neu := fuer(vorher), fuer(nachher)

	for feld, wert := range alt {
		if neu[feld] != wert {
			t.Errorf("%s, Karte %d, Feld %q: %q, erwartet %q", code, nummer+1, feld, neu[feld], wert)
		}
	}

	for feld, wert := range neu {
		if _, da := alt[feld]; !da {
			t.Errorf("%s, Karte %d: Feld %q ist dazugekommen (%q)", code, nummer+1, feld, wert)
		}
	}
}

// alsDomaene macht aus einer gelesenen Quelle Sets der Domäne.
func alsDomaene(quelle *mtgjson.Quelle) []*domain.Set {
	sets := make([]*domain.Set, 0, len(quelle.Sets))

	for _, gelesen := range quelle.Sets {
		set := &domain.Set{
			Name:   gelesen.Name,
			Game:   "magic",
			Values: []domain.Value{{Field: "code", Text: gelesen.Code}},
		}

		for _, karte := range gelesen.Karten {
			werte := append(append([]domain.Value{}, karte.Werte...), karte.Rest...)
			set.Cards = append(set.Cards, &domain.Card{Values: werte})
		}

		sets = append(sets, set)
	}

	return sets
}
