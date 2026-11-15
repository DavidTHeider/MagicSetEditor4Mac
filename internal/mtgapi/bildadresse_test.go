package mtgapi_test

// **Gatherer antwortet auch über `https`** — angestrichen vom Zweitblick über
// Phase H (#2086). magicthegathering.io nennt jedes Bild als
// `http://gatherer.wizards.com/…`; die WebView holt es nicht, denn ihre
// Schutzrichtlinie kennt nur `https`-Ursprünge, und App Transport Security
// ließe Klartext ohnehin nicht durch. Gemessen am 2026-10-07: dieselbe
// Adresse über `http` antwortet mit 301 auf ihr `https`-Gegenstück.

import (
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
)

func TestGathererBilderKommenUeberHTTPS(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, gesagt, erwartet string
	}{
		{
			"Gatherer über http wird https",
			"http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=1&type=card",
			"https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=1&type=card",
		},
		{
			"https bleibt, wie es ist",
			"https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=1&type=card",
			"https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=1&type=card",
		},
		// **Ein anderer Ursprung bleibt unangetastet**: gemessen ist nur
		// Gatherer, und ob ein fremder Rechner `https` spricht, weiß hier
		// niemand.
		{"ein anderer Ursprung bleibt", "http://beispiel.invalid/karte.png", "http://beispiel.invalid/karte.png"},
		{"kein Bild bleibt keins", "", ""},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			roh, err := json.Marshal(map[string]any{"cards": []map[string]any{{
				"name": "Probewesen", "type": "Creature — Beast", "rarity": "Rare",
				"set": "PRB", "setName": "Probeset", "number": "7", "layout": "normal",
				"id": "1111222233334444555566667777888899990000", "imageUrl": fall.gesagt,
			}}})
			if err != nil {
				t.Fatal(err)
			}

			karten, _, err := mtgapi.LiesKarten(roh)
			if err != nil {
				t.Fatal(err)
			}

			if karten[0].Bildadresse != fall.erwartet {
				t.Errorf("Bildadresse %q, erwartet %q", karten[0].Bildadresse, fall.erwartet)
			}
		})
	}
}
