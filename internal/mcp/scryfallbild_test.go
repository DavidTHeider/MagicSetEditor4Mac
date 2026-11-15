package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// bildprobe zählt mit, wonach gefragt wurde.
type bildprobe struct {
	rufe   int
	feld   string
	name   string
	art    string
	fehler error
}

func (b *bildprobe) Einsetzen(
	_ context.Context, _, _ int64, feld, name, art string,
) (mcp.Bildeinsatz, error) {
	b.rufe++
	b.feld, b.name, b.art = feld, name, art

	if b.fehler != nil {
		return mcp.Bildeinsatz{}, b.fehler
	}

	return mcp.Bildeinsatz{
		Feld: feld, Kuenstler: "Christopher Rush",
		Nennung:            "Illustration: Christopher Rush. ™ & © Wizards of the Coast",
		Adresse:            "https://scryfall.com/card/lea/161",
		IllustratorGesetzt: true, AlterStunden: 2,
	}, nil
}

// rufeBild findet das Bildwerkzeug und ruft es.
func rufeBild(t *testing.T, bilder mcp.Fremdbilder, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      bilder,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "schreiben.scryfall.bild" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("es gibt kein Werkzeug schreiben.scryfall.bild")

	return "", nil
}

func TestScryfallbildwerkzeug(t *testing.T) {
	t.Parallel()

	t.Run("setzt ein und meldet die Nennung", func(t *testing.T) {
		t.Parallel()

		probe := &bildprobe{}

		text, fehler := rufeBild(t, probe, `{"setID": 1, "karteID": 7, `+
			`"feld": "image", "name": "Lightning Bolt", "art": "ausschnitt"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if probe.feld != "image" || probe.name != "Lightning Bolt" ||
			probe.art != "ausschnitt" {
			t.Errorf("gefragt wurde %q/%q/%q", probe.feld, probe.name, probe.art)
		}

		// **Die Nennung muss in der Antwort stehen** — ein Gegenüber baut eine
		// Oberfläche, und Scryfall verlangt sie darin.
		for _, teil := range []string{
			"Christopher Rush", "Wizards of the Coast",
			"scryfall.com", "IllustratorGesetzt",
		} {
			if !strings.Contains(text, teil) {
				t.Errorf("die Antwort führt %q nicht: %s", teil, text)
			}
		}
	})

	t.Run("reicht den Fehler durch", func(t *testing.T) {
		t.Parallel()

		probe := &bildprobe{fehler: domain.ErrNotFound}

		_, fehler := rufeBild(t, probe, `{"setID": 1, "karteID": 7, `+
			`"feld": "image", "name": "Gibtesnicht", "art": "normal"}`)
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v", fehler)
		}
	})

	// **Und ein Ausfall der Auskunft kommt durch** — beide Wege.
	t.Run("reicht einen Fehler durch", func(t *testing.T) {
		t.Parallel()

		mitFehler := func(t *testing.T, name, argumente string) error {
			t.Helper()

			for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
				Leser:       probenbestand(),
				Schreiber:   probenbestand(),
				Ausfuhr:     ruhendeAusfuhr{},
				Abgleich:    nil,
				Bilder:      nil,
				Auskunft:    stilleAuskunft{fehler: domain.ErrNotFound},
				Sammeldaten: stilleSammeldaten{},
				Anlage:      nil,
				Saetze:      nil,
				Konsole:     nil,
				Wortlisten:  nil,
				Woerter:     nil,
				Pack:        nil,
				Anleger:     nil,
				Angaben:     neueAngabenprobe(),
				Stile:       neueStilprobe(),
				Bestand:     neueBestandsprobeWeit(),
			}) {
				if werkzeug.Name == name {
					_, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))

					return fehler
				}
			}

			return nil
		}

		if fehler := mitFehler(t, "lesen.scryfall.regeln",
			`{"name": "Gibtesnicht"}`); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Fehler = %v", fehler)
		}

		if fehler := mitFehler(t, "lesen.scryfall.manakosten",
			`{"kosten": "{W}"}`); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Fehler = %v", fehler)
		}
	})

	t.Run("weist unlesbare Argumente ab", func(t *testing.T) {
		t.Parallel()

		probe := &bildprobe{}

		if _, fehler := rufeBild(t, probe, `{"setID": "eins"}`); fehler == nil {
			t.Fatal("die unlesbaren Argumente blieben ohne Fehler")
		}

		if probe.rufe != 0 {
			t.Errorf("es wurde trotzdem eingesetzt (%d Rufe)", probe.rufe)
		}
	})

	t.Run("weist die Kennungen einzeln ab", func(t *testing.T) {
		t.Parallel()

		faelle := map[string]string{
			"Set null":   `{"setID": 0, "karteID": 7, "feld": "i", "name": "n", "art": "normal"}`,
			"Karte null": `{"setID": 1, "karteID": 0, "feld": "i", "name": "n", "art": "normal"}`,
		}

		for name, argumente := range faelle {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				probe := &bildprobe{}

				if _, fehler := rufeBild(t, probe, argumente); fehler == nil {
					t.Fatal("die Kennung blieb ohne Fehler")
				}

				if probe.rufe != 0 {
					t.Errorf("es wurde trotzdem eingesetzt (%d Rufe)", probe.rufe)
				}
			})
		}
	})
}

// **Die beiden lesenden Auskünfte** (#888).
//
// # Warum sie hier stehen und nicht bei den schreibenden
//
// Weil sie nichts ändern — und weil ein Gegenüber, das die Werkzeugliste
// kürzt, die harmlosen behalten soll ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
func TestScryfallauskunftswerkzeuge(t *testing.T) {
	t.Parallel()

	rufe := func(t *testing.T, name, argumente string) (string, error) {
		t.Helper()

		for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
			Leser:       probenbestand(),
			Schreiber:   probenbestand(),
			Ausfuhr:     ruhendeAusfuhr{},
			Abgleich:    nil,
			Bilder:      nil,
			Auskunft:    stilleAuskunft{},
			Sammeldaten: stilleSammeldaten{},
			Anlage:      nil,
			Saetze:      nil,
			Konsole:     nil,
			Wortlisten:  nil,
			Woerter:     nil,
			Pack:        nil,
			Anleger:     nil,
			Angaben:     neueAngabenprobe(),
			Stile:       neueStilprobe(),
			Bestand:     neueBestandsprobeWeit(),
		}) {
			if werkzeug.Name == name {
				return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
			}
		}

		t.Fatalf("es gibt kein Werkzeug %q", name)

		return "", nil
	}

	t.Run("die Regeln nennen ihre Herkunft", func(t *testing.T) {
		t.Parallel()

		text, fehler := rufe(t, "lesen.scryfall.regeln", `{"name": "Prüfstein"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		// **`Offiziell` muss in der Antwort stehen.** Für ein Sprachmodell ist
		// das die wichtigere Hälfte: eine Anmerkung des Scryfall-Teams als
		// Regel weiterzugeben wäre falsch.
		for _, teil := range []string{"Offiziell", "Prüfstein", "Anmerkung."} {
			if !strings.Contains(text, teil) {
				t.Errorf("die Antwort führt %q nicht: %s", teil, text)
			}
		}
	})

	t.Run("die Manakostenprüfung gibt eine leere Liste statt null", func(t *testing.T) {
		t.Parallel()

		text, fehler := rufe(t, "lesen.scryfall.manakosten", `{"kosten": "{2}{R}"}`)
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if !strings.Contains(text, "[]") {
			t.Errorf("die Antwort ist nicht die leere Liste: %s", text)
		}
	})

	// **Und ein Ausfall der Auskunft kommt durch** — beide Wege.
	t.Run("reicht einen Fehler durch", func(t *testing.T) {
		t.Parallel()

		mitFehler := func(t *testing.T, name, argumente string) error {
			t.Helper()

			for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
				Leser:       probenbestand(),
				Schreiber:   probenbestand(),
				Ausfuhr:     ruhendeAusfuhr{},
				Abgleich:    nil,
				Bilder:      nil,
				Auskunft:    stilleAuskunft{fehler: domain.ErrNotFound},
				Sammeldaten: stilleSammeldaten{},
				Anlage:      nil,
				Saetze:      nil,
				Konsole:     nil,
				Wortlisten:  nil,
				Woerter:     nil,
				Pack:        nil,
				Anleger:     nil,
				Angaben:     neueAngabenprobe(),
				Stile:       neueStilprobe(),
				Bestand:     neueBestandsprobeWeit(),
			}) {
				if werkzeug.Name == name {
					_, fehler := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))

					return fehler
				}
			}

			return nil
		}

		if fehler := mitFehler(t, "lesen.scryfall.regeln",
			`{"name": "Gibtesnicht"}`); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Fehler = %v", fehler)
		}

		if fehler := mitFehler(t, "lesen.scryfall.manakosten",
			`{"kosten": "{W}"}`); !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("Fehler = %v", fehler)
		}
	})

	t.Run("weist unlesbare Argumente ab", func(t *testing.T) {
		t.Parallel()

		for _, name := range []string{"lesen.scryfall.regeln", "lesen.scryfall.manakosten"} {
			if _, fehler := rufe(t, name, `{"name": 42, "kosten": 42}`); fehler == nil {
				t.Errorf("%s ließ unlesbare Argumente durch", name)
			}
		}
	})

	// **Und ein fehlendes Argument ebenso** — an allen drei Flächen gleich.
	//
	// Der erste Anlauf ließ `{"kosten": ""}` durch und antwortete „in
	// Ordnung": eine Auskunft über nichts. An der Kommandozeile und über REST
	// wurde derselbe Fall längst abgewiesen — §16 sagt, dass dieselbe
	// Fähigkeit überall dieselbe Antwort gibt.
	t.Run("weist ein fehlendes Argument ab", func(t *testing.T) {
		t.Parallel()

		for _, name := range []string{"lesen.scryfall.regeln", "lesen.scryfall.manakosten"} {
			if _, fehler := rufe(t, name, `{}`); fehler == nil {
				t.Errorf("%s ließ ein fehlendes Argument durch", name)
			}
		}
	})
}
