// **Der Rückweg trennt eine zusammengelegte Karte wieder in ihre Seiten** —
// die zweite Hälfte von Befund 4 des Zweitblicks über Phase V.
//
// # Warum das dazugehört
//
// Wenn das Lesen zwei MTGJSON-Karten zu **einer** MSE-Karte macht, muss das
// Schreiben aus **einer** MSE-Karte wieder zwei machen. Sonst gäbe der
// Rückweg `"name 2"` als MTGJSON-Schlüssel heraus — und den gibt es dort
// nicht.
//
// **Gemessen am 2026-08-24**, vor dieser Änderung, war es genau so:
//
//	"casting cost 2": "W",
//	"manaCost": "{1}{R}",
//	"name": "Wear",
//	"name 2": "Tear"
package mtgjson_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// geschriebeneKarten liest die Kartenobjekte aus einer geschriebenen Datei.
func geschriebeneKarten(t *testing.T, roh []byte) []map[string]any {
	t.Helper()

	var gelesen struct {
		Daten struct {
			Karten []map[string]any `json:"cards"`
		} `json:"data"`
	}

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		t.Fatalf("die geschriebene Datei ist unlesbar: %v", fehler)
	}

	return gelesen.Daten.Karten
}

func TestDerRueckwegTrenntDieSeitenWieder(t *testing.T) {
	t.Parallel()

	set := &domain.Set{
		Name: "Prüfsatz", Game: "magic",
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Wear"},
			{Field: "casting cost", Text: "1R"},
			{Field: "rule text", Text: "Zerstöre ein Artefakt."},
			{Field: "name 2", Text: "Tear"},
			{Field: "casting cost 2", Text: "W"},
			{Field: "rule text 2", Text: "Zerstöre eine Verzauberung."},
		}}},
	}

	roh, _ := mtgjson.ZuordnungMagic.Schreibe(set)
	karten := geschriebeneKarten(t, roh)

	// **Zwei Kartenobjekte**, wie MTGJSON sie führt.
	if len(karten) != 2 {
		t.Fatalf("%d Kartenobjekte, erwartet 2 — MTGJSON führt je Seite eines",
			len(karten))
	}

	erste, zweite := karten[0], karten[1]

	// **Der zusammengesetzte Name steht in beiden**, so wie MTGJSON ihn führt:
	// „The name of the card. Cards with multiple faces are given a `//`
	// delimiter."
	for stelle, karte := range karten {
		if karte["name"] != "Wear // Tear" {
			t.Errorf("Karte %d: name = %v, erwartet den zusammengesetzten Namen",
				stelle+1, karte["name"])
		}
	}

	if erste["faceName"] != "Wear" || zweite["faceName"] != "Tear" {
		t.Errorf("faceName = %v / %v", erste["faceName"], zweite["faceName"])
	}

	if erste["side"] != "a" || zweite["side"] != "b" {
		t.Errorf("side = %v / %v", erste["side"], zweite["side"])
	}

	if erste["manaCost"] != "{1}{R}" || zweite["manaCost"] != "{W}" {
		t.Errorf("manaCost = %v / %v", erste["manaCost"], zweite["manaCost"])
	}

	// **Und sie verweisen aufeinander.** Ohne `otherFaceIds` wüsste ein Leser
	// nicht, dass sie zusammengehören — und legte sie beim nächsten Mal nicht
	// mehr zusammen.
	if erste["uuid"] == nil || zweite["uuid"] == nil {
		t.Fatalf("eine Kennung fehlt: %v / %v", erste["uuid"], zweite["uuid"])
	}

	if !verweistAuf(erste, zweite["uuid"]) || !verweistAuf(zweite, erste["uuid"]) {
		t.Errorf("die Seiten verweisen nicht aufeinander: %v / %v",
			erste["otherFaceIds"], zweite["otherFaceIds"])
	}

	// **Kein Schlüssel mit einem Zusatz.** `name 2` ist MSEs Schreibweise, nicht
	// MTGJSONs.
	for stelle, karte := range karten {
		for schluessel := range karte {
			if len(schluessel) > 2 && schluessel[len(schluessel)-2:] == " 2" {
				t.Errorf("Karte %d führt den MSE-Schlüssel %q", stelle+1, schluessel)
			}
		}
	}
}

// verweistAuf sagt, ob eine Karte die Kennung in `otherFaceIds` führt.
func verweistAuf(karte map[string]any, kennung any) bool {
	liste, gut := karte["otherFaceIds"].([]any)
	if !gut {
		return false
	}

	for _, eine := range liste {
		if eine == kennung {
			return true
		}
	}

	return false
}

// **Eine einseitige Karte bleibt ein Kartenobjekt.**
//
// Ohne diesen Fall prüfte der Lauf darüber nur, dass **irgendwie** getrennt
// wird — eine Fassung, die jede Karte verdoppelte, bestünde ihn.
func TestEineEinseitigeKarteBleibtEinKartenobjekt(t *testing.T) {
	t.Parallel()

	set := &domain.Set{
		Name: "Prüfsatz", Game: "magic",
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Lightning Bolt"},
			{Field: "casting cost", Text: "R"},
		}}},
	}

	roh, _ := mtgjson.ZuordnungMagic.Schreibe(set)
	karten := geschriebeneKarten(t, roh)

	if len(karten) != 1 {
		t.Fatalf("%d Kartenobjekte, erwartet 1", len(karten))
	}

	if karten[0]["name"] != "Lightning Bolt" {
		t.Errorf("name = %v", karten[0]["name"])
	}

	for _, unerwartet := range []string{"faceName", "side", "otherFaceIds"} {
		if _, da := karten[0][unerwartet]; da {
			t.Errorf("eine einseitige Karte führt %q", unerwartet)
		}
	}
}

// **Und der Rundlauf trägt sie über beide Richtungen.**
//
// # Warum das der eigentliche Beleg ist
//
// Die beiden Läufe darüber prüfen je eine Richtung. Erst der Rundlauf zeigt,
// dass sie zueinander passen — und das ist die Naht, an der in Phase 5 der
// Fehler saß, den kein Test sah (#130).
func TestZweiSeitenUeberlebenDenRundlauf(t *testing.T) {
	t.Parallel()

	hinein := `{"data":{"code":"PRF","name":"Prüfsatz","cards":[
		{"name":"Wear // Tear","faceName":"Wear","side":"a","layout":"split",
		 "manaCost":"{1}{R}","otherFaceIds":["zwei"],"uuid":"eins"},
		{"name":"Wear // Tear","faceName":"Tear","side":"b","layout":"split",
		 "manaCost":"{W}","otherFaceIds":["eins"],"uuid":"zwei"}
	]}}`

	quelle, fehler := mtgjson.LiesInhalt([]byte(hinein), "probe.json", mtgjson.ZuordnungMagic)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Eine Karte in der Mitte.**
	if len(quelle.Sets[0].Karten) != 1 {
		t.Fatalf("%d Karten in der Mitte", len(quelle.Sets[0].Karten))
	}

	roh, _ := mtgjson.ZuordnungMagic.Schreibe(alsDomaene(quelle)[0])

	// **Und wieder zwei Kartenobjekte hinaus.**
	wieder, lesefehler := mtgjson.LiesInhalt(roh, "zurueck.json", mtgjson.ZuordnungMagic)
	if lesefehler != nil {
		t.Fatalf("zurücklesen: %v", lesefehler)
	}

	if len(wieder.Sets[0].Karten) != 1 {
		t.Fatalf("%d Karten nach dem Rundlauf", len(wieder.Sets[0].Karten))
	}

	werte := wieder.Sets[0].Karten[0].Werte

	erwartet := map[string]string{
		"name": "Wear", "casting cost": "1R",
		"name 2": "Tear", "casting cost 2": "W",
	}

	for feld, soll := range erwartet {
		if ist, _ := wertVon(werte, feld); ist != soll {
			t.Errorf("nach dem Rundlauf: %q = %q, erwartet %q", feld, ist, soll)
		}
	}
}

// **Ein fehlender Name lässt den Trenner weg.**
//
// # Warum das eine eigene Zusicherung ist
//
// Der erste Anlauf setzte `vorn + " // " + hinten` ohne Ansehen der Teile.
// Eine Seite ohne Namen ergab damit `" // Tear"` — eine Behauptung über eine
// Vorderseite, die keinen Namen hat. Der Trenner ohne die Seiten sagt nichts,
// und ein Leser hielte das Leerzeichen für den Namen.
//
// # Hier stand ein dritter Fall, und er hielt einen Fehler fest
//
// „nur die Vorderseite" — `name` und `casting cost 2`, **ohne** `name 2` —
// erwartete **zwei** Kartenobjekte. Das war die Zusicherung des Fehlers, den
// Befund 1 des Zweitblicks über Phase 3.6 gefunden hat: jedes Feld mit dem
// Zusatz ` 2` galt als Rückseite, und ein einseitiger Planeswalker mit
// `loyalty cost 2` wurde in zwei Karten zerlegt. Der Fall steht jetzt in
// `kartenseitentrennung_test.go` mit der **richtigen** Erwartung: eine Karte.
func TestEinFehlenderNameLaesstDenTrennerWeg(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		werte    []domain.Value
		erwartet string
	}{
		"beide Namen": {
			[]domain.Value{
				{Field: "name", Text: "Wear"},
				{Field: "name 2", Text: "Tear"},
			},
			"Wear // Tear",
		},
		"nur die Rückseite": {
			[]domain.Value{
				{Field: "casting cost", Text: "R"},
				{Field: "name 2", Text: "Tear"},
			},
			"Tear",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			set := &domain.Set{
				Name: "Prüfsatz", Game: "magic",
				Cards: []*domain.Card{{Values: fall.werte}},
			}

			roh, _ := mtgjson.ZuordnungMagic.Schreibe(set)
			karten := geschriebeneKarten(t, roh)

			if len(karten) != 2 {
				t.Fatalf("%d Kartenobjekte", len(karten))
			}

			if karten[0]["name"] != fall.erwartet {
				t.Errorf("name = %v, erwartet %q", karten[0]["name"], fall.erwartet)
			}
		})
	}
}

// **Zwei Läufe schreiben dieselbe Datei.**
//
// Die Kennungen werden aus dem Namen gebildet und nicht gewürfelt — sonst
// änderte jeder Rundlauf die Datei, und ein Vergleich zweier Läufe wäre nicht
// mehr möglich.
func TestZweiLaeufeSchreibenDieselbeDatei(t *testing.T) {
	t.Parallel()

	set := &domain.Set{
		Name: "Prüfsatz", Game: "magic",
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Wear"},
			{Field: "name 2", Text: "Tear"},
		}}},
	}

	erster, _ := mtgjson.ZuordnungMagic.Schreibe(set)
	zweiter, _ := mtgjson.ZuordnungMagic.Schreibe(set)

	if !bytes.Equal(erster, zweiter) {
		t.Error("zwei Läufe schrieben verschiedene Dateien")
	}
}
