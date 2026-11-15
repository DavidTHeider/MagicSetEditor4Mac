// **Das Zufallspack als Werkzeug** —
// #889,
// Befund 7 des Zweitblicks über Phase W.
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// packprobe merkt sich, wonach gefragt wurde.
type packprobe struct {
	angebot mcp.Packangebot
	ziehung mcp.Packziehung
	fehler  error
	setID   int64
	art     string
	same    string
	gerufen int
}

func (p *packprobe) Packarten(_ context.Context, setID int64) (mcp.Packangebot, error) {
	p.setID, p.gerufen = setID, p.gerufen+1

	return p.angebot, p.fehler
}

func (p *packprobe) PackZiehen(_ context.Context, setID int64, art, same string) (
	mcp.Packziehung, error,
) {
	p.setID, p.art, p.same = setID, art, same
	p.gerufen++

	return p.ziehung, p.fehler
}

// rufePackwerkzeug ruft eines der beiden über die echte Werkzeugliste.
//
// **Über die Liste und nicht über die Funktion**: ein Werkzeug, das gebaut,
// aber nie eingetragen wird, bestünde jeden Einzeltest und wäre für ein
// Sprachmodell nicht da.
func rufePackwerkzeug(t *testing.T, pack mcp.Pack, name, argumente string) (string, error) {
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
		Pack:        pack,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("%s steht nicht in der Liste", name)

	return "", nil
}

// **Die Packarten kommen als JSON, samt Wählbarkeit und Zahl.**
func TestLesenPackartenGibtDieListe(t *testing.T) {
	t.Parallel()

	probe := &packprobe{angebot: mcp.Packangebot{
		Arten: []mcp.Packart{
			{Name: "booster pack", Waehlbar: true, Karten: 107},
			{Name: "rare", Karten: 12},
		},
	}}

	antwort, err := rufePackwerkzeug(t, probe, "lesen.packarten", `{"setID": 3}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.setID != 3 {
		t.Errorf("gefragt wurde nach Set %d", probe.setID)
	}

	for _, erwartet := range []string{
		`"Name": "booster pack"`, `"Waehlbar": true`, `"Karten": 107`, `"Befunde": []`,
	} {
		if !strings.Contains(antwort, erwartet) {
			t.Errorf("%s fehlt in\n%s", erwartet, antwort)
		}
	}
}

// **Und die Ziehung reicht Art und Samen durch.**
//
// **Der Same als Text** — eine Zahl über 2⁵³ verlöre in JSON ihre letzten
// Stellen, und ein Same, der beim Hinweg schon ein anderer ist, wäre keine
// Wiederholbarkeit.
func TestLesenPackZiehenReichtArtUndSamenDurch(t *testing.T) {
	t.Parallel()

	probe := &packprobe{ziehung: mcp.Packziehung{
		Karten: []mcp.Packkarte{{ID: "7", Name: "Eisball"}},
	}}

	antwort, err := rufePackwerkzeug(t, probe, "lesen.pack.ziehen",
		`{"setID": 3, "art": "booster pack", "same": "18446744073709551615"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.art != "booster pack" || probe.same != "18446744073709551615" {
		t.Errorf("gefragt wurde nach %q mit Samen %q", probe.art, probe.same)
	}

	if !strings.Contains(antwort, `"Name": "Eisball"`) {
		t.Errorf("die Karte fehlt in\n%s", antwort)
	}
}

// **Eine unbrauchbare Kennung erreicht die Naht gar nicht.**
func TestEinPackwerkzeugPrueftDieKennung(t *testing.T) {
	t.Parallel()

	for name, argumente := range map[string]string{
		"lesen.packarten":   `{"setID": 0}`,
		"lesen.pack.ziehen": `{"setID": 0, "art": "alles"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := &packprobe{}

			if _, err := rufePackwerkzeug(t, probe, name, argumente); err == nil {
				t.Error("die Kennung 0 kam durch")
			}

			if probe.gerufen != 0 {
				t.Errorf("die Naht wurde %dmal gerufen", probe.gerufen)
			}
		})
	}
}

// **Und ein unbrauchbares Argument ebenso wenig.**
func TestEinPackwerkzeugWeistUnbrauchbareArgumenteAb(t *testing.T) {
	t.Parallel()

	for name := range map[string]struct{}{
		"lesen.packarten": {}, "lesen.pack.ziehen": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufePackwerkzeug(t, &packprobe{}, name, `{"setID": "drei"}`); err == nil {
				t.Error("eine Zeichenkette als Kennung kam durch")
			}
		})
	}
}

// **Ein Fehler der Naht kommt an und wird nicht geschluckt** (§8).
func TestEinPackfehlerKommtDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	for name, argumente := range map[string]string{
		"lesen.packarten":   `{"setID": 3}`,
		"lesen.pack.ziehen": `{"setID": 3, "art": "alles"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := rufePackwerkzeug(t, &packprobe{fehler: kaputt}, name, argumente)
			if !errors.Is(err, kaputt) {
				t.Errorf("der Fehler ist %v", err)
			}
		})
	}
}
