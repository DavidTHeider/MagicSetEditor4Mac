package service_test

// **Der Feldweg sagt, was er gespeichert hat** — #2535 (Nachlauf über das
// Gesamtreview über 0.9.0, Befund 1). Seit #2396 speichert er den gerechneten
// Wert, und das Fenster bestätigte seinen Entwurf mit der nächsten Änderung
// des Bestands — auch mit einer fremden, die zwischen Speichern und Nachladen
// kam. Jetzt nennt das Schreiben den Wert, den es gespeichert hat.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerFeldwegNenntDenGespeichertenWert(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		stile   *wegstile
		gewollt string
		bekannt bool
		feld    string
		eingabe string
		felder  []domain.Field
	}{
		{
			name:  "gerechnet",
			stile: &wegstile{vorlage: service.Vorlage{Kartenfelder: nachzugfelder()}, gefunden: true},
			feld:  "eins", eingabe: "Funke", gewollt: "[Funke]", bekannt: true, felder: nachzugfelder(),
		},
		{
			name:  "ohne Vorlage der eingegebene",
			stile: &wegstile{},
			feld:  "eins", eingabe: "Funke", gewollt: "Funke", bekannt: true, felder: nachzugfelder(),
		},
		{
			// **Ein Verbund steht nicht selbst im Bestand** — seine Teile tun
			// es, und was das Fenster zeigt, rechnet erst das Lesen.
			name:  "ein Verbund ist unbekannt",
			stile: &wegstile{vorlage: service.Vorlage{Kartenfelder: verbundfelder()}, gefunden: true},
			feld:  "text", eingabe: "Fliegend<sep><line>\n</line></sep>Hoch", gewollt: "", bekannt: false,
			felder: verbundfelder(),
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			weg := service.NeuerFeldweg(wegkarten{karte: nachzugkarte()}, wegfelder{felder: fall.felder}, &wegablage{}).
				MitNachzug(fall.stile, nil, nil)

			gespeichert, bekannt, fehler := weg.FeldSetzenMitErgebnis(t.Context(), 7, fall.feld, fall.eingabe)
			if fehler != nil {
				t.Fatalf("setzen: %v", fehler)
			}

			if gespeichert != fall.gewollt || bekannt != fall.bekannt {
				t.Errorf("gespeichert %q (%v), erwartet %q (%v)", gespeichert, bekannt, fall.gewollt, fall.bekannt)
			}
		})
	}
}

// verbundfelder ist ein Feld, das in zwei gespeicherte zerfällt.
func verbundfelder() []domain.Field {
	return []domain.Field{
		{Name: "text", Gespeichert: false, Skript: `combined_editor(field1: card.rule_text, ` +
			`separator: "<line>\n</line>", field2: card.flavor_text)`},
		{Name: "rule text", Gespeichert: true, Skript: `"R:" + value`},
		{Name: "flavor text", Gespeichert: true, Skript: `"F:" + value`},
	}
}
