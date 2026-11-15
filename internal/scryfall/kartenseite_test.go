package scryfall_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Woran eine Kartenseite bei Scryfall erkannt wird** —
// #1025.
//
// # Warum die Frage hierher gehört
//
// **Weil hier steht, was in eine Herkunft geschrieben wird.** Der Feldabgleich
// und der Bildweg legen `Karte.Adresse` ab — das `scryfall_uri` der Antwort.
// Wer die Prüfung anderswo nachbaute, prüfte eine Gestalt, die er nicht setzt.
//
// # Warum der Wirt verglichen und nicht gesucht wird
//
// Ein `strings.Contains` auf „scryfall.com" hielte
// `https://scryfall.com.example.org/card/lea/232` für eine Kartenseite — und
// die Anwendung schriebe fremdes Copyright unter ein Bild, das von irgendwoher
// kam. Eine falsche Nennung ist schlimmer als keine.
func TestIstKartenseite(t *testing.T) {
	for name, fall := range map[string]struct {
		adresse  string
		erwartet bool
	}{
		"die Adresse aus einer Antwort": {
			"https://scryfall.com/card/lea/232/black-lotus?utm_source=api", true,
		},
		"mit www":            {"https://www.scryfall.com/card/lea/232", true},
		"eine Setseite":      {"https://scryfall.com/sets/lea", false},
		"ein ähnlicher Wirt": {"https://scryfall.com.example.org/card/lea/232", false},
		"eine Unterdomäne":   {"https://api.scryfall.com/cards/named", false},
		"ohne TLS":           {"http://scryfall.com/card/lea/232", false},
		"eine Datei":         {"/Users/probe/bilder/lotus.png", false},
		"leer":               {"", false},

		// **Eine Adresse, die sich nicht zerlegen lässt.** `url.Parse` nimmt
		// fast alles an; eine kaputte Prozentfolge im Wirt nicht.
		"unzerlegbar": {"https://%zz/card/lea/232", false},
	} {
		t.Run(name, func(t *testing.T) {
			if gefunden := scryfall.IstKartenseite(fall.adresse); gefunden != fall.erwartet {
				t.Errorf("IstKartenseite(%q) = %v, erwartet %v",
					fall.adresse, gefunden, fall.erwartet)
			}
		})
	}
}
