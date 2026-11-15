package app_test

// **Ein leerer Wert ist kein Vorschlag — an keinem Weg** — N1 des Nachlaufs
// über Phase 6.10 (#1671). Z1 hatte den Rückfall des Vorrangs repariert;
// bei ausdrücklich gewähltem Anbieter, im Stapel und im Einzelzug blieb ein
// leerer Name ein Vorschlag, und `uebernehmen` schrieb ihn.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEinLeererWertIstAnKeinemWegEinVorschlag(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.antwort.Inhalt = `{"name": "  ", "rarity": "rare"}`

	bezug := app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b", Auftrag: "x",
	}

	antwort := welt.anwendung.KIVorschlagInEinemZug(probekontext(t), bezug, true, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, feld := range antwort.Felder {
		if feld.Feld == "name" {
			t.Errorf("ein leerer Name steht als Vorschlag da: %+v", feld)
		}
	}

	// **Und geschrieben wurde er nicht** — der alte Name bleibt.
	if werte := welt.kartenwerte(t); !strings.Contains(werte, "Alt") {
		t.Errorf("die Karte trägt %q", werte)
	}
}
